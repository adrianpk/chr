package workflow

import (
	"context"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ObserveWorkflowNativePublication binds configured provider IO and method proof
// checking to read-only publication observation. Receipt authority remains here.
func (adapter RemoteReviewAdapter) ObserveWorkflowNativePublication(ctx context.Context, binding workflowstate.NativePublicationBinding, git IntegrationGit) (workflowstate.NativePublicationObservation, error) {
	return workflowforge.ObserveNativePublication(ctx, binding, adapter, git, workflowNativePublicationProofs{})
}

type workflowNativePublicationProofs struct{}

func (workflowNativePublicationProofs) CheckNativeProof(query workflowstate.IntegrationQuery, proof workflowstate.IntegrationProof) error {
	return workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{})
}
