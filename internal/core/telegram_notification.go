package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
)

// DefaultTelegramAPIBaseURL defines a package boundary constant.
const DefaultTelegramAPIBaseURL = "https://api.telegram.org"

// TelegramMessage describes a package boundary value.
type TelegramMessage struct {
	ChatID string
	Text   string
}

// TelegramSendResult describes a package boundary value.
type TelegramSendResult struct {
	StatusCode int
	Body       string
}

// TelegramHTTPClient defines a package boundary interface.
type TelegramHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TelegramSender describes a package boundary value.
type TelegramSender struct {
	Client  TelegramHTTPClient
	BaseURL string
}

// RenderTelegramNotificationMessage renders the corresponding package data.
func RenderTelegramNotificationMessage(event notifycore.Event) TelegramMessage {
	text := event.Payload.Summary
	if event.Payload.Detail != "" {
		text = text + "\n\n" + event.Payload.Detail
	}

	return TelegramMessage{
		Text: text,
	}
}

// NewTelegramSender constructs the corresponding package service or value.
func NewTelegramSender(client TelegramHTTPClient) TelegramSender {
	if client == nil {
		client = http.DefaultClient
	}

	return TelegramSender{
		Client:  client,
		BaseURL: DefaultTelegramAPIBaseURL,
	}
}

// SendMessage sends the requested message.
func (s TelegramSender) SendMessage(
	ctx context.Context,
	token string,
	message TelegramMessage,
) (TelegramSendResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return TelegramSendResult{}, fmt.Errorf("telegram bot token is required")
	}

	if strings.TrimSpace(message.ChatID) == "" {
		return TelegramSendResult{}, fmt.Errorf("telegram chat id is required")
	}

	if strings.TrimSpace(message.Text) == "" {
		return TelegramSendResult{}, fmt.Errorf("telegram message text is required")
	}

	request, err := s.buildSendMessageRequest(ctx, token, message)
	if err != nil {
		return TelegramSendResult{}, err
	}

	response, err := s.Client.Do(request)
	if err != nil {
		return TelegramSendResult{}, fmt.Errorf("send telegram message: %w", err)
	}

	defer func() {
		//nolint:errcheck // Response body read errors are returned below.
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return TelegramSendResult{}, fmt.Errorf("read telegram response: %w", err)
	}

	result := TelegramSendResult{
		StatusCode: response.StatusCode,
		Body:       string(body),
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, fmt.Errorf("telegram send failed with status %d", response.StatusCode)
	}

	return result, nil
}

func (s TelegramSender) buildSendMessageRequest(
	ctx context.Context,
	token string,
	message TelegramMessage,
) (*http.Request, error) {
	baseURL := strings.TrimRight(s.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultTelegramAPIBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, "bot"+token, "sendMessage")
	if err != nil {
		return nil, fmt.Errorf("build telegram endpoint: %w", err)
	}

	payload := struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: message.ChatID,
		Text:   message.Text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode telegram payload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build telegram request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
