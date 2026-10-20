package process

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// Exercise real pipes so an overflow must cancel and reap a blocked subprocess.
func TestAgentRunProcessStreams(t *testing.T) {
	for _, test := range []struct {
		name      string
		script    string
		limit     int
		wantLimit bool
	}{
		{"success", "printf answer; printf diagnostic >&2", 64, false},
		{"stdout boundary", "printf 12345678", 8, false},
		{"stdout overflow", "printf 123456789; sleep 30", 8, true},
		{"stderr overflow", "printf 123456789 >&2; sleep 30", 8, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()

			_, err := (Executor{StdoutLimit: test.limit, StderrLimit: test.limit}).RunAgentCommand(ctx,
				harness.AgentCommand{Program: "sh", Args: []string{"-c", test.script}}, "", time.Now())
			if errors.Is(err, ErrLimit) != test.wantLimit || (!test.wantLimit && err != nil) {
				t.Fatalf("result = %v, want limit failure %v", err, test.wantLimit)
			}
		})
	}
}

// Cancellation and normal parent exit must stop descendants before returning.
func TestAgentRunProcessCancellation(t *testing.T) {
	for _, mode := range []string{"cancel", "exit"} {
		t.Run(mode, func(t *testing.T) { checkProcessDescendantStop(t, mode) })
	}
}

func checkProcessDescendantStop(t *testing.T, mode string) {
	t.Helper()

	dir := t.TempDir()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)
	script := "sleep 30 & child=$!; echo $child > child; wait"

	var expected error
	if mode == "cancel" {
		expected = context.Canceled
	} else {
		script = "sleep 30 >/dev/null 2>&1 & child=$!; echo $child > child"
	}

	go func() {
		_, err := (Executor{}).RunAgentCommand(ctx, harness.AgentCommand{
			Program: "sh", Dir: dir, Args: []string{"-c", script},
		}, "", time.Now())
		done <- err
	}()

	pid := waitChildPID(t, dir)

	if mode == "cancel" {
		cancel()
	}

	select {
	case err := <-done:
		if !errors.Is(err, expected) {
			t.Fatalf("cancellation = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("owned process did not exit")
	}

	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err == nil && !strings.Contains(string(stat), ") Z ") {
		t.Fatalf("descendant %d is still running", pid)
	}
}

func waitChildPID(t *testing.T, dir string) int {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(dir, "child"))
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err == nil {
				return pid
			}
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("child did not start")

	return 0
}

// Reject large input before spawn, and never accept a partial output artifact.
func TestAgentRunProcessBounds(t *testing.T) {
	_, err := (Executor{}).RunAgentCommand(t.Context(), harness.AgentCommand{Program: "missing-program"}, strings.Repeat("x", harness.MaxInput+1), time.Now())
	if !errors.Is(err, ErrLimit) {
		t.Fatalf("input error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "output")
	if err := os.WriteFile(path, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := ReadOutput(path, 8)
	if !errors.Is(err, ErrLimit) || data != nil {
		t.Fatalf("oversized output = %q, %v", data, err)
	}
}

// Sanitize before retaining excerpts, including a credential split by pipe reads.
func TestAgentRunProcessDiagnostics(t *testing.T) {
	t.Setenv("EXAMPLE_API_KEY", "example-private-value")

	data := Diagnostic([]byte("example-private-value Authorization: Bearer example-bearer token=example-token"))
	for _, secret := range []string{"example-private-value", "example-bearer", "example-token"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("diagnostic contains credential %q", secret)
		}
	}

	for _, text := range []string{"PaSsWoRd=fixture-value", "Paſſword=fixture-value", "Token=fixture-value", "API_KEY=fixture-value", "Bearer fixture-value"} {
		if strings.Contains(string(Diagnostic([]byte(text))), "fixture-value") {
			t.Fatal("credential prefilter changed case-folded redaction")
		}
	}

	if got := len(Diagnostic([]byte(strings.Repeat("x", harness.MaxDiagnostic+1)))); got != harness.MaxDiagnostic {
		t.Fatalf("excerpt length = %d", got)
	}
}
