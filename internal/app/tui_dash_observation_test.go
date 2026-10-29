package app

import (
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/app/work"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// Production refresh and Web consume the same policy-owned queue, including
// exclusions, blocked rows and empty pages, without changing canonical files.
func TestTUIDashObservation(t *testing.T) {
	for _, mode := range []string{"native", "legacy"} {
		t.Run(mode, func(t *testing.T) {
			var operations service
			if mode == "native" {
				operations, _ = nativeWorkObservationFixture(t)
			} else {
				operations = newTestService(t, newTestWorkspace(t))
				writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")
			}

			checkTUIDashObservation(t, operations)
		})
	}
}

func checkDashObservationParity(t *testing.T, observed, actual WorkObservation) {
	t.Helper()

	expected, actualView := webDashObservation(observed, 0), webDashObservation(actual, 0)

	expected.ObservedAt, actualView.ObservedAt = "", ""
	if !reflect.DeepEqual(expected, actualView) || actual.QueueNext != observed.QueueNext {
		t.Fatal("adapter observation disagreement", expected, actualView)
	}
}

// Priority and canonical ordering apply to the entire observed queue before
// pagination, so filtering cannot hide matching rows on later pages.
func TestTUIDashQueuePages(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")

	seedDashQueue(t, operations)

	for _, item := range []struct {
		name         string
		request      work.Request
		count, total int
		key          string
	}{
		{"first", work.Request{}, 128, 130, "PRJ-1"},
		{"second", work.Request{Offset: 128}, 2, 130, "PRJ-129"},
		{"priority before page", work.Request{Priority: "critical", Order: workflowstate.TicketOrderPriority}, 1, 1, "PRJ-130"},
		{"empty", work.Request{Offset: 256}, 0, 130, ""},
	} {
		t.Run(item.name, func(t *testing.T) {
			result, err := newTUIAgentWorkRefreshExecutor(operations).RefreshAgentWork(tui.AgentWorkRefreshRequest{Queue: item.request})
			if err != nil || result.Observation == nil {
				t.Fatal(result, err)
			}

			checkTUIDashPage(t, result, item.count, item.total, item.key)
		})
	}
}

func checkTUIDashObservation(t *testing.T, operations service) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	executor := newTUIAgentWorkRefreshExecutor(operations)
	for _, offset := range []int{0, 128} {
		observed, err := operations.ObserveWork(t.Context(), offset)
		if err != nil {
			t.Fatal(err)
		}

		refreshed, err := executor.RefreshAgentWork(tui.AgentWorkRefreshRequest{Queue: work.Request{Offset: offset}})
		if err != nil || refreshed.Observation == nil {
			t.Fatal(refreshed, err)
		}

		checkDashObservationParity(t, observed, *refreshed.Observation)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("observation mutated canonical work")
	}
}

func seedDashQueue(t *testing.T, operations service) {
	t.Helper()

	for index := 1; index <= 130; index++ {
		priority := "normal"
		if index == 130 {
			priority = "critical"
		}

		path := filepath.Join(operations.projectContext.ProjectDir, "ops/default/ticket", fmt.Sprintf("PRJ-%d.md", index))
		writeServiceWorkspaceFile(t, path, fmt.Sprintf("---\nkey: PRJ-%d\ntitle: Queued task %d\nstatus: ready_for_dev\npriority: %s\n---\nRetained task.\n", index, index, priority))
	}
}

func checkTUIDashPage(t *testing.T, result tui.AgentWorkRefreshResult, count, total int, key string) {
	if len(result.Tickets) != 130 {
		t.Fatal("queue pagination truncated the Board", len(result.Tickets))
	}

	t.Helper()

	observed := result.Observation
	if len(observed.Queue) != count || observed.QueueTotal != total {
		t.Fatal(observed)
	}

	if key != "" && (observed.Queue[0].Key != key || len(observed.Queue[0].Revision) != 64) {
		t.Fatal(observed.Queue[0])
	}

	model := tui.NewModel(tui.Options{Tickets: result.Tickets, WorkObservation: observed, WorkReadiness: result.Readiness})
	runGuidedTUIKey(t, model, "f7", false)

	if key != "" && !strings.Contains(model.View(), key) {
		t.Fatal(model.View())
	}
}
