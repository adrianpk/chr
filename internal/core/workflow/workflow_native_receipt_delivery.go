package workflow

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ConfirmWorkflowNativeReceipt seals delivery from retained host evidence for
// the exact submission and latest original-PR observation. It performs no remote
// mutation or task fanout; retries preserve the first immutable receipt bytes.
func (s *Store) ConfirmWorkflowNativeReceipt(ctx context.Context, guard IntegrationBudgetGuard) (receipt Record, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		receipt, readErr = invocationLocation.Owner.confirmWorkflowNativeReceipt(ctx, invocationLocation.Run, guard)

		return readErr
	})

	return receipt, err
}

func (s *Store) confirmWorkflowNativeReceipt(ctx context.Context, run workflowInvocationFile, guard IntegrationBudgetGuard) (Record, error) {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, guard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return Record{}, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, guard); err != nil {
		return Record{}, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil || len(journal.Attempts) == 0 {
		return Record{}, errorsOrConflict(err)
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil || !workflowNativeSubmissionEntryMatches(last, entry) {
		return Record{}, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, last); err != nil {
		return Record{}, err
	}

	receipt, err := s.retainWorkflowNativeReceipt(ctx, run, last)
	if err != nil {
		return receipt, fmt.Errorf("retain native receipt: %w", err)
	}

	facts := *last.Native
	facts.ReceiptRef = &workflowstate.EvidenceRef{Path: receipt.Path, Hash: receipt.SourceHash}
	update := NativeSubmissionUpdate{IntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: facts}

	_, err = s.retainWorkflowNativeSubmission(ctx, run, update)
	if err != nil {
		return receipt, fmt.Errorf("confirm native receipt submission: %w", err)
	}

	return receipt, nil
}

func (s *Store) retainWorkflowNativeReceipt(ctx context.Context, run workflowInvocationFile, submission workflowstate.Submission) (Record, error) {
	receiptEvidence, err := s.workflowNativeReceiptObservation(submission)
	observation, ref := receiptEvidence.Observation, receiptEvidence.Ref

	if err != nil {
		return Record{}, err
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

func (s *Store) workflowNativeReceiptObservation(submission workflowstate.Submission) (workflowNativeReceiptEvidence, error) {
	var observation workflowstate.NativePublicationObservation

	native := submission.Native
	if native.ApplicationState != "git_applied" || native.TransportResponseRef == nil || len(native.ObservationRefs) == 0 {
		return workflowNativeReceiptEvidence{Observation: observation, Ref: workflowstate.EvidenceRef{}}, ErrWorkflowObservationRequired
	}

	ref := native.ObservationRefs[len(native.ObservationRefs)-1]
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
		return workflowNativeReceiptEvidence{Observation: observation, Ref: ref}, err
	}

	if observation.State != "merged" || !s.workflowNativeCanonicalObservationValid(submission, observation) {
		return workflowNativeReceiptEvidence{Observation: observation, Ref: ref}, ErrWorkflowObservationRequired
	}

	return workflowNativeReceiptEvidence{Observation: observation, Ref: ref}, nil
}

func workflowNativeReceiptID(run workflowInvocationFile, submission workflowstate.Submission) string {
	return "delivery-" + contenthash.Sum([]byte(run.ID+"\x00"+submission.CandidateID+"\x00"+submission.ID))
}

func workflowNativeReceiptPath(run workflowInvocationFile, submission workflowstate.Submission) string {
	return filepath.ToSlash(filepath.Join("run", run.ID, "evidence", workflowNativeReceiptID(run, submission)+".json"))
}

func (s *Store) workflowNativeReceipt(run workflowInvocationFile, submission workflowstate.Submission, observation workflowstate.NativePublicationObservation, ref workflowstate.EvidenceRef) (Record, error) {
	reference, err := workflowforge.ParseReference(submission.Native.CanonicalPR)
	if err != nil {
		return Record{}, err
	}

	native, git := submission.Native, *observation.Git
	facts := &workflowstate.NativeReceipt{NativeRecordBinding: native.NativeRecordBinding, CandidateHash: native.CandidateHash, Mechanism: native.Mechanism, CandidateRef: native.CandidateRef, TransactionProofRef: *native.TransportResponseRef, ObservationRefs: []workflowstate.EvidenceRef{ref}, ActualMergeResult: observation.PullRequest.ResultCommit}
	facts.ObservationContract = observation.ObservationContract

	facts.GitApplicationRef, err = s.workflowNativeReceiptApplication(submission)
	if err != nil {
		return Record{}, err
	}

	now := time.Now().UTC()

	receipt := &Receipt{Schema: workflowNativeReceiptSchema, Native: facts, Unit: submission.Unit, CandidateID: submission.CandidateID, SubmissionID: submission.ID, Provider: string(reference.Provider), Repository: git.Repository, PullRequest: native.CanonicalPR, Method: git.Method, SourceCommit: native.OriginalSource, BaseCommit: native.ValidatedBase, PreparedTree: git.PreparedTree, HistoryProof: git.HistoryHash, ResultCommit: git.ResultCommit, VerifiedTree: git.VerifiedTree, ObservedTarget: git.ObservedTarget, VerificationRef: ref.Path, ValidationRefs: []string{native.CandidateRef.Path}, ConfirmedAt: now}
	for _, validation := range native.ValidationRefs {
		receipt.ValidationRefs = append(receipt.ValidationRefs, validation.Path)
	}

	identity := workflowstate.Identity{ID: workflowNativeReceiptID(run, submission), ProjectID: run.ProjectID, Namespace: run.Namespace, Revision: 1, CreatedAt: now, UpdatedAt: now}
	record := Record{Identity: identity, Kind: "receipt", Path: workflowNativeReceiptPath(run, submission), Receipt: receipt}

	return record, record.Validate()
}

func (s *Store) checkWorkflowNativeRetainedReceipt(run workflowInvocationFile, submission workflowstate.Submission, record Record) error {
	if record.ID != workflowNativeReceiptID(run, submission) || record.ProjectID != run.ProjectID || record.Namespace != run.Namespace || !workflowNativeSubmissionReceiptMatches(submission, record) {
		return workflowstate.ErrConflict
	}

	if submission.Native.ReceiptRef != nil && *submission.Native.ReceiptRef != (workflowstate.EvidenceRef{Path: record.Path, Hash: record.SourceHash}) {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowNativeReceipt(record)
}
