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
			newAuthorQuizCommand(stdout),
			newAuthorSessionCommand(stdout),
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

func authorProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
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

	return printAuthorSessionCreated(stdout, session, paths)
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

	for _, session := range sessions {
		_, err = fmt.Fprintf(stdout, "session: %s\t%s\t%s\t%s\n", session.ID, session.Status, session.Target, session.Topic)
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
		return printAuthorQuizPlan(stdout, plan, true)
	}

	err = chc.WriteStaticQuiz(plan)
	if err != nil {
		return err
	}

	paths, err := chc.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, plan.Session)
	if err != nil {
		return err
	}

	err = printAuthorSessionCreated(stdout, plan.Session, paths)
	if err != nil {
		return err
	}

	return printAuthorQuizPlan(stdout, plan, false)
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

func authoringSessionID(now time.Time) string {
	return "auth-" + now.UTC().Format("20060102-150405")
}

func printAuthorSessionCreated(
	stdout io.Writer,
	session chc.AuthoringSession,
	paths chc.AuthoringSessionPaths,
) error {
	lines := []string{
		"session: created",
		"id: " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		"path: " + paths.SessionPath,
		"transcript: " + paths.TranscriptPath,
		"artifacts: " + paths.ArtifactsPath,
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

func printAuthorQuizPlan(stdout io.Writer, plan chc.StaticQuizGenerationPlan, dryRun bool) error {
	lines := []string{
		"quiz: planned",
		fmt.Sprintf("dry_run: %t", dryRun),
		"session: " + plan.Session.ID,
		"path: " + plan.Layout.FullPath,
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

func writeAuthorLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write author command output: %w", err)
		}
	}

	return nil
}
