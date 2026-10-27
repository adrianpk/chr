package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) retainWorkflowNativeSend(ctx context.Context, submission WorkflowSubmission, request WorkflowGitPublication, advertisement WorkflowGitPublicationAdvertisement) error {
	_, err := s.updateWorkflowNativeDispatch(ctx, submission, func(owner *WorkflowStore, native *WorkflowNativeSubmission) error {
		if err := owner.checkWorkflowNativeSendStop(native.InvocationID); err != nil {
			return err
		}

		var capability WorkflowNativeCapability
		if err := owner.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &capability); err != nil {
			return err
		}

		if advertisement.Fingerprint != capability.ServerCapabilityFingerprint || advertisement.SourceCommit != native.OriginalSource || advertisement.TargetCommit != native.ValidatedBase {
			return fmt.Errorf("%w: actual advertisement changed", workflowstate.ErrConflict)
		}

		manifest := workflowNativeCommandManifest(submission)

		advertisementRef, err := owner.writeWorkflowNativeDispatchEvidence(submission, advertisement)
		if err != nil {
			return err
		}

		native.AdvertisementRef = &advertisementRef

		commandRef, err := owner.writeWorkflowNativeDispatchEvidence(submission, manifest)
		if err != nil {
			return err
		}

		native.CommandManifestRef = &commandRef
		result := WorkflowGitPublicationResult{Schema: "chr-git-publication-result/v1", State: "unknown", CommandsSent: true, OriginalSourceCommit: native.OriginalSource, AdvertisedSourceCommit: native.OriginalSource, AdvertisedTargetCommit: native.ValidatedBase, ResultCommit: native.PreparedResult, CandidateRevision: request.CandidateRevision}

		return owner.retainWorkflowNativeDispatchResponse(submission, native, result)
	})

	return err
}

func workflowNativeCommandManifest(submission WorkflowSubmission) WorkflowNativeCommandManifest {
	native := submission.Native

	return WorkflowNativeCommandManifest{SchemaVersion: 1, SubmissionID: submission.ID, Atomic: true, SourceRef: native.SourceRef, ExpectedSource: native.OriginalSource, NewSource: native.PreparedResult, TargetRef: native.TargetRef, ExpectedTarget: native.ValidatedBase, NewTarget: native.PreparedResult}
}

func (s *WorkflowStore) completeWorkflowNativeDispatch(ctx context.Context, submission WorkflowSubmission, result WorkflowGitPublicationResult) (WorkflowSubmission, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	return s.updateWorkflowNativeDispatch(ctx, submission, func(owner *WorkflowStore, native *WorkflowNativeSubmission) error {
		if !validWorkflowNativeResponse(*native, result) {
			return workflowstate.ErrConflict
		}

		if err := owner.retainWorkflowNativeDispatchResponse(submission, native, result); err != nil {
			return err
		}

		completion := WorkflowNativeDispatchCompletion{SchemaVersion: 1, SubmissionID: submission.ID, CandidateRevision: result.CandidateRevision, Identity: *native.DispatchIdentity, ResponseRef: *native.TransportResponseRef, FinishedAt: time.Now().UTC()}

		ref, err := owner.writeWorkflowNativeDispatchEvidence(submission, completion)
		if err != nil {
			return err
		}

		native.DispatchCompletionRef, native.DispatchState = &ref, "quiescent"

		return nil
	})
}

func (s *WorkflowStore) retainWorkflowNativeDispatchResponse(submission WorkflowSubmission, native *WorkflowNativeSubmission, result WorkflowGitPublicationResult) error {
	ref, err := s.writeWorkflowNativeDispatchEvidence(submission, result)
	if err != nil {
		return err
	}

	if native.TransportResponseRef != nil && *native.TransportResponseRef != ref {
		if len(native.TransportResponseHistory) >= workflowstate.WorkflowPageLimit {
			return workflowstate.ErrConflict
		}

		native.TransportResponseHistory = append(native.TransportResponseHistory, *native.TransportResponseRef)
	}

	native.TransportResponseRef = &ref

	return nil
}

func (s *WorkflowStore) writeWorkflowNativeDispatchEvidence(submission WorkflowSubmission, value any) (WorkflowEvidenceRef, error) {
	data, err := json.Marshal(value)
	if err != nil || len(data) >= workflowstate.WorkflowFileLimit {
		return WorkflowEvidenceRef{}, workflowstate.ErrConflict
	}

	hash := contenthash.Sum(data)
	ref := WorkflowEvidenceRef{Path: "run/" + submission.InvocationID + "/evidence/publication/" + submission.ID + "/" + hash + ".json", Hash: hash}

	path, err := s.safePath(ref.Path)
	if err != nil {
		return ref, err
	}

	return ref, atomicWorkflowFile(path, data)
}

func (s *WorkflowStore) updateWorkflowNativeDispatch(ctx context.Context, expected WorkflowSubmission, change func(*WorkflowStore, *WorkflowNativeSubmission) error) (submission WorkflowSubmission, err error) {
	submission = expected
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, expected.InvocationID)
		if err != nil {
			return err
		}

		journal, entry, err := owner.readWorkflowOwnedDispatch(ctx, run, expected)
		if err != nil {
			return err
		}

		last := journal.Attempts[len(journal.Attempts)-1]

		facts := *last.Native
		if err := change(owner, &facts); err != nil {
			return err
		}

		guard := WorkflowIntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, BudgetRevision: run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}

		submission, err = owner.retainWorkflowNativeSubmission(ctx, run, WorkflowNativeSubmissionUpdate{WorkflowIntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: facts})
		if err != nil {
			return fmt.Errorf("retain dispatch facts: %w", err)
		}

		return err
	})

	return submission, err
}

func (s *WorkflowStore) readWorkflowOwnedDispatch(ctx context.Context, run workflowInvocationFile, expected WorkflowSubmission) (workflowSubmissionJournal, WorkflowIntegrationEntry, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return journal, WorkflowIntegrationEntry{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeDispatchIntentMatches(last, expected) {
		return journal, WorkflowIntegrationEntry{}, fmt.Errorf("%w: dispatch intent changed", workflowstate.ErrConflict)
	}

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil || identity != *last.Native.DispatchIdentity || last.Native.DispatchState != "dispatching" {
		return journal, WorkflowIntegrationEntry{}, fmt.Errorf("%w: dispatch owner changed", workflowstate.ErrConflict)
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err == nil && !workflowNativeSubmissionEntryMatches(last, entry) {
		return journal, entry, fmt.Errorf("%w: dispatch generation or candidate changed", workflowstate.ErrConflict)
	}

	return journal, entry, err
}

func workflowNativeDispatchIntentMatches(last, expected WorkflowSubmission) bool {
	return last.ID == expected.ID && last.Native != nil && sameWorkflowNativeSubmissionIntent(*last.Native, *expected.Native) && last.Native.DispatchIdentity != nil
}

func (s *WorkflowStore) checkWorkflowNativeSendStop(id string) error {
	run, err := s.readWorkflowInvocation(id)
	if err != nil {
		return err
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || entry.StopIntent != "" {
		return ErrWorkflowStop
	}

	return nil
}
