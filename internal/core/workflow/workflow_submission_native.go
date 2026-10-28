package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativeSubmission extends the common journal with the exact Git intent.
// Git application alone retains awaiting_observation and cannot issue a receipt.
type NativeSubmission struct {
	NativeRecordBinding
	CandidateHash            string                      `json:"candidate_hash"`
	CandidateRef             workflowstate.EvidenceRef   `json:"candidate_ref"`
	ClaimGeneration          uint64                      `json:"claim_generation"`
	Mechanism                string                      `json:"mechanism"`
	ReviewGrantRef           workflowstate.EvidenceRef   `json:"review_grant_ref"`
	FastForwardProofRef      workflowstate.EvidenceRef   `json:"fast_forward_proof_ref"`
	IntentTime               time.Time                   `json:"intent_time"`
	DispatchState            string                      `json:"dispatch_state"`
	DispatchIdentity         *ProcessIdentity            `json:"dispatch_identity"`
	DispatchCompletionRef    *workflowstate.EvidenceRef  `json:"dispatch_completion_ref,omitempty"`
	AdvertisementRef         *workflowstate.EvidenceRef  `json:"advertisement_observation_ref"`
	CommandManifestRef       *workflowstate.EvidenceRef  `json:"command_manifest_ref"`
	TransportResponseRef     *workflowstate.EvidenceRef  `json:"transport_response_ref"`
	TransportResponseHistory []workflowstate.EvidenceRef `json:"transport_response_history,omitempty"`
	ApplicationState         string                      `json:"application_state"`
	ObservationDeadline      time.Time                   `json:"observation_deadline"`
	ObservationRefs          []workflowstate.EvidenceRef `json:"observation_refs"`
	GitObservationRefs       []workflowstate.EvidenceRef `json:"git_observation_refs,omitempty"`
	ReceiptRef               *workflowstate.EvidenceRef  `json:"receipt_ref"`
	NonApplicationRef        *workflowstate.EvidenceRef  `json:"non_application_ref"`
}

// NativeDispatchCompletion records a returned host-owned transport call.
// It proves local quiescence, never remote non-application by itself.
type NativeDispatchCompletion struct {
	SchemaVersion     int                       `json:"schema_version"`
	SubmissionID      string                    `json:"submission_id"`
	CandidateRevision string                    `json:"candidate_revision"`
	Identity          ProcessIdentity           `json:"identity"`
	ResponseRef       workflowstate.EvidenceRef `json:"response_ref"`
	FinishedAt        time.Time                 `json:"finished_at"`
}

// NativeGitObservation binds independently fetched refs and history to
// the same intent. It asserts no forge merge metadata or delivery receipt.
type NativeGitObservation struct {
	SchemaVersion     int               `json:"schema_version"`
	SubmissionID      string            `json:"submission_id"`
	CandidateRevision string            `json:"candidate_revision"`
	SourceCommit      string            `json:"source_commit"`
	TargetCommit      string            `json:"target_commit"`
	Condition         string            `json:"condition"`
	Git               *IntegrationProof `json:"git,omitempty"`
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
	Facts        NativeSubmission
}

// BeginWorkflowNativeSubmission retains intent before the native transport IO.
// Retrying an unknown attempt only repairs its projection, never grants dispatch.
func (s *Store) BeginWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider) (submission Submission, dispatch bool, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		submission, dispatch, readErr = owner.beginWorkflowNativeSubmission(ctx, run, guard, provider)

		return readErr
	})

	return submission, dispatch, err
}

// RetainWorkflowNativeSubmission records sanitized transport/observation facts.
// Only exact Git plus the original canonical PR can confirm native delivery.
func (s *Store) RetainWorkflowNativeSubmission(ctx context.Context, request NativeSubmissionUpdate) (submission Submission, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		submission, readErr = owner.retainWorkflowNativeSubmission(ctx, run, request)

		return readErr
	})

	return submission, err
}

func workflowNativeSubmissionIntent(candidate IntegrationCandidate, run workflowInvocationFile, entry IntegrationEntry) NativeSubmission {
	now := time.Now().UTC()

	return NativeSubmission{NativeRecordBinding: candidate.Native.NativeRecordBinding, CandidateHash: candidate.Native.CandidateHash, CandidateRef: *entry.CandidateRef, ClaimGeneration: entry.ClaimGeneration, Mechanism: workspaceconfig.WorkflowPublicationGitAtomicDualRef, ReviewGrantRef: workflowstate.EvidenceRef{Path: run.Review.ResultRef, Hash: run.Review.ResultHash}, FastForwardProofRef: candidate.Native.HistoryManifestRef, IntentTime: now, DispatchState: "not_started", ApplicationState: "not_started", ObservationDeadline: now.Add(3 * time.Minute)}
}
