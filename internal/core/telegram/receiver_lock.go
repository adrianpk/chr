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

// AcquireReceiverLock holds OS exclusion for one verified bot.
// Callers close the returned descriptor; the lock file must never be deleted.
func AcquireReceiverLock(env StateDirectory, botID int64) (*os.File, error) {
	if botID <= 0 {
		return nil, errors.New("verified telegram bot identity is required")
	}

	return acquireFileLock(env, fmt.Sprintf("bot-%d.lock", botID))
}

// AcquireConfigurationLock excludes setup replacement while a receiver runs.
func AcquireConfigurationLock(env StateDirectory) (*os.File, error) {
	return acquireFileLock(env, "configuration.lock")
}

func acquireFileLock(env StateDirectory, name string) (*os.File, error) {
	directory := filepath.Join(env.StateDir(), "telegram")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(directory, name), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, errors.Join(&APIError{Code: "local_receiver_conflict", Permanent: true}, file.Close())
	}

	return file, nil
}

// BindReceiverIdentity refuses silent bot replacement. Setup --replace
// clears only this binding while holding the configuration lock.
func BindReceiverIdentity(env StateDirectory, botID int64) error {
	path := filepath.Join(env.StateDir(), "telegram", "bot-id")

	prior, err := readPrivateFile(path, 32)
	if errors.Is(err, os.ErrNotExist) {
		return writePrivateFile(path, []byte(strconv.FormatInt(botID, 10)+"\n"))
	}

	if err != nil {
		return err
	}

	if strings.TrimSpace(string(prior)) != strconv.FormatInt(botID, 10) {
		return &APIError{Code: "bot_identity_changed", Permanent: true}
	}

	return nil
}

// ResetReceiverIdentity is called only by explicit setup replacement.
func ResetReceiverIdentity(env StateDirectory) error {
	return removeTemp(filepath.Join(env.StateDir(), "telegram", "bot-id"))
}
