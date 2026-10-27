package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

type guidedProgressActivator interface {
	schedule(service, authoringcore.GuidedUseCaseScope) error
}

func guidedControlAction(action string) bool {
	return action == "pause" || action == "resume" || action == "cancel"
}

func guidedProviderAction(action string) bool {
	switch action {
	case "quiz", "answer", "quiz-mode", "quiz-skip", "quiz-done", "spec", "prepare", "admit":
		return true
	default:
		return false
	}
}

func (s service) changeGuidedControl(ctx context.Context, command authoringcore.GuidedUseCaseCommand) (authoringcore.Session, error) {
	session, err := s.ControlGuidedUseCase(ctx, command)
	if err != nil || command.Action != "resume" {
		return session, err
	}

	if host, ok := s.guidedWorkHost.(guidedProgressActivator); ok {
		return session, host.schedule(s, command.GuidedUseCaseScope)
	}

	return s.advanceGuidedUseCase(ctx, session)
}

func (s service) changeGuidedOnHost(ctx context.Context, command authoringcore.GuidedUseCaseCommand, host guidedProgressActivator) (authoringcore.Session, error) {
	var (
		session authoringcore.Session
		err     error
	)
	if guidedProviderAction(command.Action) {
		session, err = core.ReserveGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command)
	} else {
		session, err = s.applyGuidedCommand(ctx, command)
	}

	if err != nil {
		return session, err
	}

	return session, host.schedule(s, command.GuidedUseCaseScope)
}

// RecoverGuidedUseCase reconciles pending control and stage operations before
// continuing the same scope. It is a lifecycle operation, never a query effect.
func (s service) RecoverGuidedUseCase(ctx context.Context, scope authoringcore.GuidedUseCaseScope) (authoringcore.Session, error) {
	if s.guidedHostCheck != nil {
		if err := s.guidedHostCheck(ctx); err != nil {
			return authoringcore.Session{}, err
		}
	}

	session, err := core.ReadGuidedUseCase(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return session, err
	}

	if command, ready := guidedControlCommand(session); ready {
		return s.ControlGuidedUseCase(ctx, command)
	}

	if session.Guided.Control != nil && session.Guided.Control.StopIntent != "" {
		return s.recoverGuidedStop(ctx, session, scope)
	}

	if !core.GuidedDispatchAllowed(session) {
		return session, nil
	}

	return s.advanceGuidedUseCase(ctx, session)
}

func (s service) recoverGuidedStop(ctx context.Context, session authoringcore.Session, scope authoringcore.GuidedUseCaseScope) (authoringcore.Session, error) {
	control := session.Guided.Control
	if control.NativeReceipt == nil || control.AcknowledgedRevision != 0 {
		return session, nil
	}

	run, err := core.ReadGuidedInvocation(ctx, s.projectContext.ProjectDir, session)
	if err != nil {
		return session, err
	}

	if run.Status == "running" {
		s.workflowNamespace = session.Namespace
		if _, err := s.ReconcileWorkflow(ctx, run.ID, run.Revision); err != nil {
			return session, err
		}
	}

	return core.ReconcileGuidedStop(ctx, s.projectContext.ProjectDir, scope)
}

// recoverGuidedInterruptedWork resumes a quiescent host interruption under the
// native owner and original guided authority. An operator stop is never cleared.
func (s service) recoverGuidedInterruptedWork(ctx context.Context, run core.WorkflowInvocation) error {
	if run.Status != "paused" || run.StopIntent != "" || run.Outcome != "interrupted" {
		return nil
	}

	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	check := func(ctx context.Context) error { return s.checkGuidedActivation(ctx, run) }
	_, err = store.ResumeWorkflow(ctx, run.ID, core.WorkflowResumePorts{ExpectedRevision: run.Revision, AuthorizationCheck: check})

	return err
}
