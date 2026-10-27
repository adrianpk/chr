package authoring

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileSessionRepository) CreateSession(ctx context.Context, session Session) (paths SessionPaths, err error) {
	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringSessionPaths(session.ID); err != nil {
			return err
		}

		paths, err = r.createSession(ctx, session)

		return err
	})

	return paths, err
}

func (r *fileSessionRepository) UpdateSession(ctx context.Context, session Session) (paths SessionPaths, err error) {
	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
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

func (r *fileSessionRepository) GetSession(ctx context.Context, id string) (session Session, err error) {
	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, id)

		return err
	})

	return session, err
}

func (r *fileSessionRepository) getGuardedSession(ctx context.Context, id string) (Session, error) {
	if _, err := r.authoringSessionPaths(id); err != nil {
		return Session{}, err
	}

	return r.getSession(ctx, id)
}

func (r *fileSessionRepository) ListSessions(ctx context.Context) (sessions []Session, err error) {
	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringPath(authoringRuntimeRef("sessions")); err != nil {
			return err
		}

		sessions, err = r.listSessions(ctx)

		return err
	})

	return sessions, err
}

func (r *fileSessionRepository) AppendTranscriptEvents(ctx context.Context, id string, events []AssistedTranscriptEvent) error {
	return WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
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

func (r *fileSessionRepository) ListTranscriptEvents(ctx context.Context, id string) (events []AssistedTranscriptEvent, err error) {
	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
		if _, err := r.authoringSessionPaths(id); err != nil {
			return err
		}

		events, err = r.listTranscriptEvents(ctx, id)

		return err
	})

	return events, err
}

// ReadOwnedSession reads a guarded session only inside the active
// authoring mutation. Native admission and control retain their own proof checks.
func ReadOwnedSession(ctx context.Context, projectDir, id string) (Session, error) {
	absolute, err := filepath.Abs(projectDir)
	if err != nil || !holdsAuthoringOwner(ctx, absolute) {
		return Session{}, errors.Join(workflowstate.ErrConflict, err)
	}

	r := &fileSessionRepository{projectDir: projectDir}

	return r.getGuardedSession(ctx, id)
}
