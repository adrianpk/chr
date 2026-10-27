package core

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativePreparedSnapshot accepts existing immutable upgrade records with
// either prior option set. Omitted fields cannot conceal a changed policy.
func TestNativePreparedSnapshot(t *testing.T) {
	for _, keys := range [][]string{{}, {"work.integration.max_prepared_per_target"}, {"work.integration.publication_profile"}, {"work.integration.max_prepared_per_target", "work.integration.publication_profile"}} {
		t.Run(snapshotCapacityName(keys), func(t *testing.T) { nativePreparedSnapshot(t, keys) })
	}
}

func snapshotCapacityName(keys []string) string {
	if len(keys) == 0 {
		return "current"
	}

	if len(keys) == 2 {
		return "original"
	}

	return keys[0]
}

func nativePreparedSnapshot(t *testing.T, keys []string) {
	t.Helper()

	data := []byte("review:\n  mode: human_review\n")

	config, err := parseWorkflowConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	prior := effectiveWorkflowConfiguration(config, "")
	for _, key := range keys {
		prior.Values = withoutWorkflowConfigurationKey(prior.Values, key)
	}

	if !validWorkflowUpgradeConfiguration(prior) {
		t.Fatal("retained configuration shape lost compatibility")
	}

	if err := checkWorkflowUpgradePolicy(data, prior); err != nil {
		t.Fatal("default-only projection changed retained policy", err)
	}

	changed := []byte("review:\n  mode: human_review\nwork:\n  integration:\n    max_prepared_per_target: 3\n")

	prior.Revision = contenthash.Sum(changed)
	if err := checkWorkflowUpgradePolicy(changed, prior); err == nil {
		t.Fatal("snapshot accepted a changed or omitted preparation limit")
	}

	prior.Values[0].Key = "unknown.option"
	if validWorkflowUpgradeConfiguration(prior) {
		t.Fatal("unknown key replaced mandatory snapshot evidence")
	}
}
