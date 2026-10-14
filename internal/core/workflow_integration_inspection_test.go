package core

import (
	"errors"
	"reflect"
	"testing"
)

// TestNativeInspectionRole observes a real owned slot without granting another
// launch, changing authority or losing the retained target queue position.
func TestNativeInspectionRole(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	workflowTestClaim(t, store, entry.InvocationID)

	before := nativeQueueImages(t, store.layout.Root())

	inspection, err := store.InspectWorkflow(t.Context(), entry.InvocationID)
	if err != nil || inspection.Integration == nil {
		t.Fatal("owned integration inspection failed", err)
	}

	observed := inspection.Integration
	if observed.Capacity.Maximum != 1 || observed.Capacity.Available != 0 || len(observed.Roles) != 1 || observed.Roles[0].Role != "executor" || observed.QueuePosition != 1 {
		t.Fatal("inspection lost actual role or queue occupancy", observed)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("role inspection changed authority")
	}
}

// TestNativeInspectionDecision reads the sealed ambiguity and actual consumed
// budget without resolving a choice or manufacturing a prepared candidate.
func TestNativeInspectionDecision(t *testing.T) {
	f, _, decisions := nativeDecisionFixture(t)

	before := nativeQueueImages(t, f.store.layout.Root())

	inspection, err := f.store.InspectWorkflow(t.Context(), f.run.ID)
	if err != nil || inspection.Integration == nil || inspection.Integration.Decisions == nil {
		t.Fatal("decision inspection failed", err)
	}

	observed := inspection.Integration
	if observed.Entry.State != "blocked" || observed.Entry.Blocker.RequiredAction != "answer_decision" || !reflect.DeepEqual(*observed.Decisions, decisions) || observed.Candidate != nil {
		t.Fatal("inspection resolved ambiguity or invented preparation", observed)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("decision inspection changed authority")
	}
}

// TestNativeDecisionInspection retains a real interrupted answer transaction;
// neither the decision reader nor status may replay it as a side effect of a view.
func TestNativeDecisionInspection(t *testing.T) {
	f, _, decisions := nativeDecisionFixture(t)
	request := nativeDecisionAnswer(t, f.store, f.run.ID, decisions)
	crash := errors.New("interrupted operator answer")

	f.store.afterReplace = func(int) error { return crash }
	if _, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), request, f.provider); !errors.Is(err, crash) {
		t.Fatal("answer fixture did not retain its write-ahead journal", err)
	}

	store := NewWorkflowStore(f.store.layout, f.store.projectID)

	before := nativeQueueImages(t, store.layout.Root())
	if _, err := store.ReadWorkflowIntegrationDecisions(t.Context(), f.run.ID); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("decision view replayed the interrupted answer", err)
	}

	if _, err := store.InspectWorkflow(t.Context(), f.run.ID); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("status replayed the interrupted answer", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("read-only view changed interrupted authority")
	}
}
