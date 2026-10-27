package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"strconv"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowIntegrationCLIArguments(status WorkflowWorkStatus) []string {
	current := status.Integration

	return []string{"--revision", strconv.FormatUint(status.Revision, 10), "--entry-id", current.EntryID, "--entry-revision", strconv.FormatUint(current.EntryRevision, 10), "--budget-revision", strconv.FormatUint(current.Budget.Revision, 10), "--candidate-hash", current.CandidateHash}
}

func workflowIntegrationCLIGrant(t *testing.T, operations service, status WorkflowWorkStatus, daemon bool) (WorkflowWorkStatus, error) {
	t.Helper()

	args := append(workflowIntegrationCLIArguments(status), "--record-id", "more-submissions", "--kind", "submission", "--amount", "1", "--actor", "operator", "--reason", "Authorize one additional bounded submission", status.ID)
	if !daemon {
		return workflowCLIInjectedControl(t, operations, "grant", args...)
	}

	var output, stderr bytes.Buffer

	command := append([]string{"daemon", "workflow", "grant", "--dir", operations.projectContext.ProjectDir, "--json"}, args...)
	err := run(t.Context(), command, &output, &stderr)

	var current WorkflowWorkStatus
	if output.Len() != 0 {
		err = errors.Join(err, json.Unmarshal(output.Bytes(), &current))
	}

	return current, err
}

// TestWorkflowIntegrationGrant submits actual bounded grants through work,
// daemon and manual callers, then rejects reuse of the displayed old revisions.
func TestWorkflowIntegrationGrant(t *testing.T) {
	cases := []struct {
		name   string
		submit func(*testing.T, service, WorkflowWorkStatus) (WorkflowWorkStatus, error)
	}{
		{"work CLI", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			return workflowIntegrationCLIGrant(t, s, shown, false)
		}},
		{"daemon CLI", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			return workflowIntegrationCLIGrant(t, s, shown, true)
		}},
		{"manual service", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			request := serviceIntegrationControl(shown, "grant")
			request.BudgetGrant = &workflowstate.WorkflowIntegrationGrant{ID: "manual-extension", Kind: "submission", Amount: 1, Actor: "operator", Reason: "Authorize one additional bounded submission"}

			return s.ControlWorkExtent(t.Context(), request)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operations, id := serviceIntegrationReady(t)

			before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
			if err != nil {
				t.Fatal(err)
			}

			after, err := tc.submit(t, operations, before)
			checkWorkflowIntegrationExtension(t, before, after, err)

			images := workflowCLIImage(t, operations.projectContext.ProjectDir)

			stale, err := tc.submit(t, operations, before)
			if !errors.Is(err, workflowstate.ErrConflict) || stale.Revision != after.Revision || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("stale operator action changed authority or hid current status", err)
			}
		})
	}
}

func checkWorkflowIntegrationExtension(t *testing.T, before, after WorkflowWorkStatus, err error) {
	t.Helper()

	if err != nil || after.Integration == nil {
		t.Fatal("guarded extension failed", err)
	}

	old, current := before.Integration, after.Integration
	if current.Budget.Limits.Submission != old.Budget.Limits.Submission+1 || current.Budget.Consumed != old.Budget.Consumed || current.Budget.AdmissionLimits != old.Budget.AdmissionLimits || current.Phase != old.Phase || current.Publication != "not_started" || after.IntegratedUnits != 0 {
		t.Fatal("operator extension reset usage, changed phase or fabricated publication")
	}
}

// TestWorkflowIntegrationApproval binds native CLI approval to every displayed
// guard and rejects both omitted native fences and replay of the previous form.
func TestWorkflowIntegrationApproval(t *testing.T) {
	operations, id := serviceIntegrationReady(t)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	images := workflowCLIImage(t, operations.projectContext.ProjectDir)
	if _, err := workflowCLIInjectedControl(t, operations, "approve", "--binding", status.ApprovalHash, id); err == nil || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("native approval omitted its displayed guards")
	}

	args := append(workflowIntegrationCLIArguments(status), "--binding", status.ApprovalHash, id)
	after, err := workflowCLIInjectedControl(t, operations, "approve", args...)
	checkServiceIntegrationApproval(t, status, after, err)

	images = workflowCLIImage(t, operations.projectContext.ProjectDir)
	if _, err := workflowCLIInjectedControl(t, operations, "approve", args...); !errors.Is(err, workflowstate.ErrConflict) || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("native CLI approval reused a stale displayed revision", err)
	}
}
