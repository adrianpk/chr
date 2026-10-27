package core

import (
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowCapacityRoleWait proves aggregate, review, preparation, repair and
// closing retries retain their own phase and evidence while the pool is occupied.
func TestWorkflowCapacityRoleWait(t *testing.T) {
	for _, role := range []string{"aggregate", "review", "preparation", "repair", "closing"} {
		t.Run(role, func(t *testing.T) { testWorkflowCapacityRoleWait(t, role) })
	}
}

func testWorkflowCapacityRoleWait(t *testing.T, role string) {
	t.Helper()

	store, id, action := workflowCapacityRoleFixture(t, role)
	before := workflowTestRead(t, store, id)

	legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "legacy"})
	if err != nil {
		t.Fatal(err)
	}

	if err := action(); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("role bypassed occupied capacity", err)
	}

	waiting := workflowTestRead(t, store, id)
	assertWorkflowRoleWait(t, before, waiting)

	if err := legacy.Release(); err != nil {
		t.Fatal(err)
	}

	if err := action(); err != nil {
		t.Fatal("retained role could not resume", err)
	}

	after := workflowTestRead(t, store, id)
	if after.Status == "waiting_capacity" || after.CapacityResumeStatus != "" {
		t.Fatal("completed role retained an obsolete wait", after)
	}
}

func assertWorkflowRoleWait(t *testing.T, before, waiting WorkflowInvocation) {
	t.Helper()

	if waiting.Status != "waiting_capacity" || waiting.CapacityResumeStatus != before.Status || len(waiting.Attempts) != len(before.Attempts) || waiting.CandidateRef != before.CandidateRef {
		t.Fatal("role wait changed evidence, attempts or previous phase", before, waiting)
	}
}

func workflowCapacityRoleFixture(t *testing.T, role string) (*WorkflowStore, string, func() error) {
	t.Helper()

	switch role {
	case "aggregate":
		return workflowCapacityAggregateFixture(t)
	case "preparation":
		fixture := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))

		return fixture.store, fixture.run.ID, func() error {
			_, err := fixture.store.PrepareWorkflowSliceIntegration(t.Context(), fixture.run.ID, fixture.provider, fixture.git)

			return err
		}
	case "closing":
		fixture, _, run := workflowClosingEmptyOwner(t)

		return fixture.store, run.ID, func() error {
			_, err := fixture.store.CloseWorkflowInvocationSet(t.Context(), run.ID, "dev", workflowFrontierPorts(fixture))

			return err
		}
	default:
		return workflowCapacityReviewFixture(t, role)
	}
}

func workflowCapacityAggregateFixture(t *testing.T) (*WorkflowStore, string, func() error) {
	t.Helper()

	store, _, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))

	run := workflowTestStart(t, store, request).Invocation
	for range 3 {
		claim := workflowTestClaim(t, store, run.ID)
		if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker")); err != nil {
			t.Fatal(err)
		}
	}

	return store, run.ID, func() error {
		_, err := store.ValidateWorkflowSlice(t.Context(), run.ID)

		return err
	}
}

func workflowCapacityReviewFixture(t *testing.T, role string) (*WorkflowStore, string, func() error) {
	t.Helper()

	store, hierarchy, run, provider := workflowReviewFixture(t, workflowstate.ReviewAgentic)
	if role == "repair" {
		return store, run.ID, func() error {
			_, err := store.RepairWorkflowSlice(t.Context(), run.ID, hierarchy.Tasks[0].Task.Key)

			return err
		}
	}

	return store, run.ID, func() error {
		_, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, workflowReviewWorker("pass", "", ""))

		return err
	}
}
