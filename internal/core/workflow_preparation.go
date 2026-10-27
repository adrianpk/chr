package core

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowIntegrationValidationPurpose = "integration_validation"

// ErrWorkflowAtomicGuard blocks a provider path without both immutable guards.
var ErrWorkflowAtomicGuard = errors.New("provider cannot atomically guard the source and validated target base")

// WorkflowPublicationCapability is a trusted provider contract for one exact PR
// and method. A source-only merge API does not satisfy AtomicSourceBase.
type WorkflowPublicationCapability struct {
	Repository         string `json:"repository"`
	PullRequestURL     string `json:"pull_request_url"`
	Method             string `json:"method"`
	Mechanism          string `json:"mechanism"`
	Revision           string `json:"revision"`
	AtomicSourceBase   bool   `json:"atomic_source_base"`
	ExactResult        bool   `json:"exact_result"`
	RequiredSignatures bool   `json:"required_signatures,omitempty"`
	SigningProfile     string `json:"signing_profile,omitempty"`
}

// WorkflowPublicationGuardProvider resolves current repository protections.
// Implementations must enforce both guards and the prepared result in the same
// server transaction; a local lock or observation before IO is insufficient.
type WorkflowPublicationGuardProvider interface {
	WorkflowPullRequestProvider
	WorkflowIntegrationCapability(context.Context, WorkflowPullRequestSnapshot, string) (WorkflowPublicationCapability, error)
}

// WorkflowPreparationQuery describes host-owned combined-result preparation.
type WorkflowPreparationQuery struct {
	WorkflowIntegrationQuery
	BaseCommit         string       `json:"base_commit"`
	Lane               WorkflowLane `json:"lane"`
	RequiredSignatures bool         `json:"required_signatures,omitempty"`
	SigningProfile     string       `json:"signing_profile,omitempty"`
}

// WorkflowPreparationGit prepares a clean isolated lane without publishing.
type WorkflowPreparationGit interface {
	PrepareWorkflowIntegration(context.Context, WorkflowPreparationQuery) (WorkflowIntegrationProof, error)
}

// WorkflowPreparationInput freezes source proof, current policy, protection and
// the configured method's combined result before running the host slice gate.
type WorkflowPreparationInput struct {
	Source            WorkflowReviewInput             `json:"source"`
	Capability        WorkflowPublicationCapability   `json:"capability"`
	Git               WorkflowIntegrationProof        `json:"git"`
	Native            *WorkflowNativePreparationInput `json:"native,omitempty"`
	MergerAttemptRefs []workflowstate.EvidenceRef     `json:"merger_attempt_refs,omitempty"`
}

// WorkflowIntegrationCandidate is retained host validation of combined content.
// It is neither a provider response nor proof of delivery to the target.
type WorkflowIntegrationCandidate struct {
	Native       *WorkflowNativeCandidate `json:"native,omitempty"`
	Schema       string                   `json:"schema"`
	InvocationID string                   `json:"invocation_id"`
	Round        int                      `json:"round"`
	Input        WorkflowPreparationInput `json:"input"`
	Lane         WorkflowLane             `json:"lane"`
	Validation   WorkflowTaskResult       `json:"validation"`
}

// WorkflowPreparationBinding retains bounded preparation across candidate changes.
type WorkflowPreparationBinding struct {
	Round      int               `json:"round"`
	Unit       workflowstate.Ref `json:"unit"`
	State      string            `json:"state"`
	SourceHash string            `json:"source_hash"`
	ResultRef  string            `json:"result_ref"`
	ResultHash string            `json:"result_hash"`
}

// WorkflowPublicationGuard contains the exact protected result a publisher may
// submit after journaling intent. Providers must never substitute another method.
type WorkflowPublicationGuard struct {
	PullRequest     WorkflowPullRequestSnapshot        `json:"pull_request"`
	Capability      WorkflowPublicationCapability      `json:"capability"`
	Git             WorkflowIntegrationProof           `json:"git"`
	PreparationRef  string                             `json:"preparation_ref"`
	PreparationHash string                             `json:"preparation_hash"`
	ReviewHash      string                             `json:"review_hash"`
	Policy          workflowstate.WorkflowReviewPolicy `json:"policy"`
}

// PrepareWorkflowSliceIntegration validates the configured combined result on
// a separate owned lane. It preserves task proofs and the canonical PR identity.
func (s *WorkflowStore) PrepareWorkflowSliceIntegration(ctx context.Context, id string, provider WorkflowPublicationGuardProvider, git WorkflowPreparationGit) (candidate WorkflowIntegrationCandidate, err error) {
	claim, prepared, input, round, cached, err := s.prepareWorkflowIntegrationValidation(ctx, id, provider, git)
	if err != nil || cached != nil {
		if cached != nil {
			return *cached, err
		}

		return candidate, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return candidate, err
	}

	defer endWorkflowClaimExecution(claim)

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	deadline, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	result := newWorkflowTaskResult(prepared)
	result.Schema, result.Outcome, result.IntegrationPending = "chr-integration-validation/v1", "integration_validated", true
	validationErr := s.validateWorkflowTask(deadline, claim, prepared, &result)

	result.Successful, result.FinishedAt = validationErr == nil, time.Now().UTC()
	if validationErr != nil {
		result.Outcome, result.Failure = "integration_validation_failed", workflowFailureText(validationErr)
	}

	candidate = WorkflowIntegrationCandidate{Schema: "chr-integration-candidate/v1", InvocationID: id, Round: round, Input: input, Lane: prepared.input.Lane, Validation: result}

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	err = s.finishWorkflowIntegrationValidation(cleanup, claim, prepared, &candidate, provider, validationErr, watched.Err() != nil)

	return candidate, errors.Join(validationErr, err)
}

// CheckWorkflowPublicationGuard rechecks policy, review, combined validation,
// canonical identity and current protection immediately before submission intent.
func (s *WorkflowStore) CheckWorkflowPublicationGuard(ctx context.Context, id string, provider WorkflowPublicationGuardProvider) (guard WorkflowPublicationGuard, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		guard, readErr = owner.checkWorkflowPublicationGuard(ctx, run, provider)

		return readErr
	})

	return guard, err
}
