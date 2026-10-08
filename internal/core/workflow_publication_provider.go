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
