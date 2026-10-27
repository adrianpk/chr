package core

import (
	"slices"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowLooseExample struct {
	name, area, priority, severity string
	order                          workflowstate.TicketOrder
	filtered                       bool
	expected                       []string
}

// TestWorkflowLooseExamples applies all four frozen AND-filtered requests again
// after real canonical delivery, using changed pending ranks rather than aliases.
func TestWorkflowLooseExamples(t *testing.T) {
	cases := []workflowLooseExample{
		{"frontend severity filter", "frontend", "high", "low", workflowstate.TicketOrderSeverity, true, []string{"PRJ-91"}},
		{"persistence priority filter", "persistence", "critical", "high", workflowstate.TicketOrderPriority, true, []string{"PRJ-91"}},
		{"area priority", "frontend", "critical", "normal", workflowstate.TicketOrderPriority, false, []string{"PRJ-92", "PRJ-91"}},
		{"area severity", "frontend", "normal", "critical", workflowstate.TicketOrderSeverity, false, []string{"PRJ-92", "PRJ-91"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := workflowLooseConfiguredQueue(t, false, func(store *WorkflowStore, request *WorkflowStartRequest) {
				configureWorkflowLooseExample(t, store, request, tc)
			})
			workflowLooseFields(t, f.store, "PRJ-92", tc.area, tc.priority, tc.severity)
			workflowFinalizeFrontier(t, f)
			workflowAdvanceFrontier(t, f)

			run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
			if err != nil || !slices.Equal(workflowLoosePendingKeys(run), tc.expected) || run.Cursor != 1 || run.LooseQueue.Admitted[0] != "PRJ-90" {
				t.Fatal("reselection changed filters, ordered delivered work, or lost rank changes", workflowLoosePendingKeys(run), err)
			}
		})
	}
}

func configureWorkflowLooseExample(t *testing.T, store *WorkflowStore, request *WorkflowStartRequest, example workflowLooseExample) {
	t.Helper()

	request.Selection.Order = example.order

	request.Selection.Filters = workflowstate.TicketFilters{ProjectAreas: []string{example.area}, Types: []string{"task"}, Tags: []string{"delivery", "ui"}}
	if example.filtered {
		if example.order == workflowstate.TicketOrderSeverity {
			request.Selection.Filters.Severities = []string{"high"}
		} else {
			request.Selection.Filters.Priorities = []string{"high"}
		}
	}

	for _, key := range []string{"PRJ-90", "PRJ-91", "PRJ-92"} {
		workflowLooseFields(t, store, key, example.area, "high", "high")
	}
}

func workflowLooseFields(t *testing.T, store *WorkflowStore, key, area, priority, severity string) WorkflowRecord {
	t.Helper()

	record := workflowLooseRecord(t, store, key)
	record.Body = workflowLooseBody(key, priority)
	record.Body = []byte(strings.ReplaceAll(strings.ReplaceAll(string(record.Body), "project_area: frontend", "project_area: "+area), "severity: high", "severity: "+severity))

	changed, err := store.Write(t.Context(), record, record.Revision)
	if err != nil {
		t.Fatal(err)
	}

	return changed
}

func workflowLoosePendingKeys(run WorkflowInvocation) []string {
	keys := []string{}
	for _, member := range run.Members[run.Cursor : run.Cursor+run.LooseQueue.Eligible] {
		keys = append(keys, member.Admission.Key)
	}

	return keys
}

// TestWorkflowLooseLater preserves confirmed progress and initial admissions.
// Only matching ready independent records can extend an unfinished queue.
func TestWorkflowLooseLater(t *testing.T) {
	for _, include := range []bool{false, true} {
		t.Run(map[bool]string{false: "frozen", true: "accepted later"}[include], func(t *testing.T) {
			f := workflowLooseConfiguredQueue(t, include, func(_ *WorkflowStore, request *WorkflowStartRequest) {
				request.Selection.Filters.ProjectAreas = []string{"frontend"}
			})
			original := workflowLooseRecord(t, f.store, "PRJ-91")
			workflowLooseCreate(t, f.store, original, "PRJ-93", "critical", nil)
			workflowLooseCreate(t, f.store, original, "PRJ-94", "critical", nil)
			workflowLooseFields(t, f.store, "PRJ-94", "persistence", "critical", "high")
			workflowLooseClassify(t, f.store, workflowLooseCreate(t, f.store, original, "PRJ-95", "critical", nil), "critical", TicketStatusBacklog)
			workflowFinalizeFrontier(t, f)
			workflowAdvanceFrontier(t, f)

			run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
			checkWorkflowLooseLater(t, run, include, err)
		})
	}
}

func checkWorkflowLooseLater(t *testing.T, run WorkflowInvocation, include bool, failure error) {
	t.Helper()

	if failure != nil || run.Members[0].Admission.Key != "PRJ-90" || run.ConfirmedCursor != 1 || len(run.Deliveries) != 1 {
		t.Fatal("later work crossed or changed a confirmed receipt", failure)
	}

	checkWorkflowLooseLaterScope(t, run, include)
}

func checkWorkflowLooseLaterScope(t *testing.T, run WorkflowInvocation, include bool) {
	t.Helper()

	if include {
		if len(run.Members) != 4 || run.Members[1].Admission.Key != "PRJ-93" || run.LooseQueue.Admitted[3] != "PRJ-93" || run.Excluded["later"] != 0 {
			t.Fatal("explicit later admission lost frozen scope or eligibility")
		}
	} else if len(run.Members) != 3 || run.Excluded["later"] != 1 {
		t.Fatal("default admission included or hid later work")
	}
}
