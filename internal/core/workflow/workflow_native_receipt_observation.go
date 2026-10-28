package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) observeWorkflowNativeReceipt(ctx context.Context, receipt Record, provider PullRequestProvider, git IntegrationGit) (NativePublicationObservation, error) {
	observer, ok := provider.(NativePublicationObserver)
	if !ok || git == nil {
		return NativePublicationObservation{}, ErrWorkflowAtomicGuard
	}

	binding, err := s.workflowNativeReceiptBinding(receipt)
	if err != nil {
		return NativePublicationObservation{}, err
	}

	return observeWorkflowNativeBounded(ctx, observer, binding, git)
}

func (s *Store) workflowNativeReceiptBinding(receipt Record) (NativePublicationBinding, error) {
	var binding NativePublicationBinding
	if receipt.Receipt == nil || receipt.Receipt.Native == nil {
		return binding, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeReceipt(receipt); err != nil {
		return binding, err
	}

	native := receipt.Receipt.Native

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return binding, err
	}

	binding.Git, binding.PullRequest = candidate.Input.Git, candidate.Input.Source.PullRequest

	binding.SourceDirectory, binding.ReceiptConfirmed = candidate.Lane.Worktree, true
	if err := s.readWorkflowNativeReceiptPublication(*native, &binding); err != nil {
		return binding, err
	}

	return binding, validateWorkflowNativeBinding(binding)
}

func (s *Store) readWorkflowNativeReceiptPublication(native NativeReceipt, binding *NativePublicationBinding) error {
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &binding.Capability); err != nil {
		return err
	}

	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.TransactionProofRef, &binding.Publication); err != nil {
		return err
	}

	if native.GitApplicationRef != nil {
		binding.GitApplication = &NativeGitObservation{}
		if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.GitApplicationRef, binding.GitApplication); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) retainWorkflowNativeReceiptObservation(ctx context.Context, run workflowInvocationFile, observation NativePublicationObservation) (workflowInvocationFile, error) {
	if observation.CanonicalPR == "" || observation.State == "" {
		return run, ErrWorkflowObservationRequired
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return run, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return run, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.State != "confirmed" || !workflowNativeSubmissionEntryMatches(last, entry) {
		return run, workflowstate.ErrConflict
	}

	guard := IntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, BudgetRevision: run.IntegrationBudget.Revision}

	update := NativeSubmissionUpdate{IntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: *last.Native}
	if err := s.checkWorkflowNativeCanonicalCallback(ctx, run, update); err != nil {
		return run, err
	}

	if _, err := s.recordWorkflowNativeCanonicalObservation(ctx, run, update, last, observation); err != nil {
		return run, err
	}

	return s.readWorkflowInvocation(run.ID)
}
