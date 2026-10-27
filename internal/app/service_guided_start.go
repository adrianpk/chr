package app

import (
	"context"
	"errors"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// StartNamedGuidedUseCase starts through the existing idempotent name owner.
func (s service) StartNamedGuidedUseCase(ctx context.Context, request authoringcore.GuidedUseCaseStart, name string) (authoringcore.AuthoringSession, error) {
	if strings.TrimSpace(name) == "" || len(name) > 512 {
		return authoringcore.AuthoringSession{}, errors.New("enter a name of at most 512 bytes")
	}

	session, err := s.StartGuidedUseCase(ctx, request.RunID, request.Namespace, request.Actor)
	if err != nil {
		return session, err
	}

	return s.ChangeGuidedUseCase(ctx, authoringcore.GuidedUseCaseCommand{
		GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, Namespace: session.Namespace},
		OperationID:        "name-" + request.RunID, ExpectedRevision: 1, Action: "name", Content: name,
	})
}
