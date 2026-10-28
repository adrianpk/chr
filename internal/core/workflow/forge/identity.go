package forge

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MatchesPullRequest checks the retained legacy or native publication PR identity.
func MatchesPullRequest(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	if submission.Native != nil {
		return matchesNativePullRequest(submission, pr)
	}

	if !matchesLegacyPullRequest(submission, pr) {
		return false
	}

	if pr.Merged {
		return pr.State == "closed" && workflowstate.ValidCommitID(pr.ResultCommit)
	}

	return pr.State == "open" || pr.State == "closed"
}

func matchesNativePullRequest(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	native := submission.Native
	expected := [5]string{native.CanonicalPR, native.SourceRepository, native.SourceRepository, native.SourceRef, native.TargetRef}
	actual := [5]string{pr.URL, pr.Repository, pr.HeadRepository, "refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch}

	// PR head metadata may lag at S or advance after P. The immutable receipt's
	// transaction/application evidence establishes published P independently.
	return expected == actual && workflowstate.ValidCommitID(pr.HeadCommit) && pr.Merged && pr.State == "closed" && pr.ResultCommit == native.PreparedResult
}

func matchesLegacyPullRequest(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	expected := submission.Guard.PullRequest

	return pr.URL == expected.URL && pr.Repository == expected.Repository && pr.HeadRepository == expected.HeadRepository && pr.HeadBranch == expected.HeadBranch && pr.BaseBranch == expected.BaseBranch && pr.HeadCommit == expected.HeadCommit
}
