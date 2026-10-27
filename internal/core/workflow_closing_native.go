package core

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowClosingNativeDeliveryProof(ctx context.Context, receipt WorkflowRecord) (WorkflowSliceDeliveryProof, string, error) {
	native := receipt.Receipt.Native

	run, err := s.readWorkflowInvocation(native.InvocationID)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, "", err
	}

	entry, err := s.workflowClosingNativeEntry(run, receipt)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, "", err
	}

	if err := s.checkWorkflowClosingNativeSubmission(ctx, run, entry, receipt); err != nil {
		return WorkflowSliceDeliveryProof{}, "", err
	}

	view, err := workflowClosingNativeView(run, receipt)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, "", err
	}

	proof, err := s.readWorkflowDeliveryProof(view)
	if err != nil || proof.NativeReceiptRef == nil || !workflowReceiptMatchesProof(receipt, proof, receipt.Receipt.VerificationRef) {
		return proof, "", errorsOrConflict(err)
	}

	refs := native.ObservationRefs
	if len(refs) == 0 || refs[len(refs)-1].Path != receipt.Receipt.VerificationRef {
		return proof, "", workflowstate.ErrConflict
	}

	return proof, refs[len(refs)-1].Hash, nil
}

func (s *WorkflowStore) workflowClosingNativeEntry(run workflowInvocationFile, receipt WorkflowRecord) (WorkflowIntegrationEntry, error) {
	for _, binding := range run.IntegrationEntries {
		if binding.EntryID != receipt.Receipt.Native.EntryID {
			continue
		}

		entry, err := s.readWorkflowIntegrationEntry(run, binding)
		if err != nil {
			return entry, err
		}

		ref := workflowstate.WorkflowEvidenceRef{Path: receipt.Path, Hash: receipt.SourceHash}
		if run.OwnerRoot != s.layout.Root() || entry.State != "integrated" || entry.Blocker != nil || entry.ReceiptRef == nil || *entry.ReceiptRef != ref {
			return entry, workflowstate.ErrConflict
		}

		return entry, nil
	}

	return WorkflowIntegrationEntry{}, workflowstate.ErrConflict
}

func workflowClosingNativeView(run workflowInvocationFile, receipt WorkflowRecord) (workflowInvocationFile, error) {
	if run.DeliveryRef == receipt.Path && run.DeliveryHash == receipt.SourceHash && run.Cursor == run.ConfirmedCursor {
		return run, nil
	}

	for _, delivery := range run.Deliveries {
		if delivery.DeliveryRef == receipt.Path && delivery.DeliveryHash == receipt.SourceHash && delivery.Cursor <= run.ConfirmedCursor {
			return workflowDeliveredUnitView(run, delivery), nil
		}
	}

	return run, workflowstate.ErrConflict
}

func (s *WorkflowStore) checkWorkflowClosingNativeSubmission(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, receipt WorkflowRecord) error {
	binding := entry.SubmissionRef
	if binding == nil || binding.State != "confirmed" {
		return ErrWorkflowObservationRequired
	}

	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), binding.JournalKey)
	if err != nil {
		return err
	}

	journal, err := readWorkflowSubmissionFile(path)
	if err != nil {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowClosingNativeJournalMatches(journal, run, *binding, receipt) {
		return workflowstate.ErrConflict
	}

	if last.ID != receipt.Receipt.SubmissionID || last.State != "confirmed" || !workflowNativeSubmissionEntryMatches(last, entry) {
		return ErrWorkflowObservationRequired
	}

	return s.checkWorkflowNativeRetainedReceipt(run, last, receipt)
}

func workflowClosingNativeJournalMatches(journal workflowSubmissionJournal, run workflowInvocationFile, binding WorkflowSubmissionBinding, receipt WorkflowRecord) bool {
	expected := [4]string{binding.JournalHash, run.OwnerRoot, run.Namespace, run.ID}
	actual := [4]string{journal.sourceHash, journal.OwnerRoot, journal.Namespace, journal.InvocationID}

	return actual == expected && journal.Unit == receipt.Receipt.Unit
}

func (s *WorkflowStore) observeWorkflowClosingNativeSlice(ctx context.Context, input WorkflowClosingInput, receipt WorkflowRecord, ports WorkflowFrontierPorts) (WorkflowIntegrationProof, error) {
	observation, err := s.observeWorkflowNativeReceipt(ctx, receipt, ports.Provider, ports.Observer)
	if err != nil {
		return WorkflowIntegrationProof{}, err
	}

	if observation.State != "merged" || observation.Git == nil {
		return WorkflowIntegrationProof{}, ErrWorkflowObservationRequired
	}

	proof := *observation.Git
	if proof.Repository != input.Target.Repository || proof.TargetBranch != input.Branch || !workflowFrontierReceiptProofMatches(proof, receipt, input.Target) {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func (s *WorkflowStore) checkWorkflowClosingNativeCanonical(ctx context.Context, input WorkflowClosingInput, ports WorkflowFrontierPorts) error {
	for _, slice := range input.Slices {
		receipt, err := s.readRecord(slice.Receipt.Ref)
		if err != nil {
			return err
		}

		if receipt.Receipt == nil || receipt.SourceHash != slice.Receipt.Hash {
			return workflowstate.ErrConflict
		}

		if receipt.Receipt.Native == nil {
			continue
		}

		if _, err := s.observeWorkflowClosingNativeSlice(ctx, input, receipt, ports); err != nil {
			return err
		}
	}

	return nil
}

func workflowClosingNativeEvidence(receipt WorkflowRecord, proof WorkflowSliceDeliveryProof) []string {
	native := receipt.Receipt.Native
	if native == nil {
		return nil
	}

	refs := []string{proof.CandidateRef, native.CandidateRef.Path, native.TransactionProofRef.Path}
	if native.GitApplicationRef != nil {
		refs = append(refs, native.GitApplicationRef.Path)
	}

	for _, ref := range append(native.evidenceRefs(), native.ObservationRefs...) {
		refs = append(refs, ref.Path)
	}

	return refs
}
