package core

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeSchedulePage(t *testing.T, store *WorkflowStore, q WorkflowIntegrationScheduleQuery) WorkflowIntegrationSchedulePage {
	t.Helper()

	page, err := store.QueryWorkflowIntegrationSchedule(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	return page
}

func nativeScheduleClaim(t *testing.T, store *WorkflowStore, q WorkflowIntegrationScheduleQuery) *WorkflowClaim {
	t.Helper()

	claim, err := store.ClaimNextWorkflowIntegration(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.releaseReservation(); err != nil {
			t.Error(err)
		}
	})

	return claim
}

// TestNativeScheduleOrder uses real namespace records and sequence allocation;
// read-only pagination survives restart and cannot export an owner credential.
func TestNativeScheduleOrder(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	first := nativeEntryAdmit(t, store, admission)

	peer, other := nativeEntryNextNamespace(t, store)
	second := nativeEntryAdmit(t, peer, other)
	before, peerBefore := nativeQueueImages(t, store.layout.Root()), nativeQueueImages(t, peer.layout.Root())

	page := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 1})
	if len(page.Items) != 1 || page.Items[0].Entry.ID != first.ID || page.NextOffset != 1 {
		t.Fatal("schedule lost durable admission order", page)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	next := nativeSchedulePage(t, restarted, WorkflowIntegrationScheduleQuery{Limit: 1, Offset: 1, SnapshotHash: page.SnapshotHash})
	if len(next.Items) != 1 || next.Items[0].Entry.ID != second.ID || next.NextOffset != -1 {
		t.Fatal("schedule pagination changed after restart", next)
	}

	checkNativeScheduleObservation(t, store, peer, before, peerBefore)
	checkNativeScheduleCredential(t, store, page, first.InvocationID)
}

func checkNativeScheduleObservation(t *testing.T, store, peer *WorkflowStore, before, peerBefore map[string]string) {
	t.Helper()

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) || !reflect.DeepEqual(peerBefore, nativeQueueImages(t, peer.layout.Root())) {
		t.Fatal("selection changed authoritative records")
	}
}

func checkNativeScheduleCredential(t *testing.T, store *WorkflowStore, page WorkflowIntegrationSchedulePage, id string) {
	t.Helper()

	claim := nativeScheduleClaim(t, store, WorkflowIntegrationScheduleQuery{Limit: 1, SnapshotHash: page.SnapshotHash})

	data, err := json.Marshal(page)
	if err != nil || strings.Contains(string(data), claim.Token) || strings.Contains(string(data), "owner_token_hashes") {
		t.Fatal("schedule leaked an execution credential", err)
	}

	if claim.InvocationID != id {
		t.Fatal("claim selected another queue head")
	}
}

// TestNativeScheduleWaits proves that a paused or human-blocked unit retains
// its records while an independent later namespace obtains a real task claim.
func TestNativeScheduleWaits(t *testing.T) {
	for _, state := range []string{"paused", "blocked"} {
		t.Run(state, func(t *testing.T) { nativeScheduleWait(t, state) })
	}
}

func nativeScheduleWait(t *testing.T, state string) {
	t.Helper()

	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)

	peer, other := nativeEntryNextNamespace(t, store)
	nativeEntryAdmit(t, peer, other)

	stale := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 10})

	transition := nativeEntryTransition(t, store, entry, state)
	if state == "paused" {
		transition.StopIntent = "pause"
	} else {
		transition.Blocker = &WorkflowIntegrationBlocker{Reason: "waiting_review", RequiredAction: "approve the unit", ResumeState: "queued"}
	}

	if _, err := store.TransitionWorkflowIntegration(t.Context(), transition); err != nil {
		t.Fatal(err)
	}

	checkNativeScheduleWait(t, store, peer, entry, other, stale, state)
}

func checkNativeScheduleWait(t *testing.T, store, peer *WorkflowStore, entry WorkflowIntegrationEntry, other WorkflowIntegrationAdmission, stale WorkflowIntegrationSchedulePage, state string) {
	t.Helper()

	if _, err := store.ClaimWorkflow(t.Context(), entry.InvocationID); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("direct claim bypassed native pause or blocker", err)
	}

	if _, err := store.ClaimNextWorkflowIntegration(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10, SnapshotHash: stale.SnapshotHash}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale selection acquired an owner", err)
	}

	before := workflowTestRead(t, store, entry.InvocationID)

	page := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 10})
	if len(page.Items) != 1 || page.Items[0].Entry.InvocationID != other.InvocationID || page.Excluded[state] != 1 {
		t.Fatal("waiting queue head blocked independent work", page)
	}

	claim := nativeScheduleClaim(t, store, WorkflowIntegrationScheduleQuery{Limit: 10, SnapshotHash: page.SnapshotHash})
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, workflowTestRead(t, store, entry.InvocationID)) || workflowTestRead(t, peer, other.InvocationID).Cursor != 1 {
		t.Fatal("independent progression resumed the waiting unit")
	}
}

// TestNativeScheduleTasks runs real ordered task gates on one isolated lane.
// Finishing the slice cannot schedule the next slice before its receipt barrier.
func TestNativeScheduleTasks(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, workflowPreparationGate("true"))
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
	run := workflowTestStart(t, store, request).Invocation
	admission := WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "slice", ID: hierarchy.Set.Set.Slices[0].ID}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}
	nativeEntryAdmit(t, store, admission)

	for _, record := range hierarchy.Tasks[:3] {
		page := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 10})
		if len(page.Items) != 1 || page.Items[0].Task.Key != record.Task.Key {
			t.Fatal("scheduler reordered tasks inside the unit", page)
		}

		claim := nativeScheduleClaim(t, store, WorkflowIntegrationScheduleQuery{Limit: 10, SnapshotHash: page.SnapshotHash})
		if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker")); err != nil {
			t.Fatal(err)
		}
	}

	checkNativeScheduleReceiptBarrier(t, store, hierarchy, run.ID)
}

func checkNativeScheduleReceiptBarrier(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, id string) {
	t.Helper()

	before := workflowTestRead(t, store, id)

	page := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 10})
	if len(page.Items) != 0 || before.Cursor != 3 || before.ConfirmedCursor != 0 || before.Status != "waiting_integration" {
		t.Fatal("unreceipted slice exposed a later task", page)
	}

	if _, err := store.ClaimNextWorkflowIntegration(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("scheduler crossed the receipt barrier", err)
	}

	next := WorkflowIntegrationAdmission{InvocationID: id, ExpectedRevision: before.Revision, Unit: workflowstate.Ref{Kind: "slice", ID: hierarchy.Set.Set.Slices[1].ID}, AdmissionKey: before.RequestKey, TargetRef: "refs/heads/dev"}
	if _, err := store.AdmitWorkflowIntegration(t.Context(), next); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("future slice became an admitted unit", err)
	}

	if !reflect.DeepEqual(before, workflowTestRead(t, store, id)) {
		t.Fatal("blocked scheduling changed the retained frontier")
	}
}

// TestNativeSchedulePrerequisite rejects a later task even when its queue entry
// precedes other work; the prerequisite reader requires actual cumulative proof.
func TestNativeSchedulePrerequisite(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
	run := workflowTestStart(t, store, request).Invocation
	admission := WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "slice", ID: hierarchy.Set.Set.Slices[0].ID}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}
	nativeEntryAdmit(t, store, admission)

	before := nativeQueueImages(t, store.layout.Root())

	page := nativeSchedulePage(t, store, WorkflowIntegrationScheduleQuery{Limit: 10})
	if len(page.Items) != 0 || page.Excluded["task_prerequisite"] != 1 {
		t.Fatal("queue rank bypassed task prerequisites", page)
	}

	if _, err := store.ClaimNextWorkflowIntegration(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("unproven prerequisite acquired an owner", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("prerequisite wait manufactured execution state")
	}
}

// TestNativeScheduleNoReplay keeps another namespace's interrupted transition
// intact while selection reports reconciliation instead of replaying it.
func TestNativeScheduleNoReplay(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	nativeEntryAdmit(t, store, admission)

	peer, other := nativeEntryNextNamespace(t, store)
	entry := nativeEntryAdmit(t, peer, other)
	crash := errors.New("interrupted peer transition")
	peer.afterReplace = func(int) error { return crash }

	_, err := peer.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, peer, entry, "preparing"))
	if !errors.Is(err, crash) {
		t.Fatal(err)
	}

	before := nativeQueueImages(t, peer.layout.Root())
	if _, err := store.QueryWorkflowIntegrationSchedule(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("selection ignored an interrupted peer", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, peer.layout.Root())) {
		t.Fatal("read-only scheduling replayed peer mutation")
	}
}

// TestNativeScheduleInputs retains changed source bytes and blocks claiming;
// neither pagination nor direct native claim turns stale admission into work.
func TestNativeScheduleInputs(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	nativeEntryAdmit(t, store, admission)

	run := workflowTestRead(t, store, admission.InvocationID)

	record, err := store.Read(t.Context(), run.Members[0].Admission.Path)
	if err != nil {
		t.Fatal(err)
	}

	record.Body = []byte("Changed frozen task body")
	workflowTestWrite(t, store, record)

	before := nativeQueueImages(t, store.layout.Root())

	_, err = store.QueryWorkflowIntegrationSchedule(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10})
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed admitted input became selectable", err)
	}

	if _, err := store.ClaimNextWorkflowIntegration(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("scheduler claimed stale task input", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("stale source was normalized during selection")
	}
}
