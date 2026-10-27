package telegram

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TelegramBoundBotID reads the local verified binding without claiming receiver ownership.
func TelegramBoundBotID(env TelegramStateDirectory) (int64, error) {
	data, err := readTelegramPrivateFile(filepath.Join(env.StateDir(), "telegram", "bot-id"), 32)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid retained telegram bot identity")
	}

	return id, nil
}

// InspectTelegramReceiver reads local status without migrations, network calls or state creation.
func InspectTelegramReceiver(ctx context.Context, env database.DataDirectory, botID int64) (state TelegramReceiverState, err error) {
	db, err := database.OpenReadOnly(ctx, env)
	if err != nil {
		return state, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	return (&TelegramReceiverStore{db: db}).Receiver(ctx, botID)
}
