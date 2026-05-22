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

		tickets, err := loadTUITickets(ctx, env, projectContext)
		if err != nil {
			return err
		}

		documents, err := chctui.LoadDocumentViews(ctx, projectContext.ProjectDir)
		if err != nil {
			return err
		}

		authoringSessions, err := loadTUIAuthoringSessions(ctx, projectContext)
		if err != nil {
			return err
		}

		observability, err := loadTUIObservability(ctx, env, projectContext)
		if err != nil {
			return err
		}

		return chctui.Run(ctx, chctui.Options{
			ProjectDir:        projectContext.ProjectDir,
			ProjectID:         projectContext.ProjectID,
			Stdin:             stdin,
			Stdout:            stdout,
			Tickets:           tickets,
			Documents:         documents,
			AuthoringSessions: authoringSessions,
			Observability:     observability,
			ActionExecutor:    newTUIArtifactActionExecutor(ctx, env, projectContext),
			TicketExecutor:    newTUITicketActionExecutor(ctx, env, projectContext),
		})
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

		tickets, err := loadTUITickets(ctx, env, projectContext)
		if err != nil {
			return err
		}

		documents, err := chctui.LoadDocumentViews(ctx, projectContext.ProjectDir)
		if err != nil {
			return err
		}

		authoringSessions, err := loadTUIAuthoringSessions(ctx, projectContext)
		if err != nil {
			return err
		}

		observability, err := loadTUIObservability(ctx, env, projectContext)
		if err != nil {
			return err
		}

		return chctui.Run(ctx, chctui.Options{
			ProjectDir:        projectContext.ProjectDir,
			ProjectID:         projectContext.ProjectID,
			Stdin:             stdin,
			Stdout:            stdout,
			Tickets:           tickets,
			Documents:         documents,
			AuthoringSessions: authoringSessions,
			Observability:     observability,
			ActionExecutor:    newTUIArtifactActionExecutor(ctx, env, projectContext),
			TicketExecutor:    newTUITicketActionExecutor(ctx, env, projectContext),
		})
	}
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
