package workflow

import (
	"context"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func checkWorkflowMergerRemote(ctx context.Context, provider PullRequestProvider, input workflowstate.MergerInput) error {
	if provider == nil {
		return workflowstate.ErrConflict
	}

	for _, observed := range []struct{ ref, commit string }{{input.SourceRef, input.Source}, {input.TargetRef, input.Base}} {
		commit, err := provider.ReadWorkflowBranch(ctx, input.Repository, strings.TrimPrefix(observed.ref, "refs/heads/"))
		if err != nil {
			return err
		}

		if commit != observed.commit {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func (s *Store) checkWorkflowMergerPrepared(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider) error {
	if err := checkWorkflowMergerRemote(ctx, provider, *prepared.input.Merger); err != nil {
		return err
	}

	if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	return s.withLock(ctx, func() error {
		if err := s.checkWorkflowMergerOwned(ctx, claim, prepared); err != nil {
			return err
		}

		claimObservation, err := s.checkWorkflowClaim(ctx, claim)
		if err != nil {
			return err
		}

		entry, err := claim.store.readWorkflowIntegrationEntry(claimObservation.Run, *claimObservation.Run.Integration)
		if err != nil {
			return err
		}

		return claim.store.checkWorkflowCandidateRepairAdmission(ctx, claimObservation.Run, entry, *prepared.input.Merger, provider)
	})
}

func (s *Store) checkWorkflowMergerOwned(ctx context.Context, claim *Claim, prepared workflowPreparedTask) error {
	claimObservation2, err := s.checkWorkflowClaim(ctx, claim)
	if err != nil {
		return err
	}

	if err := checkWorkflowExecutionBinding(claimObservation2.Run, claimObservation2.Record, prepared); err != nil {
		return err
	}

	if claimObservation2.Run.StopIntent != "" {
		return ErrWorkflowStop
	}

	if claimObservation2.Run.Integration == nil {
		return workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(claimObservation2.Run, *claimObservation2.Run.Integration)
	if err != nil {
		return err
	}

	if entry.StopIntent != "" || entry.State != "merging" {
		return ErrWorkflowStop
	}

	if err := checkWorkflowMergerEntry(claimObservation2.Run, entry, *prepared.input.Merger); err != nil {
		return err
	}

	if _, err := s.workflowMergerSeed(ctx, claimObservation2.Run, entry, *prepared.input.Merger); err != nil {
		return err
	}

	return s.checkWorkflowMergerCommands(ctx, claimObservation2.Run, *prepared.input.Merger)
}
