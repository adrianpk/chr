package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

const workflowCLIManualHarness = `set -eu
printf 'Delivered\n' > result.txt
git add result.txt
git -c user.name=Test -c user.email=test@example.com commit -qm 'Manual result'
input=$(sha256sum "$CHR_WORKFLOW_INPUT")
printf '{"input_hash":"%s","branch":"%s","commit":"%s","tree":"%s"}\n' "${input%% *}" "$(git symbolic-ref HEAD)" "$(git rev-parse HEAD)" "$(git rev-parse 'HEAD^{tree}')"`

func workflowCLIManualStart(t *testing.T, operations service, hierarchy core.WorkflowMaterializationResult) string {
	t.Helper()

	request := core.WorkflowStartRequest{RequestKey: "cli-manual", Selection: core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: core.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: serviceWorkflowDependencyGit(t, operations.projectContext.ProjectDir, "rev-parse", "HEAD")}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return run.Invocation.ID
}

// TestWorkflowCLIManual supervises a real harness through root command parsing,
// then host-validates its exact Git image without claiming integrated delivery.
func TestWorkflowCLIManual(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir
	id := workflowCLIManualStart(t, operations, hierarchy)
	arguments := []string{"reconcile", "--kind", "manual", "--harness", "sh", "--harness", "-c", "--harness", workflowCLIManualHarness, id}
	before := workflowCLIImage(t, project)

	previewArgs := append([]string{arguments[0], "--preview"}, arguments[1:]...)
	if _, err := workflowCLIStatus(t, project, previewArgs...); err != nil || !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("manual preview launched a harness or wrote state", err)
	}

	status, err := workflowCLIStatus(t, project, arguments...)
	checkWorkflowCLIManualStatus(t, status, err)
}

func checkWorkflowCLIManualStatus(t *testing.T, status WorkflowWorkStatus, failure error) {
	t.Helper()

	if failure != nil || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || status.IntegratedUnits != 0 || len(status.Evidence) != 1 || status.Evidence[0].ResultRef == "" || status.Evidence[0].ResultHash == "" {
		t.Fatal("manual reconciliation lost host validation or fabricated delivery", status, failure)
	}

	checkWorkflowCLIStatusPrivacy(t, status)
}

func checkWorkflowCLIStatusPrivacy(t *testing.T, status WorkflowWorkStatus) {
	t.Helper()

	data, err := json.Marshal(status)
	if err != nil || strings.Contains(string(data), "owner_token") || strings.Contains(string(data), "owner_process") {
		t.Fatal("status exposed host credentials or process identity", err)
	}
}

// TestWorkflowCLIManualFailure rejects stale or malformed harness reports,
// checkpoints the attempt without validation and releases its owned capacity.
func TestWorkflowCLIManualFailure(t *testing.T) {
	cases := []struct{ name, harness string }{
		{"wrong input", strings.Replace(workflowCLIManualHarness, "${input%% *}", strings.Repeat("0", 64), 1)},
		{"unexpected output", "printf 'untrusted success'"},
		{"extra fields", `printf '{"successful":true}'`},
		{"multiple objects", `printf '{} {}'`},
		{"large report", "head -c 4097 /dev/zero"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operations, hierarchy, _ := workflowCLIFixture(t)

			id := workflowCLIManualStart(t, operations, hierarchy)
			if _, err := workflowCLIStatus(t, operations.projectContext.ProjectDir, "reconcile", "--kind", "manual", "--harness", "sh", "--harness", "-c", "--harness", tc.harness, id); err == nil {
				t.Fatal("unverified harness output accepted")
			}

			checkWorkflowCLIManualFailure(t, operations, id)
		})
	}
}

func checkWorkflowCLIManualFailure(t *testing.T, operations service, id string) {
	t.Helper()

	retained, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || retained.Cursor != 0 || retained.Status != "paused" || retained.Attempts[0].Outcome != "checkpointed" {
		t.Fatal("failed report fabricated success or lost checkpoint", retained, err)
	}

	if _, err := operations.ResumeWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		// A changed lane with an unaccepted hash stays fenced. Its failed handoff
		// must release the host's capacity descriptors even while that conflict remains.
		if !errors.Is(err, core.ErrWorkflowConflict) {
			t.Fatal("manual failure leaked capacity", err)
		}

		return
	}

	if err := operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowCLIStart applies explicit ticket flags to shared execution and
// honors a one-step bound without publishing a PR or changing integration state.
func TestWorkflowCLIStart(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)

	preview, err := operations.PreviewWorkExtent(t.Context(), core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: record.Task.Key}, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	command := (commandFactory{deps: testAppDependencies(), io: IO{Stdout: &output}}).workTicketCommand()
	command.Action = func(ctx context.Context, cmd *cli.Command) error {
		return (workWorkflowCLI{stdout: &output}).extent(ctx, cmd, operations, "ticket")
	}

	err = command.Run(t.Context(), []string{"test", "--json", "--snapshot", preview.Preview.SnapshotHash, "--max-steps", "1", record.Task.Key})

	var status WorkflowWorkStatus
	if err != nil || json.Unmarshal(output.Bytes(), &status) != nil || status.Selector.Identity != record.Task.Key || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || status.TotalAttempts != 1 || provider.creates != 0 {
		t.Fatal("explicit CLI selection changed scope, exceeded its step bound or published", output.String(), err)
	}
}

// TestWorkflowCLIAdmission rejects a stale target-bound preview before allocating
// an invocation, consuming a worker attempt or publishing a source candidate.
func TestWorkflowCLIAdmission(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)
	request := core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: record.Task.Key}, Limit: 1}

	preview, err := operations.PreviewWorkExtent(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	project := operations.projectContext.ProjectDir
	workflowCLIAdvanceTarget(t, project, provider.remote)

	request.SnapshotHash = preview.Preview.SnapshotHash
	if _, err := operations.StartWorkExtent(t.Context(), request, 1); !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatal("stale preview admitted execution", err)
	}

	paths, err := filepath.Glob(filepath.Join(project, "ops", "default", "run", "*", "extension", "chr", "workflow.json"))
	if err != nil || len(paths) != 0 || provider.creates != 0 {
		t.Fatal("stale admission allocated or published work", err)
	}
}
