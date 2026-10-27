package runtime

import (
	"fmt"
	"path/filepath"
	"strings"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// HarnessInvocationOptions describes a package boundary value.
type HarnessInvocationOptions struct {
	Worker      harness.Worker
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

	run, err := harnessRunCommand(opts)
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

func harnessRunCommand(opts HarnessInvocationOptions) (HarnessCommand, error) {
	executionID, ticketKey, ticketPath, plan := opts.ExecutionID, opts.TicketKey, opts.TicketPath, opts.Plan

	workerCommand, err := harnessWorkerCommand(opts)
	if err != nil {
		return HarnessCommand{}, err
	}

	args := []string{
		"run",
		"--rm",
		"--name",
		cleanDockerName("chr-" + executionID),
	}

	for _, mount := range plan.Mounts {
		args = append(args, "--volume", dockerVolumeArg(mount))
	}

	args = append(args,
		"--workdir", plan.ContainerWorkdir,
		"--env", "CHR_EXECUTION_ID="+executionID,
		"--env", "CHR_TICKET_KEY="+ticketKey,
		"--env", "CHR_TICKET_PATH="+ContainerProjectPath(plan, ticketPath),
		"--env", "CHR_WORKER_CONTEXT_PATH="+ContainerProjectPath(plan, plan.WorkerContextPath),
		"--env", "CHR_RESULT_PATH="+ContainerProjectPath(plan, plan.ResultPath),
		"--env", "CHR_LOG_PATH="+ContainerProjectPath(plan, plan.LogPath),
		"--env", "CHR_VALIDATION_PATH="+ContainerProjectPath(plan, plan.ValidationPath),
	)
	for _, name := range workerCommand.ForwardEnv {
		args = append(args, "--env", name)
	}

	args = append(args,
		plan.Image,
		"/bin/sh",
		"-lc",
		process.WorkerShellScript(workerCommand),
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
		cleanDockerName("chr-" + executionID + "-validation"),
	}

	for _, mount := range plan.Mounts {
		args = append(args, "--volume", dockerVolumeArg(mount))
	}

	args = append(args,
		"--workdir", plan.ContainerWorkdir,
		"--env", "CHR_EXECUTION_ID="+executionID,
		"--env", "CHR_VALIDATION_PATH="+ContainerProjectPath(plan, plan.ValidationPath),
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

func harnessWorkerCommand(opts HarnessInvocationOptions) (harness.WorkerCommand, error) {
	executionID, ticketKey, ticketPath, plan, worker := opts.ExecutionID, opts.TicketKey, opts.TicketPath, opts.Plan, opts.Worker

	if worker == nil {
		return harness.WorkerCommand{}, fmt.Errorf("worker adapter is required")
	}

	contextPath := ContainerProjectPath(plan, plan.WorkerContextPath)

	command, err := worker.WorkerCommand(harness.WorkerRequest{
		InputPath: contextPath,
		Env: []harness.EnvironmentVariable{
			{Name: "CHR_EXECUTION_ID", Value: executionID},
			{Name: "CHR_TICKET_KEY", Value: ticketKey},
			{Name: "CHR_TICKET_PATH", Value: ContainerProjectPath(plan, ticketPath)},
			{Name: "CHR_WORKER_CONTEXT_PATH", Value: contextPath},
			{Name: "CHR_RESULT_PATH", Value: ContainerProjectPath(plan, plan.ResultPath)},
			{Name: "CHR_LOG_PATH", Value: ContainerProjectPath(plan, plan.LogPath)},
			{Name: "CHR_VALIDATION_PATH", Value: ContainerProjectPath(plan, plan.ValidationPath)},
		},
	})
	if err != nil {
		return harness.WorkerCommand{}, err
	}

	return command, nil
}

func dockerVolumeArg(mount harness.RuntimeMount) string {
	arg := mount.Source + ":" + mount.Target
	if mount.ReadOnly {
		arg += ":ro"
	}

	return arg
}

// ContainerProjectPath maps a mounted host path to its container path.
func ContainerProjectPath(plan RuntimePlan, path string) string {
	relative, err := filepath.Rel(plan.ContainerPathHostRoot(), path)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return filepath.ToSlash(path)
	}

	return cleanContainerPath(filepath.ToSlash(filepath.Join(plan.ContainerWorkdir, relative)))
}

// ContainerPathHostRoot returns the host path mounted at ContainerWorkdir.
func (plan RuntimePlan) ContainerPathHostRoot() string {
	if plan.WorkspaceMode == workspacecore.WorkWorkspaceModeWorktree && strings.TrimSpace(plan.ExecutionWorkspaceDir) != "" {
		return plan.ExecutionWorkspaceDir
	}

	return plan.ProjectDir
}

func cleanDockerName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-")

	return replacer.Replace(name)
}
