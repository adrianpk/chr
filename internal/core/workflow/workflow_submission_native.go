package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativeDispatchCompletion records a returned host-owned transport call.
// It proves local quiescence, never remote non-application by itself.
type NativeDispatchCompletion struct {
	SchemaVersion     int                           `json:"schema_version"`
	SubmissionID      string                        `json:"submission_id"`
	CandidateRevision string                        `json:"candidate_revision"`
	Identity          workflowstate.ProcessIdentity `json:"identity"`
	ResponseRef       workflowstate.EvidenceRef     `json:"response_ref"`
	FinishedAt        time.Time                     `json:"finished_at"`
}

// NativeCommandManifest is sanitized evidence, never a packet trace.
type NativeCommandManifest struct {
	SchemaVersion  int    `json:"schema_version"`
	SubmissionID   string `json:"submission_id"`
	Atomic         bool   `json:"atomic"`
	SourceRef      string `json:"source_ref"`
	ExpectedSource string `json:"expected_source"`
	NewSource      string `json:"new_source"`
	TargetRef      string `json:"target_ref"`
	ExpectedTarget string `json:"expected_target"`
	NewTarget      string `json:"new_target"`
}

// NativeNonApplicationProof seals a dispatch that cannot apply later.
// InputChanged is necessary before requeue; missing responses are insufficient.
type NativeNonApplicationProof struct {
	SchemaVersion  int    `json:"schema_version"`
	SubmissionID   string `json:"submission_id"`
	CannotApply    bool   `json:"cannot_apply"`
	InputChanged   bool   `json:"input_changed"`
	ObservedSource string `json:"observed_source"`
	ObservedTarget string `json:"observed_target"`
}

// NativeSubmissionUpdate fences a callback against the common image.
// Facts retain immutable intent and append observations; they do not execute IO.
type NativeSubmissionUpdate struct {
	IntegrationBudgetGuard
	SubmissionID string
	JournalHash  string
	Facts        workflowstate.NativeSubmission
}

// BeginWorkflowNativeSubmission retains intent before the native transport IO.
// Retrying an unknown attempt only repairs its projection, never grants dispatch.
func (s *Store) BeginWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider) (result SubmissionIntent, err error) {
	var (
		submission workflowstate.Submission
		dispatch   bool
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		submissionIntent, readErr := invocationLocation.Owner.beginWorkflowNativeSubmission(ctx, invocationLocation.Run, guard, provider)
		submission, dispatch = submissionIntent.Submission, submissionIntent.Dispatch

		return readErr
	})

	return SubmissionIntent{Submission: submission, Dispatch: dispatch}, err
}

// RetainWorkflowNativeSubmission records sanitized transport/observation facts.
// Only exact Git plus the original canonical PR can confirm native delivery.
func (s *Store) RetainWorkflowNativeSubmission(ctx context.Context, request NativeSubmissionUpdate) (submission workflowstate.Submission, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation2, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		submission, readErr = invocationLocation2.Owner.retainWorkflowNativeSubmission(ctx, invocationLocation2.Run, request)

		return readErr
	})

	return submission, err
}

func workflowNativeSubmissionIntent(candidate workflowstate.IntegrationCandidate, run workflowInvocationFile, entry IntegrationEntry) workflowstate.NativeSubmission {
	now := time.Now().UTC()

	return workflowstate.NativeSubmission{NativeRecordBinding: candidate.Native.NativeRecordBinding, CandidateHash: candidate.Native.CandidateHash, CandidateRef: *entry.CandidateRef, ClaimGeneration: entry.ClaimGeneration, Mechanism: workspaceconfig.WorkflowPublicationGitAtomicDualRef, ReviewGrantRef: workflowstate.EvidenceRef{Path: run.Review.ResultRef, Hash: run.Review.ResultHash}, FastForwardProofRef: candidate.Native.HistoryManifestRef, IntentTime: now, DispatchState: "not_started", ApplicationState: "not_started", ObservationDeadline: now.Add(3 * time.Minute)}
}
