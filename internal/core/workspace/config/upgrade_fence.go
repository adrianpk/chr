package config

import (
	"errors"
	"fmt"
	"os"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// WorkflowUpgradePending identifies the retained workflow configuration upgrade intent.
const WorkflowUpgradePending = "tmp/workflow-upgrade.json"

// RequireNoWorkflowUpgrade prevents policy reads while an explicit upgrade is pending.
func RequireNoWorkflowUpgrade(project string) error {
	path, err := workspacecore.ConfigPath(project, WorkflowUpgradePending)
	if err != nil {
		return err
	}

	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("workflow upgrade pending: retry the original explicit upgrade before reading policy")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
