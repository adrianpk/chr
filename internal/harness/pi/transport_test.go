package pi

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Real pipes exercise correlation, fragmented UTF-8, continuation and final usage.
func TestPiTransportCompletion(t *testing.T) {
	for _, scenario := range []string{"fragmented", "late acknowledgment", "continuation", "backpressure"} {
		t.Run(scenario, func(t *testing.T) {
			command, request := fixtureCall(t, scenario)

			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()

			response, err := (Transport{}).Run(ctx, command, request)
			if err != nil {
				t.Fatal(err)
			}

			if string(response.Output) != "{\"answer\":\"final\u2028line\u2029end\"}" || response.ObservedModel != "fixture/model" {
				t.Fatalf("normalized result = %q, model = %q", response.Output, response.ObservedModel)
			}

			if response.Usage.Input == nil || response.Usage.Output == nil || *response.Usage.Input != 12 || *response.Usage.Output != 7 {
				t.Fatalf("usage = %#v", response.Usage)
			}
		})
	}
}

// Acceptance, low-level completion and partial messages cannot become output.
func TestPiTransportFailures(t *testing.T) {
	for _, test := range []struct {
		name string
		want error
	}{
		{"acknowledgment only", ErrEarlyExit},
		{"agent end only", ErrEarlyExit},
		{"provider failure", ErrProvider},
		{"aborted", ErrAborted},
		{"handled", ErrProtocol},
		{"malformed", ErrProtocol},
		{"unterminated", ErrProtocol},
		{"invalid UTF-8", ErrProtocol},
		{"wrong correlation", ErrProtocol},
		{"duplicate response", ErrProtocol},
		{"wrong model", ErrCapability},
		{"thinking clamped", ErrCapability},
		{"occupied session", ErrCapability},
		{"command rejected", ErrCapability},
		{"tool event", ErrCapability},
		{"output overflow", process.ErrLimit},
		{"record overflow", process.ErrLimit},
		{"stdout overflow", process.ErrLimit},
		{"stderr overflow", process.ErrLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkFixtureFailure(t, test.name, test.want)
		})
	}
}

func checkFixtureFailure(t *testing.T, scenario string, want error) {
	t.Helper()

	command, request := fixtureCall(t, scenario)
	transport := Transport{RecordLimit: 1024, Executor: process.Executor{StdoutLimit: 4096, StderrLimit: 4096}}

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	response, err := transport.Run(ctx, command, request)
	if !errors.Is(err, want) || len(response.Output) != 0 {
		t.Fatalf("result = %q, error = %v; want %v", response.Output, err, want)
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("protocol/provider failure reported caller cancellation: %v", err)
	}
}

// A later successful message cannot invent usage missing from an earlier call.
func TestPiTransportUsage(t *testing.T) {
	for _, test := range []struct {
		name        string
		inputKnown  bool
		outputKnown bool
	}{
		{"unknown usage", false, false},
		{"partial usage", true, false},
		{"continuation unknown usage", false, false},
		{"negative usage", false, true},
		{"overflow usage", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			command, request := fixtureCall(t, test.name)

			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()

			response, err := (Transport{}).Run(ctx, command, request)
			if err != nil || (response.Usage.Input != nil) != test.inputKnown || (response.Usage.Output != nil) != test.outputKnown {
				t.Fatalf("usage = %#v, error = %v", response.Usage, err)
			}
		})
	}
}

// Cancellation must reap a child holding the protocol pipes as well as Pi itself.
func TestPiTransportCancellation(t *testing.T) {
	command, request := fixtureCall(t, "cancellation")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		_, err := (Transport{}).Run(ctx, command, request)
		done <- err
	}()

	pid := waitFixtureChild(t, command.Dir)

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("owned protocol process did not stop")
	}

	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err == nil && !strings.Contains(string(data), ") Z ") {
		t.Fatalf("owned child %d survived cancellation", pid)
	}
}

// Sanitization covers retained stderr independently of protocol error strings.
func TestPiTransportDiagnostics(t *testing.T) {
	t.Setenv("PI_FIXTURE_TOKEN", "fixture-private-token")

	command, request := fixtureCall(t, "diagnostics")

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	result, err := (Transport{}).Run(ctx, command, request)
	if err != nil || strings.Contains(string(result.Process.Stderr), "fixture-private-token") || !strings.Contains(string(result.Process.Stderr), "[REDACTED]") {
		t.Fatalf("diagnostic = %q, error = %v", result.Process.Stderr, err)
	}
}

// Deadline, failed exit and a stalled orderly shutdown all discard final output.
func TestPiTransportExit(t *testing.T) {
	for _, scenario := range []string{"cancellation", "nonzero exit", "shutdown stalled"} {
		t.Run(scenario, func(t *testing.T) {
			command, request := fixtureCall(t, scenario)

			timeout := 3 * time.Second
			if scenario == "cancellation" {
				timeout = 200 * time.Millisecond
			}

			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			defer cancel()

			result, err := (Transport{}).Run(ctx, command, request)
			if err == nil || len(result.Output) != 0 || errors.Is(err, context.DeadlineExceeded) != (scenario == "cancellation") {
				t.Fatalf("exit result = %q, error = %v", result.Output, err)
			}

			if scenario == "nonzero exit" {
				checkFixtureExit(t, err, 7)
			}
		})
	}
}

func checkFixtureExit(t *testing.T, err error, code int) {
	t.Helper()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != code {
		t.Fatalf("process exit = %v; want %d", err, code)
	}
}

// Reject unusable and oversized requests before launching any process.
func TestPiTransportAdmission(t *testing.T) {
	for _, test := range []struct {
		name    string
		request Request
		want    error
	}{
		{"missing provider", Request{Prompt: "context", Model: "model"}, ErrCapability},
		{"escaped input overflow", Request{Prompt: strings.Repeat("\x00", harness.MaxRecord/4), Provider: "fixture", Model: "model"}, process.ErrLimit},
		{"invalid input UTF-8", Request{Prompt: "invalid\xff", Provider: "fixture", Model: "model"}, ErrProtocol},
		{"model identity overflow", Request{Prompt: "context", Provider: "fixture", Model: strings.Repeat("x", 1025)}, process.ErrLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := (Transport{}).Run(t.Context(), harness.AgentCommand{Program: "/must-not-spawn"}, test.request)
			if !errors.Is(err, test.want) || len(result.Output) != 0 {
				t.Fatalf("admission result = %q, error = %v; want %v", result.Output, err, test.want)
			}
		})
	}
}

func fixtureCall(t *testing.T, scenario string) (harness.AgentCommand, Request) {
	t.Helper()
	// The race-built peer must exit within the production shutdown grace.
	t.Setenv("GORACE", "atexit_sleep_ms=0")

	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.json")

	data, err := json.Marshal(fixtureConfig{Scenario: scenario})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	request := Request{Prompt: "Return the prepared result.", Provider: "fixture", Model: "model", Thinking: "high", OutputLimit: 256}
	if scenario == "backpressure" {
		request.Prompt = strings.Repeat("prepared context ", 8192)
	}

	return harness.AgentCommand{Program: program, Args: []string{"-test.run=^TestPiTransportDriver$", "--", path}, Dir: dir}, request
}

func waitFixtureChild(t *testing.T, dir string) int {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(dir, "child.pid"))
		if err == nil {
			pid, parseErr := strconv.Atoi(string(data))
			if parseErr == nil {
				return pid
			}
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("fixture child was not ready")

	return 0
}
