package chcapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

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

	err = stateRepository.UpsertDaemonState(ctx, runningState)
	if err != nil {
		return err
	}

	result, err := runDaemonForegroundIterations(ctx, env, stateRepository, maxIterations)
	if err != nil {
		failDaemonForegroundState(ctx, stateRepository, err)

		return err
	}

	stoppedAt := time.Now().UTC()

	err = stateRepository.UpsertDaemonState(ctx, chc.NewStoppedDaemonState(stoppedAt))
	if err != nil {
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

		err = stateRepository.UpdateDaemonHeartbeat(ctx, heartbeatAt)
		if err != nil {
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

		err = refreshDaemonProjects(ctx, index, projects, &result)
		if err != nil {
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

		err := refreshProjectTicketIndex(ctx, index, projectContext)
		if err != nil {
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

	err = repository.UpsertDaemonState(ctx, recovered)
	if err != nil {
		return chc.DaemonState{}, err
	}

	return recovered, nil
}

func requestDaemonStop(ctx context.Context, stdout io.Writer, env chc.UserEnvironment) error {
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
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write daemon result: %w", err)
		}
	}

	return nil
}
