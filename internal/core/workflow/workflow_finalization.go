package workflow

import (
	"context"
	"fmt"
	"time"
)

// FinalizeWorkflowSlice journals a verified receipt and task/run images, plus
// the set descriptor only when the integration unit is a slice.
// It accepts no worker-supplied receipt and never creates or integrates a PR.
func (s *Store) FinalizeWorkflowSlice(ctx context.Context, id string, provider PullRequestProvider, git IntegrationGit) (receipt Record, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		receipt, readErr = invocationLocation.Owner.finalizeWorkflowSlice(ctx, invocationLocation.Run, provider, git)

		return readErr
	})

	return receipt, err
}

func (s *Store) finalizeWorkflowSlice(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider, git IntegrationGit) (Record, error) {
	if run.Integration != nil {
		nativeFinalization, err := s.finalizeWorkflowNativeUnit(ctx, run)

		handled, receipt := nativeFinalization.Handled, nativeFinalization.Receipt
		if handled {
			return receipt, err
		}
	}

	return s.finalizeWorkflowLegacySlice(ctx, run, provider, git)
}

func (s *Store) finalizeWorkflowLegacySlice(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider, git IntegrationGit) (Record, error) {
	if run.DeliveryRef != "" {
		return s.readWorkflowSliceReceipt(ctx, run)
	}

	proof, err := s.readWorkflowDeliveryProof(run)
	if err != nil {
		return Record{}, err
	}

	if provider == nil || git == nil {
		return Record{}, fmt.Errorf("finalization requires canonical provider and Git observation")
	}

	if _, err := s.verifyWorkflowSliceIntegration(ctx, run, proof.Git.Method, provider, git); err != nil {
		return Record{}, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return Record{}, err
	}

	deliveryPreparation, err := s.workflowDeliveryCandidate(ctx, run)

	observation, candidate := deliveryPreparation.Observation, deliveryPreparation.Candidate

	if err != nil {
		return Record{}, err
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
