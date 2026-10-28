package workflow

import (
	"context"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Bound IO keeps configuration, canonical observation and native guards with the host.
type workflowClosingInputSource struct {
	store *Store
	ports FrontierPorts
}

func (source workflowClosingInputSource) ValidationTimeout(ctx context.Context, id string, contract int) (int, error) {
	return source.store.workflowClosingValidationTimeout(ctx, id, contract)
}

func (source workflowClosingInputSource) Target(ctx context.Context, request workflowcloseout.Request) (workflowstate.PreviewTarget, error) {
	return source.store.observeWorkflowClosingTarget(ctx, request, source.ports)
}

func (source workflowClosingInputSource) Tree(ctx context.Context, commit string) (string, error) {
	reader := workflowProofReader{store: source.store}

	return reader.gitOutput(ctx, "rev-parse", "--verify", commit+"^{tree}")
}

func (source workflowClosingInputSource) Snapshot(ctx context.Context) (workflowstate.Snapshot, error) {
	return source.store.readSnapshot(ctx)
}

func (source workflowClosingInputSource) Slice(ctx context.Context, input workflowstate.ClosingInput, descriptor workflowstate.WorkflowSlice) (workflowstate.ClosingSlice, error) {
	return source.store.workflowClosingSliceProof(ctx, input, descriptor, source.ports)
}

func (source workflowClosingInputSource) Check(hashes map[string]string) error {
	return source.store.checkWorkflowPrerequisiteEvidence(hashes)
}
