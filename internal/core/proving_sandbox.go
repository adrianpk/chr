package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	"forge.adrianpk.com/chr/internal/harness"
)

// ProvingSandboxRunOptions describes a package boundary value.
type ProvingSandboxRunOptions struct {
	Worker        harness.Worker
	ProjectDir    string
	ObjectivePath string
	DryRun        bool
	Executor      HarnessExecutor
	Stdout        io.Writer
	Stderr        io.Writer
}

// ProvingSandboxRunResult describes a package boundary value.
type ProvingSandboxRunResult struct {
	Objective  ProvingObjective
	Plan       harnessruntime.RuntimePlan
	Build      harnessruntime.HarnessCommand
	Run        harnessruntime.HarnessCommand
	TicketPath string
	DryRun     bool
}

// RunProvingSandbox executes the corresponding package operation.
func RunProvingSandbox(ctx context.Context, opts ProvingSandboxRunOptions) (ProvingSandboxRunResult, error) {
	source, err := loadProvingObjectiveForProject(opts.ProjectDir, opts.ObjectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return RunProvingSandboxForObjective(ctx, opts, source.ProjectDir, source.Objective)
}

// RunProvingSandboxForObjective executes the corresponding package operation.
func RunProvingSandboxForObjective(ctx context.Context, opts ProvingSandboxRunOptions, projectDir string, objective ProvingObjective) (result ProvingSandboxRunResult, err error) {
	result, err = PlanProvingSandboxRunForObjective(projectDir, objective, opts.Worker)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	result.DryRun = opts.DryRun
	if opts.DryRun {
		return result, nil
	}

	logFile, err := prepareProvingArtifacts(result)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	defer func() {
		err = errors.Join(err, logFile.Close())
	}()

	err = runProvingSandboxCommands(ctx, result, opts, logFile)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return result, nil
}

func prepareProvingArtifacts(result ProvingSandboxRunResult) (*os.File, error) {
	artifact, err := ensureProvingOpsArtifact(result.Plan.ProjectDir, result.Objective)
	if err != nil {
		return nil, err
	}

	for _, path := range []string{result.Plan.WorkerContextPath, result.Plan.ResultPath, result.Plan.LogPath, result.Plan.ValidationPath} {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return nil, fmt.Errorf("create proving artifact directory: %w", err)
		}
	}

	err = os.WriteFile(result.Plan.WorkerContextPath, []byte(RenderProvingWorkerContext(result)), 0o644)
	if err != nil {
		return nil, fmt.Errorf("write proving worker context: %w", err)
	}

	logFile, err := os.OpenFile(result.Plan.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open proving runtime log: %w", err)
	}

	if err := opscore.RefreshArtifactMetadata(artifact.Dir, "running", time.Now()); err != nil {
		if closeErr := logFile.Close(); closeErr != nil {
			return nil, fmt.Errorf("%w; close proving runtime log: %w", err, closeErr)
		}

		return nil, err
	}

	return logFile, nil
}

func runProvingSandboxCommands(ctx context.Context, result ProvingSandboxRunResult, opts ProvingSandboxRunOptions, logFile io.Writer) error {
	executor := opts.Executor
	if executor == nil {
		executor = OSHarnessExecutor{}
	}

	stdout := provingOutput(opts.Stdout, logFile)
	stderr := provingOutput(opts.Stderr, logFile)

	err := executor.Run(ctx, result.Build, stdout, stderr)
	if err != nil {
		return err
	}

	return executor.Run(ctx, result.Run, stdout, stderr)
}

func provingOutput(configured, fallback io.Writer) io.Writer {
	if configured != nil {
		return configured
	}

	return fallback
}

// PlanProvingSandboxRun executes the corresponding package operation.
func PlanProvingSandboxRun(projectDir, objectivePath string, worker harness.Worker) (ProvingSandboxRunResult, error) {
	source, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return PlanProvingSandboxRunForObjective(source.ProjectDir, source.Objective, worker)
}

// PlanProvingSandboxRunForObjective executes the corresponding package operation.
func PlanProvingSandboxRunForObjective(projectDir string, objective ProvingObjective, worker harness.Worker) (ProvingSandboxRunResult, error) {
	if worker == nil {
		return ProvingSandboxRunResult{}, fmt.Errorf("worker adapter is required")
	}

	result, err := planProvingRuntime(projectDir, objective, worker.RuntimeMounts())
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	invocation, err := harnessruntime.NewHarnessInvocation(harnessruntime.HarnessInvocationOptions{
		ExecutionID: objective.ID, TicketKey: objective.TicketKey,
		TicketPath: result.TicketPath, Plan: result.Plan, Worker: worker,
	})
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	result.Build, result.Run = invocation.Build, invocation.Run

	return result, nil
}

func planProvingRuntime(projectDir string, objective ProvingObjective, mounts []harness.RuntimeMount) (ProvingSandboxRunResult, error) {
	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	plan, err := harnessruntime.PlanRuntime(harnessruntime.RuntimePlanOptions{
		ExtraMounts: mounts,
		ProjectDir:  projectDir,
		Config:      config.Runtime,
		TicketKey:   objective.TicketKey,
		RunID:       objective.ID,
	})
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	plan.RunArtifactDir = projectcore.ResolvePath(projectDir, objective.ArtifactRoot)
	plan.WorkerContextPath = projectcore.ResolvePath(projectDir, objective.ContextPath)
	plan.ResultPath = projectcore.ResolvePath(projectDir, objective.ResultPath)
	plan.LogPath = projectcore.ResolvePath(projectDir, objective.LogPath)
	plan.ValidationPath = projectcore.ResolvePath(projectDir, objective.ValidationPath)

	ticketPath, err := findProvingTicketPath(projectDir, objective.TicketKey)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return ProvingSandboxRunResult{Objective: objective, Plan: plan, TicketPath: ticketPath}, nil
}

func ensureProvingOpsArtifact(projectDir string, objective ProvingObjective) (opscore.Artifact, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return opscore.Artifact{}, err
	}

	id := filepath.Base(filepath.Clean(filepath.FromSlash(objective.ArtifactRoot)))

	return opscore.EnsureArtifactEnvelope(opscore.ArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       "Proving " + objective.ID,
		Producer:    "chr",
		ExternalIDs: []string{objective.ID},
	}, id)
}

// RenderProvingWorkerContext renders the corresponding package data.
func RenderProvingWorkerContext(result ProvingSandboxRunResult) string {
	var output strings.Builder

	output.WriteString("# Proving Worker Context\n\n")
	output.WriteString("## Objective Identity\n\n")
	output.WriteString("- objective_id: ")
	output.WriteString(result.Objective.ID)
	output.WriteString("\n")
	output.WriteString("- ticket_key: ")
	output.WriteString(result.Objective.TicketKey)
	output.WriteString("\n")
	output.WriteString("- kind: ")
	output.WriteString(result.Objective.Kind)
	output.WriteString("\n\n")
	output.WriteString("## Artifact Paths\n\n")
	output.WriteString("- context_path: ")
	output.WriteString(harnessruntime.ContainerProjectPath(result.Plan, result.Plan.WorkerContextPath))
	output.WriteString("\n")
	output.WriteString("- result_path: ")
	output.WriteString(harnessruntime.ContainerProjectPath(result.Plan, result.Plan.ResultPath))
	output.WriteString("\n")
	output.WriteString("- log_path: ")
	output.WriteString(harnessruntime.ContainerProjectPath(result.Plan, result.Plan.LogPath))
	output.WriteString("\n")
	output.WriteString("- validation_path: ")
	output.WriteString(harnessruntime.ContainerProjectPath(result.Plan, result.Plan.ValidationPath))
	output.WriteString("\n\n")
	output.WriteString("## Objective\n\n")
	output.WriteString(result.Objective.Body)
	output.WriteString("\n\n")
	output.WriteString("## Output Contract\n\n")
	output.WriteString("Write the result JSON to the result_path above.\n")
	output.WriteString("Do not remove or weaken tests to make validation pass.\n")

	return output.String()
}
