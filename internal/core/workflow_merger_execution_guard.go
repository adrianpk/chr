package core

import (
	"context"
	"strings"
)

func checkWorkflowMergerRemote(ctx context.Context, provider WorkflowPullRequestProvider, input WorkflowMergerInput) error {
	if provider == nil {
		return ErrWorkflowConflict
	}

	for _, observed := range []struct{ ref, commit string }{{input.SourceRef, input.Source}, {input.TargetRef, input.Base}} {
		commit, err := provider.ReadWorkflowBranch(ctx, input.Repository, strings.TrimPrefix(observed.ref, "refs/heads/"))
		if err != nil {
			return err
		}

		if commit != observed.commit {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowMergerPrepared(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider) error {
	if err := checkWorkflowMergerRemote(ctx, provider, *prepared.input.Merger); err != nil {
		return err
	}

	return s.withLock(ctx, func() error { return s.checkWorkflowMergerOwned(ctx, claim, prepared) })
}

func (s *WorkflowStore) checkWorkflowMergerOwned(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask) error {
	run, record, err := s.checkWorkflowClaim(ctx, claim)
	if err != nil {
		return err
	}

	if run.StopIntent != "" {
		return ErrWorkflowStop
	}

	if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
		return err
	}

	if run.Integration == nil {
		return ErrWorkflowConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if entry.StopIntent != "" || entry.State != "merging" {
		return ErrWorkflowStop
	}

	if err := checkWorkflowMergerEntry(run, entry, *prepared.input.Merger); err != nil {
		return err
	}

	_, err = s.workflowMergerSeed(ctx, run, entry, *prepared.input.Merger)

	return err
}
