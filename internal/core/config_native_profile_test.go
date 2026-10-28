package core

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeProfileConfig keeps publication opt-in independent of version and
// review mode, preserves adjacent keys and rejects ambiguous nested policy.
func TestNativeProfileConfig(t *testing.T) {
	cases := []struct {
		name, body, profile string
		valid               bool
	}{
		{"legacy", "", workspaceconfig.WorkflowPublicationDisabled, true},
		{"version_one", "workflow:\n  version: 1\n", workspaceconfig.WorkflowPublicationDisabled, true},
		{"explicit", "work:\n  integration:\n    publication_profile: git_atomic_dual_ref\n  merge_strategy: squash\nreview:\n  mode: human_review\n", workspaceconfig.WorkflowPublicationGitAtomicDualRef, true},
		{"disabled", "work:\n  integration:\n    publication_profile: disabled\n", workspaceconfig.WorkflowPublicationDisabled, true},
		{"unsupported", "work:\n  integration:\n    publication_profile: merge_api\n", "", false},
		{"duplicate", "work:\n  integration:\n    publication_profile: disabled\n  integration:\n    publication_profile: git_atomic_dual_ref\n", "", false},
		{"unknown_nested", "work:\n  integration:\n    rewrite_any_source: true\n", "", false},
		{"flat", "work:\n  publication_profile: git_atomic_dual_ref\n", "", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			config, err := workspaceconfig.Parse([]byte(test.body))
			if (err == nil) != test.valid {
				t.Fatalf("configuration validity: %v", err)
			}

			if test.valid && config.Work.Integration.PublicationProfile != test.profile {
				t.Fatal("publication policy default or explicit selection changed")
			}

			if test.name == "explicit" && (config.Work.MergeStrategy != "squash" || config.Review.Mode != workflowstate.ReviewHuman) {
				t.Fatal("nested publication config changed adjacent method or review policy")
			}
		})
	}
}
