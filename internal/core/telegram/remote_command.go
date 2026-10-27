package telegram

import (
	"encoding/json"
	"fmt"
	"io"
)

// UpdateEnvelope describes a package boundary value.
type UpdateEnvelope struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// Update describes a package boundary value.
type Update struct {
	UpdateID int64         `json:"update_id"`
	Message  UpdateMessage `json:"message"`
}

// UpdateMessage describes a package boundary value.
type UpdateMessage struct {
	MessageID int64      `json:"message_id"`
	From      SenderUser `json:"from"`
	Chat      Chat       `json:"chat"`
	Date      int64      `json:"date"`
	Text      string     `json:"text"`
}

// SenderUser describes a package boundary value.
type SenderUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// Chat describes a package boundary value.
type Chat struct {
	ID int64 `json:"id"`
}

// DecodeUpdates decodes the corresponding package payload.
func DecodeUpdates(reader io.Reader) ([]Update, error) {
	var envelope UpdateEnvelope

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&envelope)
	if err != nil {
		return nil, fmt.Errorf("decode telegram updates: %w", err)
	}

	return envelope.Result, nil
}
