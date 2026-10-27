package core

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func workflowUpgradeConfig(data []byte, prior workspaceconfig.WorkflowConfiguration, opts WorkflowUpgradeOptions) ([]byte, error) {
	review := opts.ReviewMode
	if review == "" {
		for _, value := range prior.Values {
			if value.Key == "review.mode" {
				review = workflowstate.ReviewMode(value.Value)
			}
		}
	}

	image := workspaceconfig.SetValue(data, "workflow", "version", "1")
	image = workspaceconfig.SetValue(image, "review", "mode", string(review))

	image = workspaceconfig.SetValue(image, "work", "state_authority", string(opts.StateAuthority))
	if opts.StateAuthority == workspaceconfig.WorkStateAuthorityTickets {
		image = workspaceconfig.SetValue(image, "work", "workspace_mode", "worktree")
	}

	_, err := workspaceconfig.Parse(image)

	return image, err
}
