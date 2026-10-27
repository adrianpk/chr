package core

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestInitWorkflowProfiles checks the shipped templates and their missing-field
// defaults together, then repeats initialization with a different profile.
func TestInitWorkflowProfiles(t *testing.T) {
	for _, profile := range []string{"", "standard", "minimal", "personal"} {
		name := profile
		if name == "" {
			name = "default"
		}

		t.Run(name, func(t *testing.T) {
			checkWorkflowProfile(t, profile)
		})
	}
}

func checkWorkflowProfile(t *testing.T, profile string) {
	t.Helper()

	project := t.TempDir()
	workspace := workflowProfileInitializer(t)
	options := InitOptions{ProjectDir: project, Profile: profile, DryRun: true}

	result, err := workspace.Init(t.Context(), options)
	if err != nil || !contains(result.Created, workspacecore.UserRelativePath("config.yml")) {
		t.Fatal("dry run lost planned configuration", result, err)
	}

	assertNotExists(t, project, workspacecore.DirectoryName)

	options.DryRun = false
	if _, err := workspace.Init(t.Context(), options); err != nil {
		t.Fatal(err)
	}

	checkInitializedWorkflow(t, project)

	before, err := os.ReadFile(workspacecore.UserPath(project, "config.yml"))
	if err != nil {
		t.Fatal(err)
	}

	options.Profile = "minimal"
	if _, err := workspace.Init(t.Context(), options); err != nil {
		t.Fatal(err)
	}

	checkWorkflowConfigDefaultBytes(t, project, string(before))
	checkInitializedWorkflow(t, project)
}

func workflowProfileInitializer(t *testing.T) Workspace {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve bundled workspace profiles")
	}

	return NewWorkspace(os.DirFS(filepath.Join(filepath.Dir(file), "..", "..", "cmd", "chr")))
}

func checkInitializedWorkflow(t *testing.T, project string) {
	t.Helper()

	config, err := workspaceconfig.Load(project)
	if err != nil || config.Work.MaxActiveWorkers != 1 || config.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationDisabled {
		t.Fatal("initialized workflow differs from approved defaults", config, err)
	}

	checkWorkflowConfigDefaultValues(t, project, workflowConfigDefaultCase{"initialized", "", "version_default", 1, workflowstate.ReviewTrustedAutoMerge, workspaceconfig.WorkStateAuthorityTickets, workspacecore.ExecutionWorktree})

	effective, err := workspaceconfig.ResolveWorkflowConfiguration(project, "")
	if err != nil || len(effective.Values) != 15 {
		t.Fatal("initialized configuration is incomplete", effective, err)
	}

	for _, value := range effective.Values {
		want := "version_default"
		if value.Key == "workflow.version" {
			want = "project"
		}

		if value.Source != want {
			t.Fatal("template overrides a version default", value)
		}
	}
}

// TestInitWorkflowPreservesPolicy proves that absent, explicit and invalid
// existing policy survive every profile and both dry-run and actual init.
func TestInitWorkflowPreservesPolicy(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"absent file", ""},
		{"absent marker", "version: 1\nnotifications:\n  telegram_enabled: false\n"},
		{"legacy explicit", "review:\n  mode: agentic_review\nwork:\n  state_authority: tracker_compat\n  workspace_mode: project\n"},
		{"versioned explicit", "workflow:\n  version: 1\nreview:\n  mode: human_review\nwork:\n  max_active_workers: 8\n"},
		{"invalid policy", "workflow:\n  version: 99\nreview:\n  mode: human_review\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			checkExistingWorkflowInitialization(t, test.body)
		})
	}
}

func checkExistingWorkflowInitialization(t *testing.T, body string) {
	t.Helper()

	project := t.TempDir()
	if err := os.MkdirAll(workspacecore.UserPath(project), 0o755); err != nil {
		t.Fatal(err)
	}

	if body != "" {
		writeWorkspaceConfig(t, project, body)
	}

	before, beforeErr := workspaceconfig.ResolveWorkflowConfiguration(project, "")
	workspace := workflowProfileInitializer(t)

	for _, profile := range []string{"standard", "minimal", "personal"} {
		for _, dryRun := range []bool{true, false} {
			checkReinitializedWorkflow(t, workspace, project, profile, dryRun, body)
		}
	}

	after, afterErr := workspaceconfig.ResolveWorkflowConfiguration(project, "")
	if (beforeErr == nil) != (afterErr == nil) || !reflect.DeepEqual(before, after) {
		t.Fatal("reinitialization changed effective policy", before, after, beforeErr, afterErr)
	}
}

func checkReinitializedWorkflow(t *testing.T, workspace Workspace, project, profile string, dryRun bool, body string) {
	t.Helper()

	result, err := workspace.Init(context.Background(), InitOptions{ProjectDir: project, Profile: profile, DryRun: dryRun})
	if err != nil || contains(result.Created, workspacecore.UserRelativePath("config.yml")) {
		t.Fatal("reinitialization attempted to create policy", result, err)
	}

	if body == "" {
		assertNotExists(t, project, workspacecore.UserRelativePath("config.yml"))

		if !contains(result.Preserved, workspacecore.UserRelativePath("config.yml")) {
			t.Fatal("missing policy preservation was not reported", result)
		}
	} else {
		checkWorkflowConfigDefaultBytes(t, project, body)
	}
}
