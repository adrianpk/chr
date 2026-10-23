package app

import (
	"context"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

// StartNamedGuidedUseCase starts through the existing idempotent name owner.
func (s service) StartNamedGuidedUseCase(ctx context.Context, request core.GuidedUseCaseStart, name string) (core.AuthoringSession, error) {
	if strings.TrimSpace(name) == "" || len(name) > 512 {
		return core.AuthoringSession{}, errors.New("enter a name of at most 512 bytes")
	}

	session, err := s.StartGuidedUseCase(ctx, request.RunID, request.Namespace, request.Actor)
	if err != nil {
		return session, err
	}

	return s.ChangeGuidedUseCase(ctx, core.GuidedUseCaseCommand{
		GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: session.ID, Namespace: session.Namespace},
		OperationID:        "name-" + request.RunID, ExpectedRevision: 1, Action: "name", Content: name,
	})
}
