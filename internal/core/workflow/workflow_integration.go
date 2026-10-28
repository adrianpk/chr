package workflow

import (
	"context"
	"fmt"
	"slices"
	"time"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
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

// VerifyWorkflowSliceIntegration observes external integration without publishing
// a target or marking any task done. Worker claims and structural receipts cannot
// substitute for canonical provider metadata and fetched method-specific proof.
func (s *Store) VerifyWorkflowSliceIntegration(ctx context.Context, id, method string, provider PullRequestProvider, git IntegrationGit) (proof workflowstate.SliceDeliveryProof, err error) {
	if provider == nil || git == nil || !slices.Contains([]string{"merge", "rebase", "squash"}, method) {
		return proof, fmt.Errorf("canonical provider, Git observer and supported integration method required")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		proof, readErr = (workflowPullRequestHost{store: invocationLocation.Owner}).verifyDelivery(ctx, invocationLocation.Run, method, provider, git)

		return readErr
	})

	return proof, err
}

func (host workflowPullRequestHost) verifyDelivery(ctx context.Context, run workflowInvocationFile, method string, provider PullRequestProvider, git IntegrationGit) (workflowstate.SliceDeliveryProof, error) {
	deliveryPreparation, err := host.store.workflowDeliveryCandidate(ctx, run)

	observation, candidate := deliveryPreparation.Observation, deliveryPreparation.Candidate

	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	pr, err := (workflowPullRequestHost{store: host.store}).observeDelivery(ctx, &run, observation, provider)
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	run, err = host.store.readWorkflowInvocation(run.ID)
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	if err := host.store.reconcileWorkflowPublishedSubmission(ctx, &run, pr, workflowDeliveryObservers{Provider: provider, Git: git}); err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	query := workflowIntegrationQuery(run, candidate, pr, method)

	gitProof, err := git.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	if err := workflowintegrate.CheckProof(query, gitProof, workflowIntegrationRepairs{}); err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	if err := checkWorkflowVerifiedPR(ctx, run, gitProof, provider); err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	if _, err := host.store.workflowDeliveryCandidate(ctx, run); err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	proof := workflowstate.SliceDeliveryProof{Schema: "chr-slice-delivery-proof/v1", InvocationID: run.ID, Unit: candidate.UnitInput().Unit, CandidateID: candidate.ID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, SubmissionID: workflowDeliverySubmissionID(run), PullRequest: pr, Git: gitProof, ConfirmedAt: time.Now().UTC()}

	return host.store.persistWorkflowDeliveryProof(&run, proof)
}

func (s *Store) workflowDeliveryCandidate(ctx context.Context, run workflowInvocationFile) (workflowDeliveryPreparation, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" || !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingReview, workflowstate.InvocationWaitingIntegration, workflowstate.InvocationBlocked, workflowstate.InvocationPaused, workflowstate.InvocationCancelled}, run.Status) {
		return workflowDeliveryPreparation{Observation: workflowSliceObservation{}, Candidate: SliceCandidate{}}, fmt.Errorf("workflow requires an unchanged bound slice PR at its delivery barrier")
	}

	if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
		return workflowDeliveryPreparation{Observation: workflowSliceObservation{}, Candidate: SliceCandidate{}}, err
	}

	if err := s.checkWorkflowDeliveryEvidence(run); err != nil {
		return workflowDeliveryPreparation{Observation: workflowSliceObservation{}, Candidate: SliceCandidate{}}, err
	}

	if run.DeliveryProofRef != "" {
		if _, err := s.readWorkflowDeliveryProof(run); err != nil {
			return workflowDeliveryPreparation{Observation: workflowSliceObservation{}, Candidate: SliceCandidate{}}, err
		}
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return workflowDeliveryPreparation{Observation: observation, Candidate: SliceCandidate{}}, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)

	return workflowDeliveryPreparation{Observation: observation, Candidate: candidate}, err
}

func workflowIntegrationQuery(run workflowInvocationFile, candidate SliceCandidate, pr workflowstate.PullRequestSnapshot, method string) workflowstate.IntegrationQuery {
	return workflowstate.IntegrationQuery{Repository: run.PullRequest.Request.Repository, TargetBranch: run.PullRequest.Request.BaseBranch, Method: method, SourceCommit: candidate.Result.Commit, SourceTree: candidate.Result.Tree, OriginCommit: candidate.Target.Commit, ResultCommit: pr.ResultCommit, SourceDirectory: candidate.Lane.Worktree}
}

func (host workflowPullRequestHost) observeDelivery(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider PullRequestProvider) (workflowstate.PullRequestSnapshot, error) {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return pr, err
	}

	if err := host.store.acceptWorkflowPullRequest(ctx, run, observation, []workflowstate.PullRequestSnapshot{pr}); err != nil {
		return pr, err
	}

	if !pr.Merged {
		return pr, fmt.Errorf("canonical PR has no confirmed integration")
	}

	return pr, nil
}
