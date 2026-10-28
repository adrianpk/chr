package workflow

import (
	"context"
	"errors"
	"time"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowIntegrationValidationPurpose = "integration_validation"

// ErrWorkflowAtomicGuard blocks a provider path without both immutable guards.
var ErrWorkflowAtomicGuard = errors.New("provider cannot atomically guard the source and validated target base")

// PublicationGuardProvider resolves current repository protections.
// Implementations must enforce both guards and the prepared result in the same
// server transaction; a local lock or observation before IO is insufficient.
type PublicationGuardProvider interface {
	PullRequestProvider
	WorkflowIntegrationCapability(context.Context, workflowstate.PullRequestSnapshot, string) (workflowstate.PublicationCapability, error)
}

// PreparationQuery describes host-owned combined-result preparation.
type PreparationQuery struct {
	workflowstate.IntegrationQuery
	BaseCommit         string             `json:"base_commit"`
	Lane               workflowstate.Lane `json:"lane"`
	RequiredSignatures bool               `json:"required_signatures,omitempty"`
	SigningProfile     string             `json:"signing_profile,omitempty"`
}

// PreparationGit prepares a clean isolated lane without publishing.
type PreparationGit interface {
	PrepareWorkflowIntegration(context.Context, PreparationQuery) (workflowstate.IntegrationProof, error)
}

// PreparationBinding retains bounded preparation across candidate changes.
type PreparationBinding struct {
	Round      int               `json:"round"`
	Unit       workflowstate.Ref `json:"unit"`
	State      string            `json:"state"`
	SourceHash string            `json:"source_hash"`
	ResultRef  string            `json:"result_ref"`
	ResultHash string            `json:"result_hash"`
}

// PrepareWorkflowSliceIntegration validates the configured combined result on
// a separate owned lane. It preserves task proofs and the canonical PR identity.
func (s *Store) PrepareWorkflowSliceIntegration(ctx context.Context, id string, provider PublicationGuardProvider, git PreparationGit) (candidate workflowstate.IntegrationCandidate, err error) {
	selection, err := s.prepareWorkflowIntegrationValidation(ctx, id, provider, git)
	if err != nil || selection.Cached != nil {
		if selection.Cached != nil {
			return *selection.Cached, err
		}

		return candidate, err
	}

	claim, prepared, input := selection.Claimed.Claim, selection.Claimed.Prepared, selection.Claimed.Input

	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return candidate, err
	}

	defer endWorkflowClaimExecution(claim)

	watched, stop := context.WithCancel(ctx)

	s.watchWorkflowStop(watched, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	deadline, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	validation := workflowIntegrationValidation{store: s, claim: claim, prepared: prepared, provider: provider}

	return workflowintegrate.ValidateCandidate(deadline, watched, workflowintegrate.ValidationInput{InvocationID: id, Round: selection.Round, Preparation: input, Lane: prepared.input.Lane, Result: newWorkflowTaskResult(prepared)}, validation)
}

// CheckWorkflowPublicationGuard rechecks policy, review, combined validation,
// canonical identity and current protection immediately before submission intent.
func (s *Store) CheckWorkflowPublicationGuard(ctx context.Context, id string, provider PublicationGuardProvider) (guard workflowstate.PublicationGuard, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		guard, readErr = invocationLocation.Owner.checkWorkflowPublicationGuard(ctx, invocationLocation.Run, provider)

		return readErr
	})

	return guard, err
}
