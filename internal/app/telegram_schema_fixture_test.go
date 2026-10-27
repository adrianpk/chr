package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/core/database"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// The image is constructed by the production migrator once per test process,
// then read only. Every fixture writes its own file and opens its own real store;
// no receiver data or live connection is shared between tests.
var telegramAppSchemaImage = sync.OnceValues(newTelegramAppSchemaImage)

func newTelegramAppSchemaImage() (_ []byte, err error) {
	root, err := os.MkdirTemp("", "chr-telegram-schema-")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(root)) }()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: root})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	store, err := telegramcore.NewTelegramReceiverStore(ctx, env, testAssetsFS())
	if err != nil {
		return nil, err
	}

	// Closing the only connection checkpoints the WAL before the image is read.
	if err := store.Close(); err != nil {
		return nil, err
	}

	return os.ReadFile(database.UserDatabasePath(env))
}

func seedTelegramAppSchema(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	data, err := telegramAppSchemaImage()
	if err != nil {
		t.Fatal(err)
	}

	path := database.UserDatabasePath(env)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		t.Fatal(err)
	}
}
