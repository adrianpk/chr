package core

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowLooseFixture(t *testing.T) *WorkflowStore {
	t.Helper()

	store := workflowTestStore(t)

	specs := []struct{ priority, severity, area string }{{"high", "low", "frontend"}, {"normal", "critical", "frontend"}, {"critical", "high", "persistence"}, {"high", "normal", "persistence"}, {"unclassified", "", "frontend"}}
	for i, spec := range specs {
		key := fmt.Sprintf("PRJ-%d", i+1)
		record := workflowTestTask(key)

		record.Body = []byte(fmt.Sprintf("---\nkey: %s\ntitle: Example\nstatus: ready_for_dev\ncreated_at: 2026-01-02T03:04:05Z\nupdated_at: 2026-01-02T03:04:05Z\npriority: %s\nseverity: %s\nproject_area: %s\ntype: task\ntags: [delivery, ui]\n---\nTask body\n", key, spec.priority, spec.severity, spec.area))
		if _, err := store.Create(t.Context(), key, record); err != nil {
			t.Fatal(err)
		}
	}

	return store
}

func workflowPreviewKeys(preview WorkflowPreview) []string {
	keys := []string{}
	for _, member := range preview.Members {
		keys = append(keys, member.Key)
	}

	return keys
}

// TestWorkflowSelectionLooseExamples covers the four required AND-filtered area/rank examples.
func TestWorkflowSelectionLooseExamples(t *testing.T) {
	store := workflowLooseFixture(t)

	cases := []struct {
		name, area string
		order      workflowstate.TicketOrder
		ranks      []string
		severity   bool
		expected   []string
	}{
		{"filtered frontend severity", "frontend", workflowstate.TicketOrderSeverity, []string{"critical", "high"}, true, []string{"PRJ-2"}},
		{"filtered persistence priority", "persistence", workflowstate.TicketOrderPriority, []string{"critical", "high"}, false, []string{"PRJ-3", "PRJ-4"}},
		{"area by severity", "frontend", workflowstate.TicketOrderSeverity, nil, true, []string{"PRJ-2", "PRJ-1", "PRJ-5"}},
		{"area by priority", "frontend", workflowstate.TicketOrderPriority, nil, false, []string{"PRJ-1", "PRJ-2", "PRJ-5"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Limit: 128, Order: tc.order, Filters: workflowstate.TicketFilters{ProjectAreas: []string{" " + tc.area + " "}, Tags: []string{"delivery", "ui"}, Types: []string{"task"}}}
			if tc.severity {
				request.Filters.Severities = tc.ranks
			} else {
				request.Filters.Priorities = tc.ranks
			}

			preview, err := store.Preview(t.Context(), request)
			if err != nil || !slices.Equal(workflowPreviewKeys(preview), tc.expected) {
				t.Fatalf("keys = %v, %v", workflowPreviewKeys(preview), err)
			}
		})
	}
}

// TestWorkflowSelectionLooseFilters checks thresholds, exact case, tags, missing ranks and input bounds.
func TestWorkflowSelectionLooseFilters(t *testing.T) {
	store := workflowLooseFixture(t)

	cases := []struct {
		name               string
		filters            workflowstate.TicketFilters
		priority, severity string
		expected           []string
		fail               bool
	}{
		{name: "both thresholds", priority: "high", severity: "normal", expected: []string{"PRJ-3", "PRJ-4"}},
		{name: "unclassified exact", filters: workflowstate.TicketFilters{Priorities: []string{"unclassified"}}, expected: []string{"PRJ-5"}},
		{name: "case sensitive", filters: workflowstate.TicketFilters{ProjectAreas: []string{"Frontend"}}, expected: []string{}},
		{name: "all tags required", filters: workflowstate.TicketFilters{Tags: []string{"delivery", "absent"}}, expected: []string{}},
		{name: "invalid priority", priority: "urgent", fail: true},
		{name: "unclassified minimum", severity: "unclassified", fail: true},
		{name: "invalid severity", filters: workflowstate.TicketFilters{Severities: []string{"medium"}}, fail: true},
		{name: "oversized list", filters: workflowstate.TicketFilters{Tags: make([]string, 129)}, fail: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Limit: 128, Filters: tc.filters, MinPriority: tc.priority, MinSeverity: tc.severity}

			preview, err := store.Preview(t.Context(), request)
			if (err != nil) != tc.fail {
				t.Fatalf("filter error = %v", err)
			}

			if !tc.fail && !slices.Equal(workflowPreviewKeys(preview), tc.expected) {
				t.Fatalf("keys = %v", workflowPreviewKeys(preview))
			}
		})
	}
}

// TestWorkflowSelectionNativeTies compares both ranks before time and natural keys; unknown times sort last.
func TestWorkflowSelectionNativeTies(t *testing.T) {
	known := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	later := known.Add(time.Hour)
	tickets := []IndexedTicket{{IndexedTicketIdentity: IndexedTicketIdentity{Key: "PRJ-10"}, IndexedTicketClassification: IndexedTicketClassification{Priority: "high", Severity: "normal"}, IndexedTicketTimes: IndexedTicketTimes{CreatedAt: &known}}, {IndexedTicketIdentity: IndexedTicketIdentity{Key: "PRJ-2"}, IndexedTicketClassification: IndexedTicketClassification{Priority: "high", Severity: "normal"}, IndexedTicketTimes: IndexedTicketTimes{CreatedAt: &known}}, {IndexedTicketIdentity: IndexedTicketIdentity{Key: "PRJ-1"}, IndexedTicketClassification: IndexedTicketClassification{Priority: "high", Severity: "critical"}, IndexedTicketTimes: IndexedTicketTimes{CreatedAt: &later}}, {IndexedTicketIdentity: IndexedTicketIdentity{Key: "PRJ-3"}, IndexedTicketClassification: IndexedTicketClassification{Priority: "high", Severity: "normal"}}, {IndexedTicketIdentity: IndexedTicketIdentity{Key: "PRJ-4"}}}
	slices.SortFunc(tickets, func(left, right IndexedTicket) int {
		return compareTicketLoopOrder(workflowstate.TicketOrderPrioritySeverity, left, right)
	})

	expected := []string{"PRJ-1", "PRJ-2", "PRJ-10", "PRJ-3", "PRJ-4"}
	for i, ticket := range tickets {
		if ticket.Key != expected[i] {
			t.Fatalf("combined rank ties = %+v", tickets)
		}
	}
}

// TestWorkflowSelectionLooseOwnership proves filters cannot admit slice-owned, terminal or legacy work.
func TestWorkflowSelectionLooseOwnership(t *testing.T) {
	store, result := workflowSelectionFixture(t)
	record := workflowTestTask("PRJ-9")
	record.Task.UseCase = "example"

	record.Task.Phase, record.Task.Status, record.Task.DeliveryRef = "integrated", TicketStatusDone, "run/evidence/receipt.json"
	if _, err := store.Create(t.Context(), "Done", record); err != nil {
		t.Fatal(err)
	}

	request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Limit: 128, Statuses: []string{TicketStatusReadyForDev, TicketStatusDone}, Filters: workflowstate.TicketFilters{UseCases: []string{"example"}}}

	preview, err := store.Preview(t.Context(), request)
	if err != nil || preview.Total != 0 || preview.Excluded["slice_owned"] != len(result.Tasks) || preview.Excluded["integrated"] != 1 {
		t.Fatalf("owned work entered loose scope: %+v, %v", preview, err)
	}

	request.Selector = workflowstate.Selector{Extent: "slice", Identity: result.Set.Set.Slices[0].ID}
	if _, err := store.Preview(t.Context(), request); err == nil {
		t.Fatal("filters can reorder hierarchy")
	}
}

// TestWorkflowSelectionConcurrent verifies previews and guarded writes share one consistent file snapshot.
func TestWorkflowSelectionConcurrent(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowTestCreate(t, store, "PRJ-1")
	result := make(chan error, 1)

	go func() {
		for i := 0; i < 10; i++ {
			record.Task.Status = TicketStatusBacklog
			if i%2 == 0 {
				record.Task.Status = TicketStatusReadyForDev
			}

			next, err := store.Write(t.Context(), record, record.Revision)
			if err != nil {
				result <- err

				return
			}

			record = next
		}

		result <- nil
	}()

	for i := 0; i < 10; i++ {
		workflowTestPreview(t, store, "ticket", "PRJ-1")
	}

	if err := <-result; err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := store.Preview(ctx, workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Limit: 1}); err == nil {
		t.Fatal("cancelled preview continued")
	}

	request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Limit: 1, MinPriority: strings.Repeat("x", 33)}
	if _, err := store.Preview(t.Context(), request); err == nil {
		t.Fatal("oversized threshold accepted")
	}
}
