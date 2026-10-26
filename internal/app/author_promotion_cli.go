package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) authorPromoteCommand() *cli.Command {
	return &cli.Command{
		Name:  "promote",
		Usage: "Promote authoring artifacts",
		Commands: []*cli.Command{
			factory.authorPromoteQuizToDraftCommand(),
		},
	}
}

func (factory commandFactory) authorPromoteQuizToDraftCommand() *cli.Command {
	return &cli.Command{
		Name:   "quiz-to-draft",
		Usage:  "Promote a quiz artifact into an editable draft",
		Flags:  authorPromoteQuizToDraftFlags(),
		Action: factory.runAuthorPromoteQuizToDraft,
	}
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

func (factory commandFactory) runAuthorPromoteQuizToDraft(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author promote quiz-to-draft does not accept positional arguments")
	}

	projectContext, err := factory.authorProjectContext(ctx, cmd)
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
		return printAuthorPromotionPlan(factory.io.Stdout, plan, true)
	}

	err = core.WriteQuizToDraftPromotionContext(ctx, plan)
	if err != nil {
		return err
	}

	_, err = repository.UpdateSession(ctx, plan.Session)
	if err != nil {
		return err
	}

	return printAuthorPromotionPlan(factory.io.Stdout, plan, false)
}

func resolveAuthoringSessionForPromotion(ctx context.Context, repository core.AuthoringSessionRepository, projectDir, handle, artifactPath string) (core.AuthoringSession, error) {
	if strings.TrimSpace(handle) != "" {
		return resolveAuthoringSession(ctx, repository, handle)
	}

	return resolveAuthoringSessionFromArtifact(ctx, repository, projectDir, artifactPath)
}

func resolveAuthoringSessionFromArtifact(ctx context.Context, repository core.AuthoringSessionRepository, projectDir, artifactPath string) (core.AuthoringSession, error) {
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

func authoringSessionHasArtifact(session core.AuthoringSession, projectDir, artifactPath string) bool {
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
