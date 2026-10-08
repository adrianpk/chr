package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowIntegrationHistory rejects a correct final tree with invalid
// merge parents, squash parents or an altered intermediate rebased commit.
func TestWorkflowIntegrationHistory(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, method)
			forgeWorkflowIntegrationHistory(t, &fixture, method)

			if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, method, fixture.provider, fixture.git); err == nil {
				t.Fatal("matching final tree concealed invalid method-specific history")
			}
		})
	}
}

// TestWorkflowIntegrationHostConfig retains trusted host credential helpers for
// canonical HTTPS fetches without allowing interactive credential prompts.
func TestWorkflowIntegrationHostConfig(t *testing.T) {
	dir := t.TempDir()

	config := filepath.Join(dir, "host.gitconfig")
	if err := os.WriteFile(config, []byte("[credential]\n\thelper = host-fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GIT_CONFIG_GLOBAL", config)

	reader := workflowIntegrationReader{dir: dir}

	value, err := reader.output(t.Context(), "config", "--get", "credential.helper")
	if err != nil || value != "host-fixture" {
		t.Fatalf("trusted host credential configuration was discarded: %s: %v", value, err)
	}
}

func forgeWorkflowIntegrationHistory(t *testing.T, fixture *workflowIntegrationFixture, method string) {
	t.Helper()

	tree := workflowIntegrationCommand(t, fixture.actor, "rev-parse", fixture.result+"^{tree}")

	var result string

	switch method {
	case "merge":
		result = workflowIntegrationCommand(t, fixture.actor, "commit-tree", tree, "-p", fixture.base, "-p", fixture.run.Lane.BaseCommit, "-m", "Wrong source parent")
	case "squash":
		result = workflowIntegrationCommand(t, fixture.actor, "commit-tree", tree, "-p", fixture.base, "-p", fixture.run.Lane.HeadCommit, "-m", "Not one combined commit")
	case "rebase":
		baseTree := workflowIntegrationCommand(t, fixture.actor, "rev-parse", fixture.base+"^{tree}")
		first := workflowIntegrationCommand(t, fixture.actor, "commit-tree", baseTree, "-p", fixture.base, "-m", "Dropped first change")
		secondTree := workflowIntegrationCommand(t, fixture.actor, "rev-parse", fixture.result+"~1^{tree}")
		second := workflowIntegrationCommand(t, fixture.actor, "commit-tree", secondTree, "-p", first, "-m", "Recovered changes")
		result = workflowIntegrationCommand(t, fixture.actor, "commit-tree", tree, "-p", second, "-m", "Matching final tree")
	}

	workflowIntegrationCommand(t, fixture.actor, "push", "--force", "origin", result+":refs/heads/dev")

	fixture.provider.pr.ResultCommit = result
}

// TestWorkflowIntegrationCancelled respects an already cancelled host operation
// without accepting provider metadata or leaving a verified delivery reference.
func TestWorkflowIntegrationCancelled(t *testing.T) {
	fixture := workflowDeliveryFixture(t, "merge")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := fixture.store.VerifyWorkflowSliceIntegration(ctx, fixture.run.ID, "merge", fixture.provider, fixture.git); err == nil {
		t.Fatal("cancelled host operation accepted delivery")
	}

	if workflowTestRead(t, fixture.store, fixture.run.ID).DeliveryProofRef != "" {
		t.Fatal("cancelled verification persisted a proof")
	}
}
