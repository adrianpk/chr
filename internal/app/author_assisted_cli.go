package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (factory commandFactory) authorAssistedCommand() *cli.Command {
	return &cli.Command{
		Name:  "assisted",
		Usage: "Run assisted authoring sessions",
		Commands: []*cli.Command{
			factory.authorAssistedAnswerCommand(),
			factory.authorAssistedFinalizeCommand(),
			factory.authorAssistedResumeCommand(),
			factory.authorAssistedStartCommand(),
			factory.authorAssistedStatusCommand(),
		},
	}
}

func (factory commandFactory) authorAssistedStartCommand() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start an assisted authoring session",
		Flags:  authorAssistedStartFlags(),
		Action: factory.runAuthorAssistedStart,
	}
}

func (factory commandFactory) authorAssistedResumeCommand() *cli.Command {
	return &cli.Command{
		Name:   "resume",
		Usage:  "Ask the next assisted authoring question",
		Flags:  authorAssistedResumeFlags(),
		Action: factory.runAuthorAssistedResume,
	}
}

func (factory commandFactory) authorAssistedAnswerCommand() *cli.Command {
	return &cli.Command{
		Name:   "answer",
		Usage:  "Submit an assisted authoring answer",
		Flags:  authorAssistedAnswerFlags(),
		Action: factory.runAuthorAssistedAnswer,
	}
}

func (factory commandFactory) authorAssistedStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Show assisted authoring status",
		Flags:  authorAssistedSessionFlags(),
		Action: factory.runAuthorAssistedStatus,
	}
}

func (factory commandFactory) authorAssistedFinalizeCommand() *cli.Command {
	return &cli.Command{
		Name:   "finalize",
		Usage:  "Finalize assisted authoring into an artifact",
		Flags:  authorAssistedFinalizeFlags(),
		Action: factory.runAuthorAssistedFinalize,
	}
}

func authorAssistedStartFlags() []cli.Flag {
	return append(authorProjectFlags(),
		&cli.StringFlag{
			Name:  "target",
			Usage: "authoring target",
			Value: string(authoringcore.TargetSpec),
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
			Name:  "operation-id",
			Usage: "stable identity for a project dump finalization retry",
		},
		&cli.StringFlag{
			Name:  "target-kind",
			Usage: "target artifact kind",
			Value: string(authoringcore.ArtifactKindSpec),
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

func (factory commandFactory) runAuthorAssistedStart(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted start does not accept positional arguments")
	}

	projectContext, err := factory.authorProjectContext(ctx, cmd)
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

	err = printAuthorSessionCreated(factory.io.Stdout, session, paths, projectContext.ProjectDir)
	if err != nil {
		return err
	}

	return writeAuthorLines(factory.io.Stdout,
		[]string{"assisted: started"})
}

func (factory commandFactory) runAuthorAssistedResume(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted resume does not accept positional arguments")
	}

	state, err := factory.loadAuthorAssistedState(ctx, cmd)
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

	return printAuthorAssistedQuestion(factory.io.Stdout, result)
}

func (factory commandFactory) runAuthorAssistedAnswer(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted answer does not accept positional arguments")
	}

	state, err := factory.loadAuthorAssistedState(ctx, cmd)
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

	return printAuthorAssistedAnswer(factory.io.Stdout, result)
}

func (factory commandFactory) runAuthorAssistedStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted status does not accept positional arguments")
	}

	state, err := factory.loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	return printAuthorAssistedStatus(factory.io.Stdout, state.Session, state.Events)
}

func (factory commandFactory) runAuthorAssistedFinalize(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author assisted finalize does not accept positional arguments")
	}

	state, err := factory.loadAuthorAssistedState(ctx, cmd)
	if err != nil {
		return err
	}

	summaryText := strings.TrimSpace(cmd.String("fake-summary"))
	if summaryText == "" {
		return fmt.Errorf("fake summary is required")
	}

	now := time.Now().UTC()

	if handled, err := factory.tryAuthorAssistedDump(ctx, cmd, state, summaryText, now); handled {
		return err
	}

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
		OperationID:    cmd.String("operation-id"),
		Session:        curation.Session,
		CuratedSummary: curation.Summary,
		SourceRef:      curation.ArtifactRef,
		TargetKind:     authoringcore.ArtifactKind(cmd.String("target-kind")),
		TargetName:     strings.TrimSpace(cmd.String("target-name")),
		AllowReplace:   cmd.Bool("allow-replace"),
		Repository:     state.Repository,
		Now:            func() time.Time { return now },
	})
	if err != nil {
		return err
	}

	return printAuthorAssistedFinalized(factory.io.Stdout, curation, synthesis)
}

type authorAssistedState struct {
	ProjectDir     string
	Session        authoringcore.Session
	Paths          core.AuthoringSessionPaths
	Events         []core.AssistedTranscriptEvent
	ContextPackage core.AssistedAuthoringContextPackage
	Repository     core.AuthoringSessionRepository
}

func (factory commandFactory) loadAuthorAssistedState(ctx context.Context, cmd *cli.Command) (authorAssistedState, error) {
	projectContext, err := factory.authorProjectContext(ctx, cmd)
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

func assistedSessionFromCommand(cmd *cli.Command, projectContext commandProjectContext, now time.Time) (authoringcore.Session, error) {
	topic := strings.TrimSpace(cmd.String("topic"))
	if topic == "" {
		return authoringcore.Session{}, fmt.Errorf("authoring topic is required")
	}

	session := authoringcore.Session{
		ID:          authoringSessionID(now),
		ProjectID:   projectContext.ProjectID,
		Status:      authoringcore.SessionStatusActive,
		Mode:        authoringcore.SessionModeAssisted,
		Target:      authoringcore.Target(cmd.String("target")),
		Topic:       topic,
		CreatedAt:   now,
		StartedAt:   &now,
		UpdatedAt:   now,
		Description: strings.TrimSpace(cmd.String("description")),
	}

	err := authoringcore.ValidateSession(session)
	if err != nil {
		return authoringcore.Session{}, err
	}

	return session, nil
}

func assistedConfigFromCommand(cmd *cli.Command, session authoringcore.Session) core.AssistedAuthoringConfig {
	return core.AssistedAuthoringConfig{
		SessionID:    session.ID,
		ProjectID:    session.ProjectID,
		Target:       session.Target,
		Intent:       core.AssistedAuthoringIntentDiscover,
		MaxTurns:     cmd.Int("max-turns"),
		MaxFollowUps: cmd.Int("max-follow-ups"),
	}
}

func authorAssistedCurationResponse(session authoringcore.Session, summary string, now time.Time) core.AssistedModelResponse {
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

func printAuthorAssistedStatus(stdout io.Writer, session authoringcore.Session, events []core.AssistedTranscriptEvent) error {
	counts := assistedEventCounts(events)

	return writeAuthorLines(stdout, []string{
		"assisted: status",
		"session: " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		fmt.Sprintf("questions: %d", counts.Questions),
		fmt.Sprintf("answers: %d", counts.Answers),
		fmt.Sprintf("results: %d", counts.Results),
	})
}

func printAuthorAssistedFinalized(stdout io.Writer, curation core.AssistedAnswerCurationResult, synthesis core.AssistedArtifactSynthesisResult) error {
	return writeAuthorLines(stdout, []string{
		"assisted: finalized",
		"session: " + synthesis.Session.ID,
		"curated_summary: " + curation.ArtifactPath,
		"artifact: " + synthesis.Layout.FullPath,
	})
}

type assistedTranscriptCounts struct {
	Questions int
	Answers   int
	Results   int
}

func assistedEventCounts(events []core.AssistedTranscriptEvent) assistedTranscriptCounts {
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

	return assistedTranscriptCounts{Questions: len(questions), Answers: len(answers), Results: results}
}
