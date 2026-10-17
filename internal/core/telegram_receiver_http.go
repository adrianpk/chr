package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const telegramReceiverBodyLimit = 1 << 20

// TelegramAPIError contains only a safe classification and bounded retry delay.
type TelegramAPIError struct {
	Code       string
	RetryAfter time.Duration
	Permanent  bool
	Uncertain  bool
}

func (e *TelegramAPIError) Error() string { return "telegram API: " + e.Code }

// TelegramReceiverAPI owns the dedicated receiver's bounded HTTP boundary.
// The existing notification sender retains its separate compatibility contract.
type TelegramReceiverAPI struct {
	Client  TelegramHTTPClient
	BaseURL string
}

// NewTelegramReceiverAPI uses a client that refuses redirects carrying credentials.
func NewTelegramReceiverAPI() TelegramReceiverAPI {
	return TelegramReceiverAPI{
		BaseURL: DefaultTelegramAPIBaseURL,
		Client:  &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

type telegramReceiverEnvelope struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

// CheckConnection verifies credentials and webhook absence without sending messages.
func (a TelegramReceiverAPI) CheckConnection(ctx context.Context, token string) (int64, error) {
	data, err := a.call(ctx, token, "getMe", struct{}{}, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var bot struct {
		ID    int64 `json:"id"`
		IsBot bool  `json:"is_bot"`
	}
	if err := json.Unmarshal(data, &bot); err != nil || bot.ID <= 0 || !bot.IsBot {
		return 0, telegramAPIInvalid()
	}

	data, err = a.call(ctx, token, "getWebhookInfo", struct{}{}, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var webhook struct {
		URL *string `json:"url"`
	}
	if err := json.Unmarshal(data, &webhook); err != nil || webhook.URL == nil {
		return 0, telegramAPIInvalid()
	}

	if *webhook.URL != "" {
		return 0, &TelegramAPIError{Code: "webhook_conflict", Permanent: true}
	}

	return bot.ID, nil
}

// Poll never uses a negative offset or changes webhook configuration.
func (a TelegramReceiverAPI) Poll(ctx context.Context, token string, next *int64, timeout int) ([]TelegramReceivedUpdate, error) {
	if timeout < 1 || timeout > 50 || (next != nil && *next < 0) {
		return nil, errors.New("invalid telegram poll bounds")
	}

	payload := struct {
		Offset  *int64   `json:"offset,omitempty"`
		Limit   int      `json:"limit"`
		Timeout int      `json:"timeout"`
		Allowed []string `json:"allowed_updates"`
	}{next, 50, timeout, []string{"message", "callback_query"}}

	data, err := a.call(ctx, token, "getUpdates", payload, time.Duration(timeout+10)*time.Second)
	if err != nil {
		return nil, err
	}

	return decodeTelegramReceiverUpdates(data)
}

// Send sends one bounded plain-text delivery to its original chat and topic.
func (a TelegramReceiverAPI) Send(ctx context.Context, token string, peer TelegramPeerIdentity, text string) (int64, error) {
	return a.SendAction(ctx, token, peer, text, "", nil)
}

func (a TelegramReceiverAPI) call(ctx context.Context, token, method string, payload any, timeout time.Duration) (json.RawMessage, error) {
	if _, err := validateTelegramReceiverToken(token); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("cannot encode telegram request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(a.BaseURL, "/")+"/bot"+token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("invalid telegram endpoint")
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := a.Client.Do(request)
	if err != nil {
		return nil, &TelegramAPIError{Code: "transport_failure", Uncertain: method == "sendMessage"}
	}

	defer func() {
		//nolint:errcheck // The bounded body read reports response failures.
		_ = response.Body.Close()
	}()

	return decodeTelegramReceiverResponse(response)
}

func decodeTelegramReceiverResponse(response *http.Response) (json.RawMessage, error) {
	body, err := io.ReadAll(io.LimitReader(response.Body, telegramReceiverBodyLimit+1))
	if err != nil || len(body) > telegramReceiverBodyLimit {
		return nil, &TelegramAPIError{Code: "response_limit_or_read_failure", Uncertain: true}
	}

	var envelope telegramReceiverEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, telegramAPIInvalid()
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 || !envelope.OK {
		return nil, classifyTelegramAPIError(response.StatusCode, envelope)
	}

	if len(envelope.Result) == 0 || bytes.Equal(envelope.Result, []byte("null")) {
		return nil, telegramAPIInvalid()
	}

	return envelope.Result, nil
}

func classifyTelegramAPIError(status int, envelope telegramReceiverEnvelope) error {
	code := envelope.ErrorCode
	if code == 0 {
		code = status
	}

	switch code {
	case 400:
		return classifyTelegramBadRequest(status, envelope)
	case 401, 403:
		return &TelegramAPIError{Code: "credentials_or_access", Permanent: true}
	case 409:
		return &TelegramAPIError{Code: "receiver_conflict", Permanent: true}
	case 429:
		if envelope.Parameters.RetryAfter > 86400 {
			return &TelegramAPIError{Code: "retry_delay_exceeds_limit", Permanent: true}
		}

		delay := max(1, min(envelope.Parameters.RetryAfter, 86400))

		return &TelegramAPIError{Code: "rate_limited", RetryAfter: time.Duration(delay) * time.Second}
	default:
		return &TelegramAPIError{Code: "request_failed", Permanent: code < 500, Uncertain: code >= 500}
	}
}

func telegramAPIInvalid() error {
	return &TelegramAPIError{Code: "invalid_response", Permanent: true}
}

// Only the bounded API envelope is inspected; raw descriptions never reach logs.
func telegramEntityRejected(description string) bool {
	description = strings.ToLower(description)

	return strings.Contains(description, "can't parse entities") || strings.Contains(description, "cannot parse entities")
}

func classifyTelegramBadRequest(status int, envelope telegramReceiverEnvelope) error {
	if (status == http.StatusBadRequest || status >= 200 && status < 300) && !envelope.OK && telegramEntityRejected(envelope.Description) {
		return &TelegramAPIError{Code: "format_rejected", Permanent: true}
	}

	return &TelegramAPIError{Code: "request_failed", Permanent: true}
}
