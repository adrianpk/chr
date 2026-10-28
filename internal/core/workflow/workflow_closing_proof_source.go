package workflow

import (
	"context"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// This assembly binds finite observations without exposing mutable Store state.
func workflowClosingProofs(store *Store, ports FrontierPorts) workflowcloseout.Proofs {
	return workflowcloseout.Proofs{ProjectID: store.projectID, Namespace: store.layout.Namespace(), Directory: store.layout.ProjectDir(), Source: workflowClosingProofSource{workflowClosingEvidence{store: store}, ports}}
}

type workflowClosingProofSource struct {
	workflowClosingEvidence
	ports FrontierPorts
}

func (source workflowClosingProofSource) Record(relative string) (workflowstate.Record, error) {
	return source.store.readRecord(relative)
}

func (source workflowClosingProofSource) Snapshot(ctx context.Context) (workflowstate.Snapshot, error) {
	return source.store.readSnapshot(ctx)
}

func (source workflowClosingProofSource) Contract(descriptor workflowstate.WorkflowSlice) error {
	_, err := source.store.readWorkflowSliceContract(descriptor)

	return err
}

func (source workflowClosingProofSource) NativeDelivery(ctx context.Context, receipt workflowstate.Record) (workflowcloseout.DeliveryProof, error) {
	return source.store.workflowClosingNativeDeliveryProof(ctx, receipt)
}

func (source workflowClosingProofSource) PullRequest(ctx context.Context, url string) (workflowstate.PullRequestSnapshot, error) {
	return source.ports.Provider.ReadWorkflowPullRequest(ctx, url)
}

func (source workflowClosingProofSource) Verify(ctx context.Context, query workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error) {
	return source.ports.Observer.VerifyWorkflowIntegration(ctx, query)
}

func (source workflowClosingProofSource) CheckProof(query workflowstate.IntegrationQuery, proof workflowstate.IntegrationProof) error {
	return workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{})
}

func (source workflowClosingProofSource) NativeObservation(ctx context.Context, receipt workflowstate.Record) (workflowstate.NativePublicationObservation, error) {
	return (workflowNativeObservationHost{store: source.store}).receipt(ctx, receipt, source.ports.Provider, source.ports.Observer)
}

func (source workflowClosingProofSource) Hash(relative string) (string, error) {
	return source.store.workflowEvidenceHash(relative)
}
