package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramJourneyAdmissionFixture(t *testing.T) (*WorkflowStore, WorkflowRecord, WorkflowStartRequest) {
	t.Helper()

	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	record.Task.Status = TicketStatusBacklog
	record = workflowTestWrite(t, store, record)
	request.Selection.Limit = workflowstate.PageLimit

	preview, err := store.Preview(t.Context(), request.Selection)
	if err != nil || len(preview.Members) != 1 {
		t.Fatal(preview, err)
	}

	ticket, err := store.PreviewWorkflowTicketAdmission(t.Context(), preview.Members[0])
	if err != nil || ticket == nil {
		t.Fatal(ticket, err)
	}

	policy, err := store.InspectWorkflowPolicy(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	execution, err := store.InspectWorkflowExecutionPolicy(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	request.Selection.SnapshotHash = preview.SnapshotHash
	request.Admission = &WorkflowAdmissionGuard{SnapshotHash: preview.SnapshotHash, Selected: &WorkflowSelectionGuard{Key: record.Task.Key, Revision: record.Revision}, Policy: policy, ExecutionPolicy: execution, ReadyTicket: ticket}

	return store, record, request
}

// TestTelegramJourneyAdmission keeps preview read-only, commits readiness with
// the invocation and reconciles an original request after its ticket advances.
func TestTelegramJourneyAdmission(t *testing.T) {
	store, record, request := telegramJourneyAdmissionFixture(t)

	before, err := store.Read(t.Context(), record.Path)
	if err != nil || before.Task.Status != TicketStatusBacklog || before.Revision != record.Revision {
		t.Fatal("preview changed backlog work", before, err)
	}

	first := workflowTestStart(t, store, request)
	checkTelegramJourneyReady(t, store, record, first.Invocation)

	claim := workflowTestClaim(t, store, first.Invocation.ID)
	restarted := NewWorkflowStore(store.layout, store.projectID)
	check := func(context.Context) error {
		t.Error("replay repeated admission authority")

		return workflowstate.ErrConflict
	}

	again, err := restarted.StartWorkflowGuarded(t.Context(), request, check)
	if err != nil || !again.Attached || again.Invocation.ID != first.Invocation.ID || len(again.Invocation.Attempts) != 1 || again.Invocation.Attempts[0].ID != claim.AttemptID {
		t.Fatal("replay repeated readiness or lost native ownership", again, err)
	}
}

func checkTelegramJourneyReady(t *testing.T, store *WorkflowStore, record WorkflowRecord, run WorkflowInvocation) {
	t.Helper()

	ready, err := store.Read(t.Context(), record.Path)
	if err != nil || ready.Task.Status != TicketStatusReadyForDev || ready.Revision != record.Revision+1 || len(run.Attempts) != 0 {
		t.Fatal("admission did not atomically prepare the ticket without execution", ready, run, err)
	}

	member := run.Members[0]
	if member.Admission.Status != TicketStatusBacklog || member.Admission.Revision != record.Revision || member.ExpectedRevision != ready.Revision || member.ExpectedSourceHash != ready.SourceHash {
		t.Fatal("admission lost original and accepted images", member)
	}
}

// TestTelegramJourneyAdmissionGuards rejects stale or altered decisions before
// either a backlog ticket or native invocation changes.
func TestTelegramJourneyAdmissionGuards(t *testing.T) {
	for _, change := range []string{"source", "policy", "authority", "contract", "validation", "selector", "selected"} {
		t.Run(change, func(t *testing.T) {
			store, record, request := telegramJourneyAdmissionFixture(t)
			check := telegramJourneyChangeAdmission(t, store, record, &request, change)

			before, err := store.Read(t.Context(), record.Path)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := store.StartWorkflowGuarded(t.Context(), request, check); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed decision accepted", err)
			}

			after, err := store.Read(t.Context(), record.Path)
			if err != nil || after.Revision != before.Revision || after.SourceHash != before.SourceHash || after.Task.Status != TicketStatusBacklog {
				t.Fatal("rejection changed ticket readiness", after, err)
			}

			runs, _, err := store.InspectWorkflowPage(t.Context(), 0, 8)
			if err != nil || len(runs) != 0 {
				t.Fatal("rejection allocated an invocation", runs, err)
			}
		})
	}
}

func telegramJourneyChangeAdmission(t *testing.T, store *WorkflowStore, record WorkflowRecord, request *WorkflowStartRequest, change string) WorkflowAdmissionCheck {
	t.Helper()

	switch change {
	case "source":
		path := filepath.Join(store.layout.Root(), record.Path)

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
			t.Fatal(err)
		}
	case "policy":
		request.Admission.Policy.Revision = strings.Repeat("f", 64)
	case "authority":
		return func(context.Context) error { return workflowstate.ErrConflict }
	case "contract":
		request.Admission.ReadyTicket.ContractHash = strings.Repeat("f", 64)
	case "validation":
		request.Admission.ReadyTicket.Validation.Commands = [][]string{{"true"}}
	case "selector":
		request.Selection.Selector = WorkflowSelector{Extent: "loose"}
	case "selected":
		request.Admission.Selected = nil
	}

	return nil
}

// TestTelegramJourneyContract preserves the input record and rejects incomplete
// descriptions, dependency expansion and validation waivers without reasons.
func TestTelegramJourneyContract(t *testing.T) {
	for _, change := range []string{"valid", "acceptance", "validation", "dependency", "status"} {
		t.Run(change, func(t *testing.T) {
			record := workflowTestTask("JRN-1")
			record.Task.Status = TicketStatusBacklog
			contract := workflowPlanFixture().Slices[0].Tasks[0]
			telegramJourneyChangeContract(&record, &contract, change)

			prepared, err := PrepareIndependentWorkflowTask(record, contract)
			if change != "valid" {
				if err == nil {
					t.Fatal("partial or widened contract accepted")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if validateWorkflowClaimContract(prepared) != nil || prepared.Task.Status != TicketStatusBacklog || prepared.Task.ContractHash == record.Task.ContractHash || len(record.Domain["execution_contract"]) != 0 {
				t.Fatal("preparation mutated its input or changed readiness", prepared, err)
			}
		})
	}
}

func telegramJourneyChangeContract(record *WorkflowRecord, contract *GeneratedWorkflowTask, change string) {
	switch change {
	case "acceptance":
		contract.Acceptance = nil
	case "validation":
		contract.Validation = WorkflowValidation{}
	case "dependency":
		contract.DependsOn = []WorkflowRef{{Kind: "ticket", ID: "OTHER-1"}}
	case "status":
		record.Task.Status = TicketStatusReadyForDev
	}
}
