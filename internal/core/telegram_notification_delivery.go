package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// SendPresentation shares the receiver's bounded HTTP boundary and multipart
// uploader while retaining the notification sender and string destination.
func (s TelegramSender) SendPresentation(ctx context.Context, token, destination string, delivery TelegramNotificationDelivery) (TelegramDocumentAcknowledgement, error) {
	if strings.TrimSpace(token) == "" {
		return TelegramDocumentAcknowledgement{}, fmt.Errorf("telegram bot token is required")
	}

	if strings.TrimSpace(destination) == "" {
		return TelegramDocumentAcknowledgement{}, fmt.Errorf("telegram chat id is required")
	}

	if err := validateTelegramNotificationDelivery(delivery); err != nil {
		return TelegramDocumentAcknowledgement{}, err
	}

	api := s.notificationAPI()
	presentation := delivery.Presentation

	entities := presentation.Entities
	if delivery.PlainFallback {
		entities = nil
	}

	if delivery.Document != nil {
		return api.sendNotificationDocument(ctx, token, destination, delivery, entities)
	}

	return api.sendNotificationText(ctx, token, destination, presentation.PlainFallback, entities)
}

func (api TelegramReceiverAPI) sendNotificationDocument(ctx context.Context, token, destination string, delivery TelegramNotificationDelivery, entities []TelegramTextEntity) (TelegramDocumentAcknowledgement, error) {
	body, contentType, err := telegramDocumentMultipartDestination(destination, 0, *delivery.Document, delivery.Presentation.PlainFallback, entities, nil)
	if err != nil {
		return TelegramDocumentAcknowledgement{}, err
	}

	data, err := api.postBody(ctx, token, "sendDocument", bytes.NewReader(body), contentType, 10*time.Second)
	if err != nil {
		return TelegramDocumentAcknowledgement{}, telegramDocumentSendFailure(err)
	}

	return decodeTelegramDocumentAcknowledgement(data)
}

func (api TelegramReceiverAPI) sendNotificationText(ctx context.Context, token, destination, text string, entities []TelegramTextEntity) (TelegramDocumentAcknowledgement, error) {
	payload := struct {
		ChatID   string               `json:"chat_id"`
		Text     string               `json:"text"`
		Entities []TelegramTextEntity `json:"entities,omitempty"`
	}{destination, text, entities}

	body, err := json.Marshal(payload)
	if err != nil {
		return TelegramDocumentAcknowledgement{}, err
	}

	data, err := api.postBody(ctx, token, "sendMessage", bytes.NewReader(body), "application/json", 10*time.Second)
	if err != nil {
		return TelegramDocumentAcknowledgement{}, telegramDocumentSendFailure(err)
	}

	var result struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.MessageID <= 0 {
		return TelegramDocumentAcknowledgement{}, &TelegramAPIError{Code: "invalid_response", Uncertain: true}
	}

	return TelegramDocumentAcknowledgement{MessageID: result.MessageID}, nil
}

func (s TelegramSender) notificationAPI() TelegramReceiverAPI {
	api := NewTelegramReceiverAPI()
	if s.BaseURL != "" {
		api.BaseURL = s.BaseURL
	}

	if client, ok := s.Client.(*http.Client); ok {
		copyClient := *client
		copyClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
		api.Client = &copyClient
	} else if s.Client != nil {
		api.Client = s.Client
	}

	return api
}

// TelegramNotificationFailure retains a safe classification and uncertainty;
// notification audit data never receives a transport error containing a token.
func TelegramNotificationFailure(failure error) (string, bool) {
	var api *TelegramAPIError
	if errors.As(failure, &api) {
		return api.Code, api.Uncertain
	}

	if errors.Is(failure, ErrTelegramNotificationTargetUnavailable) {
		return "notification_target_unavailable", false
	}

	if errors.Is(failure, ErrTelegramDocumentUnavailable) {
		return "notification_document_unavailable", false
	}

	return "notification_delivery_failed", false
}

func validateTelegramNotificationDelivery(delivery TelegramNotificationDelivery) error {
	p := delivery.Presentation
	if p.PlainFallback == "" || len(p.InitialText)+len(p.Content)+len(p.SecondaryText)+len(p.PlainFallback) > 64<<10 || len(p.Entities) > 32 {
		return ErrTelegramDocumentUnavailable
	}

	if err := validateTelegramNotificationActions(p); err != nil {
		return err
	}

	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	check := TelegramReceiverDelivery{Text: p.PlainFallback, DocumentID: p.DocumentID}
	if err := decodeTelegramPresentationHash(string(payload), contenthash.Sum(payload), &check); err != nil {
		return err
	}

	return validateTelegramNotificationAttachment(p, delivery.Document)
}

func validateTelegramNotificationAttachment(p TelegramPresentation, document *TelegramDocument) error {
	if document == nil {
		if p.DocumentID != "" || p.DocumentHash != "" {
			return ErrTelegramDocumentUnavailable
		}

		return nil
	}

	if p.TemplateRevision != "document-v1" || p.DocumentID == "" || p.DocumentHash != contenthash.Sum(document.Payload) {
		return ErrTelegramDocumentUnavailable
	}

	return validateTelegramDocument(*document)
}
