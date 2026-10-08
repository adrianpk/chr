package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func (s *WorkflowStore) workflowReviewPolicy(run workflowInvocationFile) (WorkflowReviewPolicy, error) {
	data, err := readWorkflowFile(filepath.Join(s.layout.Root(), "config.yml"))
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
	policy := WorkflowReviewPolicy{Mode: DefaultReviewWorkflowMode, Method: TicketMergeStrategyRebase, Source: "legacy_default", Revision: workflowHash(data)}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	section := ""
	for scanner.Scan() {
		key, value, next, ok := parseWorkspaceConfigLine(scanner.Text(), section)
		section = next

		if !ok {
			continue
		}

		if section == "review" && key == "mode" {
			policy.Mode, policy.Source = ReviewWorkflowMode(value), "project"
		}

		if section == "work" && key == "merge_strategy" {
			policy.Method = value
		}
	}

	if err := scanner.Err(); err != nil {
		return policy, fmt.Errorf("read current workflow policy: %w", err)
	}

	return policy, validateWorkflowReviewPolicy(policy)
}

func validateWorkflowReviewPolicy(policy WorkflowReviewPolicy) error {
	if !IntegrationCycleMode(policy.Mode).Valid() || !slices.Contains([]string{"merge", "rebase", "squash"}, policy.Method) || !slices.Contains([]string{"legacy_default", "project", "invocation_restriction"}, policy.Source) || len(policy.Revision) != 64 {
		return fmt.Errorf("invalid current workflow review policy")
	}

	return nil
}
