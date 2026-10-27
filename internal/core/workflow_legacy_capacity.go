package core

import (
	"context"
	"errors"
	"os"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

func acquireLegacyWorkflowCapacity(ctx context.Context, projectDir string) (reservation *os.File, err error) {
	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return nil, err
	}

	store := NewWorkflowStore(layout, "")
	err = store.withLock(ctx, func() error {
		runs, readErr := store.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		if _, err := store.workflowRepositoryCapacity(ctx, runs); err != nil {
			return err
		}

		if err := store.checkLegacyWorkflowCoordinator(ctx); err != nil {
			return err
		}

		reservation, readErr = acquireWorkflowCapacity(ctx, projectDir)

		return readErr
	})

	return reservation, err
}

func (s *WorkflowStore) checkLegacyWorkflowCoordinator(ctx context.Context) error {
	markers, err := s.workflowCoordinatorMarkers(ctx)
	if err != nil {
		return err
	}

	if len(markers) == 0 {
		_, err := s.workflowLocalCoordinator(ctx, false)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	_, err = s.checkWorkflowCoordinator(ctx)

	return err
}
