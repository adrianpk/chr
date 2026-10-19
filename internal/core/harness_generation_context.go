package core

import (
	"context"
	"errors"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/harness"
)

type harnessGenerationContextKey struct{}

// WithHarnessGeneration threads one frozen operation through nested model calls.
// The operation is host-owned; no command or identity is accepted from model output.
func WithHarnessGeneration(ctx context.Context, operation HarnessGenerationOperation) context.Context {
	return context.WithValue(ctx, harnessGenerationContextKey{}, operation)
}

// HarnessGenerationSelection returns a value copy without the private executable.
func HarnessGenerationSelection(ctx context.Context) (string, HarnessGenerationBinding, bool) {
	operation, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)

	return operation.ID, operation.Binding, ok
}

func prepareBoundGenerationRequest(ctx context.Context, request harness.AgentRunRequest) (harness.AgentRunRequest, error) {
	operation, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)
	if !ok {
		return request, nil
	}

	if err := validateGenerationOperation(operation); err != nil {
		return request, err
	}

	if err := checkGenerationRequestOwner(operation, request); err != nil {
		return request, err
	}

	selection := operation.Binding.Selection
	if selection.Backend == "openai" {
		return request, errors.New("direct API operation cannot dispatch through a harness runner")
	}

	request, err := applyGenerationRequestSelection(operation, request)
	if err != nil {
		return request, err
	}

	return request, checkGenerationExecutable(operation)
}

func checkGenerationRequestOwner(operation HarnessGenerationOperation, request harness.AgentRunRequest) error {
	project, err := filepath.Abs(request.ProjectDir)
	if err != nil || project != operation.project || !filepath.IsLocal(request.ID) || filepath.Base(request.ID) != request.ID || request.ID == "operations" {
		return errors.New("generation call must belong to its owning project and a local request artifact")
	}

	return nil
}

func applyGenerationRequestSelection(operation HarnessGenerationOperation, request harness.AgentRunRequest) (harness.AgentRunRequest, error) {
	if request.Program != "" && request.Program != operation.program {
		return request, ErrWorkflowConflict
	}

	model := operation.Binding.Selection.Model
	if model == "cli_default" {
		model = ""
	}

	if request.Model != "" && request.Model != model {
		return request, ErrWorkflowConflict
	}

	request.Program, request.Model, request.FrozenSelection = operation.program, model, true

	return request, nil
}
