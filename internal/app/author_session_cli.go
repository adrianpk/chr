package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) authorSessionCommand() *cli.Command {
	return &cli.Command{
		Name:  "session",
		Usage: "Manage authoring sessions",
		Commands: []*cli.Command{
			factory.authorSessionListCommand(),
			factory.authorSessionNewCommand(),
			factory.authorSessionShowCommand(),
		},
	}
}

func (factory commandFactory) authorSessionNewCommand() *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create an authoring session",
		Flags:  authorSessionFlags(),
		Action: factory.runAuthorSessionNew,
	}
}

func (factory commandFactory) authorSessionListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List authoring sessions",
		Flags:  authorProjectFlags(),
		Action: factory.runAuthorSessionList,
	}
}

func (factory commandFactory) authorSessionShowCommand() *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one authoring session",
		Flags:  authorProjectFlags(),
		Action: factory.runAuthorSessionShow,
	}
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

func (factory commandFactory) runAuthorSessionNew(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author session new does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
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

	result, err := newService(ctx, env, projectContext, factory.deps.assetsFS).CreateAuthoringSession(session)
	if err != nil {
		return err
	}

	return printAuthorSessionCreated(factory.io.Stdout, result.Session, result.Paths, projectContext.ProjectDir)
}

func (factory commandFactory) runAuthorSessionList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("author session list does not accept positional arguments")
	}

	projectContext, err := factory.authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	sessions, err := core.NewAuthoringSessionRepository(projectContext.ProjectDir).ListSessions(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(factory.io.Stdout,
		"id\tstatus\ttarget\tupdated_at\ttopic")
	if err != nil {
		return fmt.Errorf("write author session list header: %w", err)
	}

	for _, session := range sessions {
		_, err = fmt.Fprintf(factory.io.Stdout,
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

func (factory commandFactory) runAuthorSessionShow(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("author session show requires exactly one session id argument")
	}

	projectContext, err := factory.authorProjectContext(ctx, cmd)
	if err != nil {
		return err
	}

	repository := core.NewAuthoringSessionRepository(projectContext.ProjectDir)

	session, err := resolveAuthoringSession(ctx, repository, cmd.Args().First())
	if err != nil {
		return err
	}

	return printAuthorSession(factory.io.Stdout, session)
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
