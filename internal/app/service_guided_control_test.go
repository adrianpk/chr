package app

import (
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func guidedNativeControl(t *testing.T, operations service, session core.AuthoringSession, id, action string) core.GuidedUseCaseCommand {
	t.Helper()

	command := guidedQuizCommand(session, id, action, "")
	if session.Guided.Work != nil && session.Guided.Work.Invocation != nil {
		run, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, session)
		if err != nil {
			t.Fatal(err)
		}

		command.NativeRevision = run.Revision
	}

	return command
}

// TestGuidedUseCaseControlNative separates accepted stop intent from quiescence
// and resumes the same invocation with its already consumed attempt intact.
func TestGuidedUseCaseControlNative(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"pause", "cancel"} {
		t.Run(action, func(t *testing.T) { testGuidedNativeStop(t, action) })
	}
}

func testGuidedNativeStop(t *testing.T, action string) {
	t.Helper()

	operations, started := guidedAdmittedFixture(t)

	claim, err := operations.ClaimWorkflow(t.Context(), started.Guided.Work.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	command := guidedNativeControl(t, operations, started, "stop", action)

	stopping, err := operations.ControlGuidedUseCase(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedStopAccepted(t, operations, stopping, action)

	stopped := acknowledgeGuidedTestStop(t, operations, claim, command.GuidedUseCaseScope)

	if action == "pause" {
		assertGuidedNativeResume(t, operations, stopped, command)
	} else if stopped.Guided.Status != "cancelled" {
		t.Fatal("native cancellation was lost")
	}
}

func assertGuidedNativeResume(t *testing.T, operations service, paused core.AuthoringSession, original core.GuidedUseCaseCommand) {
	t.Helper()

	before := assertGuidedNativeAdmission(t, operations, paused)

	resumed, err := operations.ControlGuidedUseCase(t.Context(), guidedNativeControl(t, operations, paused, "resume", "resume"))
	if err != nil || resumed.Guided.Control.StopIntent != "" {
		t.Fatal("native resume failed", err)
	}

	replay, err := operations.ControlGuidedUseCase(t.Context(), original)
	if err != nil || replay.Revision != resumed.Revision {
		t.Fatal("old stop replay changed the resumed run", err)
	}

	assertGuidedResumedAttempts(t, operations, before, resumed)
}

func assertGuidedResumedAttempts(t *testing.T, operations service, before core.WorkflowInvocation, resumed core.AuthoringSession) {
	t.Helper()

	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	if err := host.activate(operations, before.ID); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	after := assertGuidedNativeAdmission(t, operations, resumed)
	if after.ID != before.ID || after.Policy != before.Policy || serviceWorkTaskAttempts(after) != 2 || after.Review == nil || after.Review.State != "waiting_approval" {
		t.Fatal("resume replaced work, refilled attempts or bypassed review")
	}
}

// TestGuidedUseCaseControlLostAck reconciles the native receipt before current
// revision checks, while stale fresh controls leave native and guided state intact.
func TestGuidedUseCaseControlLostAck(t *testing.T) {
	t.Parallel()

	operations, started := guidedAdmittedFixture(t)
	rejected := rejectStaleGuidedControl(t, operations, started)

	command := guidedNativeControl(t, operations, rejected, "pause", "pause")
	lost := errors.New("lost native control acknowledgement")
	operations.afterGuidedControl = func() error { return lost }

	pending, err := operations.ControlGuidedUseCase(t.Context(), command)
	if !errors.Is(err, lost) || pending.Guided.Control.PendingOperation == nil {
		t.Fatal("control fault was not retained", err)
	}

	before, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, pending)
	if err != nil || before.Status != "paused" {
		t.Fatal("native stop was not accepted before fault", err)
	}

	operations.afterGuidedControl = nil

	recovered, err := operations.RecoverGuidedUseCase(t.Context(), command.GuidedUseCaseScope)
	if err != nil || recovered.Guided.Status != "paused" || recovered.Guided.Control.PendingOperation != nil {
		t.Fatal("original native receipt was not recovered", err)
	}

	after := assertGuidedNativeAdmission(t, operations, recovered)
	if after.Revision != before.Revision || len(after.Attempts) != 0 {
		t.Fatal("control recovery repeated native effects")
	}
}

// TestGuidedUseCaseControlAdmission stops an uncertain admission by finding its
// exact existing request, and never creates work merely to pause a retained stage.
func TestGuidedUseCaseControlAdmission(t *testing.T) {
	t.Parallel()

	for _, admitted := range []bool{false, true} {
		name := "not-admitted"
		if admitted {
			name = "lost-admission-response"
		}

		t.Run(name, func(t *testing.T) { testGuidedControlAdmission(t, admitted) })
	}
}

func testGuidedControlAdmission(t *testing.T, admitted bool) {
	t.Helper()

	operations, command, pending := pendingGuidedAdmissionFixture(t, admitted)

	paused, err := operations.ControlGuidedUseCase(t.Context(), guidedNativeControl(t, operations, pending, "pause", "pause"))
	if err != nil || paused.Guided.Status != "paused" || paused.Guided.PendingOperation.ID != command.OperationID {
		t.Fatal("pause lost the pending admission", err)
	}

	assertGuidedControlAdmissionCount(t, operations, paused, admitted)
}

func assertGuidedControlAdmissionCount(t *testing.T, operations service, paused core.AuthoringSession, admitted bool) {
	t.Helper()

	page, err := operations.WorkflowStatusPage(t.Context(), 0, 8)

	expected := 0
	if admitted {
		expected = 1
	}

	if err != nil || len(page.Items) != expected {
		t.Fatal("pause created or duplicated native work", err)
	}

	if admitted && (paused.Guided.Work.Invocation == nil || page.Items[0].Status != "paused" || page.Items[0].TotalAttempts != 0) {
		t.Fatal("pause did not bind the exact uncertain invocation")
	}
}

func guidedAdmittedFixture(t *testing.T) (service, core.AuthoringSession) {
	t.Helper()

	operations, prepared, _ := guidedAdmissionFixture(t)

	started, err := operations.AdmitGuidedWork(t.Context(), guidedQuizCommand(prepared, "admit", "admit", ""))
	if err != nil {
		t.Fatal(err)
	}

	return operations, started
}

func assertGuidedStopAccepted(t *testing.T, operations service, stopping core.AuthoringSession, action string) {
	t.Helper()

	if stopping.Guided.Status != "blocked" || stopping.Guided.Control.NativeReceipt == nil || stopping.Guided.Control.AcknowledgedRevision != 0 {
		t.Fatal("stop acceptance was confused with quiescence")
	}

	run, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, stopping)
	if err != nil || run.Status != "running" || run.StopIntent != action {
		t.Fatal("native owner did not retain the stop", err)
	}
}

func acknowledgeGuidedTestStop(t *testing.T, operations service, claim *core.WorkflowClaim, scope core.GuidedUseCaseScope) core.AuthoringSession {
	t.Helper()

	unsettled, err := core.ReconcileGuidedStop(t.Context(), operations.projectContext.ProjectDir, scope)
	if err != nil || unsettled.Guided.Status != "blocked" {
		t.Fatal("running owner was falsely acknowledged", err)
	}

	if err := operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	stopped, err := operations.RecoverGuidedUseCase(t.Context(), scope)
	if err != nil || stopped.Guided.Control.AcknowledgedRevision == 0 {
		t.Fatal("native quiescence was not acknowledged", err)
	}

	return stopped
}

func rejectStaleGuidedControl(t *testing.T, operations service, started core.AuthoringSession) core.AuthoringSession {
	t.Helper()

	stale := guidedNativeControl(t, operations, started, "stale", "pause")
	stale.NativeRevision++

	rejected, err := operations.ControlGuidedUseCase(t.Context(), stale)
	if !errors.Is(err, workflowstate.ErrConflict) || rejected.Guided.Control.StopIntent != "" || rejected.Guided.Control.PendingOperation != nil {
		t.Fatal("stale control retained a stop or blocked fresh controls", err)
	}

	return rejected
}

func pendingGuidedAdmissionFixture(t *testing.T, admitted bool) (service, core.GuidedUseCaseCommand, core.AuthoringSession) {
	t.Helper()

	operations, prepared, _ := guidedAdmissionFixture(t)
	command := guidedQuizCommand(prepared, "admit", "admit", "")

	pending, err := core.ReserveGuidedUseCaseStep(t.Context(), operations.projectContext.ProjectDir, command)
	if err != nil {
		t.Fatal(err)
	}

	if !admitted {
		return operations, command, pending
	}

	lost := errors.New("lost admission acknowledgement")
	operations.afterGuidedAdmission = func(phase string) error {
		if phase == "admission" {
			return lost
		}

		return nil
	}

	pending, err = operations.AdmitGuidedWork(t.Context(), command)
	if !errors.Is(err, lost) {
		t.Fatal("admission fault not reached", err)
	}

	return operations, command, pending
}
