package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

// TestProvingLogFailure keeps cleanup failures visible alongside the command error.
func TestProvingLogFailure(t *testing.T) {
	commandError := errors.New("command failed")

	cases := []struct {
		name string
		err  error
	}{
		{"close failure", nil},
		{"command and close", commandError},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			result, err := RunProvingSandbox(t.Context(), ProvingSandboxRunOptions{
				Worker: codex.Runner{}, ProjectDir: newTestProvingProject(t),
				ObjectivePath: filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md"),
				Executor:      closedProvingLogExecutor{err: item.err},
			})
			if !errors.Is(err, os.ErrClosed) || item.err != nil && !errors.Is(err, item.err) {
				t.Fatalf("error = %v, want close failure and command error %v", err, item.err)
			}

			if item.err == nil && result.Objective.ID == "" {
				t.Fatal("successful command outcome was discarded after cleanup failed")
			}
		})
	}
}

// This fixture closes only the owned log to force a real OS cleanup error;
// Docker execution is outside the resource-release contract under test.
type closedProvingLogExecutor struct {
	err error
}

func (e closedProvingLogExecutor) Run(_ context.Context, _ HarnessCommand, stdout, _ io.Writer) error {
	file, ok := stdout.(*os.File)
	if !ok {
		return fmt.Errorf("expected owned log file, got %T", stdout)
	}

	if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		return err
	}

	return e.err
}
