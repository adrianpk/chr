package app

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"sort"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
)

type appDependencies struct {
	assetsFS fs.FS
}

const appDependenciesMetadataKey = "app.dependencies"

func appDependenciesFromCommand(cmd *cli.Command) (appDependencies, error) {
	value, ok := cmd.Root().Metadata[appDependenciesMetadataKey]
	if !ok {
		return appDependencies{}, errors.New("application dependencies are not configured")
	}

	deps, ok := value.(appDependencies)
	if !ok {
		return appDependencies{}, errors.New("application dependencies have an invalid type")
	}

	return deps, nil
}

// Run executes the chr application with explicit process IO and embedded
// command assets.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, commandAssetsFS fs.FS) error {
	deps := appDependencies{assetsFS: commandAssetsFS}
	cmd := newRootCommand(stdin, stdout, stderr, deps)

	osArgs := append([]string{"chr"}, args...)

	return cmd.Run(ctx, osArgs)
}

func newRootCommand(stdin io.Reader, stdout, stderr io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:      "chr",
		Usage:     "Coordinate agent development workflows from the current project",
		Reader:    stdin,
		Writer:    stdout,
		ErrWriter: stderr,
		Metadata: map[string]any{
			appDependenciesMetadataKey: deps,
		},
		EnableShellCompletion: true,
		Action:                runDefaultTUIAction(stdin, stdout, deps),
		Commands: []*cli.Command{
			newAcceptanceCommand(stdout, deps),
			newAuthorCommand(stdout, deps),
			newDaemonCommand(stdout, deps),
			newDumpCommand(stdout, deps),
			newGateCommand(stdout, deps),
			newHarnessCommand(stdout, deps),
			newIndexCommand(stdout, deps),
			newInitCommand(stdout, deps),
			newMaintenanceCommand(stdout, deps),
			newMetadataCommand(stdout, deps),
			newNotifyCommand(stdout, deps),
			newPlanCommand(stdout, deps),
			newProjectCommand(stdout, deps),
			newProvingCommand(stdout, deps),
			newQuizCommand(stdout, deps),
			newReportCommand(stdout, deps),
			newSpecCommand(stdout, deps),
			newRuntimeCommand(stdout, deps),
			newTicketCommand(stdout, deps),
			newTelegramCommand(stdin, stdout, deps),
			newTUICommand(stdin, stdout, deps),
			newWebCommand(stdin, stdout, deps),
			newWorkCommand(stdout, deps),
		},
	}
}

func newTUICommand(stdin io.Reader, stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "tui",
		Usage:  "Open the local Chryso operational console",
		Flags:  tuiFlags(),
		Action: runTUICommandAction(stdin, stdout, deps),
	}
}

func tuiFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func runDefaultTUIAction(stdin io.Reader, stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		env, err := newUserEnvironment("")
		if err != nil {
			return err
		}

		projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
			AllowCurrentDirectory: true,
		})
		if err != nil {
			return err
		}

		options, err := loadTUIOptions(ctx, env, projectContext, stdin, stdout, deps)
		if err != nil {
			return err
		}

		return tui.Run(ctx, options)
	}
}

func runTUICommandAction(stdin io.Reader, stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		env, err := newUserEnvironmentFromCommand(cmd)
		if err != nil {
			return err
		}

		projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
			ExplicitDir:           cmd.String("dir"),
			UseExplicitDir:        cmd.IsSet("dir"),
			AllowCurrentDirectory: true,
		})
		if err != nil {
			return err
		}

		options, err := loadTUIOptions(ctx, env, projectContext, stdin, stdout, deps)
		if err != nil {
			return err
		}

		return tui.Run(ctx, options)
	}
}

func loadTUIOptions(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	stdin io.Reader,
	stdout io.Writer,
	deps appDependencies,
) (tui.Options, error) {
	data := tuiOptionsData{
		ctx:            ctx,
		env:            env,
		projectContext: projectContext,
		stdin:          stdin,
		stdout:         stdout,
		assetsFS:       deps.assetsFS,
		operations:     newService(ctx, env, projectContext, deps.assetsFS),
	}

	if err := loadTUIContentData(&data); err != nil {
		return tui.Options{}, err
	}

	if err := loadTUIMetadataData(&data); err != nil {
		return tui.Options{}, err
	}

	if err := loadTUIIndexData(&data); err != nil {
		return tui.Options{}, err
	}

	return newTUIOptions(data), nil
}

func loadTUIContentData(data *tuiOptionsData) error {
	tickets, err := loadTUITickets(data.ctx, data.env, data.projectContext, data.assetsFS)
	if err != nil {
		return err
	}

	documents, err := tui.LoadDocumentViews(data.ctx, data.projectContext.ProjectDir)
	if err != nil {
		return err
	}

	dumps, err := loadTUIDumps(data.operations)
	if err != nil {
		return err
	}

	quizzes, err := loadTUIQuizzes(data.operations)
	if err != nil {
		return err
	}

	specs, err := loadTUISpecs(data.operations)
	if err != nil {
		return err
	}

	plans, err := loadTUIPlans(data.operations)
	if err != nil {
		return err
	}

	data.tickets = tickets
	data.documents = documents
	data.dumps = dumps
	data.quizzes = quizzes
	data.specs = specs
	data.plans = plans

	return nil
}

func loadTUIMetadataData(data *tuiOptionsData) error {
	projectMetadata, err := core.NewProjectMetadataService(data.projectContext.ProjectDir).Load()
	if err != nil {
		return err
	}

	cases, err := core.NewCaseMetadataService(data.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return err
	}

	authoringSessions, err := loadTUIAuthoringSessions(data.ctx, data.projectContext)
	if err != nil {
		return err
	}

	observability, err := loadTUIObservability(data.ctx, data.env, data.projectContext, data.assetsFS)
	if err != nil {
		return err
	}

	workReadiness, err := loadTUIWorkReadiness(data.ctx, data.projectContext)
	if err != nil {
		return err
	}

	data.projectMetadata = projectMetadata
	data.cases = cases
	data.authoringSessions = authoringSessions
	data.observability = observability
	data.workReadiness = workReadiness

	return nil
}

func loadTUIIndexData(data *tuiOptionsData) error {
	workspaceGraph, workspaceIndex, semanticDigestStates, err := loadTUIWorkspaceState(data.ctx, data.projectContext)
	if err != nil {
		return err
	}

	data.workspaceGraph = workspaceGraph
	data.workspaceIndex = workspaceIndex
	data.semanticDigestStates = semanticDigestStates

	return nil
}

func loadTUIWorkspaceState(
	ctx context.Context,
	projectContext commandProjectContext,
) (core.WorkspaceGraph, core.WorkspaceIndex, []core.WorkspaceSemanticDigestState, error) {
	result, err := core.NewWorkspaceGraphBuilder(projectContext.ProjectDir).BuildCachedIndex(ctx)
	if err != nil {
		return core.WorkspaceGraph{}, core.WorkspaceIndex{}, nil, err
	}

	cache, err := core.ReadWorkspaceSemanticDigestCache(
		core.WorkspaceSemanticDigestCachePath(projectContext.ProjectDir),
	)
	if err != nil {
		return core.WorkspaceGraph{}, core.WorkspaceIndex{}, nil, err
	}

	states := core.WorkspaceSemanticDigestStates(result.Index, cache, "", "")

	return result.Graph, result.Index, states, nil
}

type tuiOptionsData struct {
	ctx                  context.Context
	env                  core.UserEnvironment
	projectContext       commandProjectContext
	stdin                io.Reader
	stdout               io.Writer
	assetsFS             fs.FS
	tickets              []core.IndexedTicket
	documents            []tui.DocumentView
	dumps                []tui.DumpView
	quizzes              []tui.QuizView
	specs                []tui.SpecView
	plans                []tui.PlanView
	projectMetadata      core.ProjectMetadata
	cases                []core.CaseMetadata
	authoringSessions    []core.AuthoringSession
	observability        tui.ObservabilityView
	workReadiness        tui.WorkReadinessView
	workspaceGraph       core.WorkspaceGraph
	workspaceIndex       core.WorkspaceIndex
	semanticDigestStates []core.WorkspaceSemanticDigestState
	operations           service
}

func newTUIOptions(data tuiOptionsData) tui.Options {
	return tui.Options{
		ProjectDir:           data.projectContext.ProjectDir,
		ProjectID:            data.projectContext.ProjectID,
		ProjectMetadata:      data.projectMetadata,
		Cases:                data.cases,
		Stdin:                data.stdin,
		Stdout:               data.stdout,
		Tickets:              data.tickets,
		Documents:            data.documents,
		Dumps:                data.dumps,
		Quizzes:              data.quizzes,
		Specs:                data.specs,
		Plans:                data.plans,
		AuthoringSessions:    data.authoringSessions,
		Observability:        data.observability,
		WorkReadiness:        data.workReadiness,
		WorkspaceGraph:       data.workspaceGraph,
		WorkspaceIndex:       data.workspaceIndex,
		SemanticDigestStates: data.semanticDigestStates,
		ActionExecutor:       newTUIArtifactActionExecutor(data.operations),
		DumpExecutor:         newTUIDumpActionExecutor(data.operations),
		QuizExecutor:         newTUIQuizActionExecutor(data.operations),
		SpecExecutor:         newTUISpecActionExecutor(data.operations),
		PlanExecutor:         newTUIPlanActionExecutor(data.operations),
		TicketExecutor:       newTUITicketActionExecutor(data.operations),
		WorkExecutor:         newTUIWorkActionExecutor(data.operations),
		AgentWorkExecutor:    newTUIAgentWorkRefreshExecutor(data.operations),
		WorkspaceExecutor:    newTUIWorkspaceRefreshExecutor(data.operations),
		ProjectExecutor:      newTUIProjectMetadataExecutor(data.operations),
		CaseExecutor:         newTUICaseMetadataExecutor(data.operations),
		SourceExecutor:       newTUISourceOpenExecutor(data.operations),
	}
}

func loadTUITickets(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) ([]core.IndexedTicket, error) {
	board, err := newService(ctx, env, projectContext, assetsFS).Board()
	if err != nil {
		return nil, err
	}

	return board.Tickets, nil
}

func loadTUIAuthoringSessions(
	ctx context.Context,
	projectContext commandProjectContext,
) ([]core.AuthoringSession, error) {
	sessions, err := core.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func loadTUIObservability(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) (tui.ObservabilityView, error) {
	daemonState, err := loadTUIDaemonState(ctx, env, assetsFS)
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	executions, err := loadTUIExecutions(ctx, env, projectContext.ProjectID, assetsFS)
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	loops, err := loadTUILoopRuns(ctx, env, projectContext.ProjectID, assetsFS)
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	runtime, err := validateCodexRuntime(core.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	events := loadTUIAgentEvents(ctx, projectContext.ProjectDir, config.Runtime, executions, loops)

	view := tui.NewObservabilityView(projectContext.ProjectDir, daemonState, executions, loops, runtime, events)

	view.Authority = core.WorkAuthority(config.Work)
	if view.Authority == core.WorkStateAuthorityTickets {
		view.WorkflowPage, err = loadTUIWorkflowPage(ctx, projectContext)
	}

	return view, err
}

func loadTUIWorkReadiness(ctx context.Context, projectContext commandProjectContext) (tui.WorkReadinessView, error) {
	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return tui.WorkReadinessView{}, err
	}

	report, err := core.CheckWorkReadiness(ctx, core.WorkReadinessOptions{
		ProjectDir: projectContext.ProjectDir,
		Policy:     workConfig.BranchPolicy,
		Workspace:  workConfig.Workspace,
	})
	if err != nil {
		return tui.WorkReadinessView{}, err
	}

	return tui.NewWorkReadinessView(report), nil
}

func loadTUIAgentEvents(ctx context.Context, projectDir string, config core.RuntimeConfig, executions []core.ExecutionRecord, loops []core.LoopRunRecord) []core.AgentEvent {
	events := []core.AgentEvent{}

	loop := latestTUILoopRun(loops)
	if loop.ID != "" {
		plan, err := core.PlanRuntime(core.RuntimePlanOptions{
			ExtraMounts: legacyCodexMounts(),
			ProjectDir:  projectDir,
			Config:      config,
			RunID:       loop.ID,
		})
		if err == nil {
			events = append(events, readTUIAgentEvents(ctx, core.AgentEventStreamPath(plan.RunArtifactDir))...)
		}
	}

	execution := latestTUIExecution(executions)
	if execution.ID != "" {
		events = append(events, readTUIAgentEvents(ctx, core.ExecutionAgentEventStreamPath(execution))...)
	}

	return recentTUIAgentEvents(events, 50)
}

func readTUIAgentEvents(ctx context.Context, path string) []core.AgentEvent {
	stream, err := core.ReadAgentEventStream(ctx, path)
	if err != nil {
		return nil
	}

	return stream.Events
}

func recentTUIAgentEvents(events []core.AgentEvent, limit int) []core.AgentEvent {
	deduped := dedupeTUIAgentEvents(events)
	sort.SliceStable(deduped, func(left, right int) bool {
		return deduped[left].Timestamp.Before(deduped[right].Timestamp)
	})

	if limit > 0 && len(deduped) > limit {
		return deduped[len(deduped)-limit:]
	}

	return deduped
}

func dedupeTUIAgentEvents(events []core.AgentEvent) []core.AgentEvent {
	seen := map[string]bool{}
	deduped := make([]core.AgentEvent, 0, len(events))

	for _, event := range events {
		key := tuiAgentEventKey(event)
		if seen[key] {
			continue
		}

		seen[key] = true

		deduped = append(deduped, event)
	}

	return deduped
}

func tuiAgentEventKey(event core.AgentEvent) string {
	return event.Timestamp.Format(time.RFC3339Nano) + "\x00" +
		event.LoopID + "\x00" +
		event.ExecutionID + "\x00" +
		event.TicketKey + "\x00" +
		string(event.Phase) + "\x00" +
		event.Message
}

func latestTUIExecution(executions []core.ExecutionRecord) core.ExecutionRecord {
	var latest core.ExecutionRecord
	for _, execution := range executions {
		if latest.ID == "" || execution.UpdatedAt.After(latest.UpdatedAt) {
			latest = execution
		}
	}

	return latest
}

func latestTUILoopRun(loops []core.LoopRunRecord) core.LoopRunRecord {
	var latest core.LoopRunRecord
	for _, loop := range loops {
		if latest.ID == "" || loop.UpdatedAt.After(latest.UpdatedAt) {
			latest = loop
		}
	}

	return latest
}

func loadTUIDaemonState(ctx context.Context, env core.UserEnvironment, assetsFS fs.FS) (core.DaemonState, error) {
	repository, err := core.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return core.DaemonState{}, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, core.ErrDaemonStateNotFound) {
		return core.NewStoppedDaemonState(time.Now().UTC()), nil
	}

	if err != nil {
		return core.DaemonState{}, err
	}

	return state, nil
}

func loadTUIExecutions(
	ctx context.Context,
	env core.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]core.ExecutionRecord, error) {
	repository, err := core.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	records, err := repository.ListProjectExecutions(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func loadTUILoopRuns(
	ctx context.Context,
	env core.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]core.LoopRunRecord, error) {
	repository, err := core.NewLoopRunRepository(ctx, env, assetsFS)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	records, err := repository.ListProjectLoopRuns(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return records, nil
}
