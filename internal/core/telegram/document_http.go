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
func (a ReceiverAPI) SendDocument(ctx context.Context, token string, send DocumentSendRequest) (tgmessage.Acknowledgement, error) {
	var acknowledgement tgmessage.Acknowledgement
	if err := tgmessage.ValidateDocument(send.Document); err != nil {
		return acknowledgement, err
	}

	if !validReceiverText(send.Caption) || tgmessage.VisibleUnits(send.Caption) > 900 || send.Peer.ChatID == 0 || send.Peer.TopicID < 0 || len(send.Entities) > 32 {
		return acknowledgement, errors.New("invalid telegram document caption")
	}

	keyboard, err := actionKeyboard(send.ActionToken, send.Choices, send.Labels...)
	if err != nil {
		return acknowledgement, err
	}

	multipart2, err := documentMultipart(send.Peer, send.Document, send.Caption, send.Entities, keyboard)
	body, contentType := multipart2.Payload, multipart2.ContentType

	if err != nil {
		return acknowledgement, err
	}

	data, err := a.callBody(ctx, token, requestBody{Method: "sendDocument", Body: bytes.NewReader(body), ContentType: contentType, Timeout: 10 * time.Second})
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

func documentMultipart(peer PeerIdentity, document tgmessage.Document, caption string, entities []tgmessage.Entity, keyboard *inlineKeyboard) (multipartBody, error) {
	return documentMultipartDestination(documentUpload{Destination: strconv.FormatInt(peer.ChatID, 10), Topic: peer.TopicID, Document: document, Caption: caption, Entities: entities, Keyboard: keyboard})
}

func documentMultipartDestination(upload documentUpload) (multipartBody, error) {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	fields := map[string]string{"chat_id": upload.Destination, "caption": upload.Caption, "disable_content_type_detection": "true"}
	if upload.Topic > 0 {
		fields["message_thread_id"] = strconv.FormatInt(upload.Topic, 10)
	}

	if err := documentJSONFields(fields, upload.Entities, upload.Keyboard); err != nil {
		return multipartBody{Payload: nil, ContentType: ""}, err
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return multipartBody{Payload: nil, ContentType: ""}, err
		}
	}

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="document"; filename="%s"`, upload.Document.Filename))
	header.Set("Content-Type", upload.Document.MediaType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return multipartBody{Payload: nil, ContentType: ""}, err
	}

	if _, err := part.Write(upload.Document.Payload); err != nil {
		return multipartBody{Payload: nil, ContentType: ""}, err
	}

	if err := writer.Close(); err != nil {
		return multipartBody{Payload: nil, ContentType: ""}, err
	}

	return multipartBody{Payload: body.Bytes(), ContentType: writer.FormDataContentType()}, nil
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
