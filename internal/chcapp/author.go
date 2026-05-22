package chcapp

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newAuthorCommand(stdout io.Writer) *cli.Command {
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
		Usage: "Generate static authoring quizzes",
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
		Usage:  "Create a static authoring quiz",
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
			Value: string(chc.AuthoringTargetSpec),
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
			Usage: "authoring session id",
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
			Value: string(chc.AuthoringArtifactKindSpec),
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
			Value: string(chc.AuthoringSessionModeStatic),
		},
		&cli.StringFlag{
			Name:  "target",
			Usage: "authoring target",
			Value: string(chc.AuthoringTargetDraft),
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
			Value: string(chc.AuthoringTargetSpec),
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

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	session, err := authoringSessionFromCommand(cmd, projectContext, time.Now().UTC())
	if err != nil {
		return err
	}

	paths, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, session)
	if err != nil {
		return err
	}

	return printAuthorSessionCreated(stdout, session, paths, projectContext.ProjectDir)
}

func runAuthorSessionList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author session list does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	sessions, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "id\tstatus\ttarget\ttopic")
	if err != nil {
		return fmt.Errorf("write author session list header: %w", err)
	}

	for _, session := range sessions {
		_, err = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", session.ID, session.Status, session.Target, session.Topic)
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

	session, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).GetSession(ctx, cmd.Args().First())
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

	session, err := staticQuizSessionFromCommand(cmd, projectContext, now)
	if err != nil {
		return err
	}

	plan, err := chc.PlanStaticQuizGeneration(chc.StaticQuizGenerationOptions{
		ProjectDir:    projectContext.ProjectDir,
		Session:       session,
		Now:           now,
		AllowReplace:  cmd.Bool("allow-replace"),
		DryRun:        cmd.Bool("dry-run"),
		QuestionCount: cmd.Int("question-count"),
	})
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printAuthorQuizPlan(stdout, plan, true, projectContext.ProjectDir)
	}

	err = chc.WriteStaticQuiz(plan)
	if err != nil {
		return err
	}

	_, err = chc.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, plan.Session)
	if err != nil {
		return err
	}

	return printAuthorQuizPlan(stdout, plan, false, projectContext.ProjectDir)
}

func runAuthorPromoteQuizToDraft(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author promote quiz-to-draft does not accept positional arguments")
	}

	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	sessionID := strings.TrimSpace(cmd.String("session"))
	if sessionID == "" {
		return fmt.Errorf("authoring session id is required")
	}

	quizPath := strings.TrimSpace(cmd.String("quiz"))
	if quizPath == "" {
		return fmt.Errorf("quiz path is required")
	}

	repository := chc.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := repository.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	plan, err := chc.PlanQuizToDraftPromotion(chc.QuizToDraftPromotionOptions{
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

	err = chc.WriteQuizToDraftPromotion(plan)
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

	paths, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, session)
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
	response := chc.AssistedModelResponse{
		ID:        "response-" + authoringSessionID(now),
		Provider:  "fake",
		Status:    chc.AssistedModelResponseReady,
		Question:  chc.AssistedAuthoringQuestion{ID: "question-" + authoringSessionID(now), Prompt: question, Required: true},
		RawText:   question,
		CreatedAt: now,
	}

	result, err := chc.AskNextAssistedQuizQuestion(ctx, chc.AssistedQuizLoopOptions{
		Config:           assistedConfigFromCommand(cmd, state.Session),
		Session:          state.Session,
		SessionPaths:     state.Paths,
		ContextPackage:   state.ContextPackage,
		TranscriptEvents: state.Events,
		Repository:       state.Repository,
		Adapter:          &chc.FakeAssistedModelAdapter{Responses: []chc.AssistedModelResponse{response}},
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

	result, err := chc.AnswerAssistedQuizQuestion(ctx, chc.AssistedQuizAnswerOptions{
		SessionID:  state.Session.ID,
		TurnID:     strings.TrimSpace(cmd.String("turn")),
		TurnIndex:  cmd.Int("turn-index"),
		Intent:     chc.AssistedAuthoringIntentDiscover,
		Question:   chc.AssistedAuthoringQuestion{ID: strings.TrimSpace(cmd.String("question-id")), Prompt: strings.TrimSpace(cmd.String("question")), Required: true},
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

	curation, err := chc.CurateAssistedAnswers(ctx, chc.AssistedAnswerCurationOptions{
		Session:          state.Session,
		SessionPaths:     state.Paths,
		TranscriptEvents: state.Events,
		Repository:       state.Repository,
		Adapter: &chc.FakeAssistedModelAdapter{Responses: []chc.AssistedModelResponse{
			authorAssistedCurationResponse(state.Session, summaryText, now),
		}},
		Now: func() time.Time { return now },
	})
	if err != nil {
		return err
	}

	synthesis, err := chc.SynthesizeAssistedArtifact(ctx, chc.AssistedArtifactSynthesisOptions{
		ProjectDir:     state.ProjectDir,
		Session:        curation.Session,
		CuratedSummary: curation.Summary,
		SourceRef:      curation.ArtifactRef,
		TargetKind:     chc.AuthoringArtifactKind(cmd.String("target-kind")),
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
	Session        chc.AuthoringSession
	Paths          chc.AuthoringSessionPaths
	Events         []chc.AssistedTranscriptEvent
	ContextPackage chc.AssistedAuthoringContextPackage
	Repository     chc.AuthoringSessionRepository
}

func loadAuthorAssistedState(ctx context.Context, cmd *cli.Command) (authorAssistedState, error) {
	projectContext, err := authorProjectContext(ctx, cmd)
	if err != nil {
		return authorAssistedState{}, err
	}

	sessionID := strings.TrimSpace(cmd.String("session"))
	if sessionID == "" {
		return authorAssistedState{}, fmt.Errorf("authoring session id is required")
	}

	repository := chc.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := repository.GetSession(ctx, sessionID)
	if err != nil {
		return authorAssistedState{}, err
	}

	paths := chc.AuthoringSessionLayout(projectContext.ProjectDir, session.ID)

	events, err := repository.ListTranscriptEvents(ctx, session.ID)
	if err != nil {
		return authorAssistedState{}, err
	}

	contextPackage, err := chc.LoadAssistedAuthoringContextPackage(chc.AssistedAuthoringContextPackageOptions{
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

	return newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
}

func assistedSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (chc.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return chc.AuthoringSession{}, fmt.Errorf("authoring topic is required")
	}

	session := chc.AuthoringSession{
		ID:          authoringSessionID(now),
		ProjectID:   projectContext.ProjectID,
		Status:      chc.AuthoringSessionStatusActive,
		Mode:        chc.AuthoringSessionModeAssisted,
		Target:      chc.AuthoringTarget(cmd.String("target")),
		Topic:       topic,
		CreatedAt:   now,
		StartedAt:   &now,
		UpdatedAt:   now,
		Description: strings.TrimSpace(cmd.String("description")),
	}

	err := chc.ValidateAuthoringSession(session)
	if err != nil {
		return chc.AuthoringSession{}, err
	}

	return session, nil
}

func authoringSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (chc.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return chc.AuthoringSession{}, fmt.Errorf("authoring topic is required")
	}

	session := chc.AuthoringSession{
		ID:          authoringSessionID(now),
		ProjectID:   projectContext.ProjectID,
		Status:      chc.AuthoringSessionStatusDraft,
		Mode:        chc.AuthoringSessionMode(cmd.String("mode")),
		Target:      chc.AuthoringTarget(cmd.String("target")),
		Topic:       topic,
		CreatedAt:   now,
		UpdatedAt:   now,
		Description: strings.TrimSpace(cmd.String("description")),
	}

	err := chc.ValidateAuthoringSession(session)
	if err != nil {
		return chc.AuthoringSession{}, err
	}

	return session, nil
}

func staticQuizSessionFromCommand(
	cmd *cli.Command,
	projectContext commandProjectContext,
	now time.Time,
) (chc.AuthoringSession, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return chc.AuthoringSession{}, fmt.Errorf("quiz topic is required")
	}

	session := chc.AuthoringSession{
		ID:        authoringSessionID(now),
		ProjectID: projectContext.ProjectID,
		Status:    chc.AuthoringSessionStatusDraft,
		Mode:      chc.AuthoringSessionModeStatic,
		Target:    chc.AuthoringTarget(cmd.String("target")),
		Topic:     topic,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := chc.ValidateAuthoringSession(session)
	if err != nil {
		return chc.AuthoringSession{}, err
	}

	return session, nil
}

func assistedConfigFromCommand(cmd *cli.Command, session chc.AuthoringSession) chc.AssistedAuthoringConfig {
	return chc.AssistedAuthoringConfig{
		SessionID:    session.ID,
		ProjectID:    session.ProjectID,
		Target:       session.Target,
		Intent:       chc.AssistedAuthoringIntentDiscover,
		MaxTurns:     cmd.Int("max-turns"),
		MaxFollowUps: cmd.Int("max-follow-ups"),
	}
}

func authorAssistedCurationResponse(
	session chc.AuthoringSession,
	summary string,
	now time.Time,
) chc.AssistedModelResponse {
	return chc.AssistedModelResponse{
		ID:        "curation-response-" + session.ID,
		RequestID: "curation-" + session.ID,
		SessionID: session.ID,
		TurnID:    "curation",
		Provider:  "fake",
		Status:    chc.AssistedModelResponseReady,
		Question: chc.AssistedAuthoringQuestion{
			ID:       "curation-ready-" + session.ID,
			Prompt:   "Curation is ready.",
			Required: false,
		},
		RawText:   summary,
		CreatedAt: now,
	}
}

func authoringSessionID(now time.Time) string {
	return "auth-" + now.UTC().Format("20060102-150405")
}

func printAuthorAssistedQuestion(stdout io.Writer, result chc.AssistedQuizAskResult) error {
	lines := []string{
		"assisted: question",
		"turn: " + result.Turn.ID,
		fmt.Sprintf("turn_index: %d", result.Turn.Index),
		"question_id: " + result.Turn.Question.ID,
		"question: " + result.Turn.Question.Prompt,
		"request: " + result.InvocationPaths.RequestPath,
		"response: " + result.InvocationPaths.ResponsePath,
	}

	if result.StopReason != chc.AssistedQuizLoopStopNone {
		lines = append(lines, "stop_reason: "+string(result.StopReason))
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorAssistedAnswer(stdout io.Writer, result chc.AssistedQuizAnswerResult) error {
	lines := []string{
		"assisted: answer",
		"turn: " + result.Turn.ID,
		fmt.Sprintf("turn_index: %d", result.Turn.Index),
		"question_id: " + result.Turn.Question.ID,
	}

	if result.StopReason != chc.AssistedQuizLoopStopNone {
		lines = append(lines, "stop_reason: "+string(result.StopReason))
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorAssistedStatus(
	stdout io.Writer,
	session chc.AuthoringSession,
	events []chc.AssistedTranscriptEvent,
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
	curation chc.AssistedAnswerCurationResult,
	synthesis chc.AssistedArtifactSynthesisResult,
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
	session chc.AuthoringSession,
	paths chc.AuthoringSessionPaths,
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

func printAuthorSession(stdout io.Writer, session chc.AuthoringSession) error {
	lines := []string{
		"session: " + session.ID,
		"status: " + string(session.Status),
		"mode: " + string(session.Mode),
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		fmt.Sprintf("inputs: %d", len(session.Inputs)),
		fmt.Sprintf("outputs: %d", len(session.Outputs)),
		fmt.Sprintf("provenance: %d", len(session.Provenance)),
	}

	return writeAuthorLines(stdout, lines)
}

func printAuthorQuizPlan(
	stdout io.Writer,
	plan chc.StaticQuizGenerationPlan,
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
		"session: " + plan.Session.ID,
		"path: " + displayPath(projectDir, plan.Layout.FullPath),
		fmt.Sprintf("questions: %d", len(plan.Questions)),
	}

	err := writeAuthorLines(stdout, lines)
	if err != nil {
		return err
	}

	if dryRun {
		_, err = fmt.Fprintln(stdout, plan.Content)
		if err != nil {
			return fmt.Errorf("write author quiz content: %w", err)
		}
	}

	return nil
}

func printAuthorPromotionPlan(stdout io.Writer, plan chc.QuizToDraftPromotionPlan, dryRun bool) error {
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

func assistedEventCounts(events []chc.AssistedTranscriptEvent) (int, int, int) {
	questions := map[string]struct{}{}
	answers := map[string]struct{}{}

	var results int

	for _, event := range events {
		switch event.Type {
		case chc.AssistedTranscriptEventQuestion:
			questions[event.Question.ID] = struct{}{}
		case chc.AssistedTranscriptEventAnswer:
			answers[event.Answer.QuestionID] = struct{}{}
		case chc.AssistedTranscriptEventResult:
			results++
		case chc.AssistedTranscriptEventState:
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
