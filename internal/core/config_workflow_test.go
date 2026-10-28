package core

import (
	"os"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestWorkflowConfigDefaults keeps legacy absence distinct from version-one
// defaults, independent of key order, and reports explicit policy sources.
func TestWorkflowConfigDefaults(t *testing.T) {
	cases := []workflowConfigDefaultCase{
		{"missing", "", "legacy_default", 0, workflowstate.ReviewHuman, workspaceconfig.WorkStateAuthorityTrackerCompat, workspacecore.ExecutionProject},
		{"old root version", "version: 1\n", "legacy_default", 0, workflowstate.ReviewHuman, workspaceconfig.WorkStateAuthorityTrackerCompat, workspacecore.ExecutionProject},
		{"new version", "workflow:\n  version: 1\n", "version_default", 1, workflowstate.ReviewTrustedAutoMerge, workspaceconfig.WorkStateAuthorityTickets, workspacecore.ExecutionWorktree},
		{"explicit human", "workflow:\n  version: 1\nreview:\n  mode: human_review\n", "project", 1, workflowstate.ReviewHuman, workspaceconfig.WorkStateAuthorityTickets, workspacecore.ExecutionWorktree},
		{"marker last", "review:\n  mode: agentic_review\nworkflow:\n  version: 1\n", "project", 1, workflowstate.ReviewAgentic, workspaceconfig.WorkStateAuthorityTickets, workspacecore.ExecutionWorktree},
		{"legacy explicit", "review:\n  mode: trusted_auto_merge\n", "project", 0, workflowstate.ReviewTrustedAutoMerge, workspaceconfig.WorkStateAuthorityTrackerCompat, workspacecore.ExecutionProject},
		{"explicit serial", "workflow:\n  version: 1\nwork:\n  state_authority: tracker_compat\n  workspace_mode: project\n", "version_default", 1, workflowstate.ReviewTrustedAutoMerge, workspaceconfig.WorkStateAuthorityTrackerCompat, workspacecore.ExecutionProject},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			project := t.TempDir()
			if test.body != "" {
				writeWorkspaceConfig(t, project, test.body)
			}

			checkWorkflowConfigDefaultValues(t, project, test)
			checkWorkflowConfigDefaultSource(t, project, test)
			checkWorkflowConfigDefaultBytes(t, project, test.body)
		})
	}
}

func workflowConfigTestValue(t *testing.T, config workspaceconfig.WorkflowConfiguration, key string) workspaceconfig.WorkflowConfigurationValue {
	t.Helper()

	for _, value := range config.Values {
		if value.Key == key {
			return value
		}
	}

	t.Fatal("missing effective value", key)

	return workspaceconfig.WorkflowConfigurationValue{}
}

type workflowConfigDefaultCase struct {
	name, body, source string
	version            int
	mode               workflowstate.ReviewMode
	authority          workspaceconfig.WorkStateAuthority
	workspace          workspacecore.ExecutionMode
}

func checkWorkflowConfigDefaultValues(t *testing.T, project string, test workflowConfigDefaultCase) {
	t.Helper()

	config, err := workspaceconfig.Load(project)
	if err != nil || config.Workflow.Version != test.version || config.Review.Mode != test.mode || workspaceconfig.WorkAuthority(config.Work) != test.authority || config.Work.Workspace.Mode != test.workspace {
		t.Fatalf("defaults: %+v, %v", config, err)
	}
}

func checkWorkflowConfigDefaultSource(t *testing.T, project string, test workflowConfigDefaultCase) {
	t.Helper()

	resolved, err := workspaceconfig.ResolveWorkflowConfiguration(project, "")
	if err != nil || len(resolved.Values) != 15 || len(resolved.Revision) != 64 {
		t.Fatal("effective configuration is incomplete", resolved, err)
	}

	prepared := workflowConfigTestValue(t, resolved, "work.integration.max_prepared_per_target")
	if prepared.Value != "2" {
		t.Fatal("prepared default differs", prepared)
	}

	mode := workflowConfigTestValue(t, resolved, "review.mode")
	if mode.Value != string(test.mode) || mode.Source != test.source {
		t.Fatal("effective policy source differs", mode)
	}
}

func checkWorkflowConfigDefaultBytes(t *testing.T, project, body string) {
	t.Helper()

	if body == "" {
		if _, err := os.Stat(workspacecore.UserPath(project, "config.yml")); !os.IsNotExist(err) {
			t.Fatal("observing absent config initialized a project", err)
		}

		return
	}

	if data, err := os.ReadFile(workspacecore.UserPath(project, "config.yml")); err != nil || string(data) != body {
		t.Fatal("configuration inspection changed project bytes", err)
	}
}
