package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

func newAuthorCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "author",
		Usage: "Manage Markdown-first authoring sessions",
		Commands: []*cli.Command{
			newAuthorAssistedCommand(stdout),
			newAuthorPromoteCommand(stdout),
			newAuthorQuizCommand(stdout),
			newAuthorSessionCommand(stdout),
		},
	}
}

var newAuthoringQuizQuestionProvider = func() core.AuthoringQuizQuestionProvider {
	return core.CodexAuthoringQuizQuestionProvider{}
}

func newAuthorAssistedCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "assisted",
		Usage: "Run assisted authoring sessions",
		Commands: []*cli.Command{
			newAuthorAssistedAnswerCommand(stdout),
			newAuthorAssistedFinalizeCommand(stdout),
			newAuthorAssistedResumeCommand(stdout),
			newAuthorAssistedStartCommand(stdout),
			newAuthorAssistedStatusCommand(stdout),
		},
	}
}

func newAuthorPromoteCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "promote",
		Usage: "Promote authoring artifacts",
		Commands: []*cli.Command{
			newAuthorPromoteQuizToDraftCommand(stdout),
		},
	}
}

func newAuthorSessionCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "session",
		Usage: "Manage authoring sessions",
		Commands: []*cli.Command{
			newAuthorSessionListCommand(stdout),
			newAuthorSessionNewCommand(stdout),
			newAuthorSessionShowCommand(stdout),
		},
	}
}

func newAuthorQuizCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "quiz",
		Usage: "Generate authoring quizzes",
		Commands: []*cli.Command{
			newAuthorQuizNewCommand(stdout),
		},
	}
}

func newAuthorSessionNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create an authoring session",
		Flags:  authorSessionFlags(),
		Action: authorSessionNewAction(stdout),
	}
}

func newAuthorSessionListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List authoring sessions",
		Flags:  authorProjectFlags(),
		Action: authorSessionListAction(stdout),
	}
}

func newAuthorSessionShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one authoring session",
		Flags:  authorProjectFlags(),
		Action: authorSessionShowAction(stdout),
	}
}

func newAuthorQuizNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create an authoring quiz",
		Flags:  authorQuizNewFlags(),
		Action: authorQuizNewAction(stdout),
	}
}

func newAuthorPromoteQuizToDraftCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "quiz-to-draft",
		Usage:  "Promote a quiz artifact into an editable draft",
		Flags:  authorPromoteQuizToDraftFlags(),
		Action: authorPromoteQuizToDraftAction(stdout),
	}
}

func newAuthorAssistedStartCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start an assisted authoring session",
		Flags:  authorAssistedStartFlags(),
		Action: authorAssistedStartAction(stdout),
	}
}

func newAuthorAssistedResumeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "resume",
		Usage:  "Ask the next assisted authoring question",
		Flags:  authorAssistedResumeFlags(),
		Action: authorAssistedResumeAction(stdout),
	}
}

func newAuthorAssistedAnswerCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "answer",
		Usage:  "Submit an assisted authoring answer",
		Flags:  authorAssistedAnswerFlags(),
		Action: authorAssistedAnswerAction(stdout),
	}
}

func newAuthorAssistedStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Show assisted authoring status",
		Flags:  authorAssistedSessionFlags(),
		Action: authorAssistedStatusAction(stdout),
	}
}

func newAuthorAssistedFinalizeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "finalize",
		Usage:  "Finalize assisted authoring into an artifact",
		Flags:  authorAssistedFinalizeFlags(),
		Action: authorAssistedFinalizeAction(stdout),
	}
}

func authorProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func authorAssistedStartFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "target",
			Usage: "authoring target",
			Value: string(core.AuthoringTargetSpec),
		},
		&cli.StringFlag{
			Name:  "topic",
			Usage: "authoring topic",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "session description",
		},
	)
}

func authorAssistedSessionFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "session",
			Usage: "authoring session handle",
		},
	)
}

func authorAssistedResumeFlags() []cli.Flag {
	return append(authorAssistedSessionFlags(),
		&cli.StringFlag{
			Name:  "fake-question",
			Usage: "fake adapter question prompt",
		},
		&cli.IntFlag{
			Name:  "max-turns",
			Usage: "maximum assisted turns",
			Value: 8,
		},
		&cli.IntFlag{
			Name:  "max-follow-ups",
			Usage: "maximum assisted follow-ups",
			Value: 4,
		},
	)
}

func authorAssistedAnswerFlags() []cli.Flag {
	return append(authorAssistedSessionFlags(),
		&cli.StringFlag{
			Name:  "turn",
			Usage: "assisted turn id",
		},
		&cli.IntFlag{
			Name:  "turn-index",
			Usage: "assisted turn index",
		},
		&cli.StringFlag{
			Name:  "question-id",
			Usage: "question id",
		},
		&cli.StringFlag{
			Name:  "question",
			Usage: "question prompt",
		},
		&cli.StringFlag{
			Name:  "answer",
			Usage: "answer text",
		},
		&cli.BoolFlag{
			Name:  "stop",
			Usage: "stop after this answer",
		},
	)
}

func authorAssistedFinalizeFlags() []cli.Flag {
	return append(authorAssistedSessionFlags(),
		&cli.StringFlag{
			Name:  "fake-summary",
			Usage: "fake adapter curated summary",
		},
		&cli.StringFlag{
			Name:  "target-kind",
			Usage: "target artifact kind",
			Value: string(core.AuthoringArtifactKindSpec),
		},
		&cli.StringFlag{
			Name:  "target-name",
			Usage: "target artifact name",
		},
		&cli.BoolFlag{
			Name:  "allow-replace",
			Usage: "allow replacing an existing target artifact",
		},
	)
}

func authorSessionFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "mode",
			Usage: "authoring mode",
			Value: string(core.AuthoringSessionModeMarkdownDirect),
		},
		&cli.StringFlag{
			Name:  "target",
			Usage: "authoring target",
			Value: string(core.AuthoringTargetDraft),
		},
		&cli.StringFlag{
			Name:  "topic",
			Usage: "authoring topic",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "session description",
		},
	)
}

func authorQuizNewFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "target",
			Usage: "authoring target",
			Value: string(core.AuthoringTargetSpec),
		},
		&cli.StringFlag{
			Name:  "topic",
			Usage: "quiz topic",
		},
		&cli.IntFlag{
			Name:  "question-count",
			Usage: "maximum number of questions to render",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned quiz without writing files",
		},
		&cli.BoolFlag{
			Name:  "allow-replace",
			Usage: "allow replacing an existing quiz artifact",
		},
	)
}

func authorPromoteQuizToDraftFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "session",
			Usage: "authoring session id",
		},
		&cli.StringFlag{
			Name:  "quiz",
			Usage: "quiz artifact path",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned draft without writing files",
		},
		&cli.BoolFlag{
			Name:  "allow-replace",
			Usage: "allow replacing an existing draft artifact",
		},
	)
}

func authorSessionNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorSessionNew(ctx, cmd, stdout)
	}
}

func authorSessionListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorSessionList(ctx, cmd, stdout)
	}
}

func authorSessionShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorSessionShow(ctx, cmd, stdout)
	}
}

func authorQuizNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorQuizNew(ctx, cmd, stdout)
	}
}

func authorPromoteQuizToDraftAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorPromoteQuizToDraft(ctx, cmd, stdout)
	}
}

func authorAssistedStartAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorAssistedStart(ctx, cmd, stdout)
	}
}

func authorAssistedResumeAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorAssistedResume(ctx, cmd, stdout)
	}
}

func authorAssistedAnswerAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorAssistedAnswer(ctx, cmd, stdout)
	}
}

func authorAssistedStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorAssistedStatus(ctx, cmd, stdout)
	}
}

func authorAssistedFinalizeAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAuthorAssistedFinalize(ctx, cmd, stdout)
	}
}

func runAuthorSessionNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author session new does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
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

	session, err := authoringSessionFromCommand(cmd, projectContext, time.Now().UTC())
	if err != nil {
		return err
	}

	result, err := newService(ctx, env, projectContext, deps.assetsFS).CreateAuthoringSession(session)
	if err != nil {
		return err
	}

	return printAuthorSessionCreated(stdout, result.Session, result.Paths, projectContext.ProjectDir)
}

func runAuthorSessionList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author session list does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	sessions, err := core.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "id\tstatus\ttarget\tupdated_at\ttopic")
	if err != nil {
		return fmt.Errorf("write author session list header: %w", err)
	}

	for _, session := range sessions {
		_, err = fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\n",
			session.ID,
			session.Status,
			session.Target,
			session.UpdatedAt.UTC().Format(time.RFC3339),
			session.Topic,
		)
		if err != nil {
			return fmt.Errorf("write author session list: %w", err)
		}
	}

	return nil
}

func runAuthorSessionShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("author session show requires exactly one session id argument")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	repository := core.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := resolveAuthoringSession(ctx, repository, cmd.Args().First())
	if err != nil {
		return err
	}

	return printAuthorSession(stdout, session)
}

func runAuthorQuizNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author quiz new does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	session, err := quizSessionFromCommand(cmd, projectContext, now)
	if err != nil {
		return err
	}

	return runAuthorContextualQuizNew(ctx, cmd, stdout, projectContext, session, now)
}

func runAuthorContextualQuizNew(
	ctx context.Context,
	cmd *cli.Command,
	stdout io.Writer,
	projectContext commandProjectContext,
	session core.AuthoringSession,
	now time.Time,
) error {
	plan, err := core.PlanContextualQuizGeneration(ctx, core.ContextualQuizGenerationOptions{
		ProjectDir:    projectContext.ProjectDir,
		Session:       session,
		Now:           now,
		AllowReplace:  cmd.Bool("allow-replace"),
		DryRun:        cmd.Bool("dry-run"),
		QuestionCount: cmd.Int("question-count"),
		Provider:      contextualQuizProviderFromCommand(),
	})
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printAuthorContextualQuizPlan(stdout, plan, true, projectContext.ProjectDir)
	}

	err = core.WriteContextualQuiz(plan)
	if err != nil {
		return err
	}

	_, err = core.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, plan.Session)
	if err != nil {
		return err
	}

	return printAuthorContextualQuizPlan(stdout, plan, false, projectContext.ProjectDir)
}

func runAuthorPromoteQuizToDraft(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author promote quiz-to-draft does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	sessionHandle := strings.TrimSpace(cmd.String("session"))

	quizPath := strings.TrimSpace(cmd.String("quiz"))
	if quizPath == "" {
		return fmt.Errorf("quiz path is required")
	}

	repository := core.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := resolveAuthoringSessionForPromotion(
		ctx,
		repository,
		projectContext.ProjectDir,
		sessionHandle,
		quizPath,
	)
	if err != nil {
		return err
	}

	plan, err := core.PlanQuizToDraftPromotion(core.QuizToDraftPromotionOptions{
		ProjectDir:   projectContext.ProjectDir,
		Session:      session,
		QuizPath:     quizPath,
		Now:          time.Now().UTC(),
		AllowReplace: cmd.Bool("allow-replace"),
		DryRun:       cmd.Bool("dry-run"),
	})
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printAuthorPromotionPlan(stdout, plan, true)
	}

	err = core.WriteQuizToDraftPromotion(plan)
	if err != nil {
		return err
	}

	_, err = repository.UpdateSession(ctx, plan.Session)
	if err != nil {
		return err
	}

	return printAuthorPromotionPlan(stdout, plan, false)
}

func runAuthorAssistedStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted start does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	session, err := assistedSessionFromCommand(cmd, projectContext, now)
	if err != nil {
		return err
	}

	paths, err := core.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, session)
	if err != nil {
		return err
	}

	err = printAuthorSessionCreated(stdout, session, paths, projectContext.ProjectDir)
	if err != nil {
		return err
	}

	return writeAuthorLines(stdout, []string{"assisted: started"})
}

func runAuthorAssistedResume(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted resume does not accept positional arguments")
	}

	state, err := loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	question := strings.TrimSpace(cmd.String("fake-question"))
	if question == "" {
		return fmt.Errorf("fake question is required")
	}

	now := time.Now().UTC()
	response := core.AssistedModelResponse{
		ID:        "response-" + authoringSessionID(now),
		Provider:  "fake",
		Status:    core.AssistedModelResponseReady,
		Question:  core.AssistedAuthoringQuestion{ID: "question-" + authoringSessionID(now), Prompt: question, Required: true},
		RawText:   question,
		CreatedAt: now,
	}

	result, err := core.AskNextAssistedQuizQuestion(ctx, core.AssistedQuizLoopOptions{
		Config:           assistedConfigFromCommand(cmd, state.Session),
		Session:          state.Session,
		SessionPaths:     state.Paths,
		ContextPackage:   state.ContextPackage,
		TranscriptEvents: state.Events,
		Repository:       state.Repository,
		Adapter:          &core.FakeAssistedModelAdapter{Responses: []core.AssistedModelResponse{response}},
		Now:              func() time.Time { return now },
	})
	if err != nil {
		return err
	}

	return printAuthorAssistedQuestion(stdout, result)
}

func runAuthorAssistedAnswer(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted answer does not accept positional arguments")
	}

	state, err := loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := core.AnswerAssistedQuizQuestion(ctx, core.AssistedQuizAnswerOptions{
		SessionID:  state.Session.ID,
		TurnID:     strings.TrimSpace(cmd.String("turn")),
		TurnIndex:  cmd.Int("turn-index"),
		Intent:     core.AssistedAuthoringIntentDiscover,
		Question:   core.AssistedAuthoringQuestion{ID: strings.TrimSpace(cmd.String("question-id")), Prompt: strings.TrimSpace(cmd.String("question")), Required: true},
		AnswerText: strings.TrimSpace(cmd.String("answer")),
		UserStop:   cmd.Bool("stop"),
		Repository: state.Repository,
		Now:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return err
	}

	return printAuthorAssistedAnswer(stdout, result)
}

func runAuthorAssistedStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted status does not accept positional arguments")
	}

	state, err := loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	return printAuthorAssistedStatus(stdout, state.Session, state.Events)
}

func runAuthorAssistedFinalize(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted finalize does not accept positional arguments")
	}

	state, err := loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	summaryText := strings.TrimSpace(cmd.String("fake-summary"))
	if summaryText == "" {
		return fmt.Errorf("fake summary is required")
	}

	now := time.Now().UTC()

	curation, err := core.CurateAssistedAnswers(ctx, core.AssistedAnswerCurationOptions{
		Session:          state.Session,
		SessionPaths:     state.Paths,
		TranscriptEvents: state.Events,
		Repository:       state.Repository,
		Adapter: &core.FakeAssistedModelAdapter{Responses: []core.AssistedModelResponse{
			authorAssistedCurationResponse(state.Session, summaryText, now),
		}},
		Now: func() time.Time { return now },
	})
	if err != nil {
		return err
	}

	synthesis, err := core.SynthesizeAssistedArtifact(ctx, core.AssistedArtifactSynthesisOptions{
		ProjectDir:     state.ProjectDir,
		Session:        curation.Session,
		CuratedSummary: curation.Summary,
		SourceRef:      curation.ArtifactRef,
		TargetKind:     core.AuthoringArtifactKind(cmd.String("target-kind")),
		TargetName:     strings.TrimSpace(cmd.String("target-name")),
		AllowReplace:   cmd.Bool("allow-replace"),
		Repository:     state.Repository,
		Now:            func() time.Time { return now },
	})
	if err != nil {
		return err
	}

	return printAuthorAssistedFinalized(stdout, curation, synthesis)
}

type authorAssistedState struct {
	ProjectDir     string
	Session        core.AuthoringSession
	Paths          core.AuthoringSessionPaths
	Events         []core.AssistedTranscriptEvent
	ContextPackage core.AssistedAuthoringContextPackage
	Repository     core.AuthoringSessionRepository
}

func loadAuthorAssistedState(ctx context.Context, cmd *cli.Command) (authorAssistedState, error) {
	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return authorAssistedState{}, err
	}

	sessionHandle := strings.TrimSpace(cmd.String("session"))
	if sessionHandle == "" {
		return authorAssistedState{}, fmt.Errorf("authoring session is required")
	}

	repository := core.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := resolveAuthoringSession(ctx, repository, sessionHandle)
	if err != nil {
		return authorAssistedState{}, err
	}

	paths := core.AuthoringSessionLayout(projectContext.ProjectDir, session.ID)

	events, err := repository.ListTranscriptEvents(ctx, session.ID)
	if err != nil {
		return authorAssistedState{}, err
	}

	contextPackage, err := core.LoadAssistedAuthoringContextPackage(core.AssistedAuthoringContextPackageOptions{
		ProjectDir:       projectContext.ProjectDir,
		Session:          session,
		SessionPaths:     paths,
		TranscriptEvents: events,
	})
	if err != nil {
		return authorAssistedState{}, err
	}

	return authorAssistedState{
		ProjectDir:     projectContext.ProjectDir,
		Session:        session,
		Paths:          paths,
		Events:         events,
		ContextPackage: contextPackage,
		Repository:     repository,
	}, nil
}

func authorProjectContext(ctx context.Context, cmd *cli.Command) (commandProjectContext, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return commandProjectContext{}, err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return commandProjectContext{}, err
	}

	return newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
}

func assistedSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (core.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return core.AuthoringSession{}, fmt.Errorf("authoring topic is required")
	}

	session := core.AuthoringSession{
		ID:          authoringSessionID(now),
		ProjectID:   projectContext.ProjectID,
		Status:      core.AuthoringSessionStatusActive,
		Mode:        core.AuthoringSessionModeAssisted,
		Target:      core.AuthoringTarget(cmd.String("target")),
		Topic:       topic,
		CreatedAt:   now,
		StartedAt:   &now,
		UpdatedAt:   now,
		Description: strings.TrimSpace(cmd.String("description")),
	}

	err := core.ValidateAuthoringSession(session)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	return session, nil
}

func authoringSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (core.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return core.AuthoringSession{}, fmt.Errorf("authoring topic is required")
	}

	session := core.AuthoringSession{
		ID:          authoringSessionID(now),
		ProjectID:   projectContext.ProjectID,
		Status:      core.AuthoringSessionStatusDraft,
		Mode:        core.AuthoringSessionMode(cmd.String("mode")),
		Target:      core.AuthoringTarget(cmd.String("target")),
		Topic:       topic,
		CreatedAt:   now,
		UpdatedAt:   now,
		Description: strings.TrimSpace(cmd.String("description")),
	}

	err := core.ValidateAuthoringSession(session)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	return session, nil
}

func quizSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (core.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return core.AuthoringSession{}, fmt.Errorf("quiz topic is required")
	}

	session := core.AuthoringSession{
		ID:        authoringSessionID(now),
		ProjectID: projectContext.ProjectID,
		Status:    core.AuthoringSessionStatusDraft,
		Mode:      core.AuthoringSessionModeContextual,
		Target:    core.AuthoringTarget(cmd.String("target")),
		Topic:     topic,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := core.ValidateAuthoringSession(session)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	return session, nil
}

func contextualQuizProviderFromCommand() core.AuthoringQuizQuestionProvider {
	return newAuthoringQuizQuestionProvider()
}

func assistedConfigFromCommand(cmd *cli.Command, session core.AuthoringSession) core.AssistedAuthoringConfig {
	return core.AssistedAuthoringConfig{
		SessionID:    session.ID,
		ProjectID:    session.ProjectID,
		Target:       session.Target,
		Intent:       core.AssistedAuthoringIntentDiscover,
		MaxTurns:     cmd.Int("max-turns"),
		MaxFollowUps: cmd.Int("max-follow-ups"),
	}
}

func authorAssistedCurationResponse(
	session core.AuthoringSession,
	summary string,
	now time.Time,
) core.AssistedModelResponse {
	return core.AssistedModelResponse{
		ID:        "curation-response-" + session.ID,
		RequestID: "curation-" + session.ID,
		SessionID: session.ID,
		TurnID:    "curation",
		Provider:  "fake",
		Status:    core.AssistedModelResponseReady,
		Question: core.AssistedAuthoringQuestion{
			ID:       "curation-ready-" + session.ID,
			Prompt:   "Curation is ready.",
			Required: false,
		},
		RawText:   summary,
		CreatedAt: now,
	}
}

func authoringSessionID(now time.Time) string {
	return "author-" + now.UTC().Format("20060102-150405")
}

func resolveAuthoringSession(
	ctx context.Context,
	repository core.AuthoringSessionRepository,
	handle string,
) (core.AuthoringSession, error) {
	handle = strings.TrimSpace(handle)
	if handle == "" {
		return core.AuthoringSession{}, fmt.Errorf("authoring session is required")
	}

	session, err := repository.GetSession(ctx, handle)
	if err == nil {
		return session, nil
	}

	if !errors.Is(err, core.ErrAuthoringSessionNotFound) {
		return core.AuthoringSession{}, err
	}

	sessions, listErr := repository.ListSessions(ctx)
	if listErr != nil {
		return core.AuthoringSession{}, listErr
	}

	matches := matchingAuthoringSessions(sessions, handle)
	if len(matches) == 1 {
		return matches[0], nil
	}

	if len(matches) > 1 {
		return core.AuthoringSession{}, fmt.Errorf("authoring session handle is ambiguous: %s", handle)
	}

	return core.AuthoringSession{}, fmt.Errorf("%w: %s", core.ErrAuthoringSessionNotFound, handle)
}

func resolveAuthoringSessionForPromotion(
	ctx context.Context,
	repository core.AuthoringSessionRepository,
	projectDir string,
	handle string,
	artifactPath string,
) (core.AuthoringSession, error) {
	if strings.TrimSpace(handle) != "" {
		return resolveAuthoringSession(ctx, repository, handle)
	}

	return resolveAuthoringSessionFromArtifact(ctx, repository, projectDir, artifactPath)
}

func resolveAuthoringSessionFromArtifact(
	ctx context.Context,
	repository core.AuthoringSessionRepository,
	projectDir string,
	artifactPath string,
) (core.AuthoringSession, error) {
	sessions, err := repository.ListSessions(ctx)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	matches := []core.AuthoringSession{}
	normalizedPath := normalizeAuthoringArtifactPath(projectDir, artifactPath)

	for _, session := range sessions {
		if authoringSessionHasArtifact(session, projectDir, normalizedPath) {
			matches = append(matches, session)
		}
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	if len(matches) > 1 {
		return core.AuthoringSession{}, fmt.Errorf("authoring artifact path is ambiguous: %s", artifactPath)
	}

	return core.AuthoringSession{}, fmt.Errorf("authoring session is required when artifact metadata cannot identify one session")
}

func authoringSessionHasArtifact(
	session core.AuthoringSession,
	projectDir string,
	artifactPath string,
) bool {
	for _, artifact := range session.Outputs {
		if normalizeAuthoringArtifactPath(projectDir, artifact.Path) == artifactPath {
			return true
		}
	}

	for _, artifact := range session.Inputs {
		if normalizeAuthoringArtifactPath(projectDir, artifact.Path) == artifactPath {
			return true
		}
	}

	return false
}

func normalizeAuthoringArtifactPath(projectDir, artifactPath string) string {
	path := filepath.Clean(strings.TrimSpace(artifactPath))
	if filepath.IsAbs(path) {
		relativePath, err := filepath.Rel(projectDir, path)
		if err == nil {
			path = relativePath
		}
	}

	return filepath.ToSlash(path)
}

func matchingAuthoringSessions(sessions []core.AuthoringSession, handle string) []core.AuthoringSession {
	handleSlug := authoringHandleSlug(handle)
	matches := []core.AuthoringSession{}

	for _, session := range sessions {
		if authoringSessionMatchesHandle(session, handle, handleSlug) {
			matches = append(matches, session)
		}
	}

	return matches
}

func authoringSessionMatchesHandle(session core.AuthoringSession, handle, handleSlug string) bool {
	if strings.EqualFold(session.ID, handle) {
		return true
	}

	if strings.EqualFold(session.Topic, handle) {
		return true
	}

	return handleSlug != "" && authoringHandleSlug(session.Topic) == handleSlug
}

func authoringHandleSlug(value string) string {
	var builder strings.Builder

	var previousDash bool

	normalized := strings.ToLower(strings.TrimSpace(value))

	for _, char := range normalized {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)

			previousDash = false

			continue
		}

		if builder.Len() > 0 && !previousDash {
			builder.WriteByte('-')

			previousDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func printAuthorAssistedQuestion(stdout io.Writer, result core.AssistedQuizAskResult) error {
	lines := []string{
		"assisted: question",
		"turn: " + result.Turn.ID,
		fmt.Sprintf("turn_index: %d", result.Turn.Index),
		"question_id: " + result.Turn.Question.ID,
		"question: " + result.Turn.Question.Prompt,
		"request: " + result.InvocationPaths.RequestPath,
		"response: " + result.InvocationPaths.ResponsePath,
	}

	if result.StopReason != core.AssistedQuizLoopStopNone {
		lines = append(lines, "stop_reason: "+string(result.StopReason))
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorAssistedAnswer(stdout io.Writer, result core.AssistedQuizAnswerResult) error {
	lines := []string{
		"assisted: answer",
		"turn: " + result.Turn.ID,
		fmt.Sprintf("turn_index: %d", result.Turn.Index),
		"question_id: " + result.Turn.Question.ID,
	}

	if result.StopReason != core.AssistedQuizLoopStopNone {
		lines = append(lines, "stop_reason: "+string(result.StopReason))
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorAssistedStatus(
	stdout io.Writer,
	session core.AuthoringSession,
	events []core.AssistedTranscriptEvent,
) error {
	questions, answers, results := assistedEventCounts(events)

	return writeAuthorLines(stdout, []string{
		"assisted: status",
		"session: " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		fmt.Sprintf("questions: %d", questions),
		fmt.Sprintf("answers: %d", answers),
		fmt.Sprintf("results: %d", results),
	})
}

func printAuthorAssistedFinalized(
	stdout io.Writer,
	curation core.AssistedAnswerCurationResult,
	synthesis core.AssistedArtifactSynthesisResult,
) error {
	return writeAuthorLines(stdout, []string{
		"assisted: finalized",
		"session: " + synthesis.Session.ID,
		"curated_summary: " + curation.ArtifactPath,
		"artifact: " + synthesis.Layout.FullPath,
	})
}

func printAuthorSessionCreated(
	stdout io.Writer,
	session core.AuthoringSession,
	paths core.AuthoringSessionPaths,
	projectDir string,
) error {
	lines := []string{
		"session: created " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		"files:",
		"  session: " + displayPath(projectDir, paths.SessionPath),
		"  transcript: " + displayPath(projectDir, paths.TranscriptPath),
		"  artifacts: " + displayPath(projectDir, paths.ArtifactsPath),
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorSession(stdout io.Writer, session core.AuthoringSession) error {
	lines := []string{
		"session: " + session.ID,
		"status: " + string(session.Status),
		"mode: " + string(session.Mode),
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		"updated_at: " + session.UpdatedAt.UTC().Format(time.RFC3339),
		fmt.Sprintf("inputs: %d", len(session.Inputs)),
		fmt.Sprintf("outputs: %d", len(session.Outputs)),
		fmt.Sprintf("provenance: %d", len(session.Provenance)),
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorContextualQuizPlan(
	stdout io.Writer,
	plan core.ContextualQuizGenerationPlan,
	dryRun bool,
	projectDir string,
) error {
	status := "created"
	if dryRun {
		status = "planned"
	}

	lines := []string{
		"quiz: " + status + " " + displayPath(projectDir, plan.Layout.FullPath),
		fmt.Sprintf("dry_run: %t", dryRun),
		"mode: context",
		"strategy: " + plan.Result.Strategy,
		"provider: " + plan.Result.Provider,
		"session: " + plan.Session.ID,
		"path: " + displayPath(projectDir, plan.Layout.FullPath),
		fmt.Sprintf("requested_questions: %d", plan.RequestedQuestionCount),
		fmt.Sprintf("generated_questions: %d", plan.GeneratedQuestionCount),
	}

	err := writeAuthorLines(stdout, lines)
	if err != nil {
		return err
	}

	if dryRun {
		_, err = fmt.Fprintln(stdout, plan.Content)
		if err != nil {
			return fmt.Errorf("write author contextual quiz content: %w", err)
		}
	}

	return nil
}

func printAuthorPromotionPlan(stdout io.Writer, plan core.QuizToDraftPromotionPlan, dryRun bool) error {
	lines := []string{
		"promotion: quiz-to-draft",
		fmt.Sprintf("dry_run: %t", dryRun),
		"session: " + plan.Session.ID,
		"source: " + plan.Source.Path,
		"path: " + plan.Layout.FullPath,
	}

	err := writeAuthorLines(stdout, lines)
	if err != nil {
		return err
	}

	if dryRun {
		_, err = fmt.Fprintln(stdout, plan.Content)
		if err != nil {
			return fmt.Errorf("write author promotion content: %w", err)
		}
	}

	return nil
}

func assistedEventCounts(events []core.AssistedTranscriptEvent) (int, int, int) {
	questions := map[string]struct{}{}
	answers := map[string]struct{}{}

	var results int

	for _, event := range events {
		switch event.Type {
		case core.AssistedTranscriptEventQuestion:
			questions[event.Question.ID] = struct{}{}
		case core.AssistedTranscriptEventAnswer:
			answers[event.Answer.QuestionID] = struct{}{}
		case core.AssistedTranscriptEventResult:
			results++
		case core.AssistedTranscriptEventState:
		}
	}

	return len(questions), len(answers), results
}

func writeAuthorLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write author command output: %w", err)
		}
	}

	return nil
}
