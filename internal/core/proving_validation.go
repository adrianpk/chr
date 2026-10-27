package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// ProvingValidationCommandResult describes a package boundary value.
type ProvingValidationCommandResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

// ProvingValidationResult describes a package boundary value.
type ProvingValidationResult struct {
	Objective ProvingObjective
	Commands  []ProvingValidationCommandResult
	Passed    bool
	LogPath   string
}

// ProvingValidationRunner defines a package boundary type.
type ProvingValidationRunner func(context.Context, string, ProvingObjective) (ProvingValidationResult, error)

// RunProvingValidation executes the corresponding package operation.
func RunProvingValidation(ctx context.Context, projectDir, objectivePath string) (ProvingValidationResult, error) {
	source, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	return RunProvingValidationForObjective(ctx, source.ProjectDir, source.Objective)
}

// RunProvingValidationForObjective executes the corresponding package operation.
func RunProvingValidationForObjective(ctx context.Context, projectDir string, objective ProvingObjective) (ProvingValidationResult, error) {
	planned, err := planProvingRuntime(projectDir, objective, nil)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	artifact, err := ensureProvingOpsArtifact(projectDir, planned.Objective)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(planned.Plan.ValidationPath), 0o755)
	if err != nil {
		return ProvingValidationResult{}, fmt.Errorf("create proving validation directory: %w", err)
	}

	commands := ProvingValidationCommands()
	result := ProvingValidationResult{
		Objective: planned.Objective,
		Passed:    true,
		LogPath:   planned.Plan.ValidationPath,
	}

	for _, command := range commands {
		commandResult := runProvingValidationCommand(ctx, planned.Plan.ProjectDir, command)
		result.Commands = append(result.Commands, commandResult)

		if commandResult.ExitCode != 0 {
			result.Passed = false
		}
	}

	err = os.WriteFile(result.LogPath, []byte(RenderProvingValidationResult(result)), 0o644)
	if err != nil {
		return ProvingValidationResult{}, fmt.Errorf("write proving validation log: %w", err)
	}

	if err := opscore.RefreshOpsArtifactMetadata(artifact.Dir, provingValidationArtifactStatus(result), time.Now()); err != nil {
		return ProvingValidationResult{}, err
	}

	return result, nil
}

// ProvingValidationCommands executes the corresponding package operation.
func ProvingValidationCommands() []string {
	return []string{
		"make compile",
		"make lint",
		"make test",
		"make verify",
	}
}

func runProvingValidationCommand(ctx context.Context, projectDir, command string) ProvingValidationCommandResult {
	var stdout bytes.Buffer

	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	cmd.Dir = projectDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return ProvingValidationCommandResult{
		Command:  command,
		ExitCode: commandExitCode(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

// RenderProvingValidationResult renders the corresponding package data.
func RenderProvingValidationResult(result ProvingValidationResult) string {
	var output strings.Builder

	output.WriteString("# Proving Validation Result\n\n")
	output.WriteString("objective_id: ")
	output.WriteString(result.Objective.ID)
	output.WriteString("\n")
	output.WriteString("ticket_key: ")
	output.WriteString(result.Objective.TicketKey)
	output.WriteString("\n")
	output.WriteString("passed: ")
	output.WriteString(strconv.FormatBool(result.Passed))
	output.WriteString("\n\n")

	for _, command := range result.Commands {
		output.WriteString("## Command: ")
		output.WriteString(command.Command)
		output.WriteString("\n\n")
		output.WriteString("exit_code: ")
		output.WriteString(strconv.Itoa(command.ExitCode))
		output.WriteString("\n\n")
		output.WriteString("### Stdout\n\n")
		output.WriteString(command.Stdout)

		if !strings.HasSuffix(command.Stdout, "\n") {
			output.WriteString("\n")
		}

		output.WriteString("\n### Stderr\n\n")
		output.WriteString(command.Stderr)

		if !strings.HasSuffix(command.Stderr, "\n") {
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// RenderProvingValidationFeedback renders the corresponding package data.
func RenderProvingValidationFeedback(result ProvingValidationResult) string {
	var output strings.Builder

	output.WriteString("The previous attempt did not pass validation.\n\n")
	output.WriteString("Validation artifact: ")
	output.WriteString(result.LogPath)
	output.WriteString("\n\n")

	for _, command := range result.Commands {
		if command.ExitCode == 0 {
			continue
		}

		output.WriteString("## Failing Command: ")
		output.WriteString(command.Command)
		output.WriteString("\n\n")
		output.WriteString("exit_code: ")
		output.WriteString(strconv.Itoa(command.ExitCode))
		output.WriteString("\n\n")

		if strings.TrimSpace(command.Stdout) != "" {
			output.WriteString("### Stdout\n\n")
			output.WriteString(strings.TrimSpace(command.Stdout))
			output.WriteString("\n\n")
		}

		if strings.TrimSpace(command.Stderr) != "" {
			output.WriteString("### Stderr\n\n")
			output.WriteString(strings.TrimSpace(command.Stderr))
			output.WriteString("\n\n")
		}
	}

	output.WriteString("Fix the implementation and any missing tests. Do not weaken validation.\n")

	return output.String()
}

// ProvingValidationFailureSignature executes the corresponding package operation.
func ProvingValidationFailureSignature(result ProvingValidationResult) string {
	hash := sha256.New()

	for _, command := range result.Commands {
		if command.ExitCode == 0 {
			continue
		}

		hash.Write([]byte(command.Command))
		hash.Write([]byte{0})
		hash.Write([]byte(strconv.Itoa(command.ExitCode)))
		hash.Write([]byte{0})
		hash.Write([]byte(command.Stdout))
		hash.Write([]byte{0})
		hash.Write([]byte(command.Stderr))
		hash.Write([]byte{0})
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func provingValidationArtifactStatus(result ProvingValidationResult) string {
	if result.Passed {
		return "passed"
	}

	return "failed"
}
