package app

import (
	"bytes"
	"maps"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/tui"
)

// TestWorkflowIntegrationViews reads an actual prepared candidate through CLI
// and TUI entrypoints, preserving authority and the distinction between S and P.
func TestWorkflowIntegrationViews(t *testing.T) {
	operations, id := serviceIntegrationReady(t)
	project := operations.projectContext.ProjectDir
	before := workflowCLIImage(t, project)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowIntegrationJSON(t, project, status)

	cases := []struct {
		name string
		read func() (string, error)
	}{
		{"CLI", func() (string, error) {
			var output, stderr bytes.Buffer

			err := run(t.Context(), []string{"work", "status", "--dir", project, id}, &output, &stderr)

			return output.String(), err
		}},
		{"TUI", func() (string, error) {
			executor := newTUIWorkActionExecutor(operations)

			result, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowStatus, Workflow: &tui.WorkflowActionRequest{ID: id}})
			if result.Workflow == nil {
				return "", err
			}

			return strings.Join(result.Workflow.Lines, "\n"), err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text, err := tc.read()
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowIntegrationView(t, text, status.Integration)

			if !maps.Equal(before, workflowCLIImage(t, project)) {
				t.Fatal("native view changed retained authority")
			}
		})
	}
}

func checkWorkflowIntegrationJSON(t *testing.T, project string, expected WorkflowWorkStatus) {
	t.Helper()

	actual, err := workflowCLIStatus(t, project, "status", expected.ID)
	if err != nil || actual.Integration == nil {
		t.Fatal("CLI JSON omitted retained integration status", err)
	}

	shown, observed := actual.Integration, expected.Integration
	if shown.Phase != observed.Phase || shown.QueuePosition != observed.QueuePosition || shown.OriginalSource != observed.OriginalSource || shown.PreparedResult != observed.PreparedResult || actual.IntegratedUnits != 0 {
		t.Fatal("CLI JSON changed native identities or claimed delivery")
	}

	checkWorkflowCLIStatusPrivacy(t, actual)
}

func checkWorkflowIntegrationView(t *testing.T, text string, observed *WorkflowIntegrationStatus) {
	t.Helper()

	for _, fact := range []string{
		"Integration phase: reviewing",
		"Target queue: " + observed.Repository + " " + observed.TargetRef,
		"Queue position: 1",
		"Original source: " + observed.OriginalSource,
		"Validated base: " + observed.ValidatedBase,
		"Prepared result: " + observed.PreparedResult,
		"Publication: not_started",
		"Agent capacity: serial; limit 1",
	} {
		if !strings.Contains(text, fact) {
			t.Fatal("native view omitted a retained fact", fact)
		}
	}

	for _, private := range []string{"claim_generation", "lease_until", "owner_token", "credential_identity", "owned_workspace_id"} {
		if strings.Contains(text, private) {
			t.Fatal("native view exposed private ownership", private)
		}
	}
}
