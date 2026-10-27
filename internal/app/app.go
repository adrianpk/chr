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
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	daemoncore "forge.adrianpk.com/chr/internal/core/harness/daemon"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	"forge.adrianpk.com/chr/internal/tui"
)

type appDependencies struct {
	assetsFS fs.FS
}

// Run executes the chr application with explicit process IO and embedded
// command assets.
func Run(ctx context.Context, args []string, processIO IO, commandAssetsFS fs.FS) error {
	factory := commandFactory{io: processIO, deps: appDependencies{assetsFS: commandAssetsFS}}
	cmd := factory.rootCommand()

	osArgs := append([]string{"chr"}, args...)

	return cmd.Run(ctx, osArgs)
}

func (factory commandFactory) rootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "chr",
		Usage:                 "Coordinate agent development workflows from the current project",
		Reader:                factory.io.Stdin,
		Writer:                factory.io.Stdout,
		ErrWriter:             factory.io.Stderr,
		EnableShellCompletion: true,
		Action:                factory.runDefaultTUIAction,
		Commands: []*cli.Command{
			factory.piWorkerCommand(),
			factory.piBridgeCommand(),
			factory.acceptanceCommand(),
			factory.authorCommand(),
			factory.daemonCommand(),
			factory.dumpCommand(),
			factory.gateCommand(),
			factory.harnessCommand(),
			factory.indexCommand(),
			factory.initCommand(),
			factory.maintenanceCommand(),
			factory.metadataCommand(),
			factory.notifyCommand(),
			factory.planCommand(),
			factory.projectCommand(),
			factory.provingCommand(),
			factory.quizCommand(),
			factory.reportCommand(),
			factory.specCommand(),
			factory.runtimeCommand(),
			factory.ticketCommand(),
			factory.telegramCommand(),
			factory.tuiCommand(),
			factory.webCommand(),
			factory.workCommand(),
		},
	}
}

func (factory commandFactory) tuiCommand() *cli.Command {
	return &cli.Command{
		Name:   "tui",
		Usage:  "Open the local chr operational console",
		Flags:  tuiFlags(),
		Action: factory.runTUICommandAction,
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

func (factory commandFactory) runDefaultTUIAction(ctx context.Context, cmd *cli.Command) error {
	env, err := newUserEnvironment("")
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	options, err := factory.loadTUIOptions(ctx, env, projectContext)
	if err != nil {
		return err
	}

	return runManagedTUI(ctx, options, newService(ctx, env, projectContext, factory.deps.assetsFS))
}

func (factory commandFactory) runTUICommandAction(ctx context.Context, cmd *cli.Command) error {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	options, err := factory.loadTUIOptions(ctx, env, projectContext)
	if err != nil {
		return err
	}

	return runManagedTUI(ctx, options, newService(ctx, env, projectContext, factory.deps.assetsFS))
}

func (factory commandFactory) loadTUIOptions(ctx context.Context, env core.UserEnvironment, projectContext commandProjectContext) (tui.Options, error) {
	data := tuiOptionsData{
		ctx:            ctx,
		env:            env,
		projectContext: projectContext,
		stdin:          factory.io.Stdin,
		stdout:         factory.io.Stdout,
		assetsFS:       factory.deps.assetsFS,
		operations:     newService(ctx, env, projectContext, factory.deps.assetsFS),
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
	projectMetadata, err := authoringcore.NewProjectMetadataService(data.projectContext.ProjectDir).Load()
	if err != nil {
		return err
	}

	cases, err := authoringcore.NewCaseMetadataService(data.projectContext.ProjectDir).LoadAll()
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
	workspaceState, err := loadTUIWorkspaceState(data.ctx, data.projectContext)
	workspaceGraph, workspaceIndex, semanticDigestStates := workspaceState.Graph, workspaceState.Index, workspaceState.SemanticDigests

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
) (tuiWorkspaceState, error) {
	result, err := core.NewWorkspaceGraphBuilder(projectContext.ProjectDir).BuildCachedIndex(ctx)
	if err != nil {
		return tuiWorkspaceState{Graph: core.WorkspaceGraph{}, Index: core.WorkspaceIndex{}, SemanticDigests: nil}, err
	}

	cache, err := core.ReadWorkspaceSemanticDigestCache(
		core.WorkspaceSemanticDigestCachePath(projectContext.ProjectDir),
	)
	if err != nil {
		return tuiWorkspaceState{Graph: core.WorkspaceGraph{}, Index: core.WorkspaceIndex{}, SemanticDigests: nil}, err
	}

	states := core.WorkspaceSemanticDigestStates(result.Index, cache, "", "")

	return tuiWorkspaceState{Graph: result.Graph, Index: result.Index, SemanticDigests: states}, nil
}

type tuiOptionsData struct {
	ctx                  context.Context
	env                  core.UserEnvironment
	projectContext       commandProjectContext
	stdin                io.Reader
	stdout               io.Writer
	assetsFS             fs.FS
	tickets              []ticketcore.IndexedTicket
	documents            []tui.DocumentView
	dumps                []tui.DumpView
	quizzes              []tui.QuizView
	specs                []tui.SpecView
	plans                []tui.PlanView
	projectMetadata      authoringcore.ProjectMetadata
	cases                []authoringcore.CaseMetadata
	authoringSessions    []authoringcore.Session
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
		HarnessExecutor:      newTUIHarnessSelectionExecutor(data.operations),
		GuidedExecutor:       tuiGuidedExecutor{operations: data.operations},
	}
}

func loadTUITickets(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) ([]ticketcore.IndexedTicket, error) {
	board, err := newService(ctx, env, projectContext, assetsFS).Board()
	if err != nil {
		return nil, err
	}

	return board.Tickets, nil
}

func loadTUIAuthoringSessions(
	ctx context.Context,
	projectContext commandProjectContext,
) ([]authoringcore.Session, error) {
	sessions, err := authoringcore.NewSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
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

	runtime, err := validateCodexRuntime(harnessruntime.ValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	events := loadTUIAgentEvents(ctx, projectContext.ProjectDir, config.Runtime, executions, loops)

	view := tui.NewObservabilityView(projectContext.ProjectDir, tui.ObservabilityRecords{Daemon: daemonState, Executions: executions, Loops: loops, Runtime: runtime, Events: events})

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

func loadTUIAgentEvents(ctx context.Context, projectDir string, config harnessruntime.Config, executions []core.ExecutionRecord, loops []core.LoopRunRecord) []core.AgentEvent {
	events := []core.AgentEvent{}

	loop := latestTUILoopRun(loops)
	if loop.ID != "" {
		plan, err := harnessruntime.ResolvePlan(harnessruntime.PlanOptions{
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

func loadTUIDaemonState(ctx context.Context, env core.UserEnvironment, assetsFS fs.FS) (daemoncore.State, error) {
	repository, err := daemoncore.NewRepository(ctx, env, assetsFS)
	if err != nil {
		return daemoncore.State{}, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.Get(ctx)
	if errors.Is(err, daemoncore.ErrStateNotFound) {
		return daemoncore.NewStoppedState(time.Now().UTC()), nil
	}

	if err != nil {
		return daemoncore.State{}, err
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

type tuiWorkspaceState struct {
	Graph           core.WorkspaceGraph
	Index           core.WorkspaceIndex
	SemanticDigests []core.WorkspaceSemanticDigestState
}
