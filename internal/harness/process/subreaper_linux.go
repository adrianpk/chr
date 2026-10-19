package process

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// EnableChildReaping is for an independently owned bridge process only. It
// adopts children detached by its model tooling when their parent exits.
func EnableChildReaping() error {
	const setChildSubreaper = 36 // Linux PR_SET_CHILD_SUBREAPER.

	_, _, err := syscall.RawSyscall6(syscall.SYS_PRCTL, setChildSubreaper, 1, 0, 0, 0, 0)
	if err != 0 {
		return err
	}

	return nil
}

// ReapChildren stops adopted children before the bridge releases its runtime
// reservation. It must run after the supervised direct command has been waited.
func ReapChildren(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := stopAdoptedChildren(); err != nil {
			return err
		}

		var status syscall.WaitStatus

		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		if errors.Is(err, syscall.ECHILD) {
			return nil
		}

		if err != nil && !errors.Is(err, syscall.EINTR) {
			return err
		}

		if pid == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
}

func stopAdoptedChildren() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue // A child can exit between enumeration and observation.
		}

		end := strings.LastIndexByte(string(data), ')')
		if end < 0 {
			continue
		}

		parts := strings.Fields(string(data[end+1:]))
		if len(parts) < 2 || parts[1] != strconv.Itoa(os.Getpid()) {
			continue
		}

		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}

	return nil
}
