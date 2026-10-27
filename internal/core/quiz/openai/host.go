package openai

import (
	"context"

	"forge.adrianpk.com/chr/internal/harness"
)

// Host owns generation admission, retained output validation and trace recording.
// Contexts belong to calls and are never stored in the adapter.
type Host interface {
	RequestID(context.Context, string) string
	Run(context.Context, harness.AgentRunRequest, func(context.Context) ([]byte, error)) ([]byte, error)
	ValidateOutput(context.Context, harness.AgentRunRequest, error) error
	Trace(context.Context, string, map[string]any)
}
