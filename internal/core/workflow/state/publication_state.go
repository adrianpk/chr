package state

// SubmissionState describes journaled publication intent and its observed outcome.
type SubmissionState string

const (
	// SubmissionAwaitingObservation retains exclusion until independent observation.
	SubmissionAwaitingObservation SubmissionState = "awaiting_observation"
	// SubmissionUnpublished seals an operation that cannot publish later.
	SubmissionUnpublished SubmissionState = "unpublished"
	// SubmissionConfirmed requires canonical PR and exact Git proof.
	SubmissionConfirmed SubmissionState = "confirmed"
)

// SubmissionResponseState classifies an untrusted command reply, not delivery.
type SubmissionResponseState string

const (
	// SubmissionResponseAccepted records provider acceptance for later observation.
	SubmissionResponseAccepted SubmissionResponseState = "accepted"
	// SubmissionResponseUnaccepted records a reply without acceptance.
	SubmissionResponseUnaccepted SubmissionResponseState = "unaccepted"
	// SubmissionResponseUnknown retains uncertainty after an IO failure.
	SubmissionResponseUnknown SubmissionResponseState = "error_unknown"
)

// ReviewState describes retained review work and authorization for exact inputs.
type ReviewState string

const (
	// ReviewPending awaits the distinct supervised review role.
	ReviewPending ReviewState = "pending"
	// ReviewRunning retains a claimed review execution.
	ReviewRunning ReviewState = "running"
	// ReviewWaitingApproval requires a human decision for the exact input hash.
	ReviewWaitingApproval ReviewState = "waiting_approval"
	// ReviewAuthorized permits the guarded publication admission check.
	ReviewAuthorized ReviewState = "authorized"
	// ReviewNeedsRepair requires the selected bounded repair task.
	ReviewNeedsRepair ReviewState = "needs_repair"
	// ReviewInsufficientEvidence retains an evidence blocker.
	ReviewInsufficientEvidence ReviewState = "insufficient_evidence"
	// ReviewHumanRequired retains the reviewer's explicit human-decision blocker.
	ReviewHumanRequired ReviewState = "human_required"
	// ReviewStale invalidates a decision whose retained inputs changed.
	ReviewStale ReviewState = "stale"
	// ReviewRepairLimitExceeded retains exhaustion of the bounded repair budget.
	ReviewRepairLimitExceeded ReviewState = "repair_limit_exceeded"
)

// NativeDispatchState tracks the host-owned transport process lifecycle.
type NativeDispatchState string

const (
	// NativeDispatchNotStarted retains intent before a process identity is bound.
	NativeDispatchNotStarted NativeDispatchState = "not_started"
	// NativeDispatching retains the exact process that may have sent commands.
	NativeDispatching NativeDispatchState = "dispatching"
	// NativeDispatchQuiescent records completion without asserting application.
	NativeDispatchQuiescent NativeDispatchState = "quiescent"
)

// NativeApplicationState classifies Git application independently of forge delivery.
type NativeApplicationState string

const (
	// NativeApplicationNotStarted accompanies an undispatched intent.
	NativeApplicationNotStarted NativeApplicationState = "not_started"
	// NativeApplicationUnknown requires reconciliation before any retry.
	NativeApplicationUnknown NativeApplicationState = "unknown"
	// NativeApplicationNotApplied requires definitive non-application evidence.
	NativeApplicationNotApplied NativeApplicationState = "not_applied"
	// NativeApplicationGitApplied requires independently verified prepared history.
	NativeApplicationGitApplied NativeApplicationState = "git_applied"
)

// NativeObservationState distinguishes canonical metadata from Git application.
type NativeObservationState string

const (
	// NativeObserving retains an incomplete canonical observation.
	NativeObserving NativeObservationState = "observing"
	// NativeReconciliationRequired blocks finalization on missing or conflicting proof.
	NativeReconciliationRequired NativeObservationState = "reconciliation_required"
	// NativeMerged supplies both canonical merge metadata and exact Git proof.
	NativeMerged NativeObservationState = "merged"
)

// GitPublicationState describes transport output without issuing a delivery receipt.
type GitPublicationState string

const (
	// GitPublicationNotSent records rejection before transport dispatch.
	GitPublicationNotSent GitPublicationState = "not_sent"
	// GitPublicationUnknown means commands may have reached the server.
	GitPublicationUnknown GitPublicationState = "unknown"
	// GitPublicationRejected retains the transport's rejection response.
	GitPublicationRejected GitPublicationState = "rejected"
	// GitPublicationApplied retains transport application for independent verification.
	GitPublicationApplied GitPublicationState = "applied"
)
