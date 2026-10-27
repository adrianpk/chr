package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

type guidedWorkStep func(service, context.Context, *authoringcore.AuthoringSession) error

func (s service) applyGuidedWorkStep(ctx context.Context, command authoringcore.GuidedUseCaseCommand, apply guidedWorkStep) (authoringcore.AuthoringSession, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return authoringcore.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *authoringcore.AuthoringSession) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return apply(bounded, ctx, session)
	})
}
