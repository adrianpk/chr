package core

import (
	"context"
	"fmt"
	"slices"
)

// WorkflowIntegrationCapability reports the existing provider merge APIs'
// source-only guard. Native workflow publication must block these paths rather
// than reuse the legacy unguarded Merge operation or change the selected method.
func (adapter RemoteReviewAdapter) WorkflowIntegrationCapability(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowPublicationCapability, error) {
	if err := ctx.Err(); err != nil {
		return WorkflowPublicationCapability{}, err
	}

	reference, _, err := adapter.resolve(pr.URL)
	if err != nil {
		return WorkflowPublicationCapability{}, err
	}

	if !slices.Contains([]string{"merge", "rebase", "squash"}, method) {
		return WorkflowPublicationCapability{}, fmt.Errorf("unsupported workflow integration method")
	}

	mechanism := string(reference.Provider) + "_merge_api_source_guard_only"

	return WorkflowPublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: mechanism, Revision: workflowHash([]byte(mechanism + ":v1"))}, nil
}

// SubmitWorkflowIntegration refuses the current source-only merge APIs. It does
// not call legacy Merge or substitute another method for an exact prepared result.
func (adapter RemoteReviewAdapter) SubmitWorkflowIntegration(_ context.Context, _ WorkflowSubmission) (WorkflowSubmissionResponse, error) {
	return WorkflowSubmissionResponse{}, ErrWorkflowAtomicGuard
}

// ObserveWorkflowSubmission cannot seal an operation through a provider path
// that lacks a stable guarded submission contract.
func (adapter RemoteReviewAdapter) ObserveWorkflowSubmission(_ context.Context, _ WorkflowSubmission) (WorkflowSubmissionObservation, error) {
	return WorkflowSubmissionObservation{}, ErrWorkflowObservationRequired
}
