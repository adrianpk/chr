package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type guidedRecoveryModel func(context.Context, quizmodel.Request) (quizmodel.Response, error)

func (model guidedRecoveryModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	return model(ctx, request)
}

// TestGuidedUseCaseRecoveryDisconnect retains input before provider dispatch,
// survives request cancellation and keeps inspection free of recovery effects.
func TestGuidedUseCaseRecoveryDisconnect(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")
	host := newGuidedWorkHost(t.Context())
	operations.guidedWorkHost = host
	entered, unblock := guidedBlockedRecoveryModel(t, &operations, model)
	t.Cleanup(func() { unblock(); host.close() })

	requestCtx, cancel := context.WithCancel(t.Context())
	pending, err := operations.ChangeGuidedUseCase(requestCtx, guidedQuizCommand(session, "accepted-quiz", "quiz", ""))

	cancel()

	if err != nil || pending.Guided.PendingOperation == nil || pending.Guided.PendingOperation.ID != "accepted-quiz" {
		t.Fatal("input was not retained before asynchronous generation", err)
	}

	select {
	case providerCtx := <-entered:
		if providerCtx.Err() != nil {
			t.Fatal("request cancellation reached the provider", providerCtx.Err())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("managed generation did not start")
	}

	assertGuidedRecoveryInspection(t, operations, session)
	unblock()
	host.wait.Wait()

	current := readWebGuidedSession(t, operations, session.ID)
	if current.Guided.Quiz == nil || current.Guided.Status != "waiting_input" || len(model.requests) != 1 {
		t.Fatal("disconnected request lost or repeated generation")
	}
}

func assertGuidedRecoveryInspection(t *testing.T, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	scope := authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}

	before, err := core.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, scope)
	if err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if _, err := operations.InspectGuidedUseCase(t.Context(), session.ID, session.Namespace); err != nil {
			t.Fatal(err)
		}
	}

	after, err := core.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, scope)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("inspection mutated or reconciled an in-flight stage", err)
	}
}

// TestGuidedUseCaseRecoveryStage recovers the original quiz owner receipt on a
// fresh managed host without another model call or replacement operation.
func TestGuidedUseCaseRecoveryStage(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "research")
	lost := errors.New("lost quiz result")
	operations.afterGuidedQuiz = func() error { return lost }

	pending, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "quiz", "quiz", ""))
	if !errors.Is(err, lost) || pending.Guided.PendingOperation == nil {
		t.Fatal("quiz fault not retained", err)
	}

	operations.afterGuidedQuiz = nil
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host
	host.start(func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy })

	recovered := waitGuidedRecovery(t, operations, session, func(current authoringcore.AuthoringSession) bool {
		return current.Guided.Quiz != nil && current.Guided.PendingOperation == nil
	})

	host.close()

	if len(model.requests) != 1 || recovered.Namespace != "research" || recovered.Guided.CompletedOperations[len(recovered.Guided.CompletedOperations)-1].ID != "quiz" {
		t.Fatal("restart repeated generation or substituted original scope")
	}
}

// TestGuidedUseCaseRecoveryBlocked does not regenerate failed operations on
// every inventory sweep; an explicit retry retains the original request identity.
func TestGuidedUseCaseRecoveryBlocked(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host

	var calls atomic.Int32

	operations.quizModel = guidedRecoveryModel(func(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
		if calls.Add(1) == 1 {
			return quizmodel.Response{}, errors.New("provider unavailable")
		}

		return model.Invoke(ctx, request)
	})

	command := guidedQuizCommand(session, "retry-original", "quiz", "")
	if _, err := operations.ChangeGuidedUseCase(t.Context(), command); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	for range 2 {
		host.recoverPage(host.ctx, operations, 0)
		host.wait.Wait()
	}

	if calls.Load() != 1 {
		t.Fatal("recovery inventory repeated a failed provider call")
	}

	if _, err := operations.ChangeGuidedUseCase(t.Context(), command); err != nil {
		t.Fatal("original retry rejected", err)
	}

	host.wait.Wait()

	current := readWebGuidedSession(t, operations, session.ID)
	if calls.Load() != 2 || current.Guided.Quiz == nil || current.Guided.PendingOperation != nil {
		t.Fatal("explicit retry did not reconcile the original operation")
	}
}

// TestGuidedUseCaseRecoveryAdmission starts retained work on host startup after
// a lost start response, preserving membership, policy and the original admission.
func TestGuidedUseCaseRecoveryAdmission(t *testing.T) {
	operations, pending, before := guidedLostStartFixture(t)

	operations.afterGuidedAdmission = nil
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host
	host.start(func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy })

	current := waitGuidedRecovery(t, operations, pending, func(current authoringcore.AuthoringSession) bool {
		run, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, current)

		return err == nil && current.Guided.PendingOperation == nil && run.Review != nil && run.Review.State == "waiting_approval"
	})

	host.close()

	after := assertGuidedNativeAdmission(t, operations, current)
	if after.ID != before.ID || after.Policy != before.Policy || !reflect.DeepEqual(after.Admission, before.Admission) || serviceWorkTaskAttempts(after) != 1 {
		t.Fatal("host startup replaced the invocation, membership or attempt budget")
	}
}

func waitGuidedRecovery(t *testing.T, operations service, session authoringcore.AuthoringSession, ready func(authoringcore.AuthoringSession) bool) authoringcore.AuthoringSession {
	t.Helper()

	scope := authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		current, err := core.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, scope)
		if err != nil {
			t.Fatal(err)
		}

		if ready(current) {
			return current
		}

		time.Sleep(10 * time.Millisecond)
	}

	failGuidedRecovery(t, operations, scope)

	return authoringcore.AuthoringSession{}
}

func failGuidedRecovery(t *testing.T, operations service, scope authoringcore.GuidedUseCaseScope) {
	t.Helper()

	current, readErr := core.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, scope)
	if readErr == nil && current.Guided.Work != nil && current.Guided.Work.Invocation != nil {
		run, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, current)
		t.Fatalf("managed recovery did not reach its expected state: guided=%s native=%s outcome=%s resume=%s blocker=%+v error=%v", current.Guided.Status, run.Status, run.Outcome, run.ResumeStatus, run.Blocker, err)
	}

	t.Fatal("managed authoring recovery did not reach its expected owner state", readErr)
}

func guidedBlockedRecoveryModel(t *testing.T, operations *service, model *fakeServiceQuizModel) (<-chan context.Context, func()) {
	t.Helper()

	entered, release := make(chan context.Context, 1), make(chan struct{})

	var once sync.Once

	unblock := func() { once.Do(func() { close(release) }) }

	operations.quizModel = guidedRecoveryModel(func(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
		entered <- ctx

		select {
		case <-release:
			return model.Invoke(ctx, request)
		case <-ctx.Done():
			return quizmodel.Response{}, ctx.Err()
		}
	})

	return entered, unblock
}

func guidedLostStartFixture(t *testing.T) (service, authoringcore.AuthoringSession, core.WorkflowInvocation) {
	t.Helper()

	operations, prepared, _ := guidedAdmissionFixture(t)
	lost := errors.New("lost start response")
	operations.afterGuidedAdmission = func(phase string) error {
		if phase == "start" {
			return lost
		}

		return nil
	}

	pending, err := operations.AdmitGuidedWork(t.Context(), guidedQuizCommand(prepared, "admit", "admit", ""))
	if !errors.Is(err, lost) || pending.Guided.Work.Invocation == nil {
		t.Fatal("lost start fault was not retained", err)
	}

	before, err := core.ReadGuidedInvocation(t.Context(), operations.projectContext.ProjectDir, pending)
	if err != nil {
		t.Fatal(err)
	}

	return operations, pending, before
}

func guidedRecoveryReviewReady(ctx context.Context, operations service, session authoringcore.AuthoringSession) bool {
	run, err := core.ReadGuidedInvocation(ctx, operations.projectContext.ProjectDir, session)

	return err == nil && session.Guided.PendingOperation == nil && run.Review != nil && run.Review.State == "waiting_approval"
}
