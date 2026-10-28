package workflow

import (
	"errors"
	"os"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) workflowReviewPolicy(run workflowInvocationFile) (workflowstate.ReviewPolicy, error) {
	if err := workspaceconfig.RequireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
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
	config, err := workspaceconfig.Parse(data)
	if err != nil {
		return workflowstate.ReviewPolicy{}, err
	}

	value := workspaceconfig.WorkflowValue(config, "review.mode", string(config.Review.Mode))
	policy := workflowstate.ReviewPolicy{Mode: config.Review.Mode, Method: config.Work.MergeStrategy, Source: value.Source, Revision: config.Revision()}

	return policy, workflowreview.CheckPolicy(policy)
}
