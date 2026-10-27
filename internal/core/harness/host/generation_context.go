package host

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

type generationContextKey struct{}

// WithGeneration threads one frozen operation through nested model calls.
// The operation is host-owned; no command or identity is accepted from model output.
func WithGeneration(ctx context.Context, operation GenerationOperation) context.Context {
	return context.WithValue(ctx, generationContextKey{}, operation)
}

// GenerationSelection returns a value copy without the private executable.
func GenerationSelection(ctx context.Context) (harnesscore.GenerationSelection, bool) {
	operation, ok := ctx.Value(generationContextKey{}).(GenerationOperation)

	return harnesscore.GenerationSelection{ID: operation.ID, Binding: operation.Binding}, ok
}

// GenerationLocationFromContext returns retained host routing for a call.
// Adapters must recheck access to ProjectRef before using an interrupted request.
func GenerationLocationFromContext(ctx context.Context) (GenerationLocation, bool) {
	operation, ok := ctx.Value(generationContextKey{}).(GenerationOperation)

	return GenerationLocation{ConfigDir: operation.configDir, ExecutionDir: operation.executionDir, ProjectRef: operation.ProjectRef}, ok
}

// GenerationRequestID namespaces a caller's logical call within its
// owning operation. Retries are stable; independent requests cannot overwrite it.
func GenerationRequestID(ctx context.Context, id string) string {
	operation, ok := ctx.Value(generationContextKey{}).(GenerationOperation)
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
	operation, ok := ctx.Value(generationContextKey{}).(GenerationOperation)
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

func checkGenerationRequestOwner(operation GenerationOperation, request harness.AgentRunRequest) error {
	project, err := filepath.Abs(request.ProjectDir)
	if err != nil || project != operation.executionDir || !filepath.IsLocal(request.ID) || filepath.Base(request.ID) != request.ID || request.ID == "operations" {
		return errors.New("generation call must belong to its owning project and a local request artifact")
	}

	return nil
}

func applyGenerationRequestSelection(operation GenerationOperation, request harness.AgentRunRequest) (harness.AgentRunRequest, error) {
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
