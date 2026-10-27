package app

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

type guidedWorkStep func(service, context.Context, *authoringcore.Session) error

func (s service) applyGuidedWorkStep(ctx context.Context, command authoringcore.GuidedUseCaseCommand, apply guidedWorkStep) (authoringcore.Session, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return authoringcore.Session{}, err
	}

	command.GuidedUseCaseScope = scope

	return authoringcore.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *authoringcore.Session) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return apply(bounded, ctx, session)
	})
}
