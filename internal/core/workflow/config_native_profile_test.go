package workflow

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeProfileUpgrade verifies a legacy upgrade never activates source
// rewriting or replaces the retained review policy.
func TestNativeProfileUpgrade(t *testing.T) {
	data := []byte("review:\n  mode: human_review\nwork:\n  merge_strategy: squash\n")

	config, err := workspaceconfig.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	image, err := workflowUpgradeConfig(data, workspaceconfig.EffectiveWorkflowConfiguration(config, ""), UpgradeOptions{StateAuthority: workspaceconfig.WorkStateAuthorityTickets})
	if err != nil {
		t.Fatal(err)
	}

	upgraded, err := workspaceconfig.Parse(image)
	if err != nil || upgraded.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationDisabled || upgraded.Review.Mode != workflowstate.ReviewHuman || upgraded.Work.MergeStrategy != "squash" {
		t.Fatal("upgrade activated publication or changed existing policy")
	}
}

// TestNativeLegacySnapshot keeps pre-profile upgrade snapshots valid without
// granting new publication authority or accepting incomplete policy records.
func TestNativeLegacySnapshot(t *testing.T) {
	data := []byte("review:\n  mode: human_review\n")

	config, err := workspaceconfig.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	prior := workspaceconfig.EffectiveWorkflowConfiguration(config, "")

	prior.Values = withoutWorkflowConfigurationKey(prior.Values, "work.integration.publication_profile")
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
