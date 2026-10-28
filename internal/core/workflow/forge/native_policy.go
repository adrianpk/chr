package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeRepositoryPolicy struct {
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

type nativeBranchPolicy struct {
	Protected *bool `json:"protected"`
}

// NativePolicyObservation retains sanitized effective policy bytes and every denial.
type NativePolicyObservation struct {
	Data        []byte
	Unsupported []workflowstate.NativeUnsupported
	Contract    string
}

func nativeRemotePolicy(responses []json.RawMessage, method, contract string) (NativePolicyObservation, error) {
	var (
		repository     nativeRepositoryPolicy
		target, source nativeBranchPolicy
	)

	if err := json.Unmarshal(responses[0], &repository); err != nil {
		return NativePolicyObservation{Contract: contract}, err
	}

	if err := json.Unmarshal(responses[1], &target); err != nil {
		return NativePolicyObservation{Contract: contract}, err
	}

	if err := json.Unmarshal(responses[2], &source); err != nil {
		return NativePolicyObservation{Contract: contract}, err
	}

	projection, err := json.Marshal(struct {
		Repository nativeRepositoryPolicy
		Target     nativeBranchPolicy
		Source     nativeBranchPolicy
		Rules      []json.RawMessage
	}{repository, target, source, responses[3:]})
	if err != nil {
		return NativePolicyObservation{Contract: contract}, err
	}

	reasons := nativeProtectionReasons(repository, target, source, method, contract)

	for _, response := range responses[3:] {
		var rules []json.RawMessage
		if err := json.Unmarshal(response, &rules); err != nil || rules == nil {
			return NativePolicyObservation{Contract: contract}, fmt.Errorf("effective branch rules could not be completely observed")
		}

		if len(rules) != 0 {
			reasons = append(reasons, workflowstate.NativeUnsupported{Condition: "native_rules_required", Decision: "Certify equivalence for the current enforced rules without disabling them."})
		}
	}

	return NativePolicyObservation{Data: projection, Unsupported: reasons, Contract: contract}, nil
}

func nativeProtectionReasons(repository nativeRepositoryPolicy, target, source nativeBranchPolicy, method, contract string) []workflowstate.NativeUnsupported {
	allowed := nativeMethodAllowed(repository, method, contract)
	checks := []struct {
		valid     bool
		condition string
		decision  string
	}{
		{repository.Permissions.Push != nil && *repository.Permissions.Push && !repository.Archived, "publisher_access_denied", "Resolve an authorized publisher role for a writable repository."},
		{allowed, "method_disabled", "Select the configured method only after current repository policy permits it."},
		{target.Protected != nil && !*target.Protected && source.Protected != nil && !*source.Protected, "native_protections_required", "Certify an equivalent enforced policy; retain all existing protections."},
	}

	var reasons []workflowstate.NativeUnsupported

	for _, check := range checks {
		if !check.valid {
			reasons = append(reasons, workflowstate.NativeUnsupported{Condition: check.condition, Decision: check.decision})
		}
	}

	return reasons
}

func nativeMethodAllowed(repository nativeRepositoryPolicy, method, contract string) bool {
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

// NativePolicy reads all required repository and branch rules within existing response bounds.
func (client Client) NativePolicy(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (NativePolicyObservation, error) {
	paths := []string{client.Endpoint.RepositoryPath(), client.Endpoint.RepositoryPath() + "/branches/" + url.PathEscape(pr.BaseBranch), client.Endpoint.RepositoryPath() + "/branches/" + url.PathEscape(pr.HeadBranch)}

	contract := "forgejo_same_pr_git_auto_merge/v1"
	if client.Endpoint.Provider == string(GitHub) {
		contract = "github_same_pr_git_merge/v1"

		paths = append(paths, client.Endpoint.RepositoryPath()+"/rules/branches/"+url.PathEscape(pr.BaseBranch), client.Endpoint.RepositoryPath()+"/rules/branches/"+url.PathEscape(pr.HeadBranch))
	}

	var responses []json.RawMessage

	for _, path := range paths {
		data, err := client.Request(ctx, Request{Method: "GET", Path: path})
		if err != nil || len(data) > workflowstate.FileLimit || !json.Valid(data) {
			return NativePolicyObservation{Contract: contract}, fmt.Errorf("native publication policy could not be completely observed")
		}

		responses = append(responses, data)
	}

	var repository nativeRepositoryPolicy
	if err := json.Unmarshal(responses[0], &repository); err != nil || strings.TrimSuffix(repository.URL, "/") != pr.Repository {
		return NativePolicyObservation{Contract: contract}, workflowstate.ErrConflict
	}

	return nativeRemotePolicy(responses, method, contract)
}
