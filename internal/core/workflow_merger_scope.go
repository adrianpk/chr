package core

import (
	"path/filepath"
	"slices"
	"strings"
)

func validWorkflowMergerPath(path string) bool {
	return len(path) > 0 && len(path) <= 4096 && filepath.IsLocal(path) && filepath.ToSlash(filepath.Clean(path)) == path && !strings.ContainsAny(path, "\\\x00\n\r") && path != "."
}

func validateWorkflowMergerPaths(paths []string) error {
	if len(paths) > WorkflowPageLimit {
		return ErrWorkflowConflict
	}

	seen := map[string]bool{}
	for _, path := range paths {
		if !validWorkflowMergerPath(path) || seen[path] {
			return ErrWorkflowConflict
		}

		seen[path] = true
	}

	return nil
}

func workflowMergerPathWithin(path string, roots []string) bool {
	return slices.ContainsFunc(roots, func(root string) bool {
		return path == root || strings.HasPrefix(path, root+"/")
	})
}

func workflowMergerControlPath(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if slices.Contains([]string{".git", ".chr", ".agents", "AGENTS.md"}, part) {
			return true
		}
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == "ops" {
		return slices.Contains([]string{"ticket", "spec", "plan", "tracker"}, parts[2])
	}

	return false
}

func workflowMergerPathAllowed(input WorkflowMergerInput, path string) bool {
	return validWorkflowMergerPath(path) && !workflowMergerControlPath(path) && workflowMergerPathWithin(path, input.AllowedPaths) && !workflowMergerPathWithin(path, input.ProtectedPaths)
}

func validateWorkflowMergerScope(input WorkflowMergerInput) error {
	if len(input.AllowedPaths) == 0 {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowMergerPaths(input.AllowedPaths); err != nil {
		return err
	}

	if err := validateWorkflowMergerPaths(input.ProtectedPaths); err != nil {
		return err
	}

	for _, path := range input.AllowedPaths {
		if workflowMergerControlPath(path) || workflowMergerPathWithin(path, input.ProtectedPaths) {
			return ErrWorkflowConflict
		}
	}

	return validateWorkflowMergerCommands(input.ValidationCommands)
}

func validateWorkflowMergerCommands(commands [][]string) error {
	if len(commands) == 0 || len(commands) > 16 {
		return ErrWorkflowConflict
	}

	for _, command := range commands {
		if !validWorkflowMergerCommand(command) {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validWorkflowMergerCommand(command []string) bool {
	if len(command) == 0 || len(command) > WorkflowPageLimit || command[0] == "" {
		return false
	}

	for _, argument := range command {
		if len(argument) > 4096 || strings.ContainsAny(argument, "\x00\n\r") {
			return false
		}
	}

	return true
}
