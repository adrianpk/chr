package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func (factory commandFactory) daemonCommand() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Coordinate long-running local work",
		Commands: []*cli.Command{
			factory.daemonStartCommand(),
			factory.daemonStatusCommand(),
			factory.daemonStopCommand(),
			factory.daemonWorkflowCommand(),
		},
	}
}

func (factory commandFactory) daemonStartCommand() *cli.Command {
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
		Action: factory.runDaemonStart,
	}
}

func (factory commandFactory) daemonStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report daemon coordination status",
		Flags:  projectRegistryFlags(),
		Action: factory.runDaemonStatus,
	}
}

func (factory commandFactory) daemonStopCommand() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop daemon coordination",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the daemon stop plan without writing state",
		}),
		Action: factory.runDaemonStop,
	}
}

func (factory commandFactory) runDaemonStart(ctx context.Context, cmd *cli.Command) error {
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

	decision, err := daemonStartGuardDecision(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printDaemonStartPlan(factory.io.Stdout, env, decision)
	}

	return factory.runDaemonForegroundLoop(ctx, env, decision, maxIterations)
}

func (factory commandFactory) runDaemonStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("daemon status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	repository, err := core.NewDaemonStateRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, core.ErrDaemonStateNotFound) {
		return printDaemonStoppedStatus(factory.io.Stdout, env)
	}

	if err != nil {
		return err
	}

	state, err = recoverStaleDaemonState(ctx, repository, state)
	if err != nil {
		return err
	}

	return printDaemonStateStatus(factory.io.Stdout, env, state)
}

func printDaemonStoppedStatus(stdout io.Writer, env core.UserEnvironment) error {
	lines := []string{
		"daemon: stopped",
		"state: missing",
		fmt.Sprintf("data_dir: %s", env.DataDir()),
		fmt.Sprintf("state_dir: %s", env.StateDir()),
	}

	return printDaemonLines(stdout, lines)
}

func printDaemonStateStatus(stdout io.Writer, env core.UserEnvironment, state core.DaemonState) error {
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

func (factory commandFactory) runDaemonStop(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("daemon stop does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		return requestDaemonStop(ctx, factory.io.Stdout, env, factory.deps.assetsFS)
	}

	return printDaemonPlan(factory.io.Stdout, "stop", env)
}

func printDaemonPlan(stdout io.Writer, action string, env core.UserEnvironment) error {
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
	WorkflowsObserved int
	WorkflowsWaiting  int
	WorkflowSteps     int
}

func (factory commandFactory) runDaemonForegroundLoop(ctx context.Context, env core.UserEnvironment, decision core.DaemonStartGuardDecision, maxIterations int) error {
	if !decision.Allowed {
		return fmt.Errorf("daemon start blocked: %s", decision.Reason)
	}

	stateRepository, err := core.NewDaemonStateRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = stateRepository.Close()
	}()

	startedAt := time.Now().UTC()

	runningState := core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         os.Getpid(),
		StartedAt:   &startedAt,
		HeartbeatAt: &startedAt,
		UpdatedAt:   startedAt,
	}

	err = stateRepository.UpsertDaemonState(ctx, runningState)
	if err != nil {
		return err
	}

	runner := daemonWorkflowRunner{stdout: factory.io.Stdout, env: env, assets: factory.deps.assetsFS, repository: stateRepository}

	result, err := runner.foregroundIterations(ctx, maxIterations)
	if err != nil {
		failDaemonForegroundState(ctx, stateRepository, err)

		return err
	}

	stoppedAt := time.Now().UTC()

	err = stateRepository.UpsertDaemonState(ctx, core.NewStoppedDaemonState(stoppedAt))
	if err != nil {
		return err
	}

	return printDaemonForegroundLoopResult(factory.io.Stdout, result)
}

func (runner daemonWorkflowRunner) foregroundIterations(ctx context.Context, maxIterations int) (result daemonForegroundLoopResult, err error) {
	registry, err := projectcore.NewRegistry(ctx, runner.env, runner.assets)
	if err != nil {
		return daemonForegroundLoopResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	index, err := core.NewTicketIndex(ctx, runner.env, runner.assets)
	if err != nil {
		return daemonForegroundLoopResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	ctx, finish := watchDaemonHeartbeat(ctx, runner.repository)
	defer func() { err = errors.Join(err, finish()) }()

	for range maxIterations {
		if err := runner.iteration(ctx, registry, index, &result); err != nil {
			return result, err
		}

		if result.StopRequested {
			break
		}
	}

	return result, nil
}

func refreshDaemonProjects(
	ctx context.Context,
	index core.TicketIndex,
	projects []core.DaemonProject,
	result *daemonForegroundLoopResult,
) error {
	for _, project := range projects {
		if project.Status != core.DaemonProjectStatusReady {
			result.ProjectsSkipped++

			continue
		}

		projectContext := commandProjectContext{
			ProjectDir: project.Project.Path,
			ProjectID:  project.Project.ID,
		}

		config, err := core.LoadWorkspaceConfig(project.Project.Path)
		if err != nil {
			return err
		}

		result.ProjectsReady++

		if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets {
			continue
		}

		if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
			return err
		}

		result.ProjectsRefreshed++
	}

	return nil
}

func failDaemonForegroundState(
	ctx context.Context,
	stateRepository core.DaemonStateRepository,
	cause error,
) {
	failedAt := time.Now().UTC()
	failedState := core.DaemonState{
		Status:        core.DaemonStatusFailed,
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
		fmt.Sprintf("workflows_observed: %d", result.WorkflowsObserved),
		fmt.Sprintf("workflows_waiting: %d", result.WorkflowsWaiting),
		fmt.Sprintf("workflow_checkpoint_steps: %d", result.WorkflowSteps),
	}

	return printDaemonLines(stdout, lines)
}

func recoverStaleDaemonState(
	ctx context.Context,
	repository core.DaemonStateRepository,
	state core.DaemonState,
) (core.DaemonState, error) {
	decision, err := core.EvaluateDaemonStartGuard(state, time.Now().UTC(), core.DefaultDaemonStaleAfter)
	if err != nil {
		return core.DaemonState{}, err
	}

	if !decision.Stale {
		return state, nil
	}

	recoveredAt := time.Now().UTC()
	recovered := state
	recovered.Status = core.DaemonStatusStale
	recovered.StopRequestedAt = nil
	recovered.FailureReason = "daemon state is stale"
	recovered.UpdatedAt = recoveredAt

	err = repository.UpsertDaemonState(ctx, recovered)
	if err != nil {
		return core.DaemonState{}, err
	}

	return recovered, nil
}

func requestDaemonStop(ctx context.Context, stdout io.Writer, env core.UserEnvironment, assetsFS fs.FS) error {
	requestedAt := time.Now().UTC()

	result, err := requestDaemonStopState(ctx, env, assetsFS, requestedAt)
	if err != nil {
		return err
	}

	if result.AlreadyStopped {
		return printDaemonLines(stdout, []string{
			"daemon: stopped",
			"action: stop",
			"state: missing",
		})
	}

	return printDaemonLines(stdout, []string{
		"daemon: stopping",
		"action: stop",
		fmt.Sprintf("stop_requested_at: %s", result.RequestedAt.Format(time.RFC3339)),
	})
}

func daemonStopRequested(ctx context.Context, stateRepository core.DaemonStateRepository) (bool, error) {
	state, err := stateRepository.GetDaemonState(ctx)
	if err != nil {
		return false, err
	}

	return state.Status == core.DaemonStatusStopping && state.StopRequestedAt != nil, nil
}

func daemonStartGuardDecision(ctx context.Context, env core.UserEnvironment, assetsFS fs.FS) (core.DaemonStartGuardDecision, error) {
	repository, err := core.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return core.DaemonStartGuardDecision{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, core.ErrDaemonStateNotFound) {
		return core.DaemonStartGuardDecision{
			Allowed: true,
			Reason:  "daemon state is missing",
		}, nil
	}

	if err != nil {
		return core.DaemonStartGuardDecision{}, err
	}

	return core.EvaluateDaemonStartGuard(state, time.Now().UTC(), core.DefaultDaemonStaleAfter)
}

func printDaemonStartPlan(stdout io.Writer, env core.UserEnvironment, decision core.DaemonStartGuardDecision) error {
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
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write daemon result: %w", err)
		}
	}

	return nil
}
