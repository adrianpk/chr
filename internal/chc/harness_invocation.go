package chc

import (
	"fmt"
	"path/filepath"
	"strings"
)

type HarnessInvocationOptions struct {
	ExecutionID string
	TicketKey   string
	Plan        RuntimePlan
}

type HarnessInvocation struct {
	ExecutionID string
	TicketKey   string
	Build       HarnessCommand
	Run         HarnessCommand
	ResultPath  string
	LogPath     string
}

type HarnessCommand struct {
	Program string
	Args    []string
}

func NewHarnessInvocation(opts HarnessInvocationOptions) (HarnessInvocation, error) {
	if opts.ExecutionID == "" {
		return HarnessInvocation{}, fmt.Errorf("execution id is required")
	}

	if opts.TicketKey == "" {
		return HarnessInvocation{}, fmt.Errorf("ticket key is required")
	}

	if opts.Plan.ProjectDir == "" {
		return HarnessInvocation{}, fmt.Errorf("runtime project dir is required")
	}

	return HarnessInvocation{
		ExecutionID: opts.ExecutionID,
		TicketKey:   opts.TicketKey,
		Build:       harnessBuildCommand(opts.Plan),
		Run:         harnessRunCommand(opts.ExecutionID, opts.TicketKey, opts.Plan),
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

func harnessRunCommand(executionID, ticketKey string, plan RuntimePlan) HarnessCommand {
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
		"--env", "CHC_RESULT_PATH="+containerProjectPath(plan, plan.ResultPath),
		"--env", "CHC_LOG_PATH="+containerProjectPath(plan, plan.LogPath),
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
