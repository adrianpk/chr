package core

import (
	"errors"
	"fmt"
	"os"
	"slices"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func (s *WorkflowStore) workflowReviewPolicy(run workflowInvocationFile) (workflowstate.ReviewPolicy, error) {
	if err := requireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return workflowstate.ReviewPolicy{}, err
	}

	data, err := opsfile.Read(workspacecore.UserPath(s.layout.ProjectDir(), "config.yml"))
	if errors.Is(err, os.ErrNotExist) {
		data, err = nil, nil
	}

	if err != nil {
		return workflowstate.ReviewPolicy{}, err
	}

	policy, err := parseWorkflowReviewPolicy(data)
	if err != nil {
		return policy, err
	}

	if run.ReviewRestriction == workflowstate.ReviewHuman {
		policy.Mode, policy.Source = workflowstate.ReviewHuman, "invocation_restriction"
	}

	return policy, nil
}

func parseWorkflowReviewPolicy(data []byte) (workflowstate.ReviewPolicy, error) {
	config, err := parseWorkflowConfig(data)
	if err != nil {
		return workflowstate.ReviewPolicy{}, err
	}

	value := workflowConfigurationValue(config, "review.mode", string(config.Review.Mode))
	policy := workflowstate.ReviewPolicy{Mode: config.Review.Mode, Method: config.Work.MergeStrategy, Source: value.Source, Revision: config.workflowRevision}

	return policy, validateWorkflowReviewPolicy(policy)
}

func validateWorkflowReviewPolicy(policy workflowstate.ReviewPolicy) error {
	if !workflowstate.IntegrationCycleMode(policy.Mode).Valid() || !slices.Contains([]string{"merge", "rebase", "squash"}, policy.Method) || !slices.Contains([]string{"legacy_default", "version_default", "project", "invocation_restriction"}, policy.Source) || len(policy.Revision) != 64 {
		return fmt.Errorf("invalid current workflow review policy")
	}

	return nil
}
