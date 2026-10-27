package core

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeRolePeer(t *testing.T, store *WorkflowStore, namespace, key, target string) (*WorkflowStore, WorkflowIntegrationEntry) {
	t.Helper()

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: store.layout.ProjectDir(), Namespace: namespace})
	if err != nil {
		t.Fatal(err)
	}

	peer := NewWorkflowStore(layout, store.projectID)
	contract := workflowPlanFixture().Slices[0].Tasks[0]
	contract.ID, contract.Label, contract.Validation = "independent", "Independent task", workflowPreparationGate("true")

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	record := WorkflowRecord{Kind: "ticket", Task: &WorkflowTask{Schema: WorkflowSchema, Key: key, Status: TicketStatusReadyForDev, Phase: "pending", ContractHash: contenthash.Sum(append([]byte("\x00\x00"), data...))}, Body: []byte(contract.Body), Domain: map[string]json.RawMessage{"execution_contract": data}}

	record, err = peer.Create(t.Context(), "Independent task", record)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, peer)
	run := workflowTestStart(t, peer, WorkflowStartRequest{RequestKey: namespace, Selection: WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "ticket", Identity: key}, Target: WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}}}).Invocation
	entry := nativeEntryAdmit(t, peer, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: target})

	return peer, entry
}

func nativeRoleAdmit(t *testing.T, store *WorkflowStore, id string) WorkflowIntegrationEntry {
	t.Helper()

	run := workflowTestRead(t, store, id)

	return nativeEntryAdmit(t, store, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"})
}

// TestNativeRolePriority runs real source validation and candidate review. New
// implementation cannot repeatedly take the free slot ahead of either finisher.
func TestNativeRolePriority(t *testing.T) {
	for _, role := range []string{"validation", "preparation", "reviewer"} {
		t.Run(role, func(t *testing.T) { nativeRolePriority(t, role) })
	}
}

func nativeRolePriority(t *testing.T, role string) {
	t.Helper()

	store, id, finish := nativeRolePriorityFixture(t, role)
	entry := nativeRoleAdmit(t, store, id)
	peer, next := nativeRolePeer(t, store, "peer", "PRJ-91", "refs/heads/dev")
	before := nativeQueueImages(t, store.layout.Root())

	page, err := store.QueryWorkflowIntegrationRoles(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10})
	if err != nil || len(page.Items) != 1 || page.Items[0].Role != role || page.Items[0].Entry.ID != entry.ID {
		t.Fatal("finisher was not selected", page, err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("role observation changed authoritative bytes")
	}

	checkNativeRolePriorityWait(t, peer, next.InvocationID)

	if err := finish(); err != nil {
		t.Fatal("finisher could not obtain its shared slot", err)
	}

	claim := nativeScheduleClaim(t, store, WorkflowIntegrationScheduleQuery{Limit: 10, InvocationID: next.InvocationID})
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}
}

func nativeRolePriorityFixture(t *testing.T, role string) (*WorkflowStore, string, func() error) {
	t.Helper()

	if role == "validation" {
		return workflowCapacityAggregateFixture(t)
	}

	f := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	if role == "preparation" {
		return f.store, f.run.ID, func() error {
			_, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)

			return err
		}
	}

	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "review", "mode", "agentic_review")

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err != nil {
		t.Fatal(err)
	}

	return f.store, f.run.ID, func() error {
		_, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, workflowReviewWorker("pass", "", ""))

		return err
	}
}

// TestNativePreparedBound reserves preparation in the durable native graph.
// Pausing and restarting do not free speculative target capacity; an independent
// target can prepare and lowering the bound preserves previously accepted work.
func TestNativePreparedBound(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	first := nativeEntryAdmit(t, store, admission)
	peer, second := nativeRolePeer(t, store, "second", "PRJ-91", "refs/heads/dev")
	third, next := nativeRolePeer(t, store, "third", "PRJ-92", "refs/heads/dev")
	first = nativeRoleTransition(t, store, first, "preparing")
	second = nativeRoleTransition(t, peer, second, "preparing")
	nativeRoleTransition(t, peer, second, "paused")

	before := nativeQueueImages(t, third.layout.Root())

	restarted := NewWorkflowStore(third.layout, third.projectID)
	if _, err := restarted.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, restarted, next, "preparing")); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("third prepared unit bypassed default target limit", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, third.layout.Root())) {
		t.Fatal("overflow altered accepted records")
	}

	independent, other := nativeRolePeer(t, store, "independent", "PRJ-93", "refs/heads/other")
	nativeRoleTransition(t, independent, other, "preparing")
	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "integration.max_prepared_per_target", "1")

	if _, err := restarted.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, restarted, next, "preparing")); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("lowered prepared limit was ignored", err)
	}

	if workflowTestRead(t, store, first.InvocationID).Integration.EntryID != first.ID {
		t.Fatal("lowering capacity removed a retained owner")
	}

	nativeRoleTransition(t, store, first, "cancelled")
	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "integration.max_prepared_per_target", "2")
	nativeRoleTransition(t, restarted, next, "preparing")
}

func nativeRoleTransition(t *testing.T, store *WorkflowStore, entry WorkflowIntegrationEntry, state string) WorkflowIntegrationEntry {
	t.Helper()

	request := nativeEntryTransition(t, store, entry, state)
	if state == "paused" {
		request.StopIntent = "pause"
	}

	result, err := store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// TestNativeRolePool uses the production reservation boundary for every role.
// The default pool permits one owner, and another purpose cannot bypass it.
func TestNativeRolePool(t *testing.T) {
	for _, purpose := range []string{"", workflowMergerPurpose, workflowSliceReviewPurpose, "closing"} {
		t.Run(workflowIntegrationOwnerRole(purpose), func(t *testing.T) { nativeRolePool(t, purpose) })
	}
}

func nativeRolePool(t *testing.T, purpose string) {
	t.Helper()

	store, _, _ := workflowClaimFixture(t)
	firstRun := workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: "role-one"}}}
	secondRun := workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: "role-two"}}}
	// Merger payload and credential fencing remain with its role owner. This
	// boundary certifies its pool reservation using the exact production purpose.
	if purpose == workflowMergerPurpose {
		firstRun.PullRequest = &WorkflowPullRequestBinding{Request: WorkflowPullRequestRequest{WorkflowPullRequestQuery: WorkflowPullRequestQuery{Repository: "https://forge.example/test/repo", BaseBranch: "dev"}}}
	}

	var first *WorkflowClaim

	err := store.withLock(t.Context(), func() error {
		var err error

		first, err = store.reserveWorkflowRole(t.Context(), firstRun, purpose)

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Error(err)
		}
	})

	err = store.withLock(t.Context(), func() error {
		claim, err := store.reserveWorkflowRole(t.Context(), secondRun, "closing")
		if claim != nil {
			err = errors.Join(err, claim.Close())
		}

		return err
	})
	if !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("another role bypassed the serial pool", err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
}

func checkNativeRolePriorityWait(t *testing.T, peer *WorkflowStore, id string) {
	t.Helper()

	for range 4 {
		if _, err := peer.ClaimWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowCapacity) {
			t.Fatal("implementation starved integration", err)
		}
	}

	waiting := workflowTestRead(t, peer, id)
	if waiting.Status != "waiting_capacity" || len(waiting.Attempts) != 0 {
		t.Fatal("priority wait consumed task ownership", waiting)
	}
}

// TestNativeRoleHumanWait retains a real combined candidate and human approval
// wait. It reserves no worker and cannot starve an independent admitted task.
func TestNativeRoleHumanWait(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "review", "mode", "human_review")

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err != nil {
		t.Fatal(err)
	}

	nativeRoleAdmit(t, f.store, f.run.ID)

	binding, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil)
	if err != nil || binding.State != "waiting_approval" {
		t.Fatal("fixture did not retain human approval", binding, err)
	}

	peer, next := nativeRolePeer(t, f.store, "peer", "PRJ-91", "refs/heads/dev")
	before := workflowTestRead(t, f.store, f.run.ID)

	page, err := f.store.QueryWorkflowIntegrationRoles(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("human approval became a runnable agent", page, err)
	}

	claim := nativeScheduleClaim(t, peer, WorkflowIntegrationScheduleQuery{Limit: 10})
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, workflowTestRead(t, f.store, f.run.ID)) || workflowTestRead(t, peer, next.InvocationID).Cursor != 1 {
		t.Fatal("independent execution resumed human approval")
	}
}

// TestNativePreparedWait excludes an otherwise valid preparation at its full
// target bound. It must not reserve global priority that prevents task progress.
func TestNativePreparedWait(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	nativeRoleAdmit(t, f.store, f.run.ID)

	first, one := nativeRolePeer(t, f.store, "one", "PRJ-91", "refs/heads/dev")
	second, two := nativeRolePeer(t, f.store, "two", "PRJ-92", "refs/heads/dev")
	nativeRoleTransition(t, first, one, "preparing")
	nativeRoleTransition(t, second, two, "preparing")

	peer, next := nativeRolePeer(t, f.store, "peer", "PRJ-93", "refs/heads/dev")
	before := workflowTestRead(t, f.store, f.run.ID)

	page, err := f.store.QueryWorkflowIntegrationRoles(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("full target preparation retained global priority", page, err)
	}

	claim := nativeScheduleClaim(t, peer, WorkflowIntegrationScheduleQuery{Limit: 10})
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, workflowTestRead(t, f.store, f.run.ID)) || workflowTestRead(t, peer, next.InvocationID).Cursor != 1 {
		t.Fatal("target capacity wait consumed preparation or prevented independent work")
	}
}

// TestNativeRoleBlocked rejects an explicit preparation retry before it can
// change a human-blocked entry or acquire a role; independent work stays usable.
func TestNativeRoleBlocked(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	request := nativeEntryTransition(t, f.store, entry, "blocked")

	request.Blocker = &WorkflowIntegrationBlocker{Reason: "human_decision", RequiredAction: "approve continuation", ResumeState: "queued"}
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	peer, next := nativeRolePeer(t, f.store, "peer", "PRJ-91", "refs/heads/dev")

	before := nativeQueueImages(t, f.store.layout.Root())
	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("preparation resumed a blocked native owner", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("rejected preparation changed the blocked records")
	}

	claim := nativeScheduleClaim(t, peer, WorkflowIntegrationScheduleQuery{Limit: 10})
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	if workflowTestRead(t, peer, next.InvocationID).Cursor != 1 {
		t.Fatal("blocked preparation prevented independent progression")
	}
}
