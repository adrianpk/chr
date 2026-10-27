package core

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativeProfileConfig keeps publication opt-in independent of version and
// review mode, preserves adjacent keys and rejects ambiguous nested policy.
func TestNativeProfileConfig(t *testing.T) {
	cases := []struct {
		name, body, profile string
		valid               bool
	}{
		{"legacy", "", WorkflowPublicationDisabled, true},
		{"version_one", "workflow:\n  version: 1\n", WorkflowPublicationDisabled, true},
		{"explicit", "work:\n  integration:\n    publication_profile: git_atomic_dual_ref\n  merge_strategy: squash\nreview:\n  mode: human_review\n", WorkflowPublicationGitAtomicDualRef, true},
		{"disabled", "work:\n  integration:\n    publication_profile: disabled\n", WorkflowPublicationDisabled, true},
		{"unsupported", "work:\n  integration:\n    publication_profile: merge_api\n", "", false},
		{"duplicate", "work:\n  integration:\n    publication_profile: disabled\n  integration:\n    publication_profile: git_atomic_dual_ref\n", "", false},
		{"unknown_nested", "work:\n  integration:\n    rewrite_any_source: true\n", "", false},
		{"flat", "work:\n  publication_profile: git_atomic_dual_ref\n", "", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			config, err := parseWorkflowConfig([]byte(test.body))
			if (err == nil) != test.valid {
				t.Fatalf("configuration validity: %v", err)
			}

			if test.valid && config.Work.Integration.PublicationProfile != test.profile {
				t.Fatal("publication policy default or explicit selection changed")
			}

			if test.name == "explicit" && (config.Work.MergeStrategy != "squash" || config.Review.Mode != ReviewWorkflowModeHumanReview) {
				t.Fatal("nested publication config changed adjacent method or review policy")
			}
		})
	}
}

// TestNativeProfileUpgrade verifies a legacy upgrade never activates source
// rewriting or replaces the retained review policy.
func TestNativeProfileUpgrade(t *testing.T) {
	data := []byte("review:\n  mode: human_review\nwork:\n  merge_strategy: squash\n")

	config, err := parseWorkflowConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	image, err := workflowUpgradeConfig(data, effectiveWorkflowConfiguration(config, ""), WorkflowUpgradeOptions{StateAuthority: WorkStateAuthorityTickets})
	if err != nil {
		t.Fatal(err)
	}

	upgraded, err := parseWorkflowConfig(image)
	if err != nil || upgraded.Work.Integration.PublicationProfile != WorkflowPublicationDisabled || upgraded.Review.Mode != ReviewWorkflowModeHumanReview || upgraded.Work.MergeStrategy != "squash" {
		t.Fatal("upgrade activated publication or changed existing policy")
	}
}

// TestNativeLegacySnapshot keeps pre-profile upgrade snapshots valid without
// granting new publication authority or accepting incomplete policy records.
func TestNativeLegacySnapshot(t *testing.T) {
	data := []byte("review:\n  mode: human_review\n")

	config, err := parseWorkflowConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	prior := effectiveWorkflowConfiguration(config, "")

	prior.Values = withoutWorkflowPublicationProfile(prior.Values)
	if !validWorkflowUpgradeConfiguration(prior) {
		t.Fatal("legacy upgrade snapshot lost compatibility")
	}

	if err := checkWorkflowUpgradePolicy(data, prior); err != nil {
		t.Fatal("legacy disabled profile changed retained policy", err)
	}

	enabled := []byte("review:\n  mode: human_review\nwork:\n  integration:\n    publication_profile: git_atomic_dual_ref\n")
	prior.Revision = contenthash.Sum(enabled)

	if err := checkWorkflowUpgradePolicy(enabled, prior); err == nil {
		t.Fatal("legacy snapshot silently admitted enabled publication")
	}

	prior.Values[0] = prior.Values[1]
	if validWorkflowUpgradeConfiguration(prior) {
		t.Fatal("duplicate key hid an incomplete retained policy")
	}
}
