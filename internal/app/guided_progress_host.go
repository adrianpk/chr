package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func guidedProgressKey(operations service, scope core.GuidedUseCaseScope) string {
	return operations.projectContext.ProjectDir + "\x00" + scope.ProjectID + "\x00" + scope.Namespace + "\x00" + scope.RunID
}

func guidedProgressFingerprint(session core.AuthoringSession) string {
	if command, ready := guidedControlCommand(session); ready {
		return "control/" + command.OperationID
	}

	if command, ready := guidedProgressCommand(session); ready {
		return "stage/" + command.OperationID
	}

	return ""
}

func (host *guidedWorkHost) schedule(operations service, scope core.GuidedUseCaseScope) error {
	return host.scheduleProgress(operations, scope, true)
}

func (host *guidedWorkHost) scheduleProgress(operations service, scope core.GuidedUseCaseScope, explicit bool) error {
	if err := host.ready(); err != nil {
		return err
	}

	session, err := core.ReadGuidedUseCase(host.ctx, operations.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	operations.workflowNamespace = scope.Namespace

	fingerprint := guidedProgressFingerprint(session)
	if fingerprint == "" {
		return host.activateRetained(operations, session)
	}

	return host.dispatchProgress(operations, scope, fingerprint, explicit)
}

func (host *guidedWorkHost) dispatchProgress(operations service, scope core.GuidedUseCaseScope, fingerprint string, explicit bool) error {
	host.mu.Lock()
	defer host.mu.Unlock()

	if err := host.ready(); err != nil {
		return err
	}

	key := guidedProgressKey(operations, scope)
	if host.progress[key] || !explicit && host.attempts[key] == fingerprint {
		return nil
	}

	if len(host.progress) >= guidedHostActivationLimit || len(host.attempts) >= core.WorkflowRecordLimit && host.attempts[key] == "" {
		return core.ErrWorkflowCapacity
	}

	host.progress[key], host.attempts[key] = true, fingerprint

	host.wait.Add(1)
	go func() {
		defer host.wait.Done()
		defer host.releaseProgress(key)

		host.runProgress(host.ctx, operations, scope, key)
	}()

	return nil
}

func (host *guidedWorkHost) runProgress(ctx context.Context, operations service, scope core.GuidedUseCaseScope, key string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	operations.ctx = ctx

	session, err := operations.RecoverGuidedUseCase(ctx, scope)
	if err == nil {
		err = host.activateRetained(operations, session)
	}

	if err != nil {
		host.retainProgressAttempt(key, session)

		if guidedHostFailure(err) && host.ctx.Err() == nil {
			slog.WarnContext(ctx, "Guided progression needs recovery", "run", scope.RunID, "code", "guided_progression_blocked")
		}
	}
}

func (host *guidedWorkHost) retainProgressAttempt(key string, session core.AuthoringSession) {
	if session.Guided == nil {
		return
	}

	host.mu.Lock()
	defer host.mu.Unlock()

	if fingerprint := guidedProgressFingerprint(session); fingerprint != "" {
		host.attempts[key] = fingerprint
	}
}

func (host *guidedWorkHost) releaseProgress(key string) {
	host.mu.Lock()
	defer host.mu.Unlock()

	delete(host.progress, key)
}

func (host *guidedWorkHost) activateRetained(operations service, session core.AuthoringSession) error {
	if !core.GuidedDispatchAllowed(session) || session.Guided.Work == nil || session.Guided.Work.Invocation == nil {
		return nil
	}

	run, err := core.ReadGuidedInvocation(host.ctx, operations.projectContext.ProjectDir, session)
	if err != nil {
		return err
	}

	if err := operations.recoverGuidedInterruptedWork(host.ctx, run); err != nil {
		return err
	}

	status, err := operations.WorkflowStatus(host.ctx, run.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || !retainedWorkRunnable(status) {
		return err
	}

	return host.activate(operations, run.ID)
}

func (operations service) checkGuidedActivation(ctx context.Context, run core.WorkflowInvocation) error {
	if run.Admission == nil || run.Admission.Guided == nil {
		return core.ErrWorkflowConflict
	}

	scope := core.GuidedUseCaseScope{RunID: run.Admission.Guided.RunID, ProjectID: run.ProjectID, Namespace: run.Namespace}

	session, err := core.ReadGuidedUseCase(ctx, operations.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	if !guidedInvocationMatches(run, session) {
		return core.ErrWorkflowConflict
	}

	if !core.GuidedDispatchAllowed(session) {
		return core.ErrGuidedUseCaseStopped
	}

	return nil
}

func (host *guidedWorkHost) acknowledgeStop(ctx context.Context, operations service, id string) {
	if ctx.Err() != nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	run, err := operations.ReadWorkflow(ctx, id)
	if err != nil || run.Admission == nil || run.Admission.Guided == nil {
		return
	}

	scope := core.GuidedUseCaseScope{RunID: run.Admission.Guided.RunID, ProjectID: run.ProjectID, Namespace: run.Namespace}

	_, err = core.ReconcileGuidedStop(ctx, operations.projectContext.ProjectDir, scope)
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.WarnContext(ctx, "Guided stop acknowledgement needs recovery", "code", "guided_stop_acknowledgement_failed")
	}
}
