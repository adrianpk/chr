package core

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowConfigDefaults keeps legacy absence distinct from version-one
// defaults, independent of key order, and reports explicit policy sources.
func TestWorkflowConfigDefaults(t *testing.T) {
	cases := []workflowConfigDefaultCase{
		{"missing", "", "legacy_default", 0, ReviewWorkflowModeHumanReview, WorkStateAuthorityTrackerCompat, WorkWorkspaceModeProject},
		{"old root version", "version: 1\n", "legacy_default", 0, ReviewWorkflowModeHumanReview, WorkStateAuthorityTrackerCompat, WorkWorkspaceModeProject},
		{"new version", "workflow:\n  version: 1\n", "version_default", 1, ReviewWorkflowModeTrustedAutoMerge, WorkStateAuthorityTickets, WorkWorkspaceModeWorktree},
		{"explicit human", "workflow:\n  version: 1\nreview:\n  mode: human_review\n", "project", 1, ReviewWorkflowModeHumanReview, WorkStateAuthorityTickets, WorkWorkspaceModeWorktree},
		{"marker last", "review:\n  mode: agentic_review\nworkflow:\n  version: 1\n", "project", 1, ReviewWorkflowModeAgenticReview, WorkStateAuthorityTickets, WorkWorkspaceModeWorktree},
		{"legacy explicit", "review:\n  mode: trusted_auto_merge\n", "project", 0, ReviewWorkflowModeTrustedAutoMerge, WorkStateAuthorityTrackerCompat, WorkWorkspaceModeProject},
		{"explicit serial", "workflow:\n  version: 1\nwork:\n  state_authority: tracker_compat\n  workspace_mode: project\n", "version_default", 1, ReviewWorkflowModeTrustedAutoMerge, WorkStateAuthorityTrackerCompat, WorkWorkspaceModeProject},
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

// Review fixtures edit only their selected policy fields in the shared project
// file, preserving configured queue capacity and execution budgets.
func writeWorkflowTestConfigValue(t *testing.T, project, root, key, value string) {
	t.Helper()

	data, err := os.ReadFile(UserPath(project, "config.yml"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}

	writeWorkspaceConfig(t, project, workflowTestConfigValueImage(string(data), root, key, value))
}

func workflowTestConfigValueImage(body, root, key, value string) string {
	lines := strings.Split(body, "\n")

	heading, section := -1, ""
	for i, line := range lines {
		parsedLine := parseWorkspaceConfigLine(line, section)
		parsed, next, scalar := parsedLine.Key, parsedLine.Section, parsedLine.Parsed

		section = next
		if scalar && section == root && parsed == key {
			lines[i] = "  " + key + ": " + value

			return strings.Join(lines, "\n")
		}

		if line == root+":" {
			heading = i
		}
	}

	line := "  " + key + ": " + value
	if heading >= 0 {
		return strings.Join(slices.Insert(lines, heading+1, line), "\n")
	}

	return body + root + ":\n" + line + "\n"
}

// TestWorkflowConfigBounds checks the public loader and publication resolver at
// both exact edges; malformed or out-of-range values cannot use fallback policy.
func TestWorkflowConfigBounds(t *testing.T) {
	fields := []struct {
		key       string
		low, high string
		bad       []string
	}{
		{"workflow.version", "1", "1", []string{"", "0", "2", "1.0"}},
		{"work.max_active_workers", "1", "8", []string{"0", "9", "many"}},
		{"work.ready_queue_capacity", "1", "1024", []string{"0", "1025", "many"}},
		{"work.integration.max_prepared_per_target", "1", "8", []string{"0", "9", "many"}},
		{"work.worker_timeout_seconds", "1", "86400", []string{"0", "86401", "-1"}},
		{"work.validation_timeout_seconds", "1", "86400", []string{"0", "86401", "1.5"}},
		{"work.checkpoint_step_limit", "1", "128", []string{"0", "129", "many"}},
		{"work.max_worker_attempts", "1", "128", []string{"0", "129", "many"}},
		{"work.max_integration_attempts", "1", "10", []string{"0", "11", "many"}},
	}
	for _, field := range fields {
		t.Run(field.key, func(t *testing.T) {
			for _, value := range append([]string{field.low, field.high}, field.bad...) {
				t.Run(value, func(t *testing.T) {
					valid := value == field.low || value == field.high
					section, key, _ := strings.Cut(field.key, ".")
					body := section + ":\n  " + key + ": " + value + "\n"
					project := t.TempDir()
					writeWorkspaceConfig(t, project, body)

					_, loadErr := LoadWorkspaceConfig(project)

					_, policyErr := parseWorkflowReviewPolicy([]byte(body))
					if (loadErr == nil) != valid || (policyErr == nil) != valid {
						t.Fatal("configuration boundary differs", loadErr, policyErr)
					}
				})
			}
		})
	}
}

// TestWorkflowConfigErrors rejects ambiguous authority, duplicated input,
// unsupported syntax and values, and oversized files before execution.
func TestWorkflowConfigErrors(t *testing.T) {
	cases := []struct{ name, body string }{
		{"duplicate key", "work:\n  max_active_workers: 1\n  max_active_workers: 8\n"},
		{"duplicate section", "workflow:\n  version: 1\nworkflow:\n  version: 1\n"},
		{"unknown mode", "review:\n  mode: automatic\n"},
		{"empty mode", "review:\n  mode:\n"},
		{"unknown authority", "work:\n  state_authority: both\n"},
		{"method fallback", "work:\n  merge_strategy: fast\n"},
		{"bool fallback", "work:\n  include_later_tickets: yes\n"},
		{"unknown key", "workflow:\n  version: 1\nreview:\n  mdoe: human_review\n"},
		{"scalar section", "workflow: 1\n"},
		{"nested version", "workflow:\n  config:\n    version: 1\n"},
		{"tabbed review", "workflow:\n  version: 1\nreview:\n\tmode: human_review\n"},
		{"unindented review", "workflow:\n  version: 1\nreview:\nmode: human_review\n"},
		{"bad indentation", "workflow:\n version: 1\n"},
		{"native project workspace", "workflow:\n  version: 1\nwork:\n  workspace_mode: project\n"},
		{"parallel legacy project", "work:\n  workspace_mode: project\n  max_active_workers: 2\n"},
		{"oversized", strings.Repeat("#", workflowstate.FileLimit+1)},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			project := t.TempDir()
			writeWorkspaceConfig(t, project, test.body)

			if _, err := LoadWorkspaceConfig(project); err == nil {
				t.Fatal("invalid configuration used a default")
			}

			if _, err := parseWorkflowReviewPolicy([]byte(test.body)); err == nil {
				t.Fatal("publication accepted invalid configuration")
			}
		})
	}
}

func workflowConfigTestValue(t *testing.T, config WorkflowConfiguration, key string) WorkflowConfigurationValue {
	t.Helper()

	for _, value := range config.Values {
		if value.Key == key {
			return value
		}
	}

	t.Fatal("missing effective value", key)

	return WorkflowConfigurationValue{}
}

type workflowConfigDefaultCase struct {
	name, body, source string
	version            int
	mode               ReviewWorkflowMode
	authority          WorkStateAuthority
	workspace          WorkWorkspaceMode
}

func checkWorkflowConfigDefaultValues(t *testing.T, project string, test workflowConfigDefaultCase) {
	t.Helper()

	config, err := LoadWorkspaceConfig(project)
	if err != nil || config.Workflow.Version != test.version || config.Review.Mode != test.mode || WorkAuthority(config.Work) != test.authority || config.Work.Workspace.Mode != test.workspace {
		t.Fatalf("defaults: %+v, %v", config, err)
	}
}

func checkWorkflowConfigDefaultSource(t *testing.T, project string, test workflowConfigDefaultCase) {
	t.Helper()

	resolved, err := ResolveWorkflowConfiguration(project, "")
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
		if _, err := os.Stat(UserPath(project, "config.yml")); !os.IsNotExist(err) {
			t.Fatal("observing absent config initialized a project", err)
		}

		return
	}

	if data, err := os.ReadFile(UserPath(project, "config.yml")); err != nil || string(data) != body {
		t.Fatal("configuration inspection changed project bytes", err)
	}
}
