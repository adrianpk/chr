package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowUpgradePending = "tmp/workflow-upgrade.json"

// Configuration has its own owner; workflow record safePath never gains access
// to parent directories. Every upgrade file stays inside .chr/default.
func workflowUpgradePath(project, relative string) (string, error) {
	if !filepath.IsLocal(relative) || filepath.Clean(relative) != relative {
		return "", fmt.Errorf("invalid workflow upgrade path")
	}

	project, err := filepath.Abs(project)
	if err != nil {
		return "", err
	}

	path := UserPath(project, relative)
	for parent := path; parent != project; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil {
			if err := checkWorkflowUpgradeFile(info, parent == path); err != nil {
				return "", err
			}
		}
	}

	return path, nil
}

func checkWorkflowUpgradeFile(info os.FileInfo, leaf bool) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("workflow upgrade path cannot traverse a symlink")
	}

	if leaf && !info.Mode().IsRegular() {
		return fmt.Errorf("workflow upgrade requires regular files")
	}

	return nil
}

func requireNoWorkflowUpgrade(project string) error {
	path, err := workflowUpgradePath(project, workflowUpgradePending)
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

func workflowUpgradeConfig(data []byte, prior WorkflowConfiguration, opts WorkflowUpgradeOptions) ([]byte, error) {
	review := opts.ReviewMode
	if review == "" {
		for _, value := range prior.Values {
			if value.Key == "review.mode" {
				review = workflowstate.ReviewMode(value.Value)
			}
		}
	}

	image := setWorkflowUpgradeValue(data, "workflow", "version", "1")
	image = setWorkflowUpgradeValue(image, "review", "mode", string(review))

	image = setWorkflowUpgradeValue(image, "work", "state_authority", string(opts.StateAuthority))
	if opts.StateAuthority == WorkStateAuthorityTickets {
		image = setWorkflowUpgradeValue(image, "work", "workspace_mode", "worktree")
	}

	_, err := parseWorkflowConfig(image)

	return image, err
}

func setWorkflowUpgradeValue(data []byte, section, key, value string) []byte {
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	start, end := -1, len(lines)

	for i, line := range lines {
		if line == section+":" {
			start = i

			continue
		}

		if start >= 0 && line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			end = i

			break
		}
	}

	replacement := "  " + key + ": " + value
	if start < 0 {
		lines = append(lines, section+":", replacement)
	} else {
		lines = replaceWorkflowUpgradeValue(lines, start, end, key, replacement)
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

func replaceWorkflowUpgradeValue(lines []string, start, end int, key, replacement string) []string {
	for i := start + 1; i < end; i++ {
		observed := parseWorkspaceConfigLine(lines[i], "")
		if observed.Parsed && observed.Key == key {
			lines[i] = replacement

			return lines
		}
	}

	return append(lines[:end], append([]string{replacement}, lines[end:]...)...)
}
