package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

var newAuthoringQuizQuestionProvider = func() core.AuthoringQuizQuestionProvider {
	return core.HarnessAuthoringQuizQuestionProvider{Program: codex.Program, Runner: codex.Runner{}}
}

func (factory commandFactory) authorQuizCommand() *cli.Command {
	return &cli.Command{
		Name:  "quiz",
		Usage: "Generate authoring quizzes",
		Commands: []*cli.Command{
			factory.authorQuizNewCommand(),
		},
	}
}

func (factory commandFactory) authorQuizNewCommand() *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create an authoring quiz",
		Flags:  authorQuizNewFlags(),
		Action: factory.runAuthorQuizNew,
	}
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

func (factory commandFactory) runAuthorQuizNew(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author quiz new does not accept positional arguments")
	}

	projectContext, err := factory.authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	session, err := quizSessionFromCommand(cmd, projectContext, now)
	if err != nil {
		return err
	}

	return factory.runAuthorContextualQuizNew(ctx, cmd, projectContext, session, now)
}

func (factory commandFactory) runAuthorContextualQuizNew(ctx context.Context, cmd *cli.Command, projectContext commandProjectContext, session core.AuthoringSession, now time.Time) error {
	return withContextualQuizGeneration(ctx, cmd, projectContext, session.ID, func(ctx context.Context, provider core.AuthoringQuizQuestionProvider) error {
		plan, err := core.PlanContextualQuizGeneration(ctx, core.ContextualQuizGenerationOptions{
			ProjectDir:    projectContext.ProjectDir,
			Session:       session,
			Now:           now,
			AllowReplace:  cmd.Bool("allow-replace"),
			DryRun:        cmd.Bool("dry-run"),
			QuestionCount: cmd.Int("question-count"),
			Provider:      provider,
		})
		if err != nil {
			return err
		}

		if cmd.Bool("dry-run") {
			return printAuthorContextualQuizPlan(factory.io.Stdout, plan, true, projectContext.ProjectDir)
		}

		err = core.WriteContextualQuizContext(ctx, plan)
		if err != nil {
			return err
		}

		_, err = core.NewAuthoringSessionRepository(projectContext.ProjectDir).CreateSession(ctx, plan.Session)
		if err != nil {
			return err
		}

		return printAuthorContextualQuizPlan(factory.io.Stdout, plan, false, projectContext.ProjectDir)
	})
}

func quizSessionFromCommand(cmd *cli.Command, projectContext commandProjectContext, now time.Time) (core.AuthoringSession, error) {
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

func printAuthorContextualQuizPlan(stdout io.Writer, plan core.ContextualQuizGenerationPlan, dryRun bool, projectDir string) error {
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
