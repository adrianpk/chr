package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Registration is journaled with the first explicit native authoring operation.
// Existing unbound authority cannot be silently registered by a work start.
func (s *Store) workflowCoordinatorFiles(ctx context.Context) ([]workflowstate.FileChange, error) {
	owner, err := s.workflowLocalCoordinator(ctx, true)
	if err != nil {
		return nil, err
	}

	markers, err := s.workflowCoordinatorMarkers(ctx)
	if err != nil {
		return nil, err
	}

	if err := checkWorkflowCoordinatorMarkers(markers, owner); err != nil {
		return nil, err
	}

	path, err := s.safePath(workflowCoordinatorPath)
	if err != nil {
		return nil, err
	}

	if _, err := readWorkflowCoordinator(path); err == nil {
		return nil, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	data, err := json.Marshal(owner)

	return []workflowstate.FileChange{{Path: workflowCoordinatorPath, After: data}}, err
}

func (s *Store) checkWorkflowCoordinatorTarget(ctx context.Context, target workflowstate.PreviewTarget) error {
	owner, err := s.checkWorkflowCoordinator(ctx)
	if err != nil {
		return err
	}

	reader := workflowProofReader{store: s}

	files, err := reader.gitOutput(ctx, "ls-tree", "-r", "--name-only", target.Commit, "--", opscore.DirectoryName)
	if err != nil {
		return err
	}

	markers := []workflowCoordinator{}

	for _, path := range strings.Split(files, "\n") {
		if !workflowCoordinatorTargetPath(path) {
			continue
		}

		data, err := reader.gitOutput(ctx, "show", target.Commit+":"+path)
		if err != nil {
			return err
		}

		marker, err := decodeWorkflowCoordinator([]byte(data))
		if err != nil {
			return err
		}

		markers = append(markers, marker)
		if len(markers) > workflowstate.PageLimit {
			return fmt.Errorf("workflow target coordinator observation bound exceeded")
		}
	}

	if len(markers) == 0 {
		return fmt.Errorf("%w: canonical target has no coordinator registration", ErrWorkflowCoordinator)
	}

	return checkWorkflowCoordinatorMarkers(markers, owner)
}

func workflowCoordinatorTargetPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")

	return len(parts) == 3 && parts[0] == opscore.DirectoryName && opsnamespace.Valid(parts[1]) && parts[2] == workflowCoordinatorPath
}

func (s *Store) checkWorkflowFrontierOwnership(ctx context.Context, run workflowInvocationFile) error {
	return s.checkWorkflowFrontierOwnershipPrepared(ctx, run, (*Store).recoverTransition)
}

func (s *Store) checkWorkflowFrontierOwnershipPrepared(ctx context.Context, run workflowInvocationFile, prepare func(*Store) error) error {
	runs, err := s.repositoryWorkflowInvocationsPrepared(ctx, prepare)
	if err != nil {
		return err
	}

	others := []workflowInvocationFile{}

	for _, other := range runs {
		if other.ID != run.ID || other.Namespace != run.Namespace || other.OwnerRoot != run.OwnerRoot {
			others = append(others, other)
		}
	}

	return s.guardWorkflowAdmission(others, run)
}
