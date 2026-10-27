package telegram

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

const receiverBodyLimit = 1 << 20

// APIError contains only a safe classification and bounded retry delay.
type APIError struct {
	Code       string
	RetryAfter time.Duration
	Permanent  bool
	Uncertain  bool
}

func (e *APIError) Error() string { return "telegram API: " + e.Code }

// ReceiverAPI owns the dedicated receiver's bounded HTTP boundary.
// The existing notification sender retains its separate compatibility contract.
type ReceiverAPI struct {
	Client  HTTPClient
	BaseURL string
}

// NewReceiverAPI uses a client that refuses redirects carrying credentials.
func NewReceiverAPI() ReceiverAPI {
	return ReceiverAPI{
		BaseURL: DefaultAPIBaseURL,
		Client:  &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

type receiverEnvelope struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

// CheckConnection verifies credentials and webhook absence without sending messages.
func (a ReceiverAPI) CheckConnection(ctx context.Context, token string) (int64, error) {
	data, err := a.call(ctx, token, "getMe", struct{}{}, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var bot struct {
		ID    int64 `json:"id"`
		IsBot bool  `json:"is_bot"`
	}
	if err := json.Unmarshal(data, &bot); err != nil || bot.ID <= 0 || !bot.IsBot {
		return 0, aPIInvalid()
	}

	data, err = a.call(ctx, token, "getWebhookInfo", struct{}{}, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var webhook struct {
		URL *string `json:"url"`
	}
	if err := json.Unmarshal(data, &webhook); err != nil || webhook.URL == nil {
		return 0, aPIInvalid()
	}

	if *webhook.URL != "" {
		return 0, &APIError{Code: "webhook_conflict", Permanent: true}
	}

	return bot.ID, nil
}

// Poll never uses a negative offset or changes webhook configuration.
func (a ReceiverAPI) Poll(ctx context.Context, token string, next *int64, timeout int) ([]ReceivedUpdate, error) {
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

	return decodeReceiverUpdates(data)
}

// Send sends one bounded plain-text delivery to its original chat and topic.
func (a ReceiverAPI) Send(ctx context.Context, token string, peer PeerIdentity, text string) (int64, error) {
	return a.SendAction(ctx, token, peer, text, "", nil)
}

func (a ReceiverAPI) call(ctx context.Context, token, method string, payload any, timeout time.Duration) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("cannot encode telegram request")
	}

	return a.callBody(ctx, token, method, bytes.NewReader(body), "application/json", timeout)
}

func (a ReceiverAPI) callBody(ctx context.Context, token, method string, body io.Reader, contentType string, timeout time.Duration) (json.RawMessage, error) {
	if _, err := validateReceiverToken(token); err != nil {
		return nil, err
	}

	return a.postBody(ctx, token, method, body, contentType, timeout)
}

func (a ReceiverAPI) postBody(ctx context.Context, token, method string, body io.Reader, contentType string, timeout time.Duration) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(a.BaseURL, "/")+"/bot"+token+"/"+method, body)
	if err != nil {
		return nil, errors.New("invalid telegram endpoint")
	}

	request.Header.Set("Content-Type", contentType)

	response, err := a.Client.Do(request)
	if err != nil {
		return nil, &APIError{Code: "transport_failure", Uncertain: method == "sendMessage" || method == "sendDocument"}
	}

	defer func() {
		//nolint:errcheck // The bounded body reader reports response failures.
		_ = response.Body.Close()
	}()

	return decodeReceiverResponse(response)
}

func decodeReceiverResponse(response *http.Response) (json.RawMessage, error) {
	body, err := io.ReadAll(io.LimitReader(response.Body, receiverBodyLimit+1))
	if err != nil || len(body) > receiverBodyLimit {
		return nil, &APIError{Code: "response_limit_or_read_failure", Uncertain: true}
	}

	var envelope receiverEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, aPIInvalid()
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 || !envelope.OK {
		return nil, classifyAPIError(response.StatusCode, envelope)
	}

	if len(envelope.Result) == 0 || bytes.Equal(envelope.Result, []byte("null")) {
		return nil, aPIInvalid()
	}

	return envelope.Result, nil
}

func classifyAPIError(status int, envelope receiverEnvelope) error {
	code := envelope.ErrorCode
	if code == 0 {
		code = status
	}

	switch code {
	case 400:
		return classifyBadRequest(status, envelope)
	case 401, 403:
		return &APIError{Code: "credentials_or_access", Permanent: true}
	case 409:
		return &APIError{Code: "receiver_conflict", Permanent: true}
	case 429:
		if envelope.Parameters.RetryAfter > 86400 {
			return &APIError{Code: "retry_delay_exceeds_limit", Permanent: true}
		}

		delay := max(1, min(envelope.Parameters.RetryAfter, 86400))

		return &APIError{Code: "rate_limited", RetryAfter: time.Duration(delay) * time.Second}
	default:
		return &APIError{Code: "request_failed", Permanent: code < 500, Uncertain: code >= 500}
	}
}

func aPIInvalid() error {
	return &APIError{Code: "invalid_response", Permanent: true}
}

// Only the bounded API envelope is inspected; raw descriptions never reach logs.
func entityRejected(description string) bool {
	description = strings.ToLower(description)

	return strings.Contains(description, "can't parse entities") || strings.Contains(description, "cannot parse entities")
}

func classifyBadRequest(status int, envelope receiverEnvelope) error {
	if (status == http.StatusBadRequest || status >= 200 && status < 300) && !envelope.OK && entityRejected(envelope.Description) {
		return &APIError{Code: "format_rejected", Permanent: true}
	}

	return &APIError{Code: "request_failed", Permanent: true}
}
