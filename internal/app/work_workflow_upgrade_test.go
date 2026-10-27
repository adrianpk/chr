package app

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// TestWorkflowUpgradeCLI exercises the actual root command, policy observation,
// snapshot result and retry rather than invoking the core operation directly.
func TestWorkflowUpgradeCLI(t *testing.T) {
	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	prior, err := core.ResolveWorkflowConfiguration(project, "")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"work", "upgrade", "--dir", project, "--config-revision", prior.Revision, "--state-authority", "tickets", "--json"}

	var first bytes.Buffer
	if err := run(t.Context(), args, &first, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var result core.WorkflowUpgradeResult
	if err := json.Unmarshal(first.Bytes(), &result); err != nil || result.Prior.Version != 0 || result.Current.Version != 1 {
		t.Fatal("CLI omitted prior policy or new marker", result, err)
	}

	checkWorkflowUpgradeCLIResult(t, project, args, first.String(), result)
}

func checkWorkflowUpgradeCLIResult(t *testing.T, project string, args []string, first string, result core.WorkflowUpgradeResult) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(project, result.Snapshot)); err != nil {
		t.Fatal("CLI omitted persistent prior snapshot", err)
	}

	var second bytes.Buffer
	if err := run(t.Context(), args, &second, &bytes.Buffer{}); err != nil || first != second.String() {
		t.Fatal("CLI retry changed its upgrade", err)
	}

	var observed bytes.Buffer
	if err := run(t.Context(), []string{"work", "config", "--dir", project}, &observed, &bytes.Buffer{}); err != nil || !strings.Contains(observed.String(), "review.mode: human_review (source: project)") {
		t.Fatal("CLI silently enabled automatic review", observed.String(), err)
	}
}

// TestWorkflowUpgradeCLIRejects checks required choices and strict bounded JSON
// selections before the command can write a version marker.
func TestWorkflowUpgradeCLIRejects(t *testing.T) {
	cases := []struct{ name, selection string }{
		{"invalid object", "{}"},
		{"unknown field", `{"unknown":true}`},
		{"multiple objects", "{} {}"},
		{"null", "null"},
		{"oversize", strings.Repeat(" ", workflowstate.FileLimit+1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := newTestWorkspace(t)

			file := filepath.Join(t.TempDir(), "selection.json")
			if err := os.WriteFile(file, []byte(tc.selection), 0o600); err != nil {
				t.Fatal(err)
			}

			args := []string{"work", "upgrade", "--dir", project, "--config-revision", core.WorkflowContentHash(nil), "--state-authority", "tickets", "--adopt-file", file}
			if err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid adoption accepted")
			}

			assertPathNotExists(t, workspacecore.UserPath(project, "config.yml"))
		})
	}
}

// TestWorkflowUpgradeServiceOwned observes actual project-wide execution rows;
// unfinished unselected work blocks upgrade while failed history remains valid.
func TestWorkflowUpgradeServiceOwned(t *testing.T) {
	cases := []struct {
		name    string
		status  core.ExecutionStatus
		blocked bool
	}{
		{"pending", core.ExecutionStatusPending, true},
		{"running", core.ExecutionStatusRunning, true},
		{"failed history", core.ExecutionStatusFailed, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := newTestWorkspace(t)
			if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, output)
			}

			operations := newTestService(t, project)
			writeWorkflowUpgradeExecution(t, operations, tc.status)

			opts := core.WorkflowUpgradeOptions{ExpectedRevision: core.WorkflowContentHash(nil), StateAuthority: core.WorkStateAuthorityTickets}

			_, err := operations.UpgradeWorkflowConfiguration(t.Context(), opts)
			if (err != nil) != tc.blocked {
				t.Fatal("service ignored legacy project ownership", err)
			}

			if tc.blocked {
				assertPathNotExists(t, workspacecore.UserPath(project, "config.yml"))
			}
		})
	}
}

func writeWorkflowUpgradeExecution(t *testing.T, operations service, status core.ExecutionStatus) {
	t.Helper()

	repository, err := core.NewExecutionRepository(t.Context(), operations.env, operations.assetsFS)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	record := core.ExecutionRecord{ID: "unselected-attempt", ProjectID: operations.projectContext.ProjectID, TicketKey: "OLD-101", Status: status, CreatedAt: now, UpdatedAt: now}
	if err := repository.CreateExecution(t.Context(), record); err != nil {
		t.Fatal(err)
	}

	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}
}
