package telegram

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

// DefaultAPIBaseURL defines a package boundary constant.
const DefaultAPIBaseURL = "https://api.telegram.org"

// Message describes a package boundary value.
type Message struct {
	ChatID string
	Text   string
}

// SendResult describes a package boundary value.
type SendResult struct {
	StatusCode int
	Body       string
}

// HTTPClient defines a package boundary interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Sender describes a package boundary value.
type Sender struct {
	Client  HTTPClient
	BaseURL string
}

// RenderNotificationMessage renders the corresponding package data.
func RenderNotificationMessage(event notifycore.Event) Message {
	text := event.Payload.Summary
	if event.Payload.Detail != "" {
		text = text + "\n\n" + event.Payload.Detail
	}

	return Message{
		Text: text,
	}
}

// NewSender constructs the corresponding package service or value.
func NewSender(client HTTPClient) Sender {
	if client == nil {
		client = http.DefaultClient
	}

	return Sender{
		Client:  client,
		BaseURL: DefaultAPIBaseURL,
	}
}

// SendMessage sends the requested message.
func (s Sender) SendMessage(ctx context.Context, token string, message Message) (SendResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return SendResult{}, fmt.Errorf("telegram bot token is required")
	}

	if strings.TrimSpace(message.ChatID) == "" {
		return SendResult{}, fmt.Errorf("telegram chat id is required")
	}

	if strings.TrimSpace(message.Text) == "" {
		return SendResult{}, fmt.Errorf("telegram message text is required")
	}

	request, err := s.buildSendMessageRequest(ctx, token, message)
	if err != nil {
		return SendResult{}, err
	}

	response, err := s.Client.Do(request)
	if err != nil {
		return SendResult{}, fmt.Errorf("send telegram message: %w", err)
	}

	defer func() {
		//nolint:errcheck // Response body read errors are returned below.
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return SendResult{}, fmt.Errorf("read telegram response: %w", err)
	}

	result := SendResult{
		StatusCode: response.StatusCode,
		Body:       string(body),
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, fmt.Errorf("telegram send failed with status %d", response.StatusCode)
	}

	return result, nil
}

func (s Sender) buildSendMessageRequest(ctx context.Context, token string, message Message) (*http.Request, error) {
	baseURL := strings.TrimRight(s.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
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
