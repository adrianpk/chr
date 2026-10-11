package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func nativeEntryFixture(t *testing.T) (*WorkflowStore, WorkflowIntegrationAdmission) {
	t.Helper()

	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("true"))
	run := workflowTestStart(t, store, request).Invocation

	return store, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: WorkflowRef{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}
}

func nativeEntryAdmit(t *testing.T, store *WorkflowStore, request WorkflowIntegrationAdmission) WorkflowIntegrationEntry {
	t.Helper()

	entry, err := store.AdmitWorkflowIntegration(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return entry
}

func nativeEntryTransition(t *testing.T, store *WorkflowStore, entry WorkflowIntegrationEntry, next string) WorkflowIntegrationTransition {
	t.Helper()

	run := workflowTestRead(t, store, entry.InvocationID)

	request := WorkflowIntegrationTransition{InvocationID: entry.InvocationID, EntryID: entry.ID, InvocationRevision: run.Revision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: next}
	if entry.CandidateRef != nil {
		request.CandidateHash = entry.CandidateRef.Hash
	}

	return request
}

// TestNativeEntryRetries exercises concurrent stores against the actual common
// lock: one queue allocation and invocation transition survive all retries.
func TestNativeEntryRetries(t *testing.T) {
	store, request := nativeEntryFixture(t)
	entries, failures := make([]WorkflowIntegrationEntry, 8), make([]error, 8)

	var pending sync.WaitGroup
	for i := range entries {
		pending.Go(func() {
			peer := NewWorkflowStore(store.layout, store.projectID)
			entries[i], failures[i] = peer.AdmitWorkflowIntegration(t.Context(), request)
		})
	}

	pending.Wait()

	for i, entry := range entries {
		if failures[i] != nil || !reflect.DeepEqual(entry, entries[0]) || entry.QueueSequence != 1 || entry.Revision != 1 {
			t.Fatalf("retry %d duplicated or changed admission: %+v, %v", i, entry, failures[i])
		}
	}

	checkNativeEntryAdmission(t, store, request, entries[0])
}

func checkNativeEntryAdmission(t *testing.T, store *WorkflowStore, request WorkflowIntegrationAdmission, entry WorkflowIntegrationEntry) {
	t.Helper()

	run := workflowTestRead(t, store, request.InvocationID)
	if run.Revision != request.ExpectedRevision+1 || len(run.IntegrationEntries) != 1 || run.Integration == nil || run.Integration.EntryID != entry.ID || len(run.Attempts) != 0 {
		t.Fatal("duplicate admission changed invocation ownership or consumption")
	}
}

// TestNativeEntryRecovery interrupts both images of admission and transition;
// the retained workflow journal replays each image once without another entry.
func TestNativeEntryRecovery(t *testing.T) {
	for _, action := range []string{"admission", "transition"} {
		for _, index := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s %d", action, index), func(t *testing.T) {
				nativeEntryRecover(t, action, index)
			})
		}
	}
}

func nativeEntryRecover(t *testing.T, action string, index int) {
	t.Helper()

	store, request := nativeEntryFixture(t)
	transition := WorkflowIntegrationTransition{}

	if action == "transition" {
		entry := nativeEntryAdmit(t, store, request)
		transition = nativeEntryTransition(t, store, entry, "preparing")
	}

	crash := errors.New("entry transaction interrupted")
	store.afterReplace = func(at int) error {
		if at == index {
			return crash
		}

		return nil
	}

	var err error
	if action == "admission" {
		_, err = store.AdmitWorkflowIntegration(t.Context(), request)
	} else {
		_, err = store.TransitionWorkflowIntegration(t.Context(), transition)
	}

	if !errors.Is(err, crash) {
		t.Fatal("fault did not interrupt the durable transition", err)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)
	entry := nativeEntryAdmit(t, restarted, request)
	checkNativeEntryReplay(t, restarted, request, entry, action)
}

func checkNativeEntryReplay(t *testing.T, store *WorkflowStore, request WorkflowIntegrationAdmission, entry WorkflowIntegrationEntry, action string) {
	t.Helper()

	expectedRevision := uint64(1)
	if action == "transition" {
		expectedRevision++

		if entry.State != "preparing" || entry.ClaimGeneration != 1 {
			t.Fatal("replay lost phase or fencing generation")
		}
	}

	run := workflowTestRead(t, store, request.InvocationID)
	if entry.QueueSequence != 1 || entry.Revision != expectedRevision || len(run.IntegrationEntries) != 1 || run.Revision != request.ExpectedRevision+expectedRevision {
		t.Fatal("replay duplicated allocation or revision")
	}

	if _, err := os.Stat(filepath.Join(store.layout.Root(), workflowJournalPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("completed replay retained a transition journal", err)
	}
}

// TestNativeEntryAdmission rejects a changed retry target and future set unit
// without touching the retained invocation or its frozen membership.
func TestNativeEntryAdmission(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
	run := workflowTestStart(t, store, request).Invocation
	admission := WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: WorkflowRef{Kind: "slice", ID: hierarchy.Set.Set.Slices[0].ID}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}
	first := nativeEntryAdmit(t, store, admission)
	before := workflowTestRead(t, store, run.ID)

	for _, tc := range []struct {
		name   string
		change func(*WorkflowIntegrationAdmission)
	}{
		{"target", func(r *WorkflowIntegrationAdmission) { r.TargetRef = "refs/heads/other" }},
		{"key", func(r *WorkflowIntegrationAdmission) { r.AdmissionKey = "foreign" }},
		{"future", func(r *WorkflowIntegrationAdmission) { r.Unit.ID = hierarchy.Set.Set.Slices[1].ID }},
		{"unit", func(r *WorkflowIntegrationAdmission) { r.Unit.Kind = "ticket" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := admission
			tc.change(&changed)

			if _, err := store.AdmitWorkflowIntegration(t.Context(), changed); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("incompatible admission accepted", err)
			}
		})
	}

	if after := workflowTestRead(t, store, run.ID); !reflect.DeepEqual(before, after) || after.Integration.EntryID != first.ID {
		t.Fatal("rejected admission changed accepted history")
	}
}

// TestNativeEntryFences proves stale invocation, entry, claim and candidate
// callbacks cannot change a retained phase or release its ownership.
func TestNativeEntryFences(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	request := nativeEntryTransition(t, store, entry, "preparing")

	entry, err := store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil || entry.StateRevision != 2 || entry.ClaimGeneration != 1 {
		t.Fatal("cannot start guarded preparation", err)
	}

	request = nativeEntryTransition(t, store, entry, "queued")

	before := workflowTestRead(t, store, admission.InvocationID)
	for _, tc := range []struct {
		name   string
		change func(*WorkflowIntegrationTransition)
	}{
		{"invocation", func(r *WorkflowIntegrationTransition) { r.InvocationRevision-- }},
		{"entry", func(r *WorkflowIntegrationTransition) { r.EntryRevision-- }},
		{"generation", func(r *WorkflowIntegrationTransition) { r.ClaimGeneration-- }},
		{"candidate", func(r *WorkflowIntegrationTransition) { r.CandidateHash = workflowHash([]byte("stale")) }},
		{"phase", func(r *WorkflowIntegrationTransition) { r.State = "integrated" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := request
			tc.change(&changed)

			if _, err := store.TransitionWorkflowIntegration(t.Context(), changed); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("stale callback accepted", err)
			}
		})
	}

	if after := workflowTestRead(t, store, admission.InvocationID); !reflect.DeepEqual(after, before) {
		t.Fatal("stale callback changed durable state")
	}
}

func nativeEntryNextNamespace(t *testing.T, store *WorkflowStore) (*WorkflowStore, WorkflowIntegrationAdmission) {
	t.Helper()

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: store.layout.ProjectDir(), Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	peer := NewWorkflowStore(layout, store.projectID)

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	record := snapshot.Records[0]
	record.WorkflowIdentity, record.Path, record.SourceHash = WorkflowIdentity{}, "", ""
	record.Task.Key = "PRJ-91"

	record, err = peer.Create(t.Context(), "Other namespace task", record)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, peer)
	run := workflowTestStart(t, peer, WorkflowStartRequest{RequestKey: "other-unit", Selection: WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "ticket", Identity: record.Task.Key}, Target: WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}}}).Invocation

	return peer, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: WorkflowRef{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}
}

// TestNativeEntrySequence shares one coordinator order across namespaces and
// preserves admitted bytes when capacity or allocation arithmetic is exhausted.
func TestNativeEntrySequence(t *testing.T) {
	for _, limit := range []string{"none", "capacity", "sequence"} {
		t.Run(limit, func(t *testing.T) {
			nativeEntrySequence(t, limit)
		})
	}
}

func nativeEntrySequence(t *testing.T, limit string) {
	t.Helper()

	store, request := nativeEntryFixture(t)
	first := nativeEntryAdmit(t, store, request)
	peer, second := nativeEntryNextNamespace(t, store)

	switch limit {
	case "capacity":
		writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  ready_queue_capacity: 1\n")
	case "sequence":
		nativeEntrySetSequence(t, store, &first, math.MaxUint64)
	}

	before := workflowTestRead(t, store, request.InvocationID)
	next, err := peer.AdmitWorkflowIntegration(t.Context(), second)
	checkNativeEntrySequence(t, limit, first, next, err)

	if retry := nativeEntryAdmit(t, store, request); !reflect.DeepEqual(retry, first) || !reflect.DeepEqual(before, workflowTestRead(t, store, request.InvocationID)) {
		t.Fatal("allocation or retry changed the previous accepted entry")
	}
}

func checkNativeEntrySequence(t *testing.T, limit string, first, next WorkflowIntegrationEntry, err error) {
	t.Helper()

	switch limit {
	case "none":
		if err != nil || next.QueueSequence != first.QueueSequence+1 || next.CoordinatorID != first.CoordinatorID {
			t.Fatal("namespace allocated another queue order", err)
		}
	case "capacity":
		if !errors.Is(err, ErrWorkflowQueueFull) {
			t.Fatal("overflow accepted another entry", err)
		}
	case "sequence":
		if !errors.Is(err, ErrWorkflowConflict) {
			t.Fatal("sequence overflow wrapped allocation", err)
		}
	}
}

func nativeEntrySetSequence(t *testing.T, store *WorkflowStore, entry *WorkflowIntegrationEntry, sequence uint64) {
	t.Helper()

	run, err := store.readWorkflowInvocation(entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	entry.QueueSequence = sequence
	if err := store.commitWorkflowIntegrationEntry(&run, *entry, run.Integration.EntryRef.Hash); err != nil {
		t.Fatal(err)
	}
}

// TestNativeEntryBlocker admits only the reason's explicit resume phase and
// preserves optional extensions through actual guarded replacement.
func TestNativeEntryBlocker(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	entry.Extensions = map[string]json.RawMessage{"future_optional": json.RawMessage(`{"retained":true}`)}
	nativeEntrySetSequence(t, store, &entry, entry.QueueSequence)

	request := nativeEntryTransition(t, store, entry, "blocked")
	request.Blocker = &WorkflowIntegrationBlocker{Reason: "operator_reconciliation", RequiredAction: "confirm current input", ResumeState: "queued"}

	entry, err := store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	wrong := nativeEntryTransition(t, store, entry, "preparing")
	if _, err := store.TransitionWorkflowIntegration(t.Context(), wrong); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("blocked entry resumed outside its reason", err)
	}

	request = nativeEntryTransition(t, store, entry, "queued")

	entry, err = store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil || entry.Blocker != nil || !workflowNativeJSONEqual(entry.Extensions, map[string]json.RawMessage{"future_optional": json.RawMessage(`{"retained":true}`)}) {
		t.Fatal("resume lost explicit blocker or optional extension", err)
	}
}

// TestNativeEntryPause requires an explicit guarded resume and keeps the
// original admission and consumption when a stopped unit becomes eligible.
func TestNativeEntryPause(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	request := nativeEntryTransition(t, store, entry, "paused")
	request.StopIntent = "pause"

	entry, err := store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil || entry.StopIntent != "pause" {
		t.Fatal("pause lost its durable intent", err)
	}

	request = nativeEntryTransition(t, store, entry, "queued")
	if _, err := store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("stopped entry resumed implicitly", err)
	}

	request.ClearStopIntent = true

	entry, err = store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil || entry.StopIntent != "" || entry.QueueSequence != 1 || entry.ClaimGeneration != 0 {
		t.Fatal("explicit resume changed admission or invented consumption", err)
	}

	request = nativeEntryTransition(t, store, entry, "preparing")

	request.ClearStopIntent = true
	if _, err := store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("non-quiescent transition cleared stop intent", err)
	}
}
