package chcapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
)

var assetsFS fs.FS

func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, commandAssetsFS fs.FS) error {
	assetsFS = commandAssetsFS
	cmd := newRootCommand(stdin, stdout, stderr)

	osArgs := append([]string{"chc"}, args...)

	return cmd.Run(ctx, osArgs)
}

func newRootCommand(stdin io.Reader, stdout, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:                  "chc",
		Usage:                 "Coordinate agent development workflows from the current project",
		Reader:                stdin,
		Writer:                stdout,
		ErrWriter:             stderr,
		EnableShellCompletion: true,
		Action:                runDefaultTUIAction(stdin, stdout),
		Commands: []*cli.Command{
			newAuthorCommand(stdout),
			newDaemonCommand(stdout),
			newIndexCommand(stdout),
			newInitCommand(stdout),
			newMaintenanceCommand(stdout),
			newNotifyCommand(stdout),
			newProjectCommand(stdout),
			newProvingCommand(stdout),
			newReportCommand(stdout),
			newRuntimeCommand(stdout),
			newTicketCommand(stdout),
			newTUICommand(stdin, stdout),
		},
	}
}

func newTUICommand(stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "tui",
		Usage:  "Open the local ChrysoCoord operational console",
		Flags:  tuiFlags(),
		Action: runTUICommandAction(stdin, stdout),
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

func runDefaultTUIAction(stdin io.Reader, stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		env, err := newUserEnvironment("")
		if err != nil {
			return err
		}

		projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
			AllowCurrentDirectory: true,
		})
		if err != nil {
			return err
		}

		options, err := loadTUIOptions(ctx, env, projectContext, stdin, stdout)
		if err != nil {
			return err
		}

		return chctui.Run(ctx, options)
	}
}

func runTUICommandAction(stdin io.Reader, stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
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

		options, err := loadTUIOptions(ctx, env, projectContext, stdin, stdout)
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
) (chctui.Options, error) {
	tickets, err := loadTUITickets(ctx, env, projectContext)
	if err != nil {
		return chctui.Options{}, err
	}

	documents, err := chctui.LoadDocumentViews(ctx, projectContext.ProjectDir)
	if err != nil {
		return chctui.Options{}, err
	}

	projectMetadata, err := chc.NewProjectMetadataService(projectContext.ProjectDir).Load()
	if err != nil {
		return chctui.Options{}, err
	}

	cases, err := chc.NewCaseMetadataService(projectContext.ProjectDir).LoadAll()
	if err != nil {
		return chctui.Options{}, err
	}

	authoringSessions, err := loadTUIAuthoringSessions(ctx, projectContext)
	if err != nil {
		return chctui.Options{}, err
	}

	observability, err := loadTUIObservability(ctx, env, projectContext)
	if err != nil {
		return chctui.Options{}, err
	}

	workspaceGraph, err := chc.NewWorkspaceGraphBuilder(projectContext.ProjectDir).Build(ctx)
	if err != nil {
		return chctui.Options{}, err
	}

	operations := newService(ctx, env, projectContext)

	return newTUIOptions(tuiOptionsData{
		ctx:               ctx,
		env:               env,
		projectContext:    projectContext,
		stdin:             stdin,
		stdout:            stdout,
		tickets:           tickets,
		documents:         documents,
		projectMetadata:   projectMetadata,
		cases:             cases,
		authoringSessions: authoringSessions,
		observability:     observability,
		workspaceGraph:    workspaceGraph,
		operations:        operations,
	}), nil
}

type tuiOptionsData struct {
	ctx               context.Context
	env               chc.UserEnvironment
	projectContext    commandProjectContext
	stdin             io.Reader
	stdout            io.Writer
	tickets           []chc.IndexedTicket
	documents         []chctui.DocumentView
	projectMetadata   chc.ProjectMetadata
	cases             []chc.CaseMetadata
	authoringSessions []chc.AuthoringSession
	observability     chctui.ObservabilityView
	workspaceGraph    chc.WorkspaceGraph
	operations        service
}

func newTUIOptions(data tuiOptionsData) chctui.Options {
	return chctui.Options{
		ProjectDir:        data.projectContext.ProjectDir,
		ProjectID:         data.projectContext.ProjectID,
		ProjectMetadata:   data.projectMetadata,
		Cases:             data.cases,
		Stdin:             data.stdin,
		Stdout:            data.stdout,
		Tickets:           data.tickets,
		Documents:         data.documents,
		AuthoringSessions: data.authoringSessions,
		Observability:     data.observability,
		WorkspaceGraph:    data.workspaceGraph,
		ActionExecutor:    newTUIArtifactActionExecutor(data.ctx, data.env, data.projectContext),
		TicketExecutor:    newTUITicketActionExecutor(data.ctx, data.env, data.projectContext),
		WorkExecutor:      newTUIWorkActionExecutor(data.ctx, data.env, data.projectContext),
		WorkspaceExecutor: newTUIWorkspaceRefreshExecutor(data.operations),
		ProjectExecutor:   newTUIProjectMetadataExecutor(data.ctx, data.env, data.projectContext),
		CaseExecutor:      newTUICaseMetadataExecutor(data.ctx, data.env, data.projectContext),
	}
}

type tuiWorkspaceRefreshExecutor struct {
	operations service
}

func newTUIWorkspaceRefreshExecutor(
	operations service,
) tuiWorkspaceRefreshExecutor {
	return tuiWorkspaceRefreshExecutor{
		operations: operations,
	}
}

func (e tuiWorkspaceRefreshExecutor) RefreshWorkspace() (chctui.WorkspaceRefreshResult, error) {
	result, err := e.operations.RefreshWorkspace()
	if err != nil {
		return chctui.WorkspaceRefreshResult{}, err
	}

	return chctui.WorkspaceRefreshResult{
		Message: result.Message,
		Graph:   result.Graph,
	}, nil
}

type tuiProjectMetadataExecutor struct {
	projectContext commandProjectContext
}

func newTUIProjectMetadataExecutor(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) tuiProjectMetadataExecutor {
	return tuiProjectMetadataExecutor{
		projectContext: projectContext,
	}
}

func (e tuiProjectMetadataExecutor) SaveProjectMetadata(
	metadata chc.ProjectMetadata,
) (chctui.ProjectMetadataResult, error) {
	saved, err := chc.NewProjectMetadataService(e.projectContext.ProjectDir).Save(metadata)
	if err != nil {
		return chctui.ProjectMetadataResult{}, err
	}

	return chctui.ProjectMetadataResult{
		Message:  "Project metadata saved.",
		Metadata: saved,
	}, nil
}

type tuiCaseMetadataExecutor struct {
	projectContext commandProjectContext
}

func newTUICaseMetadataExecutor(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) tuiCaseMetadataExecutor {
	return tuiCaseMetadataExecutor{
		projectContext: projectContext,
	}
}

func (e tuiCaseMetadataExecutor) SaveCaseMetadata(
	metadata chc.CaseMetadata,
) (chctui.CaseMetadataResult, error) {
	saved, err := chc.NewCaseMetadataService(e.projectContext.ProjectDir).Save(metadata)
	if err != nil {
		return chctui.CaseMetadataResult{}, err
	}

	cases, err := chc.NewCaseMetadataService(e.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return chctui.CaseMetadataResult{}, err
	}

	return chctui.CaseMetadataResult{
		Message: "Case metadata saved.",
		Case:    saved,
		Cases:   cases,
	}, nil
}

type tuiArtifactActionExecutor struct {
	ctx            context.Context
	env            chc.UserEnvironment
	projectContext commandProjectContext
}

func newTUIArtifactActionExecutor(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) tuiArtifactActionExecutor {
	return tuiArtifactActionExecutor{
		ctx:            ctx,
		env:            env,
		projectContext: projectContext,
	}
}

func (e tuiArtifactActionExecutor) ExecuteArtifactAction(
	action chctui.ArtifactActionID,
) (chctui.ArtifactActionResult, error) {
	switch action {
	case chctui.ArtifactActionCreateTicket:
		return e.createTicket()
	case chctui.ArtifactActionCreateAuthoringSession:
		return e.createAuthoringSession()
	case chctui.ArtifactActionCreateContextQuiz:
		return chctui.ArtifactActionResult{}, errors.New(
			"context-aware quiz creation requires provider-backed question generation and is not available from this TUI action yet",
		)
	case chctui.ArtifactActionCreateDraft:
		return chctui.ArtifactActionResult{}, errors.New(
			"draft file creation requires a source quiz promotion; create an authoring session first",
		)
	default:
		return chctui.ArtifactActionResult{}, errors.New("unknown artifact action")
	}
}

func (e tuiArtifactActionExecutor) createTicket() (chctui.ArtifactActionResult, error) {
	config, err := chc.LoadWorkspaceConfig(e.projectContext.ProjectDir)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	repository := chc.NewTicketRepository(
		e.projectContext.ProjectDir,
		chc.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
	)

	result, err := repository.CreateTicket(e.ctx, chc.CreateTicketOptions{
		Title:       "TUI created ticket",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "task",
		ProjectArea: "tui",
	})
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	reloaded, err := e.reloadArtifactActionState()
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	reloaded.Message = "Created ticket " + result.Key + " at " + displayPath(e.projectContext.ProjectDir, result.Path)

	return reloaded, nil
}

func (e tuiArtifactActionExecutor) createAuthoringSession() (chctui.ArtifactActionResult, error) {
	now := time.Now().UTC()
	session := chc.AuthoringSession{
		ID:        authoringSessionID(now),
		ProjectID: e.projectContext.ProjectID,
		Status:    chc.AuthoringSessionStatusDraft,
		Mode:      chc.AuthoringSessionModeTUI,
		Target:    chc.AuthoringTargetDraft,
		Topic:     "TUI artifact creation",
		CreatedAt: now,
		UpdatedAt: now,
	}

	paths, err := chc.NewAuthoringSessionRepository(e.projectContext.ProjectDir).CreateSession(e.ctx, session)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	reloaded, err := e.reloadArtifactActionState()
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	reloaded.Message = "Created authoring session " + session.ID + " at " +
		displayPath(e.projectContext.ProjectDir, paths.SessionPath)

	return reloaded, nil
}

func (e tuiArtifactActionExecutor) reloadArtifactActionState() (chctui.ArtifactActionResult, error) {
	tickets, err := loadTUITickets(e.ctx, e.env, e.projectContext)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	documents, err := chctui.LoadDocumentViews(e.ctx, e.projectContext.ProjectDir)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	authoringSessions, err := loadTUIAuthoringSessions(e.ctx, e.projectContext)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	return chctui.ArtifactActionResult{
		Tickets:           tickets,
		Documents:         documents,
		AuthoringSessions: authoringSessions,
	}, nil
}

type tuiTicketActionExecutor struct {
	ctx            context.Context
	env            chc.UserEnvironment
	projectContext commandProjectContext
}

func newTUITicketActionExecutor(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) tuiTicketActionExecutor {
	return tuiTicketActionExecutor{
		ctx:            ctx,
		env:            env,
		projectContext: projectContext,
	}
}

func (e tuiTicketActionExecutor) ExecuteTicketAction(
	request chctui.TicketActionRequest,
) (chctui.TicketActionResult, error) {
	ticket, err := e.ticketByKey(request.Key)
	if err != nil {
		return chctui.TicketActionResult{}, err
	}

	switch request.Action {
	case chctui.TicketActionAddComment:
		return e.addComment(ticket)
	case chctui.TicketActionMoveStatus:
		return e.moveStatus(ticket, request.Status)
	case chctui.TicketActionToggleChecklist:
		return chctui.TicketActionResult{}, errors.New(
			"checklist toggle requires a safe checklist item writer and is not available from this TUI action yet",
		)
	case chctui.TicketActionAddChecklistItem:
		return chctui.TicketActionResult{}, errors.New(
			"checklist item creation requires a safe checklist item writer and is not available from this TUI action yet",
		)
	default:
		return chctui.TicketActionResult{}, errors.New("unknown ticket action")
	}
}

func (e tuiTicketActionExecutor) addComment(ticket chc.IndexedTicket) (chctui.TicketActionResult, error) {
	err := chc.NewTicketCommenter().AddComment(e.ctx, chc.AddTicketCommentOptions{
		Path:      ticket.Path,
		Author:    "tui",
		Body:      "TUI action recorded.",
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return chctui.TicketActionResult{}, err
	}

	tickets, err := loadTUITickets(e.ctx, e.env, e.projectContext)
	if err != nil {
		return chctui.TicketActionResult{}, err
	}

	return chctui.TicketActionResult{
		Message: "Added comment to ticket " + ticket.Key + ".",
		Tickets: tickets,
	}, nil
}

func (e tuiTicketActionExecutor) moveStatus(
	ticket chc.IndexedTicket,
	status string,
) (chctui.TicketActionResult, error) {
	err := chc.NewTicketStatusUpdater().UpdateStatus(e.ctx, chc.UpdateTicketStatusOptions{
		Path:      ticket.Path,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return chctui.TicketActionResult{}, err
	}

	tickets, err := loadTUITickets(e.ctx, e.env, e.projectContext)
	if err != nil {
		return chctui.TicketActionResult{}, err
	}

	return chctui.TicketActionResult{
		Message: "Moved ticket " + ticket.Key + " to " + status + ".",
		Tickets: tickets,
	}, nil
}

func (e tuiTicketActionExecutor) ticketByKey(key string) (chc.IndexedTicket, error) {
	index, err := chc.NewTicketIndex(e.ctx, e.env, assetsFS)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	defer func() {
		//nolint:errcheck // Lookup result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(e.ctx, index, e.projectContext)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	return index.GetProjectTicketByKey(e.ctx, e.projectContext.ProjectID, key)
}

type tuiWorkActionExecutor struct {
	ctx            context.Context
	env            chc.UserEnvironment
	projectContext commandProjectContext
}

func newTUIWorkActionExecutor(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) tuiWorkActionExecutor {
	return tuiWorkActionExecutor{
		ctx:            ctx,
		env:            env,
		projectContext: projectContext,
	}
}

func (e tuiWorkActionExecutor) ExecuteWorkAction(
	request chctui.WorkActionRequest,
) (chctui.WorkActionResult, error) {
	switch request.Action {
	case chctui.WorkActionStartTicket:
		return e.startTicket(request.TicketKey)
	case chctui.WorkActionStartReadyLoop:
		return e.startReadyLoop()
	case chctui.WorkActionRequestStop:
		return e.requestStop()
	default:
		return chctui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func (e tuiWorkActionExecutor) startTicket(ticketKey string) (chctui.WorkActionResult, error) {
	ticket, err := e.ticketByKey(ticketKey)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	preflight, err := buildTicketStartPreflightForTicket(e.ctx, e.env, e.projectContext, ticket)
	if err != nil {
		return chctui.WorkActionResult{}, chc.NewExecutionFailure(chc.ExecutionFailurePreflight, err)
	}

	if !chc.RuntimeValidationPassed(preflight.Runtime.Checks) {
		return chctui.WorkActionResult{}, chc.NewExecutionFailure(
			chc.ExecutionFailureRuntime,
			errors.New("runtime validation failed"),
		)
	}

	record, err := runTicketStartExecution(e.ctx, e.env, preflight.Record, preflight.Runtime)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	result, err := e.reloadWorkActionState()
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	result.Message = "Execution " + record.ID + " for ticket " + ticket.Key + " finished with status " +
		string(record.Status) + ". Result: " + displayPath(e.projectContext.ProjectDir, record.Artifacts.ResultPath)

	if record.FailureReason != "" {
		result.Message += " Failure: " + record.FailureReason
	}

	return result, nil
}

func (e tuiWorkActionExecutor) startReadyLoop() (chctui.WorkActionResult, error) {
	tickets, eligibleCount, err := e.readyLoopTickets()
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	result, err := runTicketLoopExecution(e.ctx, e.env, e.projectContext, tickets, eligibleCount)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	reloaded, err := e.reloadWorkActionState()
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	reloaded.Message = "Loop " + result.Record.ID + " finished with status " + string(result.Record.Status) +
		fmt.Sprintf(": %d processed, %d blocked, %d failed.", result.Record.ProcessedCount, result.Record.BlockedCount, result.Record.FailedCount)

	if result.Record.FailureReason != "" {
		reloaded.Message += " Failure: " + result.Record.FailureReason
	}

	return reloaded, nil
}

func (e tuiWorkActionExecutor) requestStop() (chctui.WorkActionResult, error) {
	repository, err := chc.NewDaemonStateRepository(e.ctx, e.env, assetsFS)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	defer func() {
		//nolint:errcheck // Stop result already comes from the main operation.
		_ = repository.Close()
	}()

	requestedAt := time.Now().UTC()

	err = repository.RequestDaemonStop(e.ctx, requestedAt)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		result, reloadErr := e.reloadWorkActionState()
		if reloadErr != nil {
			return chctui.WorkActionResult{}, reloadErr
		}

		result.Message = "Daemon is already stopped."

		return result, nil
	}

	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	result, err := e.reloadWorkActionState()
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	result.Message = "Daemon stop requested at " + requestedAt.Format(time.RFC3339) + "."

	return result, nil
}

func (e tuiWorkActionExecutor) readyLoopTickets() ([]chc.IndexedTicket, int, error) {
	index, err := chc.NewTicketIndex(e.ctx, e.env, assetsFS)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		//nolint:errcheck // Query result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(e.ctx, index, e.projectContext)
	if err != nil {
		return nil, 0, err
	}

	tickets, err := chc.NewTicketLoopQuery(index).EligibleTickets(e.ctx, chc.TicketLoopQueryOptions{
		ProjectID: e.projectContext.ProjectID,
		Statuses:  []string{"ready"},
		Order:     chc.TicketLoopOrderKey,
	})
	if err != nil {
		return nil, 0, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: 1})

	return tickets, len(eligibleTickets), nil
}

func (e tuiWorkActionExecutor) ticketByKey(key string) (chc.IndexedTicket, error) {
	index, err := chc.NewTicketIndex(e.ctx, e.env, assetsFS)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	defer func() {
		//nolint:errcheck // Lookup result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(e.ctx, index, e.projectContext)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	return index.GetProjectTicketByKey(e.ctx, e.projectContext.ProjectID, key)
}

func (e tuiWorkActionExecutor) reloadWorkActionState() (chctui.WorkActionResult, error) {
	tickets, err := loadTUITickets(e.ctx, e.env, e.projectContext)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	observability, err := loadTUIObservability(e.ctx, e.env, e.projectContext)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	return chctui.WorkActionResult{
		Tickets:       tickets,
		Observability: observability,
	}, nil
}

func loadTUITickets(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) ([]chc.IndexedTicket, error) {
	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // Load result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return nil, err
	}

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	return tickets, nil
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
) (chctui.ObservabilityView, error) {
	daemonState, err := loadTUIDaemonState(ctx, env)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	executions, err := loadTUIExecutions(ctx, env, projectContext.ProjectID)
	if err != nil {
		return chctui.ObservabilityView{}, err
	}

	loops, err := loadTUILoopRuns(ctx, env, projectContext.ProjectID)
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

func loadTUIDaemonState(ctx context.Context, env chc.UserEnvironment) (chc.DaemonState, error) {
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
