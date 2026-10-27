package file

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"
)

// Lock acquires the existing advisory exclusive lock with explicit cancellation.
// The caller owns unlock and descriptor close on every exit.
func Lock(ctx context.Context, file *os.File) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}

		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
