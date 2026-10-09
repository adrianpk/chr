package core

import (
	"errors"
	"fmt"
	"os"
	"slices"
)

func (s *WorkflowStore) workflowReviewPolicy(run workflowInvocationFile) (WorkflowReviewPolicy, error) {
	if err := requireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return WorkflowReviewPolicy{}, err
	}

	data, err := readWorkflowFile(UserPath(s.layout.ProjectDir(), "config.yml"))
	if errors.Is(err, os.ErrNotExist) {
		data, err = nil, nil
	}

	if err != nil {
		return WorkflowReviewPolicy{}, err
	}

	policy, err := parseWorkflowReviewPolicy(data)
	if err != nil {
		return policy, err
	}

	if run.ReviewRestriction == ReviewWorkflowModeHumanReview {
		policy.Mode, policy.Source = ReviewWorkflowModeHumanReview, "invocation_restriction"
	}

	return policy, nil
}

func parseWorkflowReviewPolicy(data []byte) (WorkflowReviewPolicy, error) {
	config, err := parseWorkflowConfig(data)
	if err != nil {
		return WorkflowReviewPolicy{}, err
	}

	value := workflowConfigurationValue(config, "review.mode", string(config.Review.Mode))
	policy := WorkflowReviewPolicy{Mode: config.Review.Mode, Method: config.Work.MergeStrategy, Source: value.Source, Revision: config.workflowRevision}

	return policy, validateWorkflowReviewPolicy(policy)
}

func validateWorkflowReviewPolicy(policy WorkflowReviewPolicy) error {
	if !IntegrationCycleMode(policy.Mode).Valid() || !slices.Contains([]string{"merge", "rebase", "squash"}, policy.Method) || !slices.Contains([]string{"legacy_default", "version_default", "project", "invocation_restriction"}, policy.Source) || len(policy.Revision) != 64 {
		return fmt.Errorf("invalid current workflow review policy")
	}

	return nil
}
