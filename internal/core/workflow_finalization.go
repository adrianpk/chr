package core

import (
	"context"
	"fmt"
	"time"
)

// FinalizeWorkflowSlice journals a verified receipt and task/run images, plus
// the set descriptor only when the integration unit is a slice.
// It accepts no worker-supplied receipt and never creates or integrates a PR.
func (s *WorkflowStore) FinalizeWorkflowSlice(ctx context.Context, id string, provider WorkflowPullRequestProvider, git WorkflowIntegrationGit) (receipt WorkflowRecord, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		receipt, readErr = owner.finalizeWorkflowSlice(ctx, run, provider, git)

		return readErr
	})

	return receipt, err
}

func (s *WorkflowStore) finalizeWorkflowSlice(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider, git WorkflowIntegrationGit) (WorkflowRecord, error) {
	if run.DeliveryRef != "" {
		return s.readWorkflowSliceReceipt(ctx, run)
	}

	proof, err := s.readWorkflowDeliveryProof(run)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if provider == nil || git == nil {
		return WorkflowRecord{}, fmt.Errorf("finalization requires canonical provider and Git observation")
	}

	if _, err := s.verifyWorkflowSliceIntegration(ctx, run, proof.Git.Method, provider, git); err != nil {
		return WorkflowRecord{}, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return WorkflowRecord{}, err
	}

	observation, candidate, err := s.workflowDeliveryCandidate(ctx, run)
	if err != nil {
		return WorkflowRecord{}, err
	}

	receipt, err := workflowSliceReceipt(run, proof, candidate)
	if err != nil {
		return receipt, err
	}

	if err := s.commitWorkflowSliceFinalization(ctx, &run, observation, receipt); err != nil {
		return receipt, err
	}

	return s.readWorkflowSliceReceipt(ctx, run)
}
