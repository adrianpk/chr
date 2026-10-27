package telegram

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// AcquireTelegramReceiverLock holds OS exclusion for one verified bot.
// Callers close the returned descriptor; the lock file must never be deleted.
func AcquireTelegramReceiverLock(env TelegramStateDirectory, botID int64) (*os.File, error) {
	if botID <= 0 {
		return nil, errors.New("verified telegram bot identity is required")
	}

	return acquireTelegramFileLock(env, fmt.Sprintf("bot-%d.lock", botID))
}

// AcquireTelegramConfigurationLock excludes setup replacement while a receiver runs.
func AcquireTelegramConfigurationLock(env TelegramStateDirectory) (*os.File, error) {
	return acquireTelegramFileLock(env, "configuration.lock")
}

func acquireTelegramFileLock(env TelegramStateDirectory, name string) (*os.File, error) {
	directory := filepath.Join(env.StateDir(), "telegram")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(directory, name), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, errors.Join(&TelegramAPIError{Code: "local_receiver_conflict", Permanent: true}, file.Close())
	}

	return file, nil
}

// BindTelegramReceiverIdentity refuses silent bot replacement. Setup --replace
// clears only this binding while holding the configuration lock.
func BindTelegramReceiverIdentity(env TelegramStateDirectory, botID int64) error {
	path := filepath.Join(env.StateDir(), "telegram", "bot-id")

	prior, err := readTelegramPrivateFile(path, 32)
	if errors.Is(err, os.ErrNotExist) {
		return writeTelegramPrivateFile(path, []byte(strconv.FormatInt(botID, 10)+"\n"))
	}

	if err != nil {
		return err
	}

	if strings.TrimSpace(string(prior)) != strconv.FormatInt(botID, 10) {
		return &TelegramAPIError{Code: "bot_identity_changed", Permanent: true}
	}

	return nil
}

// ResetTelegramReceiverIdentity is called only by explicit setup replacement.
func ResetTelegramReceiverIdentity(env TelegramStateDirectory) error {
	return removeTelegramTemp(filepath.Join(env.StateDir(), "telegram", "bot-id"))
}
