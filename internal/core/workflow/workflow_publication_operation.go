package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SubmissionIntent retains the journal image and permission for one dispatch.
// A retained unknown attempt has Dispatch false and must be reconciled.
type SubmissionIntent struct {
	Submission workflowstate.Submission
	Dispatch   bool
}

// NativePublicationResult retains transport response beside the journal image.
// Neither transport completion nor acceptance proves canonical delivery.
type NativePublicationResult struct {
	Submission workflowstate.Submission
	Transport  workflowstate.GitPublicationResult
}

type workflowNativePublisher struct {
	Run     workflowInvocationFile
	Owner   *Store
	Profile NativePublisherProfile
}

type workflowNativeObservation struct {
	Submission  workflowstate.Submission
	Observation workflowstate.NativePublicationObservation
	Next        IntegrationBudgetGuard
}

type workflowNativeReconciliation struct {
	Submission workflowstate.Submission
	Request    GitPublication
	Update     NativeSubmissionUpdate
}

type workflowNativeDispatch struct {
	Submission workflowstate.Submission
	Request    GitPublication
}

type workflowNativeDispatchEvidence struct {
	Journal workflowSubmissionJournal
	Entry   IntegrationEntry
}

// Handled retains the native-branch barrier, including journal/read failures.
// It does not assert that finalization or a receipt transaction succeeded.
type workflowNativeFinalization struct {
	Handled bool
	Receipt Record
}

type workflowNativeRetention struct {
	Submission workflowstate.Submission
	Next       IntegrationBudgetGuard
}

type workflowNativeReceiptEvidence struct {
	Observation workflowstate.NativePublicationObservation
	Ref         workflowstate.EvidenceRef
}

// Observed preserves the existing publication barrier even on proof errors.
// Only Receipt carries successful finalization evidence.
type workflowPublicationCheckpoint struct {
	Receipt  *Record
	Observed bool
}

type workflowPublicationObservation struct {
	Journal     workflowSubmissionJournal
	PullRequest workflowstate.PullRequestSnapshot
}

type workflowNativeContinuation struct {
	Entry IntegrationEntry
	Guard IntegrationBudgetGuard
}

type workflowClosingNativeProof struct {
	Proof            SliceDeliveryProof
	VerificationHash string
}

// The owner and callback snapshot bind inputs observed from the same image.
// Run and context stay explicit; guarded writes retain their existing Store owner.
type workflowNativeIntentOwner struct {
	Entry   IntegrationEntry
	Journal workflowSubmissionJournal
	Fence   IntegrationBudgetGuard
}

type workflowNativeCallbackSnapshot struct {
	Entry       IntegrationEntry
	Budget      IntegrationBudget
	Submission  workflowstate.Submission
	JournalHash string
}

type workflowSubmissionServices struct {
	Provider PublicationProvider
	Git      IntegrationGit
}

// Canonical metadata and Git proof are the paired read-only observation ports.
type workflowNativeObservers struct {
	Canonical NativePublicationObserver
	Git       IntegrationGit
}

// workflowDeliveryObservers supplies canonical PR and independently verified Git observations.
type workflowDeliveryObservers struct {
	Provider PullRequestProvider
	Git      IntegrationGit
}

// workflowNativeGitObservers pairs transport advertisement with prepared history proof.
type workflowNativeGitObservers struct {
	Transport GitPublicationInspector
	Prepared  PreparedIntegrationGit
}
