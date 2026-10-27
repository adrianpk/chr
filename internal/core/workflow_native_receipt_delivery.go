package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ConfirmWorkflowNativeReceipt seals delivery from retained host evidence for
// the exact submission and latest original-PR observation. It performs no remote
// mutation or task fanout; retries preserve the first immutable receipt bytes.
func (s *WorkflowStore) ConfirmWorkflowNativeReceipt(ctx context.Context, guard WorkflowIntegrationBudgetGuard) (receipt WorkflowRecord, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		receipt, readErr = owner.confirmWorkflowNativeReceipt(ctx, run, guard)

		return readErr
	})

	return receipt, err
}

func (s *WorkflowStore) confirmWorkflowNativeReceipt(ctx context.Context, run workflowInvocationFile, guard WorkflowIntegrationBudgetGuard) (WorkflowRecord, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, guard)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, guard); err != nil {
		return WorkflowRecord{}, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil || len(journal.Attempts) == 0 {
		return WorkflowRecord{}, errorsOrConflict(err)
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil || !workflowNativeSubmissionEntryMatches(last, entry) {
		return WorkflowRecord{}, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, last); err != nil {
		return WorkflowRecord{}, err
	}

	receipt, err := s.retainWorkflowNativeReceipt(ctx, run, last)
	if err != nil {
		return receipt, fmt.Errorf("retain native receipt: %w", err)
	}

	facts := *last.Native
	facts.ReceiptRef = &workflowstate.EvidenceRef{Path: receipt.Path, Hash: receipt.SourceHash}
	update := WorkflowNativeSubmissionUpdate{WorkflowIntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: facts}

	_, err = s.retainWorkflowNativeSubmission(ctx, run, update)
	if err != nil {
		return receipt, fmt.Errorf("confirm native receipt submission: %w", err)
	}

	return receipt, nil
}

func (s *WorkflowStore) retainWorkflowNativeReceipt(ctx context.Context, run workflowInvocationFile, submission WorkflowSubmission) (WorkflowRecord, error) {
	observation, ref, err := s.workflowNativeReceiptObservation(submission)
	if err != nil {
		return WorkflowRecord{}, err
	}

	record, err := s.readRecord(workflowNativeReceiptPath(run, submission))
	if err == nil {
		return record, s.checkWorkflowNativeRetainedReceipt(run, submission, record)
	}

	if !errors.Is(err, fs.ErrNotExist) || submission.Native.ReceiptRef != nil {
		return record, errorsOrConflict(err)
	}

	record, err = s.workflowNativeReceipt(run, submission, observation, ref)
	if err != nil {
		return record, err
	}

	return s.writeRecord(ctx, record, 0)
}

func (s *WorkflowStore) workflowNativeReceiptObservation(submission WorkflowSubmission) (WorkflowNativePublicationObservation, workflowstate.EvidenceRef, error) {
	var observation WorkflowNativePublicationObservation

	native := submission.Native
	if native.ApplicationState != "git_applied" || native.TransportResponseRef == nil || len(native.ObservationRefs) == 0 {
		return observation, workflowstate.EvidenceRef{}, ErrWorkflowObservationRequired
	}

	ref := native.ObservationRefs[len(native.ObservationRefs)-1]
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
		return observation, ref, err
	}

	if observation.State != "merged" || !s.workflowNativeCanonicalObservationValid(submission, observation) {
		return observation, ref, ErrWorkflowObservationRequired
	}

	return observation, ref, nil
}

func workflowNativeReceiptID(run workflowInvocationFile, submission WorkflowSubmission) string {
	return "delivery-" + contenthash.Sum([]byte(run.ID+"\x00"+submission.CandidateID+"\x00"+submission.ID))
}

func workflowNativeReceiptPath(run workflowInvocationFile, submission WorkflowSubmission) string {
	return filepath.ToSlash(filepath.Join("run", run.ID, "evidence", workflowNativeReceiptID(run, submission)+".json"))
}

func (s *WorkflowStore) workflowNativeReceipt(run workflowInvocationFile, submission WorkflowSubmission, observation WorkflowNativePublicationObservation, ref workflowstate.EvidenceRef) (WorkflowRecord, error) {
	reference, err := DetectReviewProviderReference(submission.Native.CanonicalPR)
	if err != nil {
		return WorkflowRecord{}, err
	}

	native, git := submission.Native, *observation.Git
	facts := &WorkflowNativeReceipt{WorkflowNativeRecordBinding: native.WorkflowNativeRecordBinding, CandidateHash: native.CandidateHash, Mechanism: native.Mechanism, CandidateRef: native.CandidateRef, TransactionProofRef: *native.TransportResponseRef, ObservationRefs: []workflowstate.EvidenceRef{ref}, ActualMergeResult: observation.PullRequest.ResultCommit}
	facts.ObservationContract = observation.ObservationContract

	facts.GitApplicationRef, err = s.workflowNativeReceiptApplication(submission)
	if err != nil {
		return WorkflowRecord{}, err
	}

	now := time.Now().UTC()

	receipt := &WorkflowReceipt{Schema: workflowNativeReceiptSchema, Native: facts, Unit: submission.Unit, CandidateID: submission.CandidateID, SubmissionID: submission.ID, Provider: string(reference.Provider), Repository: git.Repository, PullRequest: native.CanonicalPR, Method: git.Method, SourceCommit: native.OriginalSource, BaseCommit: native.ValidatedBase, PreparedTree: git.PreparedTree, HistoryProof: git.HistoryHash, ResultCommit: git.ResultCommit, VerifiedTree: git.VerifiedTree, ObservedTarget: git.ObservedTarget, VerificationRef: ref.Path, ValidationRefs: []string{native.CandidateRef.Path}, ConfirmedAt: now}
	for _, validation := range native.ValidationRefs {
		receipt.ValidationRefs = append(receipt.ValidationRefs, validation.Path)
	}

	identity := WorkflowIdentity{ID: workflowNativeReceiptID(run, submission), ProjectID: run.ProjectID, Namespace: run.Namespace, Revision: 1, CreatedAt: now, UpdatedAt: now}
	record := WorkflowRecord{WorkflowIdentity: identity, Kind: "receipt", Path: workflowNativeReceiptPath(run, submission), Receipt: receipt}

	return record, record.Validate()
}

func (s *WorkflowStore) checkWorkflowNativeRetainedReceipt(run workflowInvocationFile, submission WorkflowSubmission, record WorkflowRecord) error {
	if record.ID != workflowNativeReceiptID(run, submission) || record.ProjectID != run.ProjectID || record.Namespace != run.Namespace || !workflowNativeSubmissionReceiptMatches(submission, record) {
		return workflowstate.ErrConflict
	}

	if submission.Native.ReceiptRef != nil && *submission.Native.ReceiptRef != (workflowstate.EvidenceRef{Path: record.Path, Hash: record.SourceHash}) {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowNativeReceipt(record)
}
