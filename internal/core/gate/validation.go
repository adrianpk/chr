package gate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// DefaultValidationCommand defines a package boundary constant.
const DefaultValidationCommand = "go test ./..."

// ValidationRunner describes a package boundary value.
type ValidationRunner struct{}

// ValidationRequest describes a package boundary value.
type ValidationRequest struct {
	ProjectDir string
	Command    string
}

// ValidationResult describes a package boundary value.
type ValidationResult struct {
	Command  string
	ExitCode int
	Output   string
	Passed   bool
}

// NewValidationRunner constructs the corresponding package service or value.
func NewValidationRunner() ValidationRunner {
	return ValidationRunner{}
}

// Run executes the configured operation.
func (r ValidationRunner) Run(ctx context.Context, request ValidationRequest) (ValidationResult, error) {
	projectDir := strings.TrimSpace(request.ProjectDir)
	if projectDir == "" {
		return ValidationResult{}, fmt.Errorf("project directory is required")
	}

	command := validationCommand(request.Command)
	if len(command) == 0 {
		return ValidationResult{}, fmt.Errorf("validation command is required")
	}

	cmd := exec.CommandContext(ctx, validationShell(), validationShellArgs(command)...)
	cmd.Dir = projectDir

	var output bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()

	result := ValidationResult{
		Command: strings.Join(command, " "),
		Output:  strings.TrimSpace(output.String()),
		Passed:  err == nil,
	}
	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()

		return result, nil
	}

	return ValidationResult{}, fmt.Errorf("run validation command %q: %w", result.Command, err)
}

func validationCommand(command string) []string {
	command = strings.TrimSpace(command)
	if command == "" {
		command = DefaultValidationCommand
	}

	return strings.Fields(command)
}

func validationShell() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}

	return "sh"
}

func validationShellArgs(command []string) []string {
	joined := strings.Join(command, " ")
	if runtime.GOOS == "windows" {
		return []string{"/C", joined}
	}

	return []string{"-c", joined}
}
