package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// Check the host registration before replay as well as before new mutations.
// The common repository lock is held. An empty repository may still register
// through explicit authoring; copied or disconnected authority cannot do so.
func (s *WorkflowStore) guardWorkflowCoordinatorMutation(ctx context.Context) error {
	if _, err := s.workflowLocalCoordinator(ctx, false); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Read markers directly: discovering owners through invocation recovery here
	// would replay the very journal this guard must protect.
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return err
	}

	for _, layout := range layouts {
		if err := s.guardUnregisteredWorkflowLayout(ctx, layout); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) guardUnregisteredWorkflowLayout(ctx context.Context, layout opscore.Layout) error {
	_, err := readWorkflowCoordinator(filepath.Join(layout.Root(), workflowCoordinatorPath))
	if !errors.Is(err, os.ErrNotExist) {
		return errors.Join(ErrWorkflowCoordinator, err)
	}

	owner := NewWorkflowStore(layout, s.projectID)
	if err := owner.requireNoWorkflowTransition(); err != nil {
		return errors.Join(ErrWorkflowCoordinator, err)
	}

	snapshot, err := owner.readSnapshot(ctx)
	if err != nil || len(snapshot.Records) != 0 {
		return errors.Join(ErrWorkflowCoordinator, err)
	}

	return nil
}

// Reads retain their historical recovery behavior only for a registered owner.
// Without registration they can observe consistent files, but cannot replay.
func (s *WorkflowStore) prepareWorkflowRead(ctx context.Context) error {
	if err := s.guardWorkflowCoordinatorMutation(ctx); err != nil {
		if errors.Is(err, ErrWorkflowCoordinator) {
			return s.requireNoWorkflowTransition()
		}

		return err
	}

	return s.recoverTransition()
}
