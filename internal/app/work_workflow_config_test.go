package app

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

// TestWorkflowConfigCLI exercises the root parser and shared service using real
// project bytes. Inspection reports sources without initializing or claiming work.
func TestWorkflowConfigCLI(t *testing.T) {
	project := newTestWorkspace(t)
	body := "workflow:\n  version: 1\nwork:\n  max_active_workers: 2\nreview:\n  mode: human_review\n"
	writeCLIWorkspaceConfig(t, project, body)

	before := workflowCLIImage(t, project)

	cases := []struct {
		name  string
		args  []string
		valid bool
	}{
		{"effective", []string{"--json"}, true},
		{"restrictive", []string{"--json", "--review-restriction", "human_review"}, true},
		{"text", nil, true},
		{"relaxed", []string{"--review-restriction", "trusted_auto_merge"}, false},
		{"positional", []string{"extra"}, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var output, stderr bytes.Buffer

			args := append([]string{"work", "config", "--dir", project}, test.args...)

			err := run(t.Context(), args, &output, &stderr)
			if (err == nil) != test.valid {
				t.Fatal("CLI configuration boundary differs", output.String(), err)
			}

			if test.valid {
				checkWorkflowConfigCLIOutput(t, test.name, output.Bytes())
			}

			if !maps.Equal(before, workflowCLIImage(t, project)) {
				t.Fatal("configuration inspection changed workspace bytes")
			}
		})
	}
}

func checkWorkflowConfigCLIOutput(t *testing.T, name string, output []byte) {
	t.Helper()

	if name == "text" {
		if !strings.Contains(string(output), "review.mode: human_review (source: project)") {
			t.Fatal("text output lost effective policy source", string(output))
		}

		return
	}

	var result core.WorkflowConfiguration
	if err := json.Unmarshal(output, &result); err != nil || result.Version != 1 || len(result.Values) != 13 {
		t.Fatal("JSON configuration is incomplete", result, err)
	}
}

// TestWorkflowConfigRestriction carries explicit CLI human-review authority to
// shared admission and rejects an automatic override without another attempt.
func TestWorkflowConfigRestriction(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)
	project := operations.projectContext.ProjectDir
	writeCLIWorkspaceConfig(t, project, "workflow:\n  version: 1\n")

	preview, err := operations.PreviewWorkExtent(t.Context(), core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: record.Task.Key}, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	command := newWorkTicketCommand(&output, testAppDependencies())
	command.Action = func(ctx context.Context, cmd *cli.Command) error {
		return (workWorkflowCLI{stdout: &output}).extent(ctx, cmd, operations, "ticket")
	}
	err = command.Run(t.Context(), []string{"test", "--json", "--human-review", "--snapshot", preview.Preview.SnapshotHash, "--max-steps", "1", record.Task.Key})

	var status WorkflowWorkStatus
	if err != nil || json.Unmarshal(output.Bytes(), &status) != nil {
		t.Fatal("restrictive CLI failed", output.String(), err)
	}

	checkWorkflowConfigRestricted(t, operations, status, provider.creates)

	before := workflowCLIImage(t, project)
	if _, err := operations.StartWorkExtent(t.Context(), core.WorkflowPreviewRequest{Selector: status.Selector}, 1, core.ReviewWorkflowModeTrustedAutoMerge); err == nil {
		t.Fatal("invocation override enabled automatic publication")
	}

	if !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("rejected override changed accepted work")
	}
}

func checkWorkflowConfigRestricted(t *testing.T, operations service, status WorkflowWorkStatus, creates int) {
	t.Helper()

	retained, err := operations.ReadWorkflow(t.Context(), status.ID)
	if err != nil || retained.ReviewRestriction != core.ReviewWorkflowModeHumanReview {
		t.Fatal("CLI restriction was not frozen at admission", retained, err)
	}

	if status.Policy.Mode != core.ReviewWorkflowModeHumanReview || status.Policy.Source != "invocation_restriction" || status.ValidatedTasks != 1 || creates != 0 {
		t.Fatal("restriction changed validation or publication", status)
	}
}
