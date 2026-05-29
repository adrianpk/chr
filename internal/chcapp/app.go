package chcapp

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
)

type appDependencies struct {
	assetsFS fs.FS
}

const appDependenciesMetadataKey = "chcapp.dependencies"

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
			newAuthorCommand(stdout, deps),
			newDaemonCommand(stdout, deps),
			newDumpCommand(stdout, deps),
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

		return chctui.Run(ctx, options)
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

		return chctui.Run(ctx, options)
	}
}

func loadTUIOptions(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	stdin io.Reader,
	stdout io.Writer,
	deps appDependencies,
) (chctui.Options, error) {
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
		return chctui.Options{}, err
	}

	if err := loadTUIMetadataData(&data); err != nil {
		return chctui.Options{}, err
	}

	if err := loadTUIIndexData(&data); err != nil {
		return chctui.Options{}, err
	}

	return newTUIOptions(data), nil
}

func loadTUIContentData(data *tuiOptionsData) error {
	tickets, err := loadTUITickets(data.ctx, data.env, data.projectContext, data.assetsFS)
	if err != nil {
		return err
	}

	documents, err := chctui.LoadDocumentViews(data.ctx, data.projectContext.ProjectDir)
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
	projectMetadata, err := chc.NewProjectMetadataService(data.projectContext.ProjectDir).Load()
	if err != nil {
		return err
	}

	cases, err := chc.NewCaseMetadataService(data.projectContext.ProjectDir).LoadAll()
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
) (chc.WorkspaceGraph, chc.WorkspaceIndex, []chc.WorkspaceSemanticDigestState, error) {
	result, err := chc.NewWorkspaceGraphBuilder(projectContext.ProjectDir).BuildCachedIndex(ctx)
	if err != nil {
		return chc.WorkspaceGraph{}, chc.WorkspaceIndex{}, nil, err
	}

	cache, err := chc.ReadWorkspaceSemanticDigestCache(
		chc.WorkspaceSemanticDigestCachePath(projectContext.ProjectDir),
	)
	if err != nil {
		return chc.WorkspaceGraph{}, chc.WorkspaceIndex{}, nil, err
	}

	states := chc.WorkspaceSemanticDigestStates(result.Index, cache, "", "")

	return result.Graph, result.Index, states, nil
}

type tuiOptionsData struct {
	ctx                  context.Context
	env                  chc.UserEnvironment
	projectContext       commandProjectContext
	stdin                io.Reader
	stdout               io.Writer
	assetsFS             fs.FS
	tickets              []chc.IndexedTicket
	documents            []chctui.DocumentView
	dumps                []chctui.DumpView
	quizzes              []chctui.QuizView
	specs                []chctui.SpecView
	plans                []chctui.PlanView
	projectMetadata      chc.ProjectMetadata
	cases                []chc.CaseMetadata
	authoringSessions    []chc.AuthoringSession
	observability        chctui.ObservabilityView
	workspaceGraph       chc.WorkspaceGraph
	workspaceIndex       chc.WorkspaceIndex
	semanticDigestStates []chc.WorkspaceSemanticDigestState
	operations           service
}

func newTUIOptions(data tuiOptionsData) chctui.Options {
	return chctui.Options{
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
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) ([]chc.IndexedTicket, error) {
	board, err := newService(ctx, env, projectContext, assetsFS).Board()
	if err != nil {
		return nil, err
	}

	return board.Tickets, nil
}

func loadTUIAuthoringSessions(
	ctx context.Context,
	projectContext commandProjectContext,
) ([]chc.AuthoringSession, error) {
	sessions, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func loadTUIObservability(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) (chctui.ObservabilityView, error) {
	daemonState, err := loadTUIDaemonState(ctx, env, assetsFS)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	executions, err := loadTUIExecutions(ctx, env, projectContext.ProjectID, assetsFS)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	loops, err := loadTUILoopRuns(ctx, env, projectContext.ProjectID, assetsFS)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	runtime, err := chc.ValidateRuntime(chc.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	return chctui.NewObservabilityView(projectContext.ProjectDir, daemonState, executions, loops, runtime), nil
}

func loadTUIDaemonState(ctx context.Context, env chc.UserEnvironment, assetsFS fs.FS) (chc.DaemonState, error) {
	repository, err := chc.NewDaemonStateRepository(ctx, env, assetsFS)
	if err != nil {
		return chc.DaemonState{}, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = repository.Close()
	}()

	state, err := repository.GetDaemonState(ctx)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		return chc.NewStoppedDaemonState(time.Now().UTC()), nil
	}

	if err != nil {
		return chc.DaemonState{}, err
	}

	return state, nil
}

func loadTUIExecutions(
	ctx context.Context,
	env chc.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]chc.ExecutionRecord, error) {
	repository, err := chc.NewExecutionRepository(ctx, env, assetsFS)
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
	env chc.UserEnvironment,
	projectID string,
	assetsFS fs.FS,
) ([]chc.LoopRunRecord, error) {
	repository, err := chc.NewLoopRunRepository(ctx, env, assetsFS)
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
