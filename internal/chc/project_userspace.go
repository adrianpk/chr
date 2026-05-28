package chc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const ActiveOperatorIDEnv = "CHC_OPERATOR_ID"

type ProjectUserspace struct {
	ProjectDir  string
	OperatorID  string
	Path        string
	RelativeDir string
}

type ResolveProjectUserspaceOptions struct {
	ProjectDir string
	LookupEnv  func(string) string
}

type EnsureProjectUserspaceOptions struct {
	ProjectDir string
}

func ResolveProjectUserspace(opts ResolveProjectUserspaceOptions) (ProjectUserspace, error) {
	projectDir, err := resolveProjectDir(opts.ProjectDir)
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
