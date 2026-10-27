package telegram

import (
	"errors"
	"fmt"
	"os"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
)

type modelClaim struct {
	source *os.File
	slot   *os.File
}

func acquireModelClaim(env StateDirectory, botID int64, name string) (*modelClaim, error) {
	source, err := acquireFileLock(env, name)
	if err != nil {
		return nil, err
	}

	slot, err := acquireModelSlot(env, botID)
	if err != nil {
		return nil, errors.Join(err, source.Close())
	}

	return &modelClaim{source: source, slot: slot}, nil
}

func (c *modelClaim) Close() error {
	return errors.Join(c.slot.Close(), c.source.Close())
}

// Both interpretation and operation generation share two OS-owned provider
// slots. Source exclusion uses 16 fixed stripes per model purpose, so compaction
// cannot accumulate lock files. Capacity never reserves an attempt, and process
// death releases a slot. A stripe collision defers input without an attempt.
func acquireModelSlot(env StateDirectory, botID int64) (*os.File, error) {
	for slot := range 2 {
		lock, err := acquireFileLock(env, fmt.Sprintf("model-%d-%d.lock", botID, slot))
		if err == nil {
			return lock, nil
		}

		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "local_receiver_conflict" {
			return nil, err
		}
	}

	return nil, modelcore.ErrCapacity
}
