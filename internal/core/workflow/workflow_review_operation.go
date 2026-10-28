package workflow

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

// Review inputs retain the observation used to bind them, including partial IO results.
type workflowReviewObservation struct {
	Observation workflowSliceObservation
	Input       workflowstate.ReviewInput
}

type workflowReviewPullRequest struct {
	PullRequest  workflowstate.PullRequestSnapshot
	TargetCommit string
}

type workflowReviewRecord struct {
	Input    workflowstate.ReviewInput
	Decision workflowstate.ReviewDecision
}

// The role's command and optional harness remain one supervised execution choice.
type workflowReviewRole struct {
	Command []string
	Harness *AgentHarness
}

type workflowClaimedReview struct {
	Claim    *Claim
	Prepared workflowPreparedTask
}

type workflowReviewPreparation struct {
	Claim    *Claim
	Prepared workflowPreparedTask
	Binding  workflowstate.ReviewBinding
}
