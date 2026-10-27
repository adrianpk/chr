package core

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness"
)

type harnessGenerationContextKey struct{}

// WithHarnessGeneration threads one frozen operation through nested model calls.
// The operation is host-owned; no command or identity is accepted from model output.
func WithHarnessGeneration(ctx context.Context, operation HarnessGenerationOperation) context.Context {
	return context.WithValue(ctx, harnessGenerationContextKey{}, operation)
}

// HarnessGenerationSelection returns a value copy without the private executable.
func HarnessGenerationSelection(ctx context.Context) (harnesscore.GenerationSelection, bool) {
	operation, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)

	return harnesscore.GenerationSelection{ID: operation.ID, Binding: operation.Binding}, ok
}

// HarnessGenerationLocationFromContext returns retained host routing for a call.
// Adapters must recheck access to ProjectRef before using an interrupted request.
func HarnessGenerationLocationFromContext(ctx context.Context) (HarnessGenerationLocation, bool) {
	operation, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)

	return HarnessGenerationLocation{ConfigDir: operation.configDir, ExecutionDir: operation.executionDir, ProjectRef: operation.ProjectRef}, ok
}

// HarnessGenerationRequestID namespaces a caller's logical call within its
// owning operation. Retries are stable; independent requests cannot overwrite it.
func HarnessGenerationRequestID(ctx context.Context, id string) string {
	operation, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation)
	if !ok {
		return id
	}

	prefix := "generation-" + contenthash.Sum([]byte(operation.ID))[:16] + "-"
	if strings.HasPrefix(id, prefix) && len(id) == len(prefix)+64 {
		return id
	}

	return prefix + contenthash.Sum([]byte(id))
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
	if err != nil || project != operation.executionDir || !filepath.IsLocal(request.ID) || filepath.Base(request.ID) != request.ID || request.ID == "operations" {
		return errors.New("generation call must belong to its owning project and a local request artifact")
	}

	return nil
}

func applyGenerationRequestSelection(operation HarnessGenerationOperation, request harness.AgentRunRequest) (harness.AgentRunRequest, error) {
	model := operation.Binding.Selection.Model
	if model == "cli_default" {
		model = ""
	}

	selection := operation.Binding.Selection
	for _, field := range []struct{ actual, expected string }{{request.Program, operation.program}, {request.Model, model}, {request.ModelProvider, selection.ModelProvider}, {request.Thinking, selection.Thinking}} {
		if field.actual != "" && field.actual != field.expected {
			return request, workflowstate.ErrConflict
		}
	}

	if request.Resources != (harness.GenerationResources{}) && request.Resources != operation.resources {
		return request, workflowstate.ErrConflict
	}

	request.Program, request.Model, request.FrozenSelection = operation.program, model, true
	request.ModelProvider, request.Thinking, request.Resources = selection.ModelProvider, selection.Thinking, operation.resources

	return request, nil
}
