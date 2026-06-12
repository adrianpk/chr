package app

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/tui"
	"forge.adrianpk.com/chc/internal/workspace"
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

// Run executes the chc application with explicit process IO and embedded
// command assets.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, commandAssetsFS fs.FS) error {
	deps := appDependencies{assetsFS: commandAssetsFS}
	cmd := newRootCommand(stdin, stdout, stderr, deps)

	osArgs := append([]string{"chc"}, args...)

	return cmd.Run(ctx, osArgs)
}

func newRootCommand(stdin io.Reader, stdout, stderr io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:      "chc",
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
			newTUICommand(stdin, stdout, deps),
			newWorkCommand(stdout, deps),
		},
	}
}

func newTUICommand(stdin io.Reader, stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "tui",
		Usage:  "Open the local ChrysoCoord operational console",
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
	env workspace.UserEnvironment,
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
	projectMetadata, err := workspace.NewProjectMetadataService(data.projectContext.ProjectDir).Load()
	if err != nil {
		return err
	}

	cases, err := workspace.NewCaseMetadataService(data.projectContext.ProjectDir).LoadAll()
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

	data.projectMetadata = projectMetadata
	data.cases = cases
	data.authoringSessions = authoringSessions
	data.observability = observability

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
) (workspace.WorkspaceGraph, workspace.WorkspaceIndex, []workspace.WorkspaceSemanticDigestState, error) {
	result, err := workspace.NewWorkspaceGraphBuilder(projectContext.ProjectDir).BuildCachedIndex(ctx)
	if err != nil {
		return workspace.WorkspaceGraph{}, workspace.WorkspaceIndex{}, nil, err
	}

	cache, err := workspace.ReadWorkspaceSemanticDigestCache(
		workspace.WorkspaceSemanticDigestCachePath(projectContext.ProjectDir),
	)
	if err != nil {
		return workspace.WorkspaceGraph{}, workspace.WorkspaceIndex{}, nil, err
	}

	states := workspace.WorkspaceSemanticDigestStates(result.Index, cache, "", "")

	return result.Graph, result.Index, states, nil
}

type tuiOptionsData struct {
	ctx                  context.Context
	env                  workspace.UserEnvironment
	projectContext       commandProjectContext
	stdin                io.Reader
	stdout               io.Writer
	assetsFS             fs.FS
	tickets              []workspace.IndexedTicket
	documents            []tui.DocumentView
	dumps                []tui.DumpView
	quizzes              []tui.QuizView
	specs                []tui.SpecView
	plans                []tui.PlanView
	projectMetadata      workspace.ProjectMetadata
	cases                []workspace.CaseMetadata
	authoringSessions    []workspace.AuthoringSession
	observability        tui.ObservabilityView
	workspaceGraph       workspace.WorkspaceGraph
	workspaceIndex       workspace.WorkspaceIndex
	semanticDigestStates []workspace.WorkspaceSemanticDigestState
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
	env workspace.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) ([]workspace.IndexedTicket, error) {
	board, err := newService(ctx, env, projectContext, assetsFS).Board()
	if err != nil {
		return nil, err
	}

	return board.Tickets, nil
}

func loadTUIAuthoringSessions(
	ctx context.Context,
	projectContext commandProjectContext,
) ([]workspace.AuthoringSession, error) {
	sessions, err := workspace.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func loadTUIObservability(
	ctx context.Context,
	env workspace.UserEnvironment,
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

	config, err := workspace.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	runtime, err := workspace.ValidateRuntime(workspace.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return tui.ObservabilityView{}, err
	}

	return tui.NewObservabilityView(projectContext.ProjectDir, daemonState, executions, loops, runtime), nil
}

func loadTUIDaemonState(ctx context.Context, env workspace.UserEnvironment, assetsFS fs.FS) (workspace.DaemonState, error) {
	repository, err := workspace.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return workspace.DaemonState{}, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, workspace.ErrDaemonStateNotFound) {
		return workspace.NewStoppedDaemonState(time.Now().UTC()), nil
	}

	if err != nil {
		return workspace.DaemonState{}, err
	}

	return state, nil
}

func loadTUIExecutions(
	ctx context.Context,
	env workspace.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]workspace.ExecutionRecord, error) {
	repository, err := workspace.NewExecutionRepository(ctx, env, assetsFS)
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
	env workspace.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]workspace.LoopRunRecord, error) {
	repository, err := workspace.NewLoopRunRepository(ctx, env, assetsFS)
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
