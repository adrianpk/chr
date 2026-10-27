package core

import (
	"context"

	harnesstrace "forge.adrianpk.com/chr/internal/core/harness/trace"
	"forge.adrianpk.com/chr/internal/harness"
)

// QuizGenerationHost retains existing admission, retained output and trace ownership.
type QuizGenerationHost struct{}

// RequestID retains the original operation-bound request identity.
func (QuizGenerationHost) RequestID(ctx context.Context, id string) string {
	return HarnessGenerationRequestID(ctx, id)
}

// Run retains guarded API execution and run output ownership.
func (QuizGenerationHost) Run(ctx context.Context, request harness.AgentRunRequest, invoke func(context.Context) ([]byte, error)) ([]byte, error) {
	return RunAPIGeneration(ctx, request, invoke)
}

// ValidateOutput records the original semantic output outcome.
func (QuizGenerationHost) ValidateOutput(ctx context.Context, request harness.AgentRunRequest, err error) error {
	return ValidateHarnessGenerationOutput(ctx, request, err)
}

// Trace uses the existing configured redaction and payload policy.
func (QuizGenerationHost) Trace(ctx context.Context, projectDir string, event map[string]any) {
	harnesstrace.RecordLLMTraceEvent(ctx, harnesstrace.LoadLLMTraceConfigOptions{ProjectDir: projectDir}, event)
}
