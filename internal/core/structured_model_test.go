package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
	"forge.adrianpk.com/chr/internal/harness"
)

// Domain validation stays above the transport, including for injected runners.
func TestStructuredHarnessOutput(t *testing.T) {
	for _, test := range []struct {
		name    string
		output  string
		invalid bool
	}{
		{"valid", `{"answer":"ready"}`, false},
		{"malformed", `{"answer":`, true},
		{"oversized", `"` + strings.Repeat("x", harness.MaxStructuredOutput) + `"`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider := NewConfiguredStructuredModel(false)
			provider.Runner = &fakeAgentRunner{Result: harness.AgentRunResult{Output: []byte(test.output)}}

			output, err := provider.InvokeStructuredModel(t.Context(), structuredHarnessRequest(t))
			if errors.Is(err, modelcore.ErrOutput) != test.invalid || (!test.invalid && err != nil) {
				t.Fatalf("output = %q, error = %v", output, err)
			}

			if test.invalid && output != nil {
				t.Fatal("invalid partial output was accepted")
			}
		})
	}
}

// Missing injection must fail rather than silently selecting a concrete CLI.
func TestStructuredHarnessRequiresRunner(t *testing.T) {
	provider := NewConfiguredStructuredModel(false)
	if _, err := provider.InvokeStructuredModel(t.Context(), structuredHarnessRequest(t)); err == nil {
		t.Fatal("missing runner accepted")
	}
}

// A substitute runner inspects the real dispatch boundary without waiting for
// a live provider. Longer authoring calls retain caller cancellation and the
// default deadline for unrelated structured requests remains unchanged.
func TestStructuredHarnessDeadline(t *testing.T) {
	for _, test := range []struct {
		name                                  string
		timeout, parent, expected, dispatched time.Duration
	}{
		{name: "default", expected: time.Minute, dispatched: time.Minute},
		{name: "authoring", timeout: 5 * time.Minute, expected: 5 * time.Minute, dispatched: 5 * time.Minute},
		{name: "earlier parent", timeout: 5 * time.Minute, parent: 10 * time.Second, expected: 10 * time.Second, dispatched: 5 * time.Minute},
		{name: "shorter request", timeout: 10 * time.Second, expected: 10 * time.Second, dispatched: 10 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()

			if test.parent > 0 {
				var cancel context.CancelFunc

				ctx, cancel = context.WithTimeout(ctx, test.parent)
				defer cancel()
			}

			runner := &structuredDeadlineRunner{}
			runner.Result.Output = []byte(`{"answer":"ready"}`)
			provider := NewConfiguredStructuredModel(false)
			provider.Runner = runner
			request := structuredHarnessRequest(t)
			request.Timeout = test.timeout

			if _, err := provider.InvokeStructuredModel(ctx, request); err != nil {
				t.Fatal(err)
			}

			remaining := time.Until(runner.deadline)
			if remaining > test.expected || remaining < test.expected-time.Second {
				t.Fatalf("runner deadline remaining = %s, want %s", remaining, test.expected)
			}

			if runner.Request.Timeout != test.dispatched {
				t.Fatalf("runner timeout = %s, want the selected request timeout", runner.Request.Timeout)
			}
		})
	}
}

type structuredDeadlineRunner struct {
	fakeAgentRunner
	deadline time.Time
}

func (r *structuredDeadlineRunner) RunAgent(ctx context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	r.deadline, _ = ctx.Deadline()

	return r.fakeAgentRunner.RunAgent(ctx, request)
}

func structuredHarnessRequest(t *testing.T) StructuredModelRequest {
	t.Helper()

	return StructuredModelRequest{
		ID: "structured-test", Purpose: "conversation", ProjectDir: t.TempDir(),
		Instructions: "Return an answer.", Input: json.RawMessage(`{"message":"ready"}`),
		Schema: map[string]any{"type": "object"},
	}
}
