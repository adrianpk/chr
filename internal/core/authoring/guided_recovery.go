package authoring

import (
	"context"
	"errors"
	"io"
	"os"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedHostPage is a bounded inventory for a managed host, across namespaces.
// Directory offsets are observational; repeated sweeps recover concurrent additions.
type GuidedHostPage struct {
	Runs       []Session
	NextOffset int
}

// RecoverGuidedHostPage reconciles the existing authoring journal before listing
// a page. Only the managed lifecycle calls this mutating recovery operation.
func RecoverGuidedHostPage(ctx context.Context, root, project string, offset, limit int) (page GuidedHostPage, err error) {
	if offset < 0 || offset > workflowstate.RecordLimit || limit < 1 || limit > workflowstate.PageLimit {
		return page, workflowstate.ErrConflict
	}

	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: root}

		var readErr error

		page, readErr = r.guidedHostPage(ctx, project, offset, limit)

		return readErr
	})

	return page, err
}

func (r *fileAuthoringSessionRepository) guidedHostPage(ctx context.Context, project string, offset, limit int) (page GuidedHostPage, err error) {
	page = GuidedHostPage{NextOffset: -1}

	path, err := r.authoringPath(authoringRuntimeRef("sessions"))
	if err != nil {
		return page, err
	}

	directory, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return page, nil
	}

	if err != nil {
		return page, err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()

	if offset > 0 {
		if _, err := directory.ReadDir(offset); err != nil {
			return page, guidedDirectoryError(err)
		}
	}

	entries, err := directory.ReadDir(limit + 1)
	if err := guidedDirectoryError(err); err != nil {
		return page, err
	}

	if len(entries) > limit {
		page.NextOffset = offset + limit
		entries = entries[:limit]
	}

	for _, entry := range entries {
		if err := r.appendGuidedHostRun(ctx, project, entry, &page); err != nil {
			return page, err
		}
	}

	return page, nil
}

func guidedDirectoryError(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}

func (r *fileAuthoringSessionRepository) appendGuidedHostRun(ctx context.Context, project string, entry os.DirEntry, page *GuidedHostPage) error {
	if !entry.IsDir() {
		return nil
	}

	session, err := r.getGuardedSession(ctx, entry.Name())
	if err != nil {
		return err
	}

	if session.Guided != nil && session.ProjectID == project {
		page.Runs = append(page.Runs, session)
	}

	return nil
}

// RecoverGuidedHostRun reconciles the authoring journal before reading one exact
// association. Only a managed lifecycle calls this mutating recovery operation.
func RecoverGuidedHostRun(ctx context.Context, root string, scope GuidedUseCaseScope) (session Session, err error) {
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		repository := &fileAuthoringSessionRepository{projectDir: root}

		var readErr error

		session, readErr = repository.getGuardedSession(ctx, scope.RunID)
		if readErr != nil {
			return readErr
		}

		return ValidateGuidedScope(session, scope)
	})

	return session, err
}
