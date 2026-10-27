// Package workspace owns workspace operator resolution and contained configuration files.
package workspace

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

// Userspace describes a package boundary value.
type Userspace struct {
	ProjectDir  string
	OperatorID  string
	Path        string
	RelativeDir string
}

// ResolveUserspaceOptions describes a package boundary value.
type ResolveUserspaceOptions struct {
	ProjectDir string
	LookupEnv  func(string) string
}

// EnsureUserspaceOptions describes a package boundary value.
type EnsureUserspaceOptions struct {
	ProjectDir string
}

// ResolveUserspace executes the corresponding package operation.
func ResolveUserspace(opts ResolveUserspaceOptions) (Userspace, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return Userspace{}, err
	}

	lookupEnv := opts.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	operatorID := strings.TrimSpace(lookupEnv(ActiveOperatorIDEnv))
	if operatorID == "" {
		return Userspace{}, fmt.Errorf("%s is required", ActiveOperatorIDEnv)
	}

	if !ValidOperatorID(operatorID) {
		return Userspace{}, fmt.Errorf("invalid operator id: %s", operatorID)
	}

	relativeDir := UserRelativePath()

	return Userspace{
		ProjectDir:  projectDir,
		OperatorID:  operatorID,
		Path:        UserPath(projectDir),
		RelativeDir: relativeDir,
	}, nil
}

// EnsureUserspace executes the corresponding package operation.
func EnsureUserspace(ctx context.Context, opts EnsureUserspaceOptions) (Userspace, error) {
	err := ctx.Err()
	if err != nil {
		return Userspace{}, err
	}

	userspace, err := ResolveUserspace(ResolveUserspaceOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return Userspace{}, err
	}

	err = os.MkdirAll(userspace.Path, 0o755)
	if err != nil {
		return Userspace{}, fmt.Errorf("create project userspace: %w", err)
	}

	return userspace, nil
}

// ValidOperatorID executes the corresponding package operation.
func ValidOperatorID(value string) bool {
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
