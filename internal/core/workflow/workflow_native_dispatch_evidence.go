package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowNativeSend(ctx context.Context, submission Submission, request GitPublication, advertisement GitPublicationAdvertisement) error {
	_, err := s.updateWorkflowNativeDispatch(ctx, submission, func(owner *Store, native *NativeSubmission) error {
		if err := owner.checkWorkflowNativeSendStop(native.InvocationID); err != nil {
			return err
		}

		var capability NativeCapability
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
		result := GitPublicationResult{Schema: "chr-git-publication-result/v1", State: "unknown", CommandsSent: true, OriginalSourceCommit: native.OriginalSource, AdvertisedSourceCommit: native.OriginalSource, AdvertisedTargetCommit: native.ValidatedBase, ResultCommit: native.PreparedResult, CandidateRevision: request.CandidateRevision}

		return owner.retainWorkflowNativeDispatchResponse(submission, native, result)
	})

	return err
}

func workflowNativeCommandManifest(submission Submission) NativeCommandManifest {
	native := submission.Native

	return NativeCommandManifest{SchemaVersion: 1, SubmissionID: submission.ID, Atomic: true, SourceRef: native.SourceRef, ExpectedSource: native.OriginalSource, NewSource: native.PreparedResult, TargetRef: native.TargetRef, ExpectedTarget: native.ValidatedBase, NewTarget: native.PreparedResult}
}

func (s *Store) completeWorkflowNativeDispatch(ctx context.Context, submission Submission, result GitPublicationResult) (Submission, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	return s.updateWorkflowNativeDispatch(ctx, submission, func(owner *Store, native *NativeSubmission) error {
		if !validWorkflowNativeResponse(*native, result) {
			return workflowstate.ErrConflict
		}

		if err := owner.retainWorkflowNativeDispatchResponse(submission, native, result); err != nil {
			return err
		}

		completion := NativeDispatchCompletion{SchemaVersion: 1, SubmissionID: submission.ID, CandidateRevision: result.CandidateRevision, Identity: *native.DispatchIdentity, ResponseRef: *native.TransportResponseRef, FinishedAt: time.Now().UTC()}

		ref, err := owner.writeWorkflowNativeDispatchEvidence(submission, completion)
		if err != nil {
			return err
		}

		native.DispatchCompletionRef, native.DispatchState = &ref, "quiescent"

		return nil
	})
}

func (s *Store) retainWorkflowNativeDispatchResponse(submission Submission, native *NativeSubmission, result GitPublicationResult) error {
	ref, err := s.writeWorkflowNativeDispatchEvidence(submission, result)
	if err != nil {
		return err
	}

	if native.TransportResponseRef != nil && *native.TransportResponseRef != ref {
		if len(native.TransportResponseHistory) >= workflowstate.PageLimit {
			return workflowstate.ErrConflict
		}

		native.TransportResponseHistory = append(native.TransportResponseHistory, *native.TransportResponseRef)
	}

	native.TransportResponseRef = &ref

	return nil
}

func (s *Store) writeWorkflowNativeDispatchEvidence(submission Submission, value any) (workflowstate.EvidenceRef, error) {
	data, err := json.Marshal(value)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.EvidenceRef{}, workflowstate.ErrConflict
	}

	hash := contenthash.Sum(data)
	ref := workflowstate.EvidenceRef{Path: "run/" + submission.InvocationID + "/evidence/publication/" + submission.ID + "/" + hash + ".json", Hash: hash}

	path, err := s.safePath(ref.Path)
	if err != nil {
		return ref, err
	}

	return ref, opsfile.WriteAtomic(path, data)
}

func (s *Store) updateWorkflowNativeDispatch(ctx context.Context, expected Submission, change func(*Store, *NativeSubmission) error) (submission Submission, err error) {
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

		guard := IntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, BudgetRevision: run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}

		submission, err = owner.retainWorkflowNativeSubmission(ctx, run, NativeSubmissionUpdate{IntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: facts})
		if err != nil {
			return fmt.Errorf("retain dispatch facts: %w", err)
		}

		return err
	})

	return submission, err
}

func (s *Store) readWorkflowOwnedDispatch(ctx context.Context, run workflowInvocationFile, expected Submission) (workflowSubmissionJournal, IntegrationEntry, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return journal, IntegrationEntry{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeDispatchIntentMatches(last, expected) {
		return journal, IntegrationEntry{}, fmt.Errorf("%w: dispatch intent changed", workflowstate.ErrConflict)
	}

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil || identity != *last.Native.DispatchIdentity || last.Native.DispatchState != "dispatching" {
		return journal, IntegrationEntry{}, fmt.Errorf("%w: dispatch owner changed", workflowstate.ErrConflict)
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err == nil && !workflowNativeSubmissionEntryMatches(last, entry) {
		return journal, entry, fmt.Errorf("%w: dispatch generation or candidate changed", workflowstate.ErrConflict)
	}

	return journal, entry, err
}

func workflowNativeDispatchIntentMatches(last, expected Submission) bool {
	return last.ID == expected.ID && last.Native != nil && sameWorkflowNativeSubmissionIntent(*last.Native, *expected.Native) && last.Native.DispatchIdentity != nil
}

func (s *Store) checkWorkflowNativeSendStop(id string) error {
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
