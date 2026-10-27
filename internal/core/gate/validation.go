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

// DefaultGateValidationCommand defines a package boundary constant.
const DefaultGateValidationCommand = "go test ./..."

// GateValidationRunner describes a package boundary value.
type GateValidationRunner struct{}

// GateValidationRequest describes a package boundary value.
type GateValidationRequest struct {
	ProjectDir string
	Command    string
}

// GateValidationResult describes a package boundary value.
type GateValidationResult struct {
	Command  string
	ExitCode int
	Output   string
	Passed   bool
}

// NewGateValidationRunner constructs the corresponding package service or value.
func NewGateValidationRunner() GateValidationRunner {
	return GateValidationRunner{}
}

// Run executes the configured operation.
func (r GateValidationRunner) Run(ctx context.Context, request GateValidationRequest) (GateValidationResult, error) {
	projectDir := strings.TrimSpace(request.ProjectDir)
	if projectDir == "" {
		return GateValidationResult{}, fmt.Errorf("project directory is required")
	}

	command := gateValidationCommand(request.Command)
	if len(command) == 0 {
		return GateValidationResult{}, fmt.Errorf("validation command is required")
	}

	cmd := exec.CommandContext(ctx, gateValidationShell(), gateValidationShellArgs(command)...)
	cmd.Dir = projectDir

	var output bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()

	result := GateValidationResult{
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

	return GateValidationResult{}, fmt.Errorf("run validation command %q: %w", result.Command, err)
}

func gateValidationCommand(command string) []string {
	command = strings.TrimSpace(command)
	if command == "" {
		command = DefaultGateValidationCommand
	}

	return strings.Fields(command)
}

func gateValidationShell() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}

	return "sh"
}

func gateValidationShellArgs(command []string) []string {
	joined := strings.Join(command, " ")
	if runtime.GOOS == "windows" {
		return []string{"/C", joined}
	}

	return []string{"-c", joined}
}
