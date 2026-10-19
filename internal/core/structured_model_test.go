package core

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

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
			if errors.Is(err, ErrStructuredModelOutput) != test.invalid || (!test.invalid && err != nil) {
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

func structuredHarnessRequest(t *testing.T) StructuredModelRequest {
	t.Helper()

	return StructuredModelRequest{
		ID: "structured-test", Purpose: "conversation", ProjectDir: t.TempDir(),
		Instructions: "Return an answer.", Input: json.RawMessage(`{"message":"ready"}`),
		Schema: map[string]any{"type": "object"},
	}
}
