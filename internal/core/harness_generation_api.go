package core

import (
	"context"
	"errors"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/harness"
)

// RunAPIGeneration extends the existing private agent-run bundle to direct API
// calls. The original provider owns its network request, limits and accounting.
func RunAPIGeneration(ctx context.Context, request harness.AgentRunRequest, invoke func(context.Context) ([]byte, error)) ([]byte, error) {
	operation, bound := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)
	if !bound {
		return invoke(ctx)
	}

	if err := validateGenerationOperation(operation); err != nil {
		return nil, err
	}

	if operation.Binding.Selection.Backend != "openai" {
		return nil, errors.New("harness operation cannot dispatch through a direct API")
	}

	if err := checkGenerationRequestOwner(operation, request); err != nil {
		return nil, err
	}

	request.ID = HarnessGenerationRequestID(ctx, request.ID)
	if request.RunDir == "" {
		request.RunDir = UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	if err := harness.ValidateAgentRunRequest(request); err != nil {
		return nil, err
	}

	metadata, err := startHarnessGenerationRun(ctx, request)
	if err != nil {
		return nil, err
	}

	if err := writeAPIGenerationInputs(request); err != nil {
		return nil, finishHarnessGenerationRun(request, harness.AgentRunResult{}, metadata, err)
	}

	output, err := invoke(ctx)

	return output, finishHarnessGenerationRun(request, harness.AgentRunResult{Output: output}, metadata, err)
}

func writeAPIGenerationInputs(request harness.AgentRunRequest) error {
	if err := atomicBoundedWorkflowFile(filepath.Join(request.RunDir, "prompt.md"), []byte(request.Prompt), harness.MaxInput); err != nil {
		return err
	}

	if len(request.OutputSchema) != 0 {
		return atomicBoundedWorkflowFile(filepath.Join(request.RunDir, "output-schema.json"), request.OutputSchema, harness.MaxStructuredOutput)
	}

	return nil
}

// ValidateStructuredModelOutput records the original caller's typed/schema
// validation, independently of JSON framing and process/network success.
func ValidateStructuredModelOutput(ctx context.Context, request StructuredModelRequest, validationErr error) error {
	return ValidateHarnessGenerationOutput(ctx, harness.AgentRunRequest{ID: request.ID, Purpose: request.Purpose, ProjectDir: request.ProjectDir}, validationErr)
}
