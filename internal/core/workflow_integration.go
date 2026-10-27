package core

import (
	"context"
	"fmt"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationQuery binds host Git verification to canonical provider facts.
type WorkflowIntegrationQuery struct {
	Repository      string `json:"repository"`
	TargetBranch    string `json:"target_branch"`
	Method          string `json:"method"`
	SourceCommit    string `json:"source_commit"`
	SourceTree      string `json:"source_tree"`
	OriginCommit    string `json:"origin_commit"`
	ResultCommit    string `json:"result_commit"`
	SourceDirectory string `json:"-"`
}

// WorkflowIntegrationCommit retains ordered source/result tree and parent proof.
type WorkflowIntegrationCommit struct {
	Source string `json:"source"`
	Commit string `json:"commit"`
	Tree   string `json:"tree"`
	Parent string `json:"parent"`
}

// WorkflowIntegrationProof contains host observations of an actual remote target.
type WorkflowIntegrationProof struct {
	WorkflowIntegrationQuery
	BaseCommit     string                         `json:"base_commit"`
	PreparedTree   string                         `json:"prepared_tree"`
	VerifiedTree   string                         `json:"verified_tree"`
	ObservedTarget string                         `json:"observed_target"`
	History        []WorkflowIntegrationCommit    `json:"history"`
	HistoryHash    string                         `json:"history_hash"`
	Provenance     *WorkflowIntegrationProvenance `json:"provenance,omitempty"`
}

// WorkflowIntegrationGit is a trusted host IO port, never a worker evidence port.
// It fetches the named canonical target and verifies method-specific Git proof.
type WorkflowIntegrationGit interface {
	VerifyWorkflowIntegration(context.Context, WorkflowIntegrationQuery) (WorkflowIntegrationProof, error)
}

// WorkflowPreparedIntegrationGit re-proves a retained history version from the
// actual canonical target, including declared fixes and attribution metadata.
type WorkflowPreparedIntegrationGit interface {
	VerifyPreparedWorkflowIntegration(context.Context, WorkflowIntegrationQuery, WorkflowIntegrationProof) (WorkflowIntegrationProof, error)
}

// WorkflowSliceDeliveryProof retains verified external delivery before finalization.
type WorkflowSliceDeliveryProof struct {
	Schema           string                      `json:"schema"`
	InvocationID     string                      `json:"invocation_id"`
	Unit             workflowstate.Ref           `json:"unit"`
	CandidateID      string                      `json:"candidate_id"`
	CandidateRef     string                      `json:"candidate_ref"`
	CandidateHash    string                      `json:"candidate_hash"`
	SubmissionID     string                      `json:"submission_id"`
	PullRequest      WorkflowPullRequestSnapshot `json:"pull_request"`
	Git              WorkflowIntegrationProof    `json:"git"`
	ConfirmedAt      time.Time                   `json:"confirmed_at"`
	NativeReceiptRef *workflowstate.EvidenceRef  `json:"native_receipt_ref,omitempty"`
}

// VerifyWorkflowSliceIntegration observes external integration without publishing
// a target or marking any task done. Worker claims and structural receipts cannot
// substitute for canonical provider metadata and fetched method-specific proof.
func (s *WorkflowStore) VerifyWorkflowSliceIntegration(ctx context.Context, id, method string, provider WorkflowPullRequestProvider, git WorkflowIntegrationGit) (proof WorkflowSliceDeliveryProof, err error) {
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

func (s *WorkflowStore) verifyWorkflowSliceIntegration(ctx context.Context, run workflowInvocationFile, method string, provider WorkflowPullRequestProvider, git WorkflowIntegrationGit) (WorkflowSliceDeliveryProof, error) {
	observation, candidate, err := s.workflowDeliveryCandidate(ctx, run)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	pr, err := s.observeWorkflowDeliveryPR(ctx, &run, observation, provider)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	if err := s.reconcileWorkflowPublishedSubmission(ctx, &run, pr, provider, git); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	query := workflowIntegrationQuery(run, candidate, pr, method)

	gitProof, err := git.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	if err := validateWorkflowIntegrationProof(query, gitProof); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	if err := checkWorkflowVerifiedPR(ctx, run, gitProof, provider); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	if _, _, err := s.workflowDeliveryCandidate(ctx, run); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	proof := WorkflowSliceDeliveryProof{Schema: "chr-slice-delivery-proof/v1", InvocationID: run.ID, Unit: candidate.UnitInput().Unit, CandidateID: candidate.ID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, SubmissionID: workflowDeliverySubmissionID(run), PullRequest: pr, Git: gitProof, ConfirmedAt: time.Now().UTC()}

	return s.persistWorkflowDeliveryProof(&run, proof)
}

func (s *WorkflowStore) workflowDeliveryCandidate(ctx context.Context, run workflowInvocationFile) (workflowSliceObservation, WorkflowSliceCandidate, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" || !slices.Contains([]string{"waiting_review", "waiting_integration", "blocked", "paused", "cancelled"}, run.Status) {
		return workflowSliceObservation{}, WorkflowSliceCandidate{}, fmt.Errorf("workflow requires an unchanged bound slice PR at its delivery barrier")
	}

	if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
		return workflowSliceObservation{}, WorkflowSliceCandidate{}, err
	}

	if err := s.checkWorkflowDeliveryEvidence(run); err != nil {
		return workflowSliceObservation{}, WorkflowSliceCandidate{}, err
	}

	if run.DeliveryProofRef != "" {
		if _, err := s.readWorkflowDeliveryProof(run); err != nil {
			return workflowSliceObservation{}, WorkflowSliceCandidate{}, err
		}
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return observation, WorkflowSliceCandidate{}, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)

	return observation, candidate, err
}

func workflowIntegrationQuery(run workflowInvocationFile, candidate WorkflowSliceCandidate, pr WorkflowPullRequestSnapshot, method string) WorkflowIntegrationQuery {
	return WorkflowIntegrationQuery{Repository: run.PullRequest.Request.Repository, TargetBranch: run.PullRequest.Request.BaseBranch, Method: method, SourceCommit: candidate.Result.Commit, SourceTree: candidate.Result.Tree, OriginCommit: candidate.Target.Commit, ResultCommit: pr.ResultCommit, SourceDirectory: candidate.Lane.Worktree}
}

func (s *WorkflowStore) observeWorkflowDeliveryPR(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider WorkflowPullRequestProvider) (WorkflowPullRequestSnapshot, error) {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return pr, err
	}

	if err := s.acceptWorkflowPullRequest(ctx, run, observation, []WorkflowPullRequestSnapshot{pr}); err != nil {
		return pr, err
	}

	if !pr.Merged {
		return pr, fmt.Errorf("canonical PR has no confirmed integration")
	}

	return pr, nil
}
