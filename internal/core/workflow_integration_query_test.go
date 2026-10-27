package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeQueuePage(t *testing.T, store *WorkflowStore, q WorkflowIntegrationQueueQuery) WorkflowIntegrationQueuePage {
	t.Helper()

	page, err := store.QueryWorkflowIntegrationQueue(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	return page
}

func nativeQueueImages(t *testing.T, root string) map[string]string {
	t.Helper()

	images := map[string]string{}

	err := filepath.WalkDir(root, func(path string, file os.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		if err != nil {
			return err
		}

		if file.Type().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			images[path] = contenthash.Sum(data)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return images
}

// TestNativeQueueQueries checks coordinator ordering, sealed budgets, filtering
// and stale pagination across namespaces without changing authoritative bytes.
func TestNativeQueueQueries(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	first := nativeEntryAdmit(t, store, admission)
	peer, next := nativeEntryNextNamespace(t, store)
	second := nativeEntryAdmit(t, peer, next)
	before := nativeQueueImages(t, store.layout.Root())
	q := WorkflowIntegrationQueueQuery{Limit: 1, EligibleOnly: true}

	page := nativeQueuePage(t, store, q)
	if len(page.Entries) != 1 || page.Entries[0].Entry.ID != first.ID || page.NextOffset != 1 {
		t.Fatal("queue did not preserve authoritative sequence")
	}

	q.Offset, q.SnapshotHash = page.NextOffset, page.SnapshotHash

	page = nativeQueuePage(t, peer, q)
	if len(page.Entries) != 1 || page.Entries[0].Entry.ID != second.ID || page.NextOffset != -1 {
		t.Fatal("namespace query lost coordinator history")
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("query changed source bytes")
	}

	nativeQueueStalePage(t, store, first, second, q)
}

func nativeQueueStalePage(t *testing.T, store *WorkflowStore, first, second WorkflowIntegrationEntry, q WorkflowIntegrationQueueQuery) {
	t.Helper()

	_, err := store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, first, "preparing"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.QueryWorkflowIntegrationQueue(t.Context(), q); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale pagination returned runnable work", err)
	}

	page := nativeQueuePage(t, store, WorkflowIntegrationQueueQuery{Namespace: "other", Limit: 10})
	if len(page.Entries) != 1 || page.Entries[0].Budget.UnitCounters[0].Unit != second.Unit {
		t.Fatal("queue query lost owning budget")
	}
}

// TestNativeQueueNoReplay retains an interrupted namespace transition while
// every query and both projection rebuild paths report reconciliation required.
func TestNativeQueueNoReplay(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	crash := errors.New("retained transition")
	store.afterReplace = func(int) error { return crash }

	_, err := store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "preparing"))
	if !errors.Is(err, crash) {
		t.Fatal(err)
	}

	store.afterReplace = nil
	index := workflowTestIndex(t, ":memory:")

	native, ok := index.(WorkflowIntegrationIndex)
	if !ok {
		t.Fatal("index lacks native projection")
	}

	before := nativeQueueImages(t, store.layout.Root())

	checks := []struct {
		name  string
		query func() error
	}{
		{"queue", func() error {
			_, err := store.QueryWorkflowIntegrationQueue(t.Context(), WorkflowIntegrationQueueQuery{Limit: 10})

			return err
		}},
		{"native index", func() error { return native.RebuildWorkflowIntegrationIndex(t.Context(), store) }},
		{"workflow index", func() error { return index.RebuildWorkflowIndex(t.Context(), store) }},
		{"capacity", func() error { _, err := store.QueryWorkflowIntegrationCapacity(t.Context()); return err }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.query(); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("query replayed transition", err)
			}

			if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
				t.Fatal("query changed pending authority")
			}
		})
	}
}

// TestNativeQueueCorruption reseals invalid sources to prove graph validation,
// rather than an obsolete file hash, blocks mutation and preserves their bytes.
func TestNativeQueueCorruption(t *testing.T) {
	for _, damage := range []string{"budget generation", "missing counter", "token grant", "orphan counter", "invocation project", "orphan entry"} {
		t.Run(damage, func(t *testing.T) { nativeQueueCorruption(t, damage) })
	}
}

func nativeQueueCorruption(t *testing.T, damage string) {
	t.Helper()

	store, entry := nativeBudgetFixture(t, nil)
	guard := nativeBudgetGuard(t, store, entry.InvocationID)

	request := WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: guard, Reservation: nativeBudgetReservation(entry, "call-1", "executor", 10, 10)}
	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	nativeQueueDamage(t, store, entry.InvocationID, damage)

	before := nativeQueueImages(t, store.layout.Root())
	_, err := store.QueryWorkflowIntegrationQueue(t.Context(), WorkflowIntegrationQueueQuery{Limit: 10})

	var blocker *WorkflowIntegrationRecordError
	if !errors.As(err, &blocker) {
		t.Fatal("invalid authority lacks typed blocker", err)
	}

	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err == nil {
		t.Fatal("invalid authority allowed mutation")
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("failure normalized source bytes")
	}
}

func nativeQueueDamage(t *testing.T, store *WorkflowStore, id, damage string) {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	if nativeQueueDamageOwner(t, store, run, damage) {
		return
	}

	budget, err := store.readWorkflowIntegrationBudget(run)
	if err != nil {
		t.Fatal(err)
	}

	nativeQueueDamageBudget(&budget, damage)

	data, err := json.Marshal(budget)
	if err != nil {
		t.Fatal(err)
	}

	change := workflowFileChange{Path: run.IntegrationBudget.Ref.Path, Before: run.IntegrationBudget.Ref.Hash, After: data}
	run.IntegrationBudget.Ref.Hash = contenthash.Sum(data)

	invocation, err := workflowInvocationChange(run)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.commitFiles([]workflowFileChange{change, invocation}); err != nil {
		t.Fatal(err)
	}
}

func nativeQueueDamageOwner(t *testing.T, store *WorkflowStore, run workflowInvocationFile, damage string) bool {
	t.Helper()

	if damage == "invocation project" {
		run.ProjectID = "foreign"

		data, err := json.Marshal(run)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(store.layout.Root(), workflowInvocationPath(run.ID)), data, 0o600); err != nil {
			t.Fatal(err)
		}

		return true
	}

	if damage == "orphan entry" {
		path := filepath.Join(store.layout.Root(), "run", run.ID, "extension", "chr", "integration", "unbound")
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}

		return true
	}

	return false
}

func nativeQueueDamageBudget(budget *WorkflowIntegrationBudget, damage string) {
	switch damage {
	case "budget generation":
		budget.Reservations[0].ClaimGeneration++
	case "missing counter":
		budget.UnitCounters = nil
	case "orphan counter":
		budget.UnitCounters = append(budget.UnitCounters, WorkflowIntegrationCounter{Unit: WorkflowRef{Kind: "ticket", ID: "unadmitted"}})
	case "token grant":
		budget.Grants = append(budget.Grants, WorkflowIntegrationGrant{ID: "illegal", Kind: "tokens", Amount: 1})
	}
}
