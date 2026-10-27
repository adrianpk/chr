package telegram

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

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// SendDocument uploads frozen bytes and a bounded caption through the same
// receiver HTTP boundary. The outbox alone owns fallback and attempt allowance.
func (a ReceiverAPI) SendDocument(ctx context.Context, token string, peer PeerIdentity, document tgmessage.Document, caption string, entities []tgmessage.Entity, actionToken string, choices []string, labels ...[]string) (tgmessage.Acknowledgement, error) {
	var acknowledgement tgmessage.Acknowledgement
	if err := tgmessage.ValidateDocument(document); err != nil {
		return acknowledgement, err
	}

	if !validReceiverText(caption) || tgmessage.VisibleUnits(caption) > 900 || peer.ChatID == 0 || peer.TopicID < 0 || len(entities) > 32 {
		return acknowledgement, errors.New("invalid telegram document caption")
	}

	keyboard, err := actionKeyboard(actionToken, choices, labels...)
	if err != nil {
		return acknowledgement, err
	}

	body, contentType, err := documentMultipart(peer, document, caption, entities, keyboard)
	if err != nil {
		return acknowledgement, err
	}

	data, err := a.callBody(ctx, token, "sendDocument", bytes.NewReader(body), contentType, 10*time.Second)
	if err != nil {
		return acknowledgement, documentSendFailure(err)
	}

	return decodeDocumentAcknowledgement(data)
}

func decodeDocumentAcknowledgement(data []byte) (tgmessage.Acknowledgement, error) {
	var (
		acknowledgement tgmessage.Acknowledgement
		message         struct {
			MessageID int64 `json:"message_id"`
			Document  struct {
				FileID string `json:"file_id"`
			} `json:"document"`
		}
	)
	if err := json.Unmarshal(data, &message); err != nil || message.MessageID <= 0 || message.Document.FileID == "" || len(message.Document.FileID) > 256 {
		return acknowledgement, &APIError{Code: "invalid_document_response", Uncertain: true}
	}

	return tgmessage.Acknowledgement{MessageID: message.MessageID, FileID: message.Document.FileID}, nil
}

func documentMultipart(peer PeerIdentity, document tgmessage.Document, caption string, entities []tgmessage.Entity, keyboard *inlineKeyboard) ([]byte, string, error) {
	return documentMultipartDestination(strconv.FormatInt(peer.ChatID, 10), peer.TopicID, document, caption, entities, keyboard)
}

func documentMultipartDestination(destination string, topic int64, document tgmessage.Document, caption string, entities []tgmessage.Entity, keyboard *inlineKeyboard) ([]byte, string, error) {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	fields := map[string]string{"chat_id": destination, "caption": caption, "disable_content_type_detection": "true"}
	if topic > 0 {
		fields["message_thread_id"] = strconv.FormatInt(topic, 10)
	}

	if err := documentJSONFields(fields, entities, keyboard); err != nil {
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

func documentJSONFields(fields map[string]string, entities []tgmessage.Entity, keyboard *inlineKeyboard) error {
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

func documentSendFailure(err error) error {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Code == "invalid_response" {
		return &APIError{Code: "invalid_document_response", Uncertain: true}
	}

	return err
}
