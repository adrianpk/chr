package core

import (
	"encoding/json"
	"fmt"
)

type workflowNativeRepositoryPolicy struct {
	URL          string `json:"html_url"`
	Archived     bool   `json:"archived"`
	AllowMerge   *bool  `json:"allow_merge_commit"`
	ForgejoMerge *bool  `json:"allow_merge_commits"`
	AllowRebase  *bool  `json:"allow_rebase"`
	GitHubRebase *bool  `json:"allow_rebase_merge"`
	AllowSquash  *bool  `json:"allow_squash_merge"`
	Permissions  struct {
		Push *bool `json:"push"`
	} `json:"permissions"`
}

type workflowNativeBranchPolicy struct {
	Protected *bool `json:"protected"`
}

func workflowNativeRemotePolicy(responses []json.RawMessage, method, contract string) ([]byte, []WorkflowNativeUnsupported, string, error) {
	var (
		repository     workflowNativeRepositoryPolicy
		target, source workflowNativeBranchPolicy
	)

	if err := json.Unmarshal(responses[0], &repository); err != nil {
		return nil, nil, contract, err
	}

	if err := json.Unmarshal(responses[1], &target); err != nil {
		return nil, nil, contract, err
	}

	if err := json.Unmarshal(responses[2], &source); err != nil {
		return nil, nil, contract, err
	}

	projection, err := json.Marshal(struct {
		Repository workflowNativeRepositoryPolicy
		Target     workflowNativeBranchPolicy
		Source     workflowNativeBranchPolicy
		Rules      []json.RawMessage
	}{repository, target, source, responses[3:]})
	if err != nil {
		return nil, nil, contract, err
	}

	reasons := workflowNativeProtectionReasons(repository, target, source, method, contract)

	for _, response := range responses[3:] {
		var rules []json.RawMessage
		if err := json.Unmarshal(response, &rules); err != nil || rules == nil {
			return nil, nil, contract, fmt.Errorf("effective branch rules could not be completely observed")
		}

		if len(rules) != 0 {
			reasons = append(reasons, WorkflowNativeUnsupported{"native_rules_required", "Certify equivalence for the current enforced rules without disabling them."})
		}
	}

	return projection, reasons, contract, nil
}

func workflowNativeProtectionReasons(repository workflowNativeRepositoryPolicy, target, source workflowNativeBranchPolicy, method, contract string) []WorkflowNativeUnsupported {
	allowed := workflowNativeMethodAllowed(repository, method, contract)
	checks := []struct {
		valid     bool
		condition string
		decision  string
	}{
		{repository.Permissions.Push != nil && *repository.Permissions.Push && !repository.Archived, "publisher_access_denied", "Resolve an authorized publisher role for a writable repository."},
		{allowed, "method_disabled", "Select the configured method only after current repository policy permits it."},
		{target.Protected != nil && !*target.Protected && source.Protected != nil && !*source.Protected, "native_protections_required", "Certify an equivalent enforced policy; retain all existing protections."},
	}

	var reasons []WorkflowNativeUnsupported

	for _, check := range checks {
		if !check.valid {
			reasons = append(reasons, WorkflowNativeUnsupported{check.condition, check.decision})
		}
	}

	return reasons
}

func workflowNativeMethodAllowed(repository workflowNativeRepositoryPolicy, method, contract string) bool {
	allowed := repository.AllowMerge
	if method == "merge" && contract == "forgejo_same_pr_git_auto_merge/v1" {
		allowed = repository.ForgejoMerge
	}

	if method == "rebase" {
		allowed = repository.AllowRebase
		if allowed == nil {
			allowed = repository.GitHubRebase
		}
	}

	if method == "squash" {
		allowed = repository.AllowSquash
	}

	return allowed != nil && *allowed
}
