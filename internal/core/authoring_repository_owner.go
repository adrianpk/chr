package core

import (
	"context"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileAuthoringSessionRepository) CreateSession(ctx context.Context, session authoringcore.AuthoringSession) (paths AuthoringSessionPaths, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringSessionPaths(session.ID); err != nil {
			return err
		}

		paths, err = r.createSession(ctx, session)

		return err
	})

	return paths, err
}

func (r *fileAuthoringSessionRepository) UpdateSession(ctx context.Context, session authoringcore.AuthoringSession) (paths AuthoringSessionPaths, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		current, err := r.getGuardedSession(ctx, session.ID)
		if err != nil {
			return err
		}

		if current.Guided != nil || session.Guided != nil {
			return workflowstate.ErrConflict
		}

		if current.Revision > 0 {
			if session.Revision != current.Revision || session.Namespace != current.Namespace {
				return workflowstate.ErrConflict
			}

			session.Revision++
		}

		paths, err = r.updateSession(ctx, session)

		return err
	})

	return paths, err
}

func (r *fileAuthoringSessionRepository) GetSession(ctx context.Context, id string) (session authoringcore.AuthoringSession, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, id)

		return err
	})

	return session, err
}

func (r *fileAuthoringSessionRepository) getGuardedSession(ctx context.Context, id string) (authoringcore.AuthoringSession, error) {
	if _, err := r.authoringSessionPaths(id); err != nil {
		return authoringcore.AuthoringSession{}, err
	}

	return r.getSession(ctx, id)
}

func (r *fileAuthoringSessionRepository) ListSessions(ctx context.Context) (sessions []authoringcore.AuthoringSession, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringPath(authoringRuntimeRef("sessions")); err != nil {
			return err
		}

		sessions, err = r.listSessions(ctx)

		return err
	})

	return sessions, err
}

func (r *fileAuthoringSessionRepository) AppendTranscriptEvents(ctx context.Context, id string, events []AssistedTranscriptEvent) error {
	return WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		session, err := r.getGuardedSession(ctx, id)
		if err != nil {
			return err
		}

		if session.Guided != nil {
			return workflowstate.ErrConflict
		}

		if err := r.appendTranscriptEvents(ctx, id, events); err != nil {
			return err
		}

		if session.Revision > 0 {
			session.Revision++
			session.UpdatedAt = time.Now().UTC()
			_, err = r.updateSession(ctx, session)
		}

		return err
	})
}

func (r *fileAuthoringSessionRepository) ListTranscriptEvents(ctx context.Context, id string) (events []AssistedTranscriptEvent, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringSessionPaths(id); err != nil {
			return err
		}

		events, err = r.listTranscriptEvents(ctx, id)

		return err
	})

	return events, err
}
