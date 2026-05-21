package chc

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

type TelegramUpdateEnvelope struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

type TelegramUpdate struct {
	UpdateID int64                 `json:"update_id"`
	Message  TelegramUpdateMessage `json:"message"`
}

type TelegramUpdateMessage struct {
	MessageID int64              `json:"message_id"`
	From      TelegramSenderUser `json:"from"`
	Chat      TelegramChat       `json:"chat"`
	Date      int64              `json:"date"`
	Text      string             `json:"text"`
}

type TelegramSenderUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type TelegramChat struct {
	ID int64 `json:"id"`
}

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

func RemoteCommandRequestFromTelegramUpdate(
	update TelegramUpdate,
	now time.Time,
) (RemoteCommandRequest, error) {
	sender := RemoteCommandSenderFromTelegramUpdate(update)
	receivedAt := TelegramUpdateReceivedAt(update, now)

	return ParseRemoteCommandRequest(RemoteCommandParseOptions{
		Sender:     sender,
		Text:       update.Message.Text,
		ReceivedAt: receivedAt,
		CreatedAt:  now,
	})
}

func RemoteCommandRecordFromTelegramUpdate(
	update TelegramUpdate,
	now time.Time,
) (RemoteCommandRecord, error) {
	request, err := RemoteCommandRequestFromTelegramUpdate(update, now)
	if err == nil {
		return NewAcceptedRemoteCommandRecord(request)
	}

	sender := RemoteCommandSenderFromTelegramUpdate(update)
	receivedAt := TelegramUpdateReceivedAt(update, now)

	return NewRejectedRemoteCommandRecord(
		RemoteCommandRequest{},
		update.Message.Text,
		sender,
		err.Error(),
		receivedAt,
		now,
	)
}

func RemoteCommandSenderFromTelegramUpdate(update TelegramUpdate) RemoteCommandSender {
	externalID := strconv.FormatInt(update.Message.From.ID, 10)
	if update.Message.From.ID == 0 {
		externalID = strconv.FormatInt(update.Message.Chat.ID, 10)
	}

	return RemoteCommandSender{
		Channel:     "telegram",
		ExternalID:  externalID,
		DisplayName: telegramSenderDisplayName(update.Message.From),
	}
}

func TelegramUpdateReceivedAt(update TelegramUpdate, fallback time.Time) time.Time {
	if update.Message.Date <= 0 {
		return fallback
	}

	return time.Unix(update.Message.Date, 0).UTC()
}

func telegramSenderDisplayName(sender TelegramSenderUser) string {
	if sender.Username != "" {
		return sender.Username
	}

	return sender.FirstName
}
