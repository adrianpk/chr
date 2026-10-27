package app

import (
	"strconv"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// RemoteCommandRequestFromTelegramUpdate executes the corresponding package operation.
func remoteCommandRequestFromTelegramUpdate(update telegramcore.TelegramUpdate, now time.Time) (core.RemoteCommandRequest, error) {
	sender := remoteCommandSenderFromTelegramUpdate(update)
	receivedAt := telegramUpdateReceivedAt(update, now)

	return core.ParseRemoteCommandRequest(core.RemoteCommandParseOptions{
		Sender:     sender,
		Text:       update.Message.Text,
		ReceivedAt: receivedAt,
		CreatedAt:  now,
	})
}

// RemoteCommandRecordFromTelegramUpdate executes the corresponding package operation.
func remoteCommandRecordFromTelegramUpdate(update telegramcore.TelegramUpdate, now time.Time) (core.RemoteCommandRecord, error) {
	request, err := remoteCommandRequestFromTelegramUpdate(update, now)
	if err == nil {
		return core.NewAcceptedRemoteCommandRecord(request)
	}

	sender := remoteCommandSenderFromTelegramUpdate(update)
	receivedAt := telegramUpdateReceivedAt(update, now)

	return core.NewRejectedRemoteCommandRecord(
		core.RemoteCommandRequest{},
		update.Message.Text,
		sender,
		err.Error(),
		receivedAt,
		now,
	)
}

// RemoteCommandSenderFromTelegramUpdate executes the corresponding package operation.
func remoteCommandSenderFromTelegramUpdate(update telegramcore.TelegramUpdate) core.RemoteCommandSender {
	externalID := strconv.FormatInt(update.Message.From.ID, 10)
	if update.Message.From.ID == 0 {
		externalID = strconv.FormatInt(update.Message.Chat.ID, 10)
	}

	return core.RemoteCommandSender{
		Channel:     "telegram",
		ExternalID:  externalID,
		DisplayName: telegramSenderDisplayName(update.Message.From),
	}
}

// TelegramUpdateReceivedAt executes the corresponding package operation.
func telegramUpdateReceivedAt(update telegramcore.TelegramUpdate, fallback time.Time) time.Time {
	if update.Message.Date <= 0 {
		return fallback
	}

	return time.Unix(update.Message.Date, 0).UTC()
}

func telegramSenderDisplayName(sender telegramcore.TelegramSenderUser) string {
	if sender.Username != "" {
		return sender.Username
	}

	return sender.FirstName
}
