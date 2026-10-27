package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strconv"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowCLIStatus(t *testing.T, project string, args ...string) (WorkflowWorkStatus, error) {
	t.Helper()

	var output, stderr bytes.Buffer

	command := append([]string{"work", args[0], "--dir", project, "--json"}, args[1:]...)
	err := run(t.Context(), command, &output, &stderr)

	var status WorkflowWorkStatus
	if output.Len() != 0 {
		err = errors.Join(err, json.Unmarshal(output.Bytes(), &status))
	}

	return status, err
}

// TestWorkflowCLIControls drives real status, preview, stop and resume commands;
// stale status pages fail and terminal continuation cannot restart the extent.
func TestWorkflowCLIControls(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir
	request := core.WorkflowStartRequest{RequestKey: "cli-control", Selection: workflowstate.WorkflowPreviewRequest{Selector: workflowstate.WorkflowSelector{Extent: "slice", Identity: hierarchy.Set.Set.Slices[0].ID}, Target: workflowstate.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: serviceWorkflowDependencyGit(t, project, "rev-parse", "HEAD")}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	id := run.Invocation.ID
	checkWorkflowCLIStatusPreview(t, project, id)

	status, err := workflowCLIStatus(t, project, "status", "--limit", "1", id)
	if err != nil {
		t.Fatal(err)
	}

	paused, err := workflowCLIStatus(t, project, "pause", id)
	if err != nil || paused.Status != "paused" {
		t.Fatal("pause did not reach shared lifecycle", paused, err)
	}

	if _, err := workflowCLIStatus(t, project, "status", "--revision", strconv.FormatUint(status.Revision, 10), "--offset", "1", id); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale status pagination accepted", err)
	}

	checkWorkflowCLITerminal(t, project, id)
}

func checkWorkflowCLIStatusPreview(t *testing.T, project, id string) {
	t.Helper()

	before := workflowCLIImage(t, project)

	status, err := workflowCLIStatus(t, project, "status", "--limit", "1", id)
	if err != nil || status.ID != id || status.Total != 3 || len(status.Members) != 1 || status.NextOffset != 1 {
		t.Fatal("bounded retained status", status, err)
	}

	preview, err := workflowCLIStatus(t, project, "cancel", "--preview", id)
	if err != nil || preview.Revision != status.Revision || !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("control preview changed retained state", err)
	}
}

func checkWorkflowCLITerminal(t *testing.T, project, id string) {
	t.Helper()

	resumed, err := workflowCLIStatus(t, project, "resume", id)
	if err != nil || resumed.Status != "pending" || resumed.TotalAttempts != 0 {
		t.Fatal("resume changed retained budgets or scope", resumed, err)
	}

	cancelled, err := workflowCLIStatus(t, project, "cancel", id)
	if err != nil || cancelled.Status != "cancelled" {
		t.Fatal("cancel did not reach shared lifecycle", cancelled, err)
	}

	continued, err := workflowCLIStatus(t, project, "continue", "--max-steps", "1", id)
	if err != nil || continued.Status != "cancelled" || continued.Revision != cancelled.Revision || continued.TotalAttempts != 0 {
		t.Fatal("continuation restarted terminal work", continued, err)
	}
}

func workflowCLIInjectedControl(t *testing.T, operations service, action string, args ...string) (WorkflowWorkStatus, error) {
	t.Helper()

	var output bytes.Buffer

	command := (commandFactory{deps: testAppDependencies(), io: IO{Stdout: &output}}).workControlCommand(action)
	command.Action = func(ctx context.Context, cmd *cli.Command) error {
		return (workWorkflowCLI{stdout: &output}).control(ctx, cmd, operations, action)
	}

	arguments := append([]string{"test", "--json"}, args...)
	err := command.Run(t.Context(), arguments)

	var status WorkflowWorkStatus
	if output.Len() != 0 {
		err = errors.Join(err, json.Unmarshal(output.Bytes(), &status))
	}

	return status, err
}

// TestWorkflowCLIApproval sends the exact displayed binding through real flag
// parsing and rejects reuse after the canonical PR head moves.
func TestWorkflowCLIApproval(t *testing.T) {
	operations, id, provider := serviceWorkflowReviewReady(t)

	binding, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowCLIApprovalPreview(t, operations, id, binding.InputHash)

	approved, err := workflowCLIInjectedControl(t, operations, "approve", "--binding", binding.InputHash, id)
	if err != nil || approved.ReviewState != "authorized" || approved.IntegratedUnits != 0 {
		t.Fatal("exact approval failed or fabricated delivery", approved, err)
	}

	provider.pr.HeadCommit = strings.Repeat("a", 40)

	if _, err := workflowCLIInjectedControl(t, operations, "approve", "--binding", binding.InputHash, id); err == nil {
		t.Fatal("approval reused after canonical PR head changed")
	}
}

func checkWorkflowCLIApprovalPreview(t *testing.T, operations service, id, binding string) {
	t.Helper()

	before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 1})
	if err != nil || before.ApprovalHash != binding || before.ValidatedTasks != 3 || before.IntegratedUnits != 0 || before.ConfirmedTickets != 0 || before.PullRequest == nil {
		t.Fatal("status confused task validation with confirmed integration", before, err)
	}

	preview, err := workflowCLIInjectedControl(t, operations, "approve", "--preview", "--binding", binding, id)
	if err != nil || preview.Revision != before.Revision || preview.ReviewState != "waiting_approval" {
		t.Fatal("approval preview consumed authority", preview, err)
	}
}

// TestWorkflowCLIInput rejects ambiguous identities and unsafe control arguments
// before project resolution or any admission can occur.
func TestWorkflowCLIInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"empty identity", []string{"slice", ""}, "requires one identity"},
		{"extra identity", []string{"slice", "one", "two"}, "requires one identity"},
		{"long identity", []string{"ticket", strings.Repeat("a", 129)}, "requires one identity"},
		{"loose identity", []string{"loose", "unexpected"}, "does not accept an identity"},
		{"unbound approval", []string{"approve", "invocation"}, "64-character binding"},
		{"negative revision", []string{"status", "--revision", "-1", "invocation"}, "revision must be nonnegative"},
		{"excessive steps", []string{"continue", "--max-steps", "129", "invocation"}, "0..128"},
		{"missing harness", []string{"reconcile", "--kind", "manual", "invocation"}, "1..64 arguments"},
		{"wrong kind", []string{"reconcile", "--kind", "unknown", "invocation"}, "unsupported reconciliation"},
		{"empty page", []string{"slice", "--preview", "--limit", "0", "slice-1"}, "page bounds exceeded"},
		{"large page", []string{"slice", "--preview", "--limit", "129", "slice-1"}, "page bounds exceeded"},
		{"malformed snapshot", []string{"slice", "--preview", "--snapshot", "unbound", "slice-1"}, "64-character binding"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output, stderr bytes.Buffer

			err := run(t.Context(), append([]string{"work"}, tc.args...), &output, &stderr)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal("unbounded or ambiguous input accepted", err)
			}
		})
	}
}
