package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (adapter RemoteReviewAdapter) readWorkflowNativeIdentity(ctx context.Context, pr workflowstate.PullRequestSnapshot) (string, error) {
	resolved, err := adapter.resolve(pr.URL)
	if err != nil {
		return "", err
	}

	client := adapter.workflowForgeClient(pr.Repository, resolved)

	return client.NativeIdentity(ctx)
}
