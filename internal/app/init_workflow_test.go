package app

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// TestInitWorkflowCLI exercises actual root-command initialization and the
// configuration observation boundary against every bundled profile.
func TestInitWorkflowCLI(t *testing.T) {
	for _, profile := range []string{"standard", "minimal", "personal"} {
		t.Run(profile, func(t *testing.T) {
			project := t.TempDir()
			if err := run(t.Context(), []string{"init", "--dir", project, "--profile", profile}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}

			var output bytes.Buffer

			if err := run(t.Context(), []string{"work", "config", "--dir", project, "--json"}, &output, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}

			var result core.WorkflowConfiguration
			if err := json.Unmarshal(output.Bytes(), &result); err != nil || result.Version != 1 {
				t.Fatal("CLI omitted version-one defaults", result, err)
			}

			checkInitWorkflowCLIValues(t, result)
			checkInitWorkflowRuntime(t, project)
		})
	}
}

func checkInitWorkflowCLIValues(t *testing.T, result core.WorkflowConfiguration) {
	t.Helper()

	want := map[string]string{"review.mode": "trusted_auto_merge", "work.state_authority": "tickets", "work.workspace_mode": "worktree", "work.max_active_workers": "1"}
	for _, value := range result.Values {
		if expected, ok := want[value.Key]; ok {
			if value.Value != expected || value.Source != "version_default" {
				t.Fatal("CLI initialization policy differs", value)
			}

			delete(want, value.Key)
		}
	}

	if len(want) != 0 {
		t.Fatal("CLI omitted effective policy", want)
	}
}

func checkInitWorkflowRuntime(t *testing.T, project string) {
	t.Helper()

	before, err := core.ResolveWorkflowConfiguration(project, "")
	if err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), []string{"runtime", "init", "--dir", project, "--dev-root", t.TempDir()}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	after, err := core.ResolveWorkflowConfiguration(project, "")
	if err != nil || before.Revision != after.Revision || before.Version != after.Version {
		t.Fatal("runtime bootstrap changed project policy", before, after, err)
	}
}

// TestInitWorkflowAbsentCLI keeps legacy missing-file policy absent and explains
// the preserved choice through both dry-run and actual command output.
func TestInitWorkflowAbsentCLI(t *testing.T) {
	project := t.TempDir()
	if err := os.MkdirAll(workspacecore.UserPath(project), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, extra := range [][]string{{"--dry-run"}, nil} {
		var output bytes.Buffer

		args := append([]string{"init", "--dir", project}, extra...)
		if err := run(t.Context(), args, &output, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(output.String(), "preserve absent .chr/default/config.yml") || strings.Contains(output.String(), "create .chr/default/config.yml") {
			t.Fatal("CLI misreported policy preservation", output.String())
		}

		assertPathNotExists(t, workspacecore.UserPath(project, "config.yml"))
	}

	checkInitWorkflowRuntime(t, project)
	assertPathNotExists(t, workspacecore.UserPath(project, "config.yml"))
}
