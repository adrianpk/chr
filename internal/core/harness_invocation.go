package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// HarnessInvocationOptions describes a package boundary value.
type HarnessInvocationOptions struct {
	ExecutionID string
	TicketKey   string
	TicketPath  string
	Plan        RuntimePlan
}

// HarnessInvocation describes a package boundary value.
type HarnessInvocation struct {
	ExecutionID string
	TicketKey   string
	Build       HarnessCommand
	Run         HarnessCommand
	Validation  HarnessCommand
	ResultPath  string
	LogPath     string
}

// HarnessCommand describes a package boundary value.
type HarnessCommand struct {
	Program string
	Args    []string
}

// NewHarnessInvocation constructs the corresponding package service or value.
func NewHarnessInvocation(opts HarnessInvocationOptions) (HarnessInvocation, error) {
	if opts.ExecutionID == "" {
		return HarnessInvocation{}, fmt.Errorf("execution id is required")
	}

	if opts.TicketKey == "" {
		return HarnessInvocation{}, fmt.Errorf("ticket key is required")
	}

	if opts.TicketPath == "" {
		return HarnessInvocation{}, fmt.Errorf("ticket path is required")
	}

	if opts.Plan.ProjectDir == "" {
		return HarnessInvocation{}, fmt.Errorf("runtime project dir is required")
	}

	run, err := harnessRunCommand(opts.ExecutionID, opts.TicketKey, opts.TicketPath, opts.Plan)
	if err != nil {
		return HarnessInvocation{}, err
	}

	return HarnessInvocation{
		ExecutionID: opts.ExecutionID,
		TicketKey:   opts.TicketKey,
		Build:       harnessBuildCommand(opts.Plan),
		Run:         run,
		Validation:  harnessValidationCommand(opts.ExecutionID, opts.Plan),
		ResultPath:  opts.Plan.ResultPath,
		LogPath:     opts.Plan.LogPath,
	}, nil
}

func harnessBuildCommand(plan RuntimePlan) HarnessCommand {
	return HarnessCommand{
		Program: "docker",
		Args: []string{
			"build",
			"--file",
			plan.DockerfilePath,
			"--tag",
			plan.Image,
			plan.ProjectDir,
		},
	}
}

func harnessRunCommand(executionID, ticketKey, ticketPath string, plan RuntimePlan) (HarnessCommand, error) {
	workerScript, err := harnessWorkerShellScript(executionID, ticketKey, ticketPath, plan)
	if err != nil {
		return HarnessCommand{}, err
	}

	args := []string{
		"run",
		"--rm",
		"--name",
		cleanDockerName("chc-" + executionID),
	}

	for _, mount := range plan.Mounts {
		args = append(args, "--volume", dockerVolumeArg(mount))
	}

	args = append(args,
		"--workdir", plan.ContainerWorkdir,
		"--env", "CHC_EXECUTION_ID="+executionID,
		"--env", "CHC_TICKET_KEY="+ticketKey,
		"--env", "CHC_TICKET_PATH="+containerProjectPath(plan, ticketPath),
		"--env", "CHC_WORKER_CONTEXT_PATH="+containerProjectPath(plan, plan.WorkerContextPath),
		"--env", "CHC_RESULT_PATH="+containerProjectPath(plan, plan.ResultPath),
		"--env", "CHC_LOG_PATH="+containerProjectPath(plan, plan.LogPath),
		"--env", "CHC_VALIDATION_PATH="+containerProjectPath(plan, plan.ValidationPath),
		"--env", "OPENAI_API_KEY",
		plan.Image,
		"/bin/sh",
		"-lc",
		workerScript,
	)

	return HarnessCommand{
		Program: "docker",
		Args:    args,
	}, nil
}

func harnessValidationCommand(executionID string, plan RuntimePlan) HarnessCommand {
	args := []string{
		"run",
		"--rm",
		"--name",
		cleanDockerName("chc-" + executionID + "-validation"),
	}

	for _, mount := range plan.Mounts {
		args = append(args, "--volume", dockerVolumeArg(mount))
	}

	args = append(args,
		"--workdir", plan.ContainerWorkdir,
		"--env", "CHC_EXECUTION_ID="+executionID,
		"--env", "CHC_VALIDATION_PATH="+containerProjectPath(plan, plan.ValidationPath),
		plan.Image,
		"/bin/sh",
		"-lc",
		plan.ValidationCommand,
	)

	return HarnessCommand{
		Program: "docker",
		Args:    args,
	}
}

func harnessWorkerShellScript(executionID, ticketKey, ticketPath string, plan RuntimePlan) (string, error) {
	command, err := NewCodexWorkerCommand(CodexWorkerCommandOptions{
		ExecutionID:    executionID,
		TicketKey:      ticketKey,
		TicketPath:     containerProjectPath(plan, ticketPath),
		ContextPath:    containerProjectPath(plan, plan.WorkerContextPath),
		ResultPath:     containerProjectPath(plan, plan.ResultPath),
		LogPath:        containerProjectPath(plan, plan.LogPath),
		ValidationPath: containerProjectPath(plan, plan.ValidationPath),
	})
	if err != nil {
		return "", err
	}

	return RenderCodexWorkerShellScript(command), nil
}

func dockerVolumeArg(mount RuntimeMount) string {
	arg := mount.Source + ":" + mount.Target
	if mount.ReadOnly {
		arg += ":ro"
	}

	return arg
}

func containerProjectPath(plan RuntimePlan, path string) string {
	relative, err := filepath.Rel(plan.ProjectDir, path)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return filepath.ToSlash(path)
	}

	return cleanContainerPath(filepath.ToSlash(filepath.Join(plan.ContainerWorkdir, relative)))
}

func cleanDockerName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-")

	return replacer.Replace(name)
}
