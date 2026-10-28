package forge

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PullRequestMismatch requires one exact canonical PR without confirming delivery.
func PullRequestMismatch(binding workflowstate.PullRequestBinding, snapshots []workflowstate.PullRequestSnapshot) string {
	if len(snapshots) != 1 {
		return "canonical_pr_ambiguous"
	}

	pr := snapshots[0]
	if !pullRequestIdentityMatches(binding, pr) {
		return "canonical_pr_mismatch"
	}

	if pr.State == "closed" && !pr.Merged {
		return "canonical_pr_closed_unmerged"
	}

	if pr.State != "open" && pr.State != "closed" {
		return "canonical_pr_state_unknown"
	}

	if pr.Merged && (pr.State != "closed" || !workflowstate.ValidCommitID(pr.ResultCommit)) {
		return "canonical_pr_merge_unverifiable"
	}

	return ""
}

func pullRequestIdentityMatches(binding workflowstate.PullRequestBinding, pr workflowstate.PullRequestSnapshot) bool {
	if pr.Repository != binding.Request.Repository || pr.HeadRepository != pr.Repository || pr.HeadBranch != binding.Request.HeadBranch || pr.BaseBranch != binding.Request.BaseBranch || pr.HeadCommit != binding.SourceCommit || !workflowstate.ValidCommitID(pr.BaseCommit) {
		return false
	}

	if binding.URL != "" && binding.URL != pr.URL {
		return false
	}

	repository, err := PullRequestRepository(pr.URL)

	return err == nil && repository == binding.Request.Repository
}

// PullRequestRepository accepts only the canonical provider PR URL spelling.
func PullRequestRepository(raw string) (string, error) {
	reference, err := ParseReference(raw)
	if err != nil {
		return "", err
	}

	marker := "/pulls/"
	if reference.Provider == GitHub {
		marker = "/pull/"
	}

	pos := len(raw) - len(marker) - len(fmt.Sprint(reference.Number))
	if pos < 0 || raw[pos:] != marker+fmt.Sprint(reference.Number) {
		return "", workflowstate.ErrConflict
	}

	return workflowstate.CanonicalRepository(raw[:pos])
}
