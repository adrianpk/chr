package merge

import (
	"path/filepath"
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidPath accepts only bounded, canonical local pathnames.
func ValidPath(path string) bool {
	return len(path) > 0 && len(path) <= 4096 && filepath.IsLocal(path) && filepath.ToSlash(filepath.Clean(path)) == path && !strings.ContainsAny(path, "\\\x00\n\r") && path != "."
}

// CheckPaths rejects duplicate, invalid or excessive scoped paths.
func CheckPaths(paths []string) error {
	if len(paths) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, path := range paths {
		if !ValidPath(path) || seen[path] {
			return workflowstate.ErrConflict
		}

		seen[path] = true
	}

	return nil
}

func pathWithin(path string, roots []string) bool {
	return slices.ContainsFunc(roots, func(root string) bool {
		return path == root || strings.HasPrefix(path, root+"/")
	})
}

func controlPath(path string) bool {
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

// PathAllowed excludes control records and protected paths from the editable scope.
func PathAllowed(input workflowstate.MergerInput, path string) bool {
	return ValidPath(path) && !controlPath(path) && pathWithin(path, input.AllowedPaths) && !pathWithin(path, input.ProtectedPaths)
}

// CheckScope checks allowed/protected paths and bounded host validation commands.
func CheckScope(input workflowstate.MergerInput) error {
	if len(input.AllowedPaths) == 0 {
		return workflowstate.ErrConflict
	}

	if err := CheckPaths(input.AllowedPaths); err != nil {
		return err
	}

	if err := CheckPaths(input.ProtectedPaths); err != nil {
		return err
	}

	for _, path := range input.AllowedPaths {
		if controlPath(path) || pathWithin(path, input.ProtectedPaths) {
			return workflowstate.ErrConflict
		}
	}

	return validateWorkflowMergerCommands(input.ValidationCommands)
}

func validateWorkflowMergerCommands(commands [][]string) error {
	if len(commands) == 0 || len(commands) > 16 {
		return workflowstate.ErrConflict
	}

	for _, command := range commands {
		if !ValidCommand(command) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

// ValidCommand checks bounded command arguments without granting execution.
func ValidCommand(command []string) bool {
	if len(command) == 0 || len(command) > workflowstate.PageLimit || command[0] == "" {
		return false
	}

	for _, argument := range command {
		if len(argument) > 4096 || strings.ContainsAny(argument, "\x00\n\r") {
			return false
		}
	}

	return true
}
