package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strconv"
	"time"
)

// TelegramDocumentAcknowledgement is transport evidence for the original send.
type TelegramDocumentAcknowledgement struct {
	MessageID int64
	FileID    string
}

// SendDocument uploads frozen bytes and a bounded caption through the same
// receiver HTTP boundary. The outbox alone owns fallback and attempt allowance.
func (a TelegramReceiverAPI) SendDocument(ctx context.Context, token string, peer TelegramPeerIdentity, document TelegramDocument, caption string, entities []TelegramTextEntity, actionToken string, choices []string, labels ...[]string) (TelegramDocumentAcknowledgement, error) {
	var acknowledgement TelegramDocumentAcknowledgement
	if err := validateTelegramDocument(document); err != nil {
		return acknowledgement, err
	}

	if !validTelegramReceiverText(caption) || telegramVisibleUnits(caption) > 900 || peer.ChatID == 0 || peer.TopicID < 0 || len(entities) > 32 {
		return acknowledgement, errors.New("invalid telegram document caption")
	}

	keyboard, err := telegramActionKeyboard(actionToken, choices, labels...)
	if err != nil {
		return acknowledgement, err
	}

	body, contentType, err := telegramDocumentMultipart(peer, document, caption, entities, keyboard)
	if err != nil {
		return acknowledgement, err
	}

	data, err := a.callBody(ctx, token, "sendDocument", bytes.NewReader(body), contentType, 10*time.Second)
	if err != nil {
		return acknowledgement, telegramDocumentSendFailure(err)
	}

	return decodeTelegramDocumentAcknowledgement(data)
}

func decodeTelegramDocumentAcknowledgement(data []byte) (TelegramDocumentAcknowledgement, error) {
	var (
		acknowledgement TelegramDocumentAcknowledgement
		message         struct {
			MessageID int64 `json:"message_id"`
			Document  struct {
				FileID string `json:"file_id"`
			} `json:"document"`
		}
	)
	if err := json.Unmarshal(data, &message); err != nil || message.MessageID <= 0 || message.Document.FileID == "" || len(message.Document.FileID) > 256 {
		return acknowledgement, &TelegramAPIError{Code: "invalid_document_response", Uncertain: true}
	}

	return TelegramDocumentAcknowledgement{MessageID: message.MessageID, FileID: message.Document.FileID}, nil
}

func telegramDocumentMultipart(peer TelegramPeerIdentity, document TelegramDocument, caption string, entities []TelegramTextEntity, keyboard *telegramInlineKeyboard) ([]byte, string, error) {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	fields := map[string]string{"chat_id": strconv.FormatInt(peer.ChatID, 10), "caption": caption, "disable_content_type_detection": "true"}
	if peer.TopicID > 0 {
		fields["message_thread_id"] = strconv.FormatInt(peer.TopicID, 10)
	}

	if err := telegramDocumentJSONFields(fields, entities, keyboard); err != nil {
		return nil, "", err
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="document"; filename="%s"`, document.Filename))
	header.Set("Content-Type", document.MediaType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", err
	}

	if _, err := part.Write(document.Payload); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body.Bytes(), writer.FormDataContentType(), nil
}

func telegramDocumentJSONFields(fields map[string]string, entities []TelegramTextEntity, keyboard *telegramInlineKeyboard) error {
	if len(entities) > 0 {
		payload, err := json.Marshal(entities)
		if err != nil {
			return err
		}

		fields["caption_entities"] = string(payload)
	}

	if keyboard != nil {
		payload, err := json.Marshal(keyboard)
		if err != nil {
			return err
		}

		fields["reply_markup"] = string(payload)
	}

	return nil
}

func telegramDocumentSendFailure(err error) error {
	var apiErr *TelegramAPIError
	if errors.As(err, &apiErr) && apiErr.Code == "invalid_response" {
		return &TelegramAPIError{Code: "invalid_document_response", Uncertain: true}
	}

	return err
}
