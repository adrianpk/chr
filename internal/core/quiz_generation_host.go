package core

import (
	"context"
	"forge.adrianpk.com/chr/internal/harness"
)

// QuizGenerationHost retains existing admission, retained output and trace ownership.
type QuizGenerationHost struct{}

func (QuizGenerationHost) RequestID(ctx context.Context, id string) string {
	return HarnessGenerationRequestID(ctx, id)
}
func (QuizGenerationHost) Run(ctx context.Context, request harness.AgentRunRequest, invoke func(context.Context) ([]byte, error)) ([]byte, error) {
	return RunAPIGeneration(ctx, request, invoke)
}
func (QuizGenerationHost) ValidateOutput(ctx context.Context, request harness.AgentRunRequest, err error) error {
	return ValidateHarnessGenerationOutput(ctx, request, err)
}
func (QuizGenerationHost) Trace(ctx context.Context, projectDir string, event map[string]any) {
	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: projectDir}, event)
}
