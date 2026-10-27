package core

import (
	"errors"
	"os"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileAuthoringSessionRepository) checkPlanWorkflowPending(reference string) error {
	root := filepath.Dir(filepath.Dir(filepath.Dir(reference)))

	path, err := r.authoringPath(filepath.ToSlash(filepath.Join(root, workflowstate.JournalPath)))
	if err != nil {
		return err
	}

	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	return nil
}
