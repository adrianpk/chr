package workflow

import (
	"context"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationCapability reports the existing provider merge APIs'
// source-only guard. Native workflow publication must block these paths rather
// than reuse the legacy unguarded Merge operation or change the selected method.
func (adapter RemoteReviewAdapter) WorkflowIntegrationCapability(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowstate.PublicationCapability, error) {
	if err := ctx.Err(); err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	resolvedReview, err := adapter.resolve(pr.URL)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	reference := resolvedReview.reference

	if !slices.Contains([]string{"merge", "rebase", "squash"}, method) {
		return workflowstate.PublicationCapability{}, fmt.Errorf("unsupported workflow integration method")
	}

	mechanism := string(reference.Provider) + "_merge_api_source_guard_only"

	return workflowstate.PublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: mechanism, Revision: contenthash.Sum([]byte(mechanism + ":v1"))}, nil
}

// SubmitWorkflowIntegration refuses the current source-only merge APIs. It does
// not call legacy Merge or substitute another method for an exact prepared result.
func (adapter RemoteReviewAdapter) SubmitWorkflowIntegration(_ context.Context, _ workflowstate.Submission) (workflowstate.SubmissionResponse, error) {
	return workflowstate.SubmissionResponse{}, workflowstate.ErrAtomicGuard
}

// ObserveWorkflowSubmission cannot seal an operation through a provider path
// that lacks a stable guarded submission contract.
func (adapter RemoteReviewAdapter) ObserveWorkflowSubmission(_ context.Context, _ workflowstate.Submission) (workflowstate.SubmissionObservation, error) {
	return workflowstate.SubmissionObservation{}, ErrWorkflowObservationRequired
}
