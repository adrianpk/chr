package core

import (
	"encoding/json"
	"fmt"
	"io"
)

// TelegramUpdateEnvelope describes a package boundary value.
type TelegramUpdateEnvelope struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

// TelegramUpdate describes a package boundary value.
type TelegramUpdate struct {
	UpdateID int64                 `json:"update_id"`
	Message  TelegramUpdateMessage `json:"message"`
}

// TelegramUpdateMessage describes a package boundary value.
type TelegramUpdateMessage struct {
	MessageID int64              `json:"message_id"`
	From      TelegramSenderUser `json:"from"`
	Chat      TelegramChat       `json:"chat"`
	Date      int64              `json:"date"`
	Text      string             `json:"text"`
}

// TelegramSenderUser describes a package boundary value.
type TelegramSenderUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// TelegramChat describes a package boundary value.
type TelegramChat struct {
	ID int64 `json:"id"`
}

// DecodeTelegramUpdates decodes the corresponding package payload.
func DecodeTelegramUpdates(reader io.Reader) ([]TelegramUpdate, error) {
	var envelope TelegramUpdateEnvelope

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&envelope)
	if err != nil {
		return nil, fmt.Errorf("decode telegram updates: %w", err)
	}

	return envelope.Result, nil
}
