package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

//go:embed assets
var assetsFS embed.FS

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := newRootCommand(stdout, stderr)

	osArgs := append([]string{"chc"}, args...)

	return cmd.Run(ctx, osArgs)
}

func newRootCommand(stdout, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:                  "chc",
		Usage:                 "Coordinate agent development workflows from the current project",
		Writer:                stdout,
		ErrWriter:             stderr,
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			newDaemonCommand(stdout),
			newIndexCommand(stdout),
			newInitCommand(stdout),
			newProjectCommand(stdout),
			newRuntimeCommand(stdout),
			newTicketCommand(stdout),
		},
	}
}

func newDaemonCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Coordinate long-running local work",
		Commands: []*cli.Command{
			newDaemonStartCommand(stdout),
			newDaemonStatusCommand(stdout),
			newDaemonStopCommand(stdout),
		},
	}
}

func newDaemonStartCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start daemon coordination",
		Flags: append(projectRegistryFlags(),
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print the daemon start plan without writing state",
			},
			&cli.IntFlag{
				Name:  "max-iterations",
				Usage: "maximum foreground daemon loop iterations",
				Value: 1,
			},
		),
		Action: daemonStartAction(stdout),
	}
}

func newDaemonStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report daemon coordination status",
		Flags:  projectRegistryFlags(),
		Action: daemonStatusAction(stdout),
	}
}

func newDaemonStopCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop daemon coordination",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the daemon stop plan without writing state",
		}),
		Action: daemonStopAction(stdout),
	}
}

func daemonStartAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDaemonStart(ctx, cmd, stdout)
	}
}

func daemonStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDaemonStatus(ctx, cmd, stdout)
	}
}

func daemonStopAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDaemonStop(ctx, cmd, stdout)
	}
}

func runDaemonStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("daemon start does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	maxIterations := cmd.Int("max-iterations")
	if maxIterations <= 0 {
		return fmt.Errorf("max-iterations must be greater than zero")
	}

	decision, err := daemonStartGuardDecision(ctx, env)
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printDaemonStartPlan(stdout, env, decision)
	}

	return runDaemonForegroundLoop(ctx, stdout, env, decision, maxIterations)
}

func runDaemonStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("daemon status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	repository, err := chc.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		return printDaemonStoppedStatus(stdout, env)
	}

	if err != nil {
		return err
	}

	state, err = recoverStaleDaemonState(ctx, repository, state)
	if err != nil {
		return err
	}

	return printDaemonStateStatus(stdout, env, state)
}

func printDaemonStoppedStatus(stdout io.Writer, env chc.UserEnvironment) error {
	lines := []string{
		"daemon: stopped",
		"state: missing",
		fmt.Sprintf("data_dir: %s", env.DataDir()),
		fmt.Sprintf("state_dir: %s", env.StateDir()),
	}

	return printDaemonLines(stdout, lines)
}

func printDaemonStateStatus(stdout io.Writer, env chc.UserEnvironment, state chc.DaemonState) error {
	lines := []string{
		fmt.Sprintf("daemon: %s", state.Status),
		"state: stored",
		fmt.Sprintf("pid: %d", state.PID),
		fmt.Sprintf("started_at: %s", formatDaemonStatusTime(state.StartedAt)),
		fmt.Sprintf("heartbeat_at: %s", formatDaemonStatusTime(state.HeartbeatAt)),
		fmt.Sprintf("stop_requested_at: %s", formatDaemonStatusTime(state.StopRequestedAt)),
		fmt.Sprintf("failure_reason: %s", state.FailureReason),
		fmt.Sprintf("updated_at: %s", state.UpdatedAt.Format(time.RFC3339)),
		fmt.Sprintf("data_dir: %s", env.DataDir()),
		fmt.Sprintf("state_dir: %s", env.StateDir()),
	}

	return printDaemonLines(stdout, lines)
}

func formatDaemonStatusTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format(time.RFC3339)
}

func runDaemonStop(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("daemon stop does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		return requestDaemonStop(ctx, stdout, env)
	}

	return printDaemonPlan(stdout, "stop", env)
}

func printDaemonPlan(stdout io.Writer, action string, env chc.UserEnvironment) error {
	lines := []string{
		"daemon: planned",
		fmt.Sprintf("action: %s", action),
		"dry_run: true",
		fmt.Sprintf("data_dir: %s", env.DataDir()),
		fmt.Sprintf("state_dir: %s", env.StateDir()),
	}

	return printDaemonLines(stdout, lines)
}

type daemonForegroundLoopResult struct {
	Iterations        int
	ProjectsLoaded    int
	ProjectsReady     int
	ProjectsSkipped   int
	ProjectsRefreshed int
	StopRequested     bool
}

func runDaemonForegroundLoop(
	ctx context.Context,
	stdout io.Writer,
	env chc.UserEnvironment,
	decision chc.DaemonStartGuardDecision,
	maxIterations int,
) error {
	if !decision.Allowed {
		return fmt.Errorf("daemon start blocked: %s", decision.Reason)
	}

	stateRepository, err := chc.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = stateRepository.Close()
	}()

	startedAt := time.Now().UTC()

	runningState := chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         os.Getpid(),
		StartedAt:   &startedAt,
		HeartbeatAt: &startedAt,
		UpdatedAt:   startedAt,
	}
	if err := stateRepository.UpsertDaemonState(ctx, runningState); err != nil {
		return err
	}

	result, err := runDaemonForegroundIterations(ctx, env, stateRepository, maxIterations)
	if err != nil {
		failDaemonForegroundState(ctx, stateRepository, err)

		return err
	}

	stoppedAt := time.Now().UTC()
	if err := stateRepository.UpsertDaemonState(ctx, chc.NewStoppedDaemonState(stoppedAt)); err != nil {
		return err
	}

	return printDaemonForegroundLoopResult(stdout, result)
}

func runDaemonForegroundIterations(
	ctx context.Context,
	env chc.UserEnvironment,
	stateRepository chc.DaemonStateRepository,
	maxIterations int,
) (daemonForegroundLoopResult, error) {
	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return daemonForegroundLoopResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return daemonForegroundLoopResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	result := daemonForegroundLoopResult{}

	for range maxIterations {
		heartbeatAt := time.Now().UTC()
		if err := stateRepository.UpdateDaemonHeartbeat(ctx, heartbeatAt); err != nil {
			return daemonForegroundLoopResult{}, err
		}

		stopRequested, err := daemonStopRequested(ctx, stateRepository)
		if err != nil {
			return daemonForegroundLoopResult{}, err
		}

		if stopRequested {
			result.StopRequested = true

			break
		}

		projects, err := chc.LoadDaemonProjects(ctx, registry)
		if err != nil {
			return daemonForegroundLoopResult{}, err
		}

		result.Iterations++
		result.ProjectsLoaded += len(projects)

		if err := refreshDaemonProjects(ctx, index, projects, &result); err != nil {
			return daemonForegroundLoopResult{}, err
		}
	}

	return result, nil
}

func refreshDaemonProjects(
	ctx context.Context,
	index chc.TicketIndex,
	projects []chc.DaemonProject,
	result *daemonForegroundLoopResult,
) error {
	for _, project := range projects {
		if project.Status != chc.DaemonProjectStatusReady {
			result.ProjectsSkipped++

			continue
		}

		projectContext := commandProjectContext{
			ProjectDir: project.Project.Path,
			ProjectID:  project.Project.ID,
		}
		if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
			return err
		}

		result.ProjectsReady++
		result.ProjectsRefreshed++
	}

	return nil
}

func failDaemonForegroundState(
	ctx context.Context,
	stateRepository chc.DaemonStateRepository,
	cause error,
) {
	failedAt := time.Now().UTC()
	failedState := chc.DaemonState{
		Status:        chc.DaemonStatusFailed,
		FailureReason: cause.Error(),
		UpdatedAt:     failedAt,
	}

	//nolint:errcheck // The original daemon run error is the command result.
	_ = stateRepository.UpsertDaemonState(ctx, failedState)
}

func printDaemonForegroundLoopResult(stdout io.Writer, result daemonForegroundLoopResult) error {
	lines := []string{
		"daemon: stopped",
		fmt.Sprintf("iterations: %d", result.Iterations),
		fmt.Sprintf("stop_requested: %t", result.StopRequested),
		fmt.Sprintf("projects_loaded: %d", result.ProjectsLoaded),
		fmt.Sprintf("projects_ready: %d", result.ProjectsReady),
		fmt.Sprintf("projects_skipped: %d", result.ProjectsSkipped),
		fmt.Sprintf("projects_refreshed: %d", result.ProjectsRefreshed),
	}

	return printDaemonLines(stdout, lines)
}

func recoverStaleDaemonState(
	ctx context.Context,
	repository chc.DaemonStateRepository,
	state chc.DaemonState,
) (chc.DaemonState, error) {
	decision, err := chc.EvaluateDaemonStartGuard(state, time.Now().UTC(), chc.DefaultDaemonStaleAfter)
	if err != nil {
		return chc.DaemonState{}, err
	}

	if !decision.Stale {
		return state, nil
	}

	recoveredAt := time.Now().UTC()
	recovered := state
	recovered.Status = chc.DaemonStatusStale
	recovered.StopRequestedAt = nil
	recovered.FailureReason = "daemon state is stale"
	recovered.UpdatedAt = recoveredAt

	if err := repository.UpsertDaemonState(ctx, recovered); err != nil {
		return chc.DaemonState{}, err
	}

	return recovered, nil
}

func requestDaemonStop(ctx context.Context, stdout io.Writer, env chc.UserEnvironment) error {
	repository, err := chc.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	requestedAt := time.Now().UTC()

	err = repository.RequestDaemonStop(ctx, requestedAt)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		return printDaemonLines(stdout, []string{
			"daemon: stopped",
			"action: stop",
			"state: missing",
		})
	}

	if err != nil {
		return err
	}

	return printDaemonLines(stdout, []string{
		"daemon: stopping",
		"action: stop",
		fmt.Sprintf("stop_requested_at: %s", requestedAt.Format(time.RFC3339)),
	})
}

func daemonStopRequested(ctx context.Context, stateRepository chc.DaemonStateRepository) (bool, error) {
	state, err := stateRepository.GetDaemonState(ctx)
	if err != nil {
		return false, err
	}

	return state.Status == chc.DaemonStatusStopping && state.StopRequestedAt != nil, nil
}

func daemonStartGuardDecision(ctx context.Context, env chc.UserEnvironment) (chc.DaemonStartGuardDecision, error) {
	repository, err := chc.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return chc.DaemonStartGuardDecision{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		return chc.DaemonStartGuardDecision{
			Allowed: true,
			Reason:  "daemon state is missing",
		}, nil
	}

	if err != nil {
		return chc.DaemonStartGuardDecision{}, err
	}

	return chc.EvaluateDaemonStartGuard(state, time.Now().UTC(), chc.DefaultDaemonStaleAfter)
}

func printDaemonStartPlan(stdout io.Writer, env chc.UserEnvironment, decision chc.DaemonStartGuardDecision) error {
	lines := []string{
		"daemon: planned",
		"action: start",
		"dry_run: true",
		fmt.Sprintf("allowed: %t", decision.Allowed),
		fmt.Sprintf("stale: %t", decision.Stale),
		fmt.Sprintf("reason: %s", decision.Reason),
		fmt.Sprintf("data_dir: %s", env.DataDir()),
		fmt.Sprintf("state_dir: %s", env.StateDir()),
	}

	if decision.State.Status != "" {
		lines = append(lines,
			fmt.Sprintf("current_status: %s", decision.State.Status),
			fmt.Sprintf("current_pid: %d", decision.State.PID),
		)
	}

	return printDaemonLines(stdout, lines)
}

func printDaemonLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write daemon result: %w", err)
		}
	}

	return nil
}

func newIndexCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "index",
		Usage: "Manage derived project indexes",
		Commands: []*cli.Command{
			newIndexRebuildCommand(stdout),
			newIndexStatusCommand(stdout),
		},
	}
}

func newIndexRebuildCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "rebuild",
		Usage:  "Rebuild the resolved project's ticket index",
		Flags:  ticketProjectFlags(),
		Action: indexRebuildAction(stdout),
	}
}

func newIndexStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report the resolved project's ticket index status",
		Flags:  ticketProjectFlags(),
		Action: indexStatusAction(stdout),
	}
}

func newInitCommand(stdout io.Writer) *cli.Command {
	workspace := chc.NewWorkspace(assetsFS)

	return &cli.Command{
		Name:  "init",
		Usage: "Create a .chc workspace in the current project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "profile",
				Usage: "workspace profile: standard, personal, minimal",
				Value: chc.DefaultProfile,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print planned changes without writing files",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 0 {
				return fmt.Errorf("init does not accept positional arguments")
			}

			result, err := workspace.Init(ctx, chc.InitOptions{
				ProjectDir: cmd.String("dir"),
				Profile:    cmd.String("profile"),
				DryRun:     cmd.Bool("dry-run"),
			})
			if err != nil {
				return err
			}

			return printInitResult(stdout, result)
		},
	}
}

func newRuntimeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "runtime",
		Usage: "Manage project runtime assets",
		Commands: []*cli.Command{
			newRuntimeInitCommand(stdout),
			newRuntimeValidateCommand(stdout),
		},
	}
}

func newRuntimeInitCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Install missing runtime defaults in the resolved project",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print planned changes without writing files",
			},
		),
		Action: runtimeInitAction(stdout),
	}
}

func newRuntimeValidateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate the resolved project's runtime boundary",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print the validation plan without running Docker",
			},
		),
		Action: runtimeValidateAction(stdout),
	}
}

func runtimeInitAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runRuntimeInit(ctx, cmd, stdout)
	}
}

func runtimeValidateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runRuntimeValidate(ctx, cmd, stdout)
	}
}

func newTicketCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "ticket",
		Usage: "Manage workspace tickets",
		Commands: []*cli.Command{
			newTicketApplyResultCommand(stdout),
			newTicketCommentCommand(stdout),
			newTicketListCommand(stdout),
			newTicketLoopCommand(stdout),
			newTicketNewCommand(stdout),
			newTicketShowCommand(stdout),
			newTicketStartCommand(stdout),
			newTicketStatusCommand(stdout),
		},
	}
}

func newProjectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage registered projects",
		Commands: []*cli.Command{
			newProjectAddCommand(stdout),
			newProjectCurrentCommand(stdout),
			newProjectListCommand(stdout),
			newProjectSelectCommand(stdout),
		},
	}
}

func projectRegistryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "dev-root",
			Usage: "development user environment root",
		},
	}
}

func newUserEnvironmentFromCommand(cmd *cli.Command) (chc.UserEnvironment, error) {
	devRoot := cmd.String("dev-root")
	if devRoot != "" {
		return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
			DevMode: true,
			DevRoot: devRoot,
		})
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		HomeDir:   homeDir,
		XDGConfig: os.Getenv("XDG_CONFIG_HOME"),
		XDGData:   os.Getenv("XDG_DATA_HOME"),
		XDGCache:  os.Getenv("XDG_CACHE_HOME"),
		XDGState:  os.Getenv("XDG_STATE_HOME"),
	})
}

func newProjectAddCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Register a local project",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "name",
			Usage: "registered project name",
		}),
		Action: projectAddAction(stdout),
	}
}

func projectAddAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectAdd(ctx, cmd, stdout)
	}
}

func runProjectAdd(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project add requires exactly one project directory argument")
	}

	projectDir, err := filepath.Abs(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}

	if err := requireWorkspace(projectDir); err != nil {
		return err
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projectName := cmd.String("name")
	if projectName == "" {
		projectName = filepath.Base(projectDir)
	}

	project := chc.Project{
		ID:           projectID(projectDir),
		Name:         projectName,
		Path:         projectDir,
		Enabled:      true,
		RegisteredAt: time.Now().UTC(),
	}

	if err := registry.AddProject(ctx, project); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project add result: %w", err)
	}

	return nil
}

func requireWorkspace(projectDir string) error {
	info, err := os.Stat(filepath.Join(projectDir, chc.DirectoryName))
	if err != nil {
		return fmt.Errorf("inspect project workspace: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("project workspace is not a directory: %s", filepath.Join(projectDir, chc.DirectoryName))
	}

	return nil
}

func projectID(projectDir string) string {
	sum := sha256.Sum256([]byte(projectDir))

	return "project-" + hex.EncodeToString(sum[:8])
}

func newProjectListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List registered projects",
		Flags:  projectRegistryFlags(),
		Action: projectListAction(stdout),
	}
}

func projectListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectList(ctx, cmd, stdout)
	}
}

func runProjectList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projects, err := registry.ListProjects(ctx)
	if err != nil {
		return err
	}

	for _, project := range projects {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\t%t\t%s\n", project.Name, project.Path, project.Enabled, project.ID); err != nil {
			return fmt.Errorf("write project list result: %w", err)
		}
	}

	return nil
}

func newProjectCurrentCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "current",
		Usage:  "Print the active project",
		Flags:  projectRegistryFlags(),
		Action: projectCurrentAction(stdout),
	}
}

func projectCurrentAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectCurrent(ctx, cmd, stdout)
	}
}

func runProjectCurrent(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project current does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	activeProject, err := chc.NewActiveProjectStore(env).LoadActiveProject()
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := registry.GetProjectByID(ctx, activeProject.ProjectID)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project current result: %w", err)
	}

	return nil
}

func newProjectSelectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "select",
		Usage:  "Select the active project",
		Flags:  projectRegistryFlags(),
		Action: projectSelectAction(stdout),
	}
}

func projectSelectAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectSelect(ctx, cmd, stdout)
	}
}

func runProjectSelect(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project select requires exactly one project name or path argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := resolveProjectSelection(ctx, registry, cmd.Args().First())
	if err != nil {
		return err
	}

	if !project.Enabled {
		return fmt.Errorf("project is disabled: %s", project.Name)
	}

	if _, err := resolveRegisteredProjectDirectory(project); err != nil {
		return err
	}

	activeProject, err := chc.NewActiveProject(project.ID)
	if err != nil {
		return err
	}

	store := chc.NewActiveProjectStore(env)
	if err := store.SetActiveProject(activeProject); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project select result: %w", err)
	}

	return nil
}

func resolveProjectSelection(ctx context.Context, registry chc.ProjectRegistry, selector string) (chc.Project, error) {
	project, err := registry.GetProjectByName(ctx, selector)
	if err == nil {
		return project, nil
	}

	if !errors.Is(err, chc.ErrProjectNotFound) {
		return chc.Project{}, err
	}

	projectPath, pathErr := filepath.Abs(selector)
	if pathErr != nil {
		return chc.Project{}, fmt.Errorf("resolve project selector path: %w", pathErr)
	}

	project, err = registry.GetProjectByPath(ctx, projectPath)
	if err != nil {
		return chc.Project{}, fmt.Errorf("resolve project selection: %w", err)
	}

	return project, nil
}

func newTicketNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a ticket in the current workspace",
		Flags:  ticketNewFlags(),
		Action: ticketNewAction(stdout),
	}
}

func newTicketListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List tickets for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: ticketListAction(stdout),
	}
}

func newTicketLoopCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "loop",
		Usage:  "Preview a foreground loop over eligible tickets",
		Flags:  ticketLoopFlags(),
		Action: ticketLoopAction(stdout),
	}
}

func newTicketShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one ticket by key",
		Flags:  ticketProjectFlags(),
		Action: ticketShowAction(stdout),
	}
}

func newTicketCommentCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "comment",
		Usage:  "Append a comment to one ticket",
		Flags:  ticketCommentFlags(),
		Action: ticketCommentAction(stdout),
	}
}

func newTicketApplyResultCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "apply-result",
		Usage:  "Apply one completed execution result to its ticket",
		Flags:  ticketApplyResultFlags(),
		Action: ticketApplyResultAction(stdout),
	}
}

func newTicketStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Update one ticket status",
		Flags:  ticketProjectFlags(),
		Action: ticketStatusAction(stdout),
	}
}

func newTicketStartCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Prepare one ticket execution",
		Flags:  ticketStartFlags(),
		Action: ticketStartAction(stdout),
	}
}

func ticketProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func ticketStartFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the execution preflight without creating an execution record",
		},
	)
}

func ticketLoopFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned loop selection without executing tickets",
		},
		&cli.StringSliceFlag{
			Name:  "status",
			Usage: "eligible ticket status",
		},
		&cli.StringSliceFlag{
			Name:  "priority",
			Usage: "ticket priority filter",
		},
		&cli.StringSliceFlag{
			Name:  "severity",
			Usage: "ticket severity filter",
		},
		&cli.StringSliceFlag{
			Name:  "project-area",
			Usage: "ticket project area filter",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag filter",
		},
		&cli.StringSliceFlag{
			Name:  "type",
			Usage: "ticket type filter",
		},
		&cli.StringSliceFlag{
			Name:  "use-case",
			Usage: "ticket use case filter",
		},
		&cli.StringFlag{
			Name:  "order",
			Usage: "ticket order: key, created_at, priority, severity, priority_severity",
		},
		&cli.IntFlag{
			Name:  "max-iterations",
			Usage: "maximum tickets to execute in one loop run",
		},
	)
}

func ticketApplyResultFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned result application without writing ticket state",
		},
	)
}

func ticketCommentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "author",
			Usage: "comment author",
			Value: "user",
		},
	)
}

func ticketNewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "priority",
			Usage: "ticket priority",
		},
		&cli.StringFlag{
			Name:  "severity",
			Usage: "ticket severity",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "ticket type",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag",
		},
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "ticket use case",
		},
		&cli.StringFlag{
			Name:  "project-area",
			Usage: "ticket project area",
		},
	)
}

func ticketListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketList(ctx, cmd, stdout)
	}
}

func ticketShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketShow(ctx, cmd, stdout)
	}
}

func ticketCommentAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketComment(ctx, cmd, stdout)
	}
}

func ticketStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketStatus(ctx, cmd, stdout)
	}
}

func ticketStartAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketStart(ctx, cmd, stdout)
	}
}

func ticketLoopAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketLoop(ctx, cmd, stdout)
	}
}

func ticketApplyResultAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketApplyResult(ctx, cmd, stdout)
	}
}

type ticketStartPreflight struct {
	Env     chc.UserEnvironment
	Ticket  chc.IndexedTicket
	Record  chc.ExecutionRecord
	Runtime chc.RuntimeValidationResult
}

type ticketStartSelection struct {
	ProjectContext commandProjectContext
	Ticket         chc.IndexedTicket
}

type ticketApplyResultPlan struct {
	ProjectContext commandProjectContext
	Record         chc.ExecutionRecord
	Result         chc.ExecutionResult
	Application    chc.ResultApplication
	ReportPath     string
}

type ticketLoopExecutionResult struct {
	Record       chc.LoopRunRecord
	Executions   []chc.ExecutionRecord
	Applications []chc.ResultApplication
}

type ticketLoopRunOptions struct {
	MaxIterations int
}

var newTicketStartHarnessExecutor = func() chc.HarnessExecutor {
	return chc.OSHarnessExecutor{}
}

func runTicketStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket status requires exactly one ticket key and one status argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	status := cmd.Args().Get(1)
	if err := chc.NewTicketStatusUpdater().UpdateStatus(ctx, chc.UpdateTicketStatusOptions{
		Path:      ticket.Path,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s %s\n", ticket.Key, status); err != nil {
		return fmt.Errorf("write ticket status result: %w", err)
	}

	return nil
}

func runTicketStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket start requires exactly one ticket key argument")
	}

	preflight, err := buildTicketStartPreflight(ctx, cmd)
	if err != nil {
		return chc.NewExecutionFailure(chc.ExecutionFailurePreflight, err)
	}

	if !chc.RuntimeValidationPassed(preflight.Runtime.Checks) {
		preflight.Record.Status = chc.ExecutionStatusBlocked
		preflight.Record.FailureReason = chc.NewExecutionFailure(
			chc.ExecutionFailureRuntime,
			fmt.Errorf("runtime validation failed"),
		).Error()

		if err := printTicketStartResult(stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run")); err != nil {
			return err
		}

		return chc.NewExecutionFailure(chc.ExecutionFailureRuntime, fmt.Errorf("runtime validation failed"))
	}

	if !cmd.Bool("dry-run") {
		record, err := runAndPrintTicketStartExecution(ctx, stdout, preflight)
		if err != nil {
			return err
		}

		preflight.Record = record
	}

	return printTicketStartResult(stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
}

func runTicketLoop(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket loop does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	queryOptions := ticketLoopQueryOptionsFromCommand(cmd, projectContext.ProjectID)

	runOptions, err := ticketLoopRunOptionsFromCommand(cmd)
	if err != nil {
		return err
	}

	tickets, err := chc.NewTicketLoopQuery(index).EligibleTickets(ctx, queryOptions)
	if err != nil {
		return err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, runOptions)

	if !cmd.Bool("dry-run") {
		return runAndPrintTicketLoopExecution(ctx, stdout, env, projectContext, tickets, len(eligibleTickets))
	}

	return printTicketLoopPlan(stdout, projectContext, queryOptions, tickets, cmd.Bool("dry-run"))
}

func ticketLoopQueryOptionsFromCommand(cmd *cli.Command, projectID string) chc.TicketLoopQueryOptions {
	return chc.TicketLoopQueryOptions{
		ProjectID: projectID,
		Statuses:  cmd.StringSlice("status"),
		Filters: chc.TicketLoopFilters{
			Priorities:   cmd.StringSlice("priority"),
			Severities:   cmd.StringSlice("severity"),
			ProjectAreas: cmd.StringSlice("project-area"),
			Tags:         cmd.StringSlice("tag"),
			Types:        cmd.StringSlice("type"),
			UseCases:     cmd.StringSlice("use-case"),
		},
		Order: chc.TicketLoopOrder(cmd.String("order")),
	}
}

func ticketLoopRunOptionsFromCommand(cmd *cli.Command) (ticketLoopRunOptions, error) {
	opts := ticketLoopRunOptions{
		MaxIterations: cmd.Int("max-iterations"),
	}
	if opts.MaxIterations < 0 {
		return ticketLoopRunOptions{}, fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	return opts, nil
}

func ticketsForTicketLoopRun(tickets []chc.IndexedTicket, opts ticketLoopRunOptions) []chc.IndexedTicket {
	if opts.MaxIterations == 0 || len(tickets) <= opts.MaxIterations {
		return tickets
	}

	return tickets[:opts.MaxIterations]
}

func deduplicateTicketLoopTickets(tickets []chc.IndexedTicket) []chc.IndexedTicket {
	selected := make([]chc.IndexedTicket, 0, len(tickets))
	seen := map[string]struct{}{}

	for _, ticket := range tickets {
		if _, ok := seen[ticket.Key]; ok {
			continue
		}

		seen[ticket.Key] = struct{}{}
		selected = append(selected, ticket)
	}

	return selected
}

func runAndPrintTicketLoopExecution(
	ctx context.Context,
	stdout io.Writer,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	tickets []chc.IndexedTicket,
	eligibleCount int,
) error {
	result, err := runTicketLoopExecution(ctx, env, projectContext, tickets, eligibleCount)
	if result.Record.ID == "" {
		return err
	}

	if printErr := printTicketLoopExecution(stdout, result); printErr != nil {
		return printErr
	}

	return err
}

func printTicketLoopPlan(
	stdout io.Writer,
	projectContext commandProjectContext,
	opts chc.TicketLoopQueryOptions,
	tickets []chc.IndexedTicket,
	dryRun bool,
) error {
	order := opts.Order
	if order == "" {
		order = chc.TicketLoopOrderKey
	}

	lines := []string{
		"loop: planned",
		fmt.Sprintf("project: %s", projectContext.ProjectID),
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("order: %s", order),
		fmt.Sprintf("eligible: %d", len(tickets)),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket loop result: %w", err)
		}
	}

	for _, ticket := range tickets {
		if _, err := fmt.Fprintf(
			stdout,
			"ticket: %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
		); err != nil {
			return fmt.Errorf("write ticket loop result: %w", err)
		}
	}

	return nil
}

func runTicketLoopExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	tickets []chc.IndexedTicket,
	eligibleCount int,
) (ticketLoopExecutionResult, error) {
	repository, err := chc.NewLoopRunRepository(ctx, env, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	record := newRunningLoopRunRecord(projectContext.ProjectID, eligibleCount)
	if err := repository.CreateLoopRun(ctx, record); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result := ticketLoopExecutionResult{Record: record}

	for _, ticket := range tickets {
		execution, application, err := executeLoopTicket(ctx, env, projectContext, ticket)
		if err != nil {
			record.FailedCount++

			record = failLoopRunRecord(record, err)
			if updateErr := repository.UpdateLoopRun(ctx, record); updateErr != nil {
				return ticketLoopExecutionResult{}, updateErr
			}

			result.Record = record

			return result, err
		}

		result.Executions = append(result.Executions, execution)
		result.Applications = append(result.Applications, application)

		record.ProcessedCount++
		if application.TicketStatus == chc.TicketResultStatusBlocked {
			record.BlockedCount++

			break
		}
	}

	record = finishLoopRunRecord(record)
	if err := repository.UpdateLoopRun(ctx, record); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result.Record = record

	return result, nil
}

func executeLoopTicket(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	ticket chc.IndexedTicket,
) (chc.ExecutionRecord, chc.ResultApplication, error) {
	preflight, err := buildTicketStartPreflightForTicket(ctx, env, projectContext, ticket)
	if err != nil {
		return chc.ExecutionRecord{}, chc.ResultApplication{}, err
	}

	record, err := runTicketStartExecution(ctx, env, preflight.Record, preflight.Runtime)
	if err != nil {
		return record, chc.ResultApplication{}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(projectContext, record)
	if err != nil {
		return record, chc.ResultApplication{}, err
	}

	if err := applyTicketResult(ctx, env, plan); err != nil {
		return record, chc.ResultApplication{}, err
	}

	return record, plan.Application, nil
}

func newRunningLoopRunRecord(projectID string, eligibleCount int) chc.LoopRunRecord {
	now := time.Now().UTC()

	return chc.LoopRunRecord{
		ID:            loopRunID(projectID, now),
		ProjectID:     projectID,
		Status:        chc.LoopRunStatusRunning,
		EligibleCount: eligibleCount,
		CreatedAt:     now,
		StartedAt:     &now,
		UpdatedAt:     now,
	}
}

func finishLoopRunRecord(record chc.LoopRunRecord) chc.LoopRunRecord {
	finishedAt := time.Now().UTC()

	record.Status = chc.LoopRunStatusSucceeded
	if record.BlockedCount > 0 {
		record.Status = chc.LoopRunStatusBlocked
	}

	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	return record
}

func failLoopRunRecord(record chc.LoopRunRecord, cause error) chc.LoopRunRecord {
	finishedAt := time.Now().UTC()
	record.Status = chc.LoopRunStatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	return record
}

func loopRunID(projectID string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00loop\x00" + createdAt.Format(time.RFC3339Nano)))

	return "loop-" + hex.EncodeToString(sum[:8])
}

func printTicketLoopExecution(stdout io.Writer, result ticketLoopExecutionResult) error {
	skippedCount := result.Record.EligibleCount - result.Record.ProcessedCount - result.Record.FailedCount
	if skippedCount < 0 {
		skippedCount = 0
	}

	lines := []string{
		fmt.Sprintf("loop: %s", result.Record.ID),
		fmt.Sprintf("status: %s", result.Record.Status),
		fmt.Sprintf("eligible: %d", result.Record.EligibleCount),
		fmt.Sprintf("processed: %d", result.Record.ProcessedCount),
		fmt.Sprintf("skipped: %d", skippedCount),
		fmt.Sprintf("blocked: %d", result.Record.BlockedCount),
		fmt.Sprintf("failed: %d", result.Record.FailedCount),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket loop execution result: %w", err)
		}
	}

	for _, execution := range result.Executions {
		if _, err := fmt.Fprintf(stdout, "execution: %s\t%s\t%s\n", execution.ID, execution.TicketKey, execution.Status); err != nil {
			return fmt.Errorf("write ticket loop execution result: %w", err)
		}
	}

	return nil
}

func runTicketApplyResult(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket apply-result requires exactly one execution id argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	plan, err := buildTicketApplyResultPlan(ctx, cmd, env)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		if err := applyTicketResult(ctx, env, plan); err != nil {
			return err
		}
	}

	return printTicketApplyResult(stdout, plan, cmd.Bool("dry-run"))
}

func buildTicketApplyResultPlan(
	ctx context.Context,
	cmd *cli.Command,
	env chc.UserEnvironment,
) (ticketApplyResultPlan, error) {
	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	repository, err := chc.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	record, err := repository.GetExecution(ctx, cmd.Args().First())
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	return buildTicketApplyResultPlanForRecord(projectContext, record)
}

func buildTicketApplyResultPlanForRecord(
	projectContext commandProjectContext,
	record chc.ExecutionRecord,
) (ticketApplyResultPlan, error) {
	if record.ProjectID != projectContext.ProjectID {
		return ticketApplyResultPlan{}, fmt.Errorf("execution project mismatch: %s", record.ProjectID)
	}

	result, err := chc.LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	application, err := chc.NewResultApplication(chc.ResultApplicationOptions{
		Execution:     record,
		Result:        result,
		CommentAuthor: "chc",
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	return ticketApplyResultPlan{
		ProjectContext: projectContext,
		Record:         record,
		Result:         result,
		Application:    application,
		ReportPath:     filepath.Join(record.Artifacts.RunDir, "result-application.md"),
	}, nil
}

func applyTicketResult(ctx context.Context, env chc.UserEnvironment, plan ticketApplyResultPlan) error {
	appliedAt := time.Now().UTC()

	if err := chc.NewTicketResultStateWriter().WriteResultState(ctx, chc.WriteResultStateOptions{
		Application: plan.Application,
		UpdatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := chc.NewTicketResultCommentWriter().WriteResultComment(ctx, chc.WriteResultCommentOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := chc.NewTicketResultValidationWriter().WriteResultValidation(ctx, chc.WriteResultValidationOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := chc.NewTicketResultChangedFilesWriter().WriteChangedFiles(ctx, chc.WriteChangedFilesOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := chc.NewTicketResultBlockedReasonWriter().WriteBlockedReason(ctx, chc.WriteBlockedReasonOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := refreshAppliedTicketIndex(ctx, env, plan, appliedAt); err != nil {
		return err
	}

	return writeTicketApplyResultReport(plan, appliedAt)
}

func refreshAppliedTicketIndex(
	ctx context.Context,
	env chc.UserEnvironment,
	plan ticketApplyResultPlan,
	appliedAt time.Time,
) error {
	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := chc.NewTicketReconciler(index).ReconcileProjectTicket(ctx, chc.ReconcileProjectTicketOptions{
		ProjectID:    plan.ProjectContext.ProjectID,
		ProjectDir:   plan.ProjectContext.ProjectDir,
		TicketPath:   plan.Application.TicketPath,
		ReconciledAt: appliedAt,
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return fmt.Errorf("refresh applied ticket index: %d blockers: %s %w", len(result.Blockers), firstBlocker.RelativePath, firstBlocker.Err)
	}

	return nil
}

func writeTicketApplyResultReport(plan ticketApplyResultPlan, appliedAt time.Time) error {
	report := fmt.Sprintf(
		"# Result Application\n\nexecution: %s\nticket: %s\nstatus: %s\napplied_at: %s\nresult: %s\nlog: %s\nvalidation: %s\n",
		plan.Application.ExecutionID,
		plan.Application.TicketKey,
		plan.Application.TicketStatus,
		appliedAt.Format(time.RFC3339),
		plan.Application.Evidence.ResultPath,
		plan.Application.Evidence.LogPath,
		plan.Application.Evidence.ValidationPath,
	)

	if err := os.WriteFile(plan.ReportPath, []byte(report), 0o644); err != nil {
		return fmt.Errorf("write result application report: %w", err)
	}

	return nil
}

func printTicketApplyResult(stdout io.Writer, plan ticketApplyResultPlan, dryRun bool) error {
	lines := []string{
		fmt.Sprintf("execution: %s", plan.Application.ExecutionID),
		fmt.Sprintf("ticket: %s", plan.Application.TicketKey),
		fmt.Sprintf("ticket_path: %s", plan.Application.TicketPath),
		fmt.Sprintf("status: %s", plan.Application.TicketStatus),
		fmt.Sprintf("result_status: %s", plan.Result.Status),
		fmt.Sprintf("report_path: %s", plan.ReportPath),
		fmt.Sprintf("dry_run: %t", dryRun),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket apply-result output: %w", err)
		}
	}

	return nil
}

func runAndPrintTicketStartExecution(
	ctx context.Context,
	stdout io.Writer,
	preflight ticketStartPreflight,
) (chc.ExecutionRecord, error) {
	record, err := runTicketStartExecution(ctx, preflight.Env, preflight.Record, preflight.Runtime)
	if err == nil {
		return record, nil
	}

	if printErr := printTicketStartResult(stdout, preflight.Ticket, record, preflight.Runtime, false); printErr != nil {
		return chc.ExecutionRecord{}, printErr
	}

	return record, err
}

func buildTicketStartPreflight(ctx context.Context, cmd *cli.Command) (ticketStartPreflight, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	selection, err := selectTicketStartTicket(ctx, cmd, env)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	return buildTicketStartPreflightForTicket(ctx, env, selection.ProjectContext, selection.Ticket)
}

func buildTicketStartPreflightForTicket(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	ticket chc.IndexedTicket,
) (ticketStartPreflight, error) {
	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record := newPendingExecutionRecord(projectContext.ProjectID, ticket)

	result, err := chc.ValidateRuntime(chc.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
		TicketKey:  ticket.Key,
		RunID:      record.ID,
	})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record = chc.ExecutionRecordFromRuntimePlan(record, result.Plan)

	return ticketStartPreflight{
		Env:     env,
		Ticket:  ticket,
		Record:  record,
		Runtime: result,
	}, nil
}

func selectTicketStartTicket(ctx context.Context, cmd *cli.Command, env chc.UserEnvironment) (ticketStartSelection, error) {
	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketStartSelection{}, err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return ticketStartSelection{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return ticketStartSelection{}, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return ticketStartSelection{}, err
	}

	return ticketStartSelection{
		ProjectContext: projectContext,
		Ticket:         ticket,
	}, nil
}

func newPendingExecutionRecord(projectID string, ticket chc.IndexedTicket) chc.ExecutionRecord {
	now := time.Now().UTC()

	return chc.ExecutionRecord{
		ID:         executionID(projectID, ticket.Key, now),
		ProjectID:  projectID,
		TicketKey:  ticket.Key,
		TicketPath: ticket.Path,
		Status:     chc.ExecutionStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func runTicketStartExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	record chc.ExecutionRecord,
	result chc.RuntimeValidationResult,
) (chc.ExecutionRecord, error) {
	repository, err := chc.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return chc.ExecutionRecord{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	if err := repository.CreateExecution(ctx, record); err != nil {
		return chc.ExecutionRecord{}, err
	}

	invocation, err := chc.NewHarnessInvocation(chc.HarnessInvocationOptions{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		TicketPath:  record.TicketPath,
		Plan:        result.Plan,
	})
	if err != nil {
		return chc.ExecutionRecord{}, err
	}

	runner := chc.NewExecutionRunner(repository, newTicketStartHarnessExecutor())

	return runner.Run(ctx, chc.RunExecutionOptions{
		Record:     record,
		Invocation: invocation,
	})
}

func executionID(projectID, ticketKey string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00" + ticketKey + "\x00" + createdAt.Format(time.RFC3339Nano)))

	return "exec-" + hex.EncodeToString(sum[:8])
}

func printTicketStartResult(
	stdout io.Writer,
	ticket chc.IndexedTicket,
	record chc.ExecutionRecord,
	result chc.RuntimeValidationResult,
	dryRun bool,
) error {
	lines := []string{
		fmt.Sprintf("execution: %s", record.ID),
		fmt.Sprintf("ticket: %s", ticket.Key),
		fmt.Sprintf("ticket_path: %s", ticket.Path),
		fmt.Sprintf("ticket_relative_path: %s", ticket.RelativePath),
		fmt.Sprintf("status: %s", record.Status),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket start result: %w", err)
		}
	}

	if err := printTicketStartFailure(stdout, record); err != nil {
		return err
	}

	return printRuntimeValidateResult(stdout, result, dryRun)
}

func printTicketStartFailure(stdout io.Writer, record chc.ExecutionRecord) error {
	if record.FailureReason == "" {
		return nil
	}

	kind := chc.ExecutionFailureKindFromReason(record.FailureReason)
	if kind != "" {
		if _, err := fmt.Fprintf(stdout, "failure_kind: %s\n", kind); err != nil {
			return fmt.Errorf("write ticket start failure kind: %w", err)
		}
	}

	if _, err := fmt.Fprintf(stdout, "failure_reason: %s\n", record.FailureReason); err != nil {
		return fmt.Errorf("write ticket start failure reason: %w", err)
	}

	return printTicketStartDiagnostics(stdout, record)
}

func printTicketStartDiagnostics(stdout io.Writer, record chc.ExecutionRecord) error {
	diagnostics := []struct {
		name string
		path string
	}{
		{name: "diagnostic_log_path", path: record.Artifacts.LogPath},
		{name: "diagnostic_result_path", path: record.Artifacts.ResultPath},
		{name: "diagnostic_validation_path", path: record.Artifacts.ValidationPath},
	}

	for _, diagnostic := range diagnostics {
		if diagnostic.path == "" {
			continue
		}

		if _, err := fmt.Fprintf(stdout, "%s: %s\n", diagnostic.name, diagnostic.path); err != nil {
			return fmt.Errorf("write ticket start diagnostic path: %w", err)
		}
	}

	return nil
}

func runTicketComment(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket comment requires exactly one ticket key and one comment argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	if err := chc.NewTicketCommenter().AddComment(ctx, chc.AddTicketCommentOptions{
		Path:      ticket.Path,
		Author:    cmd.String("author"),
		Body:      cmd.Args().Get(1),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s\n", ticket.Path); err != nil {
		return fmt.Errorf("write ticket comment result: %w", err)
	}

	return nil
}

func runTicketShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket show requires exactly one ticket key argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return err
	}

	return printTicketShow(stdout, ticket)
}

func printTicketShow(stdout io.Writer, ticket chc.IndexedTicket) error {
	lines := []string{
		"key: " + ticket.Key,
		"title: " + ticket.Title,
		"status: " + ticket.Status,
		"priority: " + ticket.Priority,
		"severity: " + ticket.Severity,
		"type: " + ticket.Type,
		"use_case: " + ticket.UseCase,
		"project_area: " + ticket.ProjectArea,
		"path: " + ticket.RelativePath,
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket show result: %w", err)
		}
	}

	return nil
}

func runTicketList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		if _, err := fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
		); err != nil {
			return fmt.Errorf("write ticket list result: %w", err)
		}
	}

	return nil
}

func refreshProjectTicketIndex(ctx context.Context, index chc.TicketIndex, projectContext commandProjectContext) error {
	result, err := chc.NewTicketReconciler(index).ReconcileProjectTickets(ctx, chc.ReconcileProjectTicketsOptions{
		ProjectID:    projectContext.ProjectID,
		ProjectDir:   projectContext.ProjectDir,
		ReconciledAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return fmt.Errorf("refresh ticket index: %d blockers: %s %w", len(result.Blockers), firstBlocker.RelativePath, firstBlocker.Err)
	}

	return nil
}

func ticketNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketNew(ctx, cmd, stdout)
	}
}

func indexRebuildAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runIndexRebuild(ctx, cmd, stdout)
	}
}

func indexStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runIndexStatus(ctx, cmd, stdout)
	}
}

func runIndexRebuild(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index rebuild does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := chc.NewTicketIndexRebuilder(index).RebuildProjectTickets(ctx, chc.RebuildProjectTicketsOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return printIndexRebuildResult(stdout, result)
}

func printIndexRebuildResult(stdout io.Writer, result chc.RebuildProjectTicketsResult) error {
	if _, err := fmt.Fprintf(stdout, "indexed: %d\n", result.IndexedTickets); err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "parse_errors: %d\n", len(result.ParseErrors)); err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	for _, parseErr := range result.ParseErrors {
		if _, err := fmt.Fprintf(stdout, "parse_error: %s %v\n", parseErr.RelativePath, parseErr.Err); err != nil {
			return fmt.Errorf("write index rebuild result: %w", err)
		}
	}

	return nil
}

func runIndexStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	status, err := chc.NewTicketIndexStatusChecker(index).CheckProjectTicketIndex(ctx, chc.CheckProjectTicketIndexOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if err := printIndexStatusResult(stdout, status); err != nil {
		return err
	}

	if ticketIndexHasDrift(status) {
		return fmt.Errorf("ticket index is not current")
	}

	return nil
}

func printIndexStatusResult(stdout io.Writer, status chc.TicketIndexStatus) error {
	if err := printIndexStatusCounts(stdout, status); err != nil {
		return err
	}

	if err := printIndexStatusDetails(stdout, status); err != nil {
		return err
	}

	return nil
}

func printIndexStatusCounts(stdout io.Writer, status chc.TicketIndexStatus) error {
	if _, err := fmt.Fprintf(stdout, "missing: %d\n", len(status.Missing)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "stale: %d\n", len(status.Stale)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "orphaned: %d\n", len(status.Orphaned)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "invalid: %d\n", len(status.Invalid)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	return nil
}

func printIndexStatusDetails(stdout io.Writer, status chc.TicketIndexStatus) error {
	for _, missing := range status.Missing {
		if _, err := fmt.Fprintf(stdout, "missing_ticket: %s\n", missing.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, stale := range status.Stale {
		if _, err := fmt.Fprintf(stdout, "stale_ticket: %s\n", stale.Parsed.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, orphaned := range status.Orphaned {
		if _, err := fmt.Fprintf(stdout, "orphaned_ticket: %s\n", orphaned.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, invalid := range status.Invalid {
		if _, err := fmt.Fprintf(stdout, "invalid_ticket: %s %v\n", invalid.RelativePath, invalid.Err); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	return nil
}

func ticketIndexHasDrift(status chc.TicketIndexStatus) bool {
	return len(status.Missing) > 0 ||
		len(status.Stale) > 0 ||
		len(status.Orphaned) > 0 ||
		len(status.Invalid) > 0
}

func runTicketNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket new requires exactly one title argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	repository := chc.NewTicketRepository(
		projectContext.ProjectDir,
		chc.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
	)

	result, err := repository.CreateTicket(ctx, chc.CreateTicketOptions{
		Title:       cmd.Args().First(),
		Priority:    cmd.String("priority"),
		Severity:    cmd.String("severity"),
		Type:        cmd.String("type"),
		Tags:        cmd.StringSlice("tag"),
		UseCase:     cmd.String("use-case"),
		ProjectArea: cmd.String("project-area"),
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s\n", result.Path); err != nil {
		return fmt.Errorf("write ticket result: %w", err)
	}

	return nil
}

func runRuntimeInit(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("runtime init does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	result, err := chc.NewRuntimeAssets(assetsFS).Init(ctx, chc.RuntimeInitOptions{
		ProjectDir: projectContext.ProjectDir,
		DryRun:     cmd.Bool("dry-run"),
	})
	if err != nil {
		return err
	}

	return printRuntimeInitResult(stdout, result)
}

func runRuntimeValidate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("runtime validate does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	result, err := chc.ValidateRuntime(chc.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return err
	}

	if err := printRuntimeValidateResult(stdout, result, cmd.Bool("dry-run")); err != nil {
		return err
	}

	if !chc.RuntimeValidationPassed(result.Checks) {
		return fmt.Errorf("runtime validation failed")
	}

	return nil
}

func printRuntimeValidateResult(
	stdout io.Writer,
	result chc.RuntimeValidationResult,
	dryRun bool,
) error {
	plan := result.Plan

	if _, err := fmt.Fprintf(stdout, "runtime: %s\n", filepath.Join(plan.ProjectDir, chc.DirectoryName, "runtime")); err != nil {
		return fmt.Errorf("write runtime result: %w", err)
	}

	lines := []string{
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("image: %s", plan.Image),
		fmt.Sprintf("dockerfile: %s", plan.DockerfilePath),
		fmt.Sprintf("workdir: %s", plan.ContainerWorkdir),
		fmt.Sprintf("artifact_dir: %s", plan.ArtifactDir),
		fmt.Sprintf("run_artifact_dir: %s", plan.RunArtifactDir),
		fmt.Sprintf("result_path: %s", plan.ResultPath),
		fmt.Sprintf("log_path: %s", plan.LogPath),
		fmt.Sprintf("validation_command: %s", plan.ValidationCommand),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write runtime plan: %w", err)
		}
	}

	for _, mount := range plan.Mounts {
		if _, err := fmt.Fprintf(stdout, "mount: %s -> %s read_only=%t\n", mount.Source, mount.Target, mount.ReadOnly); err != nil {
			return fmt.Errorf("write runtime mount: %w", err)
		}
	}

	for _, check := range result.Checks {
		status := "fail"
		if check.OK {
			status = "ok"
		}

		if _, err := fmt.Fprintf(stdout, "check: %s %s %s %s\n", status, check.Name, check.Path, check.Message); err != nil {
			return fmt.Errorf("write runtime check: %w", err)
		}
	}

	return nil
}

func printRuntimeInitResult(stdout io.Writer, result chc.RuntimeInitResult) error {
	if _, err := fmt.Fprintf(stdout, "runtime: %s\n", result.Root); err != nil {
		return fmt.Errorf("write runtime result: %w", err)
	}

	for _, created := range result.Created {
		if _, err := fmt.Fprintf(stdout, "create %s\n", created); err != nil {
			return fmt.Errorf("write created path: %w", err)
		}
	}

	for _, skipped := range result.Skipped {
		if _, err := fmt.Fprintf(stdout, "skip existing %s\n", skipped); err != nil {
			return fmt.Errorf("write skipped path: %w", err)
		}
	}

	if _, err := fmt.Fprintln(stdout, "done"); err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}

	return nil
}

func printInitResult(stdout io.Writer, result chc.InitResult) error {
	if _, err := fmt.Fprintf(stdout, "workspace: %s\n", result.Root); err != nil {
		return fmt.Errorf("write workspace result: %w", err)
	}

	for _, created := range result.Created {
		if _, err := fmt.Fprintf(stdout, "create %s\n", created); err != nil {
			return fmt.Errorf("write created path: %w", err)
		}
	}

	for _, skipped := range result.Skipped {
		if _, err := fmt.Fprintf(stdout, "skip existing %s\n", skipped); err != nil {
			return fmt.Errorf("write skipped path: %w", err)
		}
	}

	if _, err := fmt.Fprintln(stdout, "done"); err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}

	return nil
}
