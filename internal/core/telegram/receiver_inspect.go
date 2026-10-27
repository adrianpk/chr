package telegram

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/database"
)

// BoundBotID reads the local verified binding without claiming receiver ownership.
func BoundBotID(env StateDirectory) (int64, error) {
	data, err := readPrivateFile(filepath.Join(env.StateDir(), "telegram", "bot-id"), 32)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid retained telegram bot identity")
	}

	return id, nil
}

// InspectReceiver reads local status without migrations, network calls or state creation.
func InspectReceiver(ctx context.Context, env database.DataDirectory, botID int64) (state ReceiverState, err error) {
	db, err := database.OpenReadOnly(ctx, env)
	if err != nil {
		return state, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	return (&ReceiverStore{db: db}).Receiver(ctx, botID)
}
