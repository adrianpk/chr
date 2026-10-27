package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

// ActiveOperatorIDEnv defines a package boundary constant.
const ActiveOperatorIDEnv = "CHR_OPERATOR_ID"

// ProjectUserspace describes a package boundary value.
type ProjectUserspace struct {
	ProjectDir  string
	OperatorID  string
	Path        string
	RelativeDir string
}

// ResolveProjectUserspaceOptions describes a package boundary value.
type ResolveProjectUserspaceOptions struct {
	ProjectDir string
	LookupEnv  func(string) string
}

// EnsureProjectUserspaceOptions describes a package boundary value.
type EnsureProjectUserspaceOptions struct {
	ProjectDir string
}

// ResolveProjectUserspace executes the corresponding package operation.
func ResolveProjectUserspace(opts ResolveProjectUserspaceOptions) (ProjectUserspace, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return ProjectUserspace{}, err
	}

	lookupEnv := opts.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	operatorID := strings.TrimSpace(lookupEnv(ActiveOperatorIDEnv))
	if operatorID == "" {
		return ProjectUserspace{}, fmt.Errorf("%s is required", ActiveOperatorIDEnv)
	}

	if !IsValidOperatorID(operatorID) {
		return ProjectUserspace{}, fmt.Errorf("invalid operator id: %s", operatorID)
	}

	relativeDir := UserRelativePath()

	return ProjectUserspace{
		ProjectDir:  projectDir,
		OperatorID:  operatorID,
		Path:        UserPath(projectDir),
		RelativeDir: relativeDir,
	}, nil
}

// EnsureProjectUserspace executes the corresponding package operation.
func EnsureProjectUserspace(ctx context.Context, opts EnsureProjectUserspaceOptions) (ProjectUserspace, error) {
	err := ctx.Err()
	if err != nil {
		return ProjectUserspace{}, err
	}

	userspace, err := ResolveProjectUserspace(ResolveProjectUserspaceOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return ProjectUserspace{}, err
	}

	err = os.MkdirAll(userspace.Path, 0o755)
	if err != nil {
		return ProjectUserspace{}, fmt.Errorf("create project userspace: %w", err)
	}

	return userspace, nil
}

// IsValidOperatorID executes the corresponding package operation.
func IsValidOperatorID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." {
		return false
	}

	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '_' || char == '.' {
			continue
		}

		return false
	}

	return true
}
