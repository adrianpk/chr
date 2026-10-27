package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

type guidedWorkStep func(service, context.Context, *core.AuthoringSession) error

func (s service) applyGuidedWorkStep(ctx context.Context, command core.GuidedUseCaseCommand, apply guidedWorkStep) (core.AuthoringSession, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *core.AuthoringSession) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return apply(bounded, ctx, session)
	})
}
