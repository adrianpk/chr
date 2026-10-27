package core

import (
	"errors"
	"fmt"
	"os"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
)

type telegramModelClaim struct {
	source *os.File
	slot   *os.File
}

func acquireTelegramModelClaim(env TelegramStateDirectory, botID int64, name string) (*telegramModelClaim, error) {
	source, err := acquireTelegramFileLock(env, name)
	if err != nil {
		return nil, err
	}

	slot, err := acquireTelegramModelSlot(env, botID)
	if err != nil {
		return nil, errors.Join(err, source.Close())
	}

	return &telegramModelClaim{source: source, slot: slot}, nil
}

func (c *telegramModelClaim) Close() error {
	return errors.Join(c.slot.Close(), c.source.Close())
}

// Both interpretation and operation generation share two OS-owned provider
// slots. Source exclusion uses 16 fixed stripes per model purpose, so compaction
// cannot accumulate lock files. Capacity never reserves an attempt, and process
// death releases a slot. A stripe collision defers input without an attempt.
func acquireTelegramModelSlot(env TelegramStateDirectory, botID int64) (*os.File, error) {
	for slot := range 2 {
		lock, err := acquireTelegramFileLock(env, fmt.Sprintf("model-%d-%d.lock", botID, slot))
		if err == nil {
			return lock, nil
		}

		var apiErr *TelegramAPIError
		if !errors.As(err, &apiErr) || apiErr.Code != "local_receiver_conflict" {
			return nil, err
		}
	}

	return nil, modelcore.ErrCapacity
}
