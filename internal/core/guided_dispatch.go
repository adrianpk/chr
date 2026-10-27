package core

import (
	"context"
	"errors"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// ErrGuidedUseCaseStopped prevents new dispatch after a guided stop is accepted.
var ErrGuidedUseCaseStopped = errors.New("guided progression is paused or terminal")

// GuidedDispatchAllowed observes retained authorization without recovering state.
func GuidedDispatchAllowed(session authoringcore.AuthoringSession) bool {
	state := session.Guided
	if state == nil || !state.ProgressionIntent {
		return false
	}

	if state.Control != nil && state.Control.StopIntent != "" {
		return false
	}

	switch state.Status {
	case "paused", "cancelled", "completed", "failed":
		return false
	default:
		return true
	}
}

// ReserveGuidedUseCaseStep retains accepted input before asynchronous provider
// dispatch. Replays retain the original arguments and owner request identity.
func ReserveGuidedUseCaseStep(ctx context.Context, root string, command authoringcore.GuidedUseCaseCommand) (session authoringcore.AuthoringSession, err error) {
	operation, err := guidedOperation(command)
	if err != nil {
		return session, err
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, command.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, command.GuidedUseCaseScope); err != nil {
			return err
		}

		replay, err := guardGuidedCommand(session, command, operation)
		if err != nil || replay {
			return err
		}

		if session.Guided.PendingOperation == nil {
			if err := guardGuidedStepInput(session, operation); err != nil {
				return err
			}

			session.Guided.PendingOperation = &operation
		}

		session.Guided.Status, session.Guided.Blocker = "running", ""

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}
