package process

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SIGKILL is asynchronous. Waiting for the direct child alone does not prove
// that every group member has stopped, even after its pipe descriptors close.
// Ignore zombies, which cannot execute; their owning parent performs reaping.
func waitKilledGroup(ctx context.Context, group int) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()

	for {
		alive, err := killedGroupAlive(group)
		if err != nil || !alive {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
}

func killedGroupAlive(group int) (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}

		data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue // A group member can exit while its stat is being opened.
		}

		end := strings.LastIndexByte(string(data), ')')
		if end < 0 {
			continue
		}

		fields := strings.Fields(string(data[end+1:]))
		if len(fields) >= 3 && fields[2] == strconv.Itoa(group) && fields[0] != "Z" && fields[0] != "X" {
			return true, nil
		}
	}

	return false, nil
}
