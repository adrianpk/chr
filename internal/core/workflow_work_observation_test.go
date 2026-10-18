package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// Native observation retains shared ordering/readiness and ignores unrelated
// legacy Board failures. Concurrent readers cannot modify authoritative bytes.
func TestWorkflowWorkObservation(t *testing.T) {
	store, hierarchy := workflowSelectionFixture(t)
	expected := workflowTestPreview(t, store, "slice_set", hierarchy.Set.ID)

	legacy := filepath.Join(ticketsDirectory(store.layout.ProjectDir()), "broken.md")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(legacy, []byte("---\nkey: [broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	before := nativeQueueImages(t, store.layout.ProjectDir())

	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			observed, err := store.InspectWorkflowWork(t.Context(), WorkflowPreviewTarget{})
			if err != nil || !reflect.DeepEqual(observed.Queue, expected.Members) {
				t.Errorf("observation differs from shared selection: %v", err)
			}
		})
	}

	group.Wait()

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.ProjectDir())) {
		t.Fatal("observation mutated retained state")
	}
}

// A pending journal is unavailable state; observation must not recover it.
func TestWorkflowWorkPending(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowTestCreate(t, store, "PRJ-1")
	store.afterReplace = func(int) error { return errors.New("interrupted replacement") }

	record.Task.Status = TicketStatusBacklog
	if _, err := store.Write(t.Context(), record, record.Revision); err == nil {
		t.Fatal("interruption not injected")
	}

	before := nativeQueueImages(t, store.layout.ProjectDir())
	if _, err := store.InspectWorkflowWork(t.Context(), WorkflowPreviewTarget{}); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.ProjectDir())) {
		t.Fatal("read recovered a transition")
	}
}
