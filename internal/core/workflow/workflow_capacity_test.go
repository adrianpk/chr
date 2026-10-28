package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowCapacityBounds rejects malformed configuration instead of falling
// back to a larger or separate pool; defaults and exact bounds remain usable.
func TestWorkflowCapacityBounds(t *testing.T) {
	cases := []struct {
		name, fields             string
		workers, queue, prepared int
		fail                     bool
	}{
		{"defaults", "", 1, 64, 2, false},
		{"minimum", "max_active_workers: 1\n  ready_queue_capacity: 1", 1, 1, 1, false},
		{"maximum", "max_active_workers: 8\n  ready_queue_capacity: 1024", 8, 1024, 2, false},
		{"prepared lower", "integration.max_prepared_per_target: 1", 1, 64, 1, false},
		{"prepared maximum", "integration.max_prepared_per_target: 8", 1, 64, 8, false},
		{"prepared admitted bound", "ready_queue_capacity: 1\n  integration.max_prepared_per_target: 8", 1, 1, 1, false},
		{"zero workers", "max_active_workers: 0", 0, 0, 0, true},
		{"negative workers", "max_active_workers: -1", 0, 0, 0, true},
		{"oversized workers", "max_active_workers: 9", 0, 0, 0, true},
		{"malformed workers", "max_active_workers: two", 0, 0, 0, true},
		{"zero queue", "ready_queue_capacity: 0", 0, 0, 0, true},
		{"oversized queue", "ready_queue_capacity: 1025", 0, 0, 0, true},
		{"malformed queue", "ready_queue_capacity: 64.5", 0, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeWorkspaceConfig(t, dir, "work:\n  "+tc.fields+"\n")

			limits, err := workflowCapacityConfig(dir)
			if (err != nil) != tc.fail || !tc.fail && limits != (workflowCapacityLimits{Workers: tc.workers, Queue: tc.queue, Prepared: tc.prepared}) {
				t.Fatalf("capacity = %+v, %v", limits, err)
			}
		})
	}
}

// TestWorkflowCapacityQueue exercises actual durable admission: overflow and
// duplicate attachment cannot discard accepted work or reset queue ownership.
func TestWorkflowCapacityQueue(t *testing.T) {
	store, first, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("true"))
	second := workflowLooseCreate(t, store, first, "PRJ-91", "normal", nil)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  ready_queue_capacity: 1\n")

	admitted := workflowTestStart(t, store, request).Invocation
	request.RequestKey = "second-ticket"

	request.Selection.Selector.Identity = second.Task.Key
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowQueueFull) {
		t.Fatal("queue overflow admitted another lane", err)
	}

	if retained := workflowTestRead(t, store, admitted.ID); !reflect.DeepEqual(retained, admitted) {
		t.Fatal("queue overflow changed accepted work")
	}

	request.RequestKey, request.Selection = admitted.RequestKey, admitted.Selection
	if retry := workflowTestStart(t, store, request); !retry.Attached || retry.Invocation.ID != admitted.ID {
		t.Fatal("full queue prevented idempotent attachment")
	}

	if _, err := store.StopWorkflow(t.Context(), admitted.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	request.RequestKey, request.Selection.Selector.Identity = "after-cancel", second.Task.Key
	if next := workflowTestStart(t, store, request); next.Attached || next.Invocation.ID == admitted.ID {
		t.Fatal("terminal lane retained queue capacity")
	}
}

// TestWorkflowCapacityQueueSlices counts integration lanes, not the ordered
// tasks within them; the configured bound still rejects larger frozen extents.
func TestWorkflowCapacityQueueSlices(t *testing.T) {
	for _, tc := range []struct {
		name  string
		queue int
		fail  bool
	}{{"two lanes", 2, false}, {"one lane", 1, true}} {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			writeWorkspaceConfig(t, store.layout.ProjectDir(), fmt.Sprintf("work:\n  ready_queue_capacity: %d\n", tc.queue))

			request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")

			result, err := store.StartWorkflow(t.Context(), request)
			if (err != nil) != tc.fail || tc.fail && !errors.Is(err, ErrWorkflowQueueFull) {
				t.Fatal("slice queue admission", err)
			}

			if !tc.fail && len(result.Invocation.Members) <= tc.queue {
				t.Fatal("fixture did not distinguish tasks from lanes")
			}
		})
	}
}

// TestWorkflowCapacityAliases rejects conflicting policy across linked roots.
// A project alias and ops namespace cannot create another capacity pool.
func TestWorkflowCapacityAliases(t *testing.T) {
	store, _, _ := workflowClaimFixture(t)
	linkedPath := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "capacity-linked", linkedPath, "HEAD")
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: linkedPath, Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	linked := NewWorkflowStore(layout, "alias")
	if _, err := linked.acquireWorkflowWorkerSlot(t.Context()); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("alias accepted conflicting capacity", err)
	}

	writeWorkspaceConfig(t, linkedPath, "work:\n  max_active_workers: 2\n")

	first, firstGuard := workflowTestSlot(t, store)

	second, secondGuard := workflowTestSlot(t, linked)
	if first.Name() == second.Name() {
		t.Fatal("independent reservations shared one slot")
	}

	if _, err := store.acquireWorkflowWorkerSlot(t.Context()); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("alias created another pool", err)
	}

	if err := errors.Join(first.Close(), firstGuard.Close(), second.Close(), secondGuard.Close()); err != nil {
		t.Fatal(err)
	}
}

func workflowTestSlot(t *testing.T, store *Store) (*os.File, *os.File) {
	t.Helper()

	reservation, err := store.acquireWorkflowWorkerSlot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	slot, guard := reservation.Slot, reservation.Guard

	t.Cleanup(func() {
		for _, file := range []*os.File{slot, guard} {
			if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				t.Error(err)
			}
		}
	})

	return slot, guard
}

// TestWorkflowCapacityReduction counts live reservations above a lowered limit,
// without terminating them or admitting replacement work too soon.
func TestWorkflowCapacityReduction(t *testing.T) {
	store, _, _ := workflowClaimFixture(t)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

	first, firstGuard := workflowTestSlot(t, store)
	second, secondGuard := workflowTestSlot(t, store)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 1\n")

	for step := 0; step < 2; step++ {
		if _, err := store.acquireWorkflowWorkerSlot(t.Context()); !errors.Is(err, ErrWorkflowCapacity) {
			t.Fatal("reduction ignored a live slot", err)
		}

		if step == 0 {
			if err := errors.Join(first.Close(), firstGuard.Close()); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := errors.Join(second.Close(), secondGuard.Close()); err != nil {
		t.Fatal(err)
	}

	slot, guard := workflowTestSlot(t, store)
	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowCapacityLater refuses later work above the repository queue bound
// after confirmed integration, preserving the already admitted immutable facts.
func TestWorkflowCapacityLater(t *testing.T) {
	f := workflowLooseConfiguredQueue(t, true, func(store *Store, _ *StartRequest) {
		writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  ready_queue_capacity: 3\n")
	})
	workflowFinalizeFrontier(t, f)

	before := workflowAdvanceFrontier(t, f)
	original := workflowLooseRecord(t, f.store, "PRJ-91")
	workflowLooseCreate(t, f.store, original, "PRJ-93", "normal", nil)
	workflowLooseCreate(t, f.store, original, "PRJ-94", "normal", nil)

	if _, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID); !errors.Is(err, ErrWorkflowQueueFull) {
		t.Fatal("later admission exceeded the queue", err)
	}

	after := workflowTestRead(t, f.store, f.run.ID)
	if !reflect.DeepEqual(before.Members, after.Members) || before.ConfirmedCursor != after.ConfirmedCursor {
		t.Fatal("overflow replaced the accepted frontier")
	}
}
