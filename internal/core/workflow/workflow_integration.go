package workflow

import (
	"context"
	"fmt"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationGit is a trusted host IO port, never a worker evidence port.
// It fetches the named canonical target and verifies method-specific Git proof.
type IntegrationGit interface {
	VerifyWorkflowIntegration(context.Context, workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error)
}

// PreparedIntegrationGit re-proves a retained history version from the
// actual canonical target, including declared fixes and attribution metadata.
type PreparedIntegrationGit interface {
	VerifyPreparedWorkflowIntegration(context.Context, workflowstate.IntegrationQuery, workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error)
}

// SliceDeliveryProof retains verified external delivery before finalization.
type SliceDeliveryProof struct {
	Schema           string                            `json:"schema"`
	InvocationID     string                            `json:"invocation_id"`
	Unit             workflowstate.Ref                 `json:"unit"`
	CandidateID      string                            `json:"candidate_id"`
	CandidateRef     string                            `json:"candidate_ref"`
	CandidateHash    string                            `json:"candidate_hash"`
	SubmissionID     string                            `json:"submission_id"`
	PullRequest      workflowstate.PullRequestSnapshot `json:"pull_request"`
	Git              workflowstate.IntegrationProof    `json:"git"`
	ConfirmedAt      time.Time                         `json:"confirmed_at"`
	NativeReceiptRef *workflowstate.EvidenceRef        `json:"native_receipt_ref,omitempty"`
}

// VerifyWorkflowSliceIntegration observes external integration without publishing
// a target or marking any task done. Worker claims and structural receipts cannot
// substitute for canonical provider metadata and fetched method-specific proof.
func (s *Store) VerifyWorkflowSliceIntegration(ctx context.Context, id, method string, provider PullRequestProvider, git IntegrationGit) (proof SliceDeliveryProof, err error) {
	if provider == nil || git == nil || !slices.Contains([]string{"merge", "rebase", "squash"}, method) {
		return proof, fmt.Errorf("canonical provider, Git observer and supported integration method required")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		proof, readErr = owner.verifyWorkflowSliceIntegration(ctx, run, method, provider, git)

		return readErr
	})

	return proof, err
}

func (s *Store) verifyWorkflowSliceIntegration(ctx context.Context, run workflowInvocationFile, method string, provider PullRequestProvider, git IntegrationGit) (SliceDeliveryProof, error) {
	observation, candidate, err := s.workflowDeliveryCandidate(ctx, run)
	if err != nil {
		return SliceDeliveryProof{}, err
	}

	pr, err := s.observeWorkflowDeliveryPR(ctx, &run, observation, provider)
	if err != nil {
		return SliceDeliveryProof{}, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return SliceDeliveryProof{}, err
	}

	if err := s.reconcileWorkflowPublishedSubmission(ctx, &run, pr, provider, git); err != nil {
		return SliceDeliveryProof{}, err
	}

	query := workflowIntegrationQuery(run, candidate, pr, method)

	gitProof, err := git.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return SliceDeliveryProof{}, err
	}

	if err := validateWorkflowIntegrationProof(query, gitProof); err != nil {
		return SliceDeliveryProof{}, err
	}

	if err := checkWorkflowVerifiedPR(ctx, run, gitProof, provider); err != nil {
		return SliceDeliveryProof{}, err
	}

	if _, _, err := s.workflowDeliveryCandidate(ctx, run); err != nil {
		return SliceDeliveryProof{}, err
	}

	proof := SliceDeliveryProof{Schema: "chr-slice-delivery-proof/v1", InvocationID: run.ID, Unit: candidate.UnitInput().Unit, CandidateID: candidate.ID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, SubmissionID: workflowDeliverySubmissionID(run), PullRequest: pr, Git: gitProof, ConfirmedAt: time.Now().UTC()}

	return s.persistWorkflowDeliveryProof(&run, proof)
}

func (s *Store) workflowDeliveryCandidate(ctx context.Context, run workflowInvocationFile) (workflowSliceObservation, SliceCandidate, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" || !slices.Contains([]string{"waiting_review", "waiting_integration", "blocked", "paused", "cancelled"}, run.Status) {
		return workflowSliceObservation{}, SliceCandidate{}, fmt.Errorf("workflow requires an unchanged bound slice PR at its delivery barrier")
	}

	if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
		return workflowSliceObservation{}, SliceCandidate{}, err
	}

	if err := s.checkWorkflowDeliveryEvidence(run); err != nil {
		return workflowSliceObservation{}, SliceCandidate{}, err
	}

	if run.DeliveryProofRef != "" {
		if _, err := s.readWorkflowDeliveryProof(run); err != nil {
			return workflowSliceObservation{}, SliceCandidate{}, err
		}
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return observation, SliceCandidate{}, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)

	return observation, candidate, err
}

func workflowIntegrationQuery(run workflowInvocationFile, candidate SliceCandidate, pr workflowstate.PullRequestSnapshot, method string) workflowstate.IntegrationQuery {
	return workflowstate.IntegrationQuery{Repository: run.PullRequest.Request.Repository, TargetBranch: run.PullRequest.Request.BaseBranch, Method: method, SourceCommit: candidate.Result.Commit, SourceTree: candidate.Result.Tree, OriginCommit: candidate.Target.Commit, ResultCommit: pr.ResultCommit, SourceDirectory: candidate.Lane.Worktree}
}

func (s *Store) observeWorkflowDeliveryPR(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider PullRequestProvider) (workflowstate.PullRequestSnapshot, error) {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return pr, err
	}

	if err := s.acceptWorkflowPullRequest(ctx, run, observation, []workflowstate.PullRequestSnapshot{pr}); err != nil {
		return pr, err
	}

	if !pr.Merged {
		return pr, fmt.Errorf("canonical PR has no confirmed integration")
	}

	return pr, nil
}
