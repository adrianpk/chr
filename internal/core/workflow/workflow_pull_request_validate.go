package workflow

import (
	"fmt"
	"slices"
	"strings"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowPullRequestBinding(run workflowInvocationFile) error {
	binding := run.PullRequest
	if binding == nil {
		return nil
	}

	if !validWorkflowPullRequestIdentity(run, *binding) {
		return fmt.Errorf("invalid workflow canonical PR identity")
	}

	repository, err := workflowstate.CanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil || repository != binding.Request.Repository || !workflowforge.ValidPullRequestContent(workflowstate.PullRequestContent{BaseBranch: binding.Request.BaseBranch, Title: binding.Request.Title, Body: binding.Request.Body}) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowPullRequestState(*binding); err != nil {
		return err
	}

	if !validWorkflowPullRequestCandidate(run, *binding) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowPullRequestObservationRef(run.ID, *binding)
}

func validWorkflowPullRequestIdentity(run workflowInvocationFile, binding PullRequestBinding) bool {
	if run.Lane == nil {
		return false
	}

	return binding.Schema == "chr-slice-pr/v1" && binding.ID == workflowPullRequestBindingID(run, binding.Request.Repository) && (binding.Unit.Kind == "slice" || binding.Unit.Kind == "ticket") && binding.Unit == run.Lane.Unit && binding.TargetCommit == run.ExecutionSelection().Target.Commit && binding.Request.HeadBranch == strings.TrimPrefix(run.Lane.Branch, "refs/heads/")
}

func validateWorkflowPullRequestState(binding PullRequestBinding) error {
	if !slices.Contains([]string{"prepared", "creating", "awaiting_observation", "bound", "blocked"}, binding.State) {
		return workflowstate.ErrConflict
	}

	if binding.State == "bound" && binding.URL == "" {
		return workflowstate.ErrConflict
	}

	if binding.URL == "" {
		return nil
	}

	repository, err := workflowPullRequestRepository(binding.URL)
	if err != nil || repository != binding.Request.Repository {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowPullRequestCandidate(run workflowInvocationFile, binding PullRequestBinding) bool {
	if !workflowstate.ValidCommitID(binding.SourceCommit) || !workflowstate.ValidCommitID(binding.SourceTree) {
		return false
	}

	if binding.Unit.Kind == "ticket" {
		return validWorkflowTicketPullRequestCandidate(run, binding)
	}

	run.CandidateRef, run.CandidateHash = binding.CandidateRef, binding.CandidateHash
	for _, attempt := range run.Attempts {
		if attempt.ID == binding.CandidateID && workflowAttemptOwnsCandidate(run, attempt) {
			return true
		}
	}

	return false
}

func validateWorkflowPullRequestObservationRef(id string, binding PullRequestBinding) error {
	if binding.ObservationCount == 0 && binding.ObservationRef == "" && binding.ObservationHash == "" {
		return nil
	}

	legacy := workflowPullRequestObservationPath(id, binding.ObservationCount)

	scoped := workflowUnitPullRequestObservationPath(id, binding.ID, binding.ObservationCount)
	if binding.ObservationCount < 1 || binding.ObservationCount > workflowstate.PageLimit || !slices.Contains([]string{legacy, scoped}, binding.ObservationRef) || len(binding.ObservationHash) != 64 {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowTicketPullRequestCandidate(run workflowInvocationFile, binding PullRequestBinding) bool {
	run.CandidateRef, run.CandidateHash = binding.CandidateRef, binding.CandidateHash

	return validWorkflowTicketCandidateRef(run) && binding.CandidateRef == workflowTicketCandidatePath(run.ID, binding.CandidateID)
}
