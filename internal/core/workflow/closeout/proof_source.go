package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Proofs verifies ordered closing evidence against the bound repository scope.
// The host owns all IO, native authority guards and atomic confirmation.
type Proofs struct {
	ProjectID string
	Namespace string
	Directory string
	Source    ProofSource
}

// ProofSource supplies bounded retained images and canonical observations.
// CheckProof delegates integration history policy to its existing owner.
type ProofSource interface {
	Read(string) ([]byte, error)
	Record(string) (workflowstate.Record, error)
	Snapshot(context.Context) (workflowstate.Snapshot, error)
	Contract(workflowstate.WorkflowSlice) error
	NativeDelivery(context.Context, workflowstate.Record) (DeliveryProof, error)
	PullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error)
	Verify(context.Context, workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error)
	CheckProof(workflowstate.IntegrationQuery, workflowstate.IntegrationProof) error
	NativeObservation(context.Context, workflowstate.Record) (workflowstate.NativePublicationObservation, error)
	Hash(string) (string, error)
}

func proofError(err error) error {
	if err != nil {
		return err
	}

	return workflowstate.ErrConflict
}
