package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

//go:embed assets
var assetsFS embed.FS

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := newRootCommand(stdout, stderr)

	osArgs := append([]string{"chc"}, args...)

	return cmd.Run(ctx, osArgs)
}

func newRootCommand(stdout, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:                  "chc",
		Usage:                 "Coordinate agent development workflows from the current project",
		Writer:                stdout,
		ErrWriter:             stderr,
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			newIndexCommand(stdout),
			newInitCommand(stdout),
			newProjectCommand(stdout),
			newTicketCommand(stdout),
		},
	}
}

func newIndexCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "index",
		Usage: "Manage derived project indexes",
		Commands: []*cli.Command{
			newIndexRebuildCommand(stdout),
			newIndexStatusCommand(stdout),
		},
	}
}

func newIndexRebuildCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "rebuild",
		Usage:  "Rebuild the resolved project's ticket index",
		Flags:  ticketProjectFlags(),
		Action: indexRebuildAction(stdout),
	}
}

func newIndexStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report the resolved project's ticket index status",
		Flags:  ticketProjectFlags(),
		Action: indexStatusAction(stdout),
	}
}

func newInitCommand(stdout io.Writer) *cli.Command {
	workspace := chc.NewWorkspace(assetsFS)

	return &cli.Command{
		Name:  "init",
		Usage: "Create a .chc workspace in the current project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "profile",
				Usage: "workspace profile: standard, personal, minimal",
				Value: chc.DefaultProfile,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print planned changes without writing files",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 0 {
				return fmt.Errorf("init does not accept positional arguments")
			}

			result, err := workspace.Init(ctx, chc.InitOptions{
				ProjectDir: cmd.String("dir"),
				Profile:    cmd.String("profile"),
				DryRun:     cmd.Bool("dry-run"),
			})
			if err != nil {
				return err
			}

			return printInitResult(stdout, result)
		},
	}
}

func newTicketCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "ticket",
		Usage: "Manage workspace tickets",
		Commands: []*cli.Command{
			newTicketCommentCommand(stdout),
			newTicketListCommand(stdout),
			newTicketNewCommand(stdout),
			newTicketShowCommand(stdout),
			newTicketStatusCommand(stdout),
		},
	}
}

func newProjectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage registered projects",
		Commands: []*cli.Command{
			newProjectAddCommand(stdout),
			newProjectCurrentCommand(stdout),
			newProjectListCommand(stdout),
			newProjectSelectCommand(stdout),
		},
	}
}

func projectRegistryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "dev-root",
			Usage: "development user environment root",
		},
	}
}

func newUserEnvironmentFromCommand(cmd *cli.Command) (chc.UserEnvironment, error) {
	devRoot := cmd.String("dev-root")
	if devRoot != "" {
		return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
			DevMode: true,
			DevRoot: devRoot,
		})
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		HomeDir:   homeDir,
		XDGConfig: os.Getenv("XDG_CONFIG_HOME"),
		XDGData:   os.Getenv("XDG_DATA_HOME"),
		XDGCache:  os.Getenv("XDG_CACHE_HOME"),
		XDGState:  os.Getenv("XDG_STATE_HOME"),
	})
}

func newProjectAddCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Register a local project",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "name",
			Usage: "registered project name",
		}),
		Action: projectAddAction(stdout),
	}
}

func projectAddAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectAdd(ctx, cmd, stdout)
	}
}

func runProjectAdd(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project add requires exactly one project directory argument")
	}

	projectDir, err := filepath.Abs(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}

	if err := requireWorkspace(projectDir); err != nil {
		return err
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projectName := cmd.String("name")
	if projectName == "" {
		projectName = filepath.Base(projectDir)
	}

	project := chc.Project{
		ID:           projectID(projectDir),
		Name:         projectName,
		Path:         projectDir,
		Enabled:      true,
		RegisteredAt: time.Now().UTC(),
	}

	if err := registry.AddProject(ctx, project); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project add result: %w", err)
	}

	return nil
}

func requireWorkspace(projectDir string) error {
	info, err := os.Stat(filepath.Join(projectDir, chc.DirectoryName))
	if err != nil {
		return fmt.Errorf("inspect project workspace: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("project workspace is not a directory: %s", filepath.Join(projectDir, chc.DirectoryName))
	}

	return nil
}

func projectID(projectDir string) string {
	sum := sha256.Sum256([]byte(projectDir))

	return "project-" + hex.EncodeToString(sum[:8])
}

func newProjectListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List registered projects",
		Flags:  projectRegistryFlags(),
		Action: projectListAction(stdout),
	}
}

func projectListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectList(ctx, cmd, stdout)
	}
}

func runProjectList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projects, err := registry.ListProjects(ctx)
	if err != nil {
		return err
	}

	for _, project := range projects {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\t%t\t%s\n", project.Name, project.Path, project.Enabled, project.ID); err != nil {
			return fmt.Errorf("write project list result: %w", err)
		}
	}

	return nil
}

func newProjectCurrentCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "current",
		Usage:  "Print the active project",
		Flags:  projectRegistryFlags(),
		Action: projectCurrentAction(stdout),
	}
}

func projectCurrentAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectCurrent(ctx, cmd, stdout)
	}
}

func runProjectCurrent(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project current does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	activeProject, err := chc.NewActiveProjectStore(env).LoadActiveProject()
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := registry.GetProjectByID(ctx, activeProject.ProjectID)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project current result: %w", err)
	}

	return nil
}

func newProjectSelectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "select",
		Usage:  "Select the active project",
		Flags:  projectRegistryFlags(),
		Action: projectSelectAction(stdout),
	}
}

func projectSelectAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectSelect(ctx, cmd, stdout)
	}
}

func runProjectSelect(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project select requires exactly one project name or path argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := resolveProjectSelection(ctx, registry, cmd.Args().First())
	if err != nil {
		return err
	}

	if !project.Enabled {
		return fmt.Errorf("project is disabled: %s", project.Name)
	}

	if _, err := resolveRegisteredProjectDirectory(project); err != nil {
		return err
	}

	activeProject, err := chc.NewActiveProject(project.ID)
	if err != nil {
		return err
	}

	store := chc.NewActiveProjectStore(env)
	if err := store.SetActiveProject(activeProject); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path); err != nil {
		return fmt.Errorf("write project select result: %w", err)
	}

	return nil
}

func resolveProjectSelection(ctx context.Context, registry chc.ProjectRegistry, selector string) (chc.Project, error) {
	project, err := registry.GetProjectByName(ctx, selector)
	if err == nil {
		return project, nil
	}

	if !errors.Is(err, chc.ErrProjectNotFound) {
		return chc.Project{}, err
	}

	projectPath, pathErr := filepath.Abs(selector)
	if pathErr != nil {
		return chc.Project{}, fmt.Errorf("resolve project selector path: %w", pathErr)
	}

	project, err = registry.GetProjectByPath(ctx, projectPath)
	if err != nil {
		return chc.Project{}, fmt.Errorf("resolve project selection: %w", err)
	}

	return project, nil
}

func newTicketNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a ticket in the current workspace",
		Flags:  ticketNewFlags(),
		Action: ticketNewAction(stdout),
	}
}

func newTicketListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List tickets for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: ticketListAction(stdout),
	}
}

func newTicketShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one ticket by key",
		Flags:  ticketProjectFlags(),
		Action: ticketShowAction(stdout),
	}
}

func newTicketCommentCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "comment",
		Usage:  "Append a comment to one ticket",
		Flags:  ticketCommentFlags(),
		Action: ticketCommentAction(stdout),
	}
}

func newTicketStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Update one ticket status",
		Flags:  ticketProjectFlags(),
		Action: ticketStatusAction(stdout),
	}
}

func ticketProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func ticketCommentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "author",
			Usage: "comment author",
			Value: "user",
		},
	)
}

func ticketNewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "priority",
			Usage: "ticket priority",
		},
		&cli.StringFlag{
			Name:  "severity",
			Usage: "ticket severity",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "ticket type",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag",
		},
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "ticket use case",
		},
		&cli.StringFlag{
			Name:  "project-area",
			Usage: "ticket project area",
		},
	)
}

func ticketListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketList(ctx, cmd, stdout)
	}
}

func ticketShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketShow(ctx, cmd, stdout)
	}
}

func ticketCommentAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketComment(ctx, cmd, stdout)
	}
}

func ticketStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketStatus(ctx, cmd, stdout)
	}
}

func runTicketStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket status requires exactly one ticket key and one status argument")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	status := cmd.Args().Get(1)
	if err := chc.NewTicketStatusUpdater().UpdateStatus(ctx, chc.UpdateTicketStatusOptions{
		Path:      ticket.Path,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s %s\n", ticket.Key, status); err != nil {
		return fmt.Errorf("write ticket status result: %w", err)
	}

	return nil
}

func runTicketComment(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket comment requires exactly one ticket key and one comment argument")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	if err := chc.NewTicketCommenter().AddComment(ctx, chc.AddTicketCommentOptions{
		Path:      ticket.Path,
		Author:    cmd.String("author"),
		Body:      cmd.Args().Get(1),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s\n", ticket.Path); err != nil {
		return fmt.Errorf("write ticket comment result: %w", err)
	}

	return nil
}

func runTicketShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket show requires exactly one ticket key argument")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return err
	}

	return printTicketShow(stdout, ticket)
}

func printTicketShow(stdout io.Writer, ticket chc.IndexedTicket) error {
	lines := []string{
		"key: " + ticket.Key,
		"title: " + ticket.Title,
		"status: " + ticket.Status,
		"priority: " + ticket.Priority,
		"severity: " + ticket.Severity,
		"type: " + ticket.Type,
		"use_case: " + ticket.UseCase,
		"project_area: " + ticket.ProjectArea,
		"path: " + ticket.RelativePath,
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket show result: %w", err)
		}
	}

	return nil
}

func runTicketList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket list does not accept positional arguments")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		if _, err := fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
		); err != nil {
			return fmt.Errorf("write ticket list result: %w", err)
		}
	}

	return nil
}

func ticketNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketNew(ctx, cmd, stdout)
	}
}

func indexRebuildAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runIndexRebuild(ctx, cmd, stdout)
	}
}

func indexStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runIndexStatus(ctx, cmd, stdout)
	}
}

func runIndexRebuild(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index rebuild does not accept positional arguments")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := chc.NewTicketIndexRebuilder(index).RebuildProjectTickets(ctx, chc.RebuildProjectTicketsOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return printIndexRebuildResult(stdout, result)
}

func printIndexRebuildResult(stdout io.Writer, result chc.RebuildProjectTicketsResult) error {
	if _, err := fmt.Fprintf(stdout, "indexed: %d\n", result.IndexedTickets); err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "parse_errors: %d\n", len(result.ParseErrors)); err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	for _, parseErr := range result.ParseErrors {
		if _, err := fmt.Fprintf(stdout, "parse_error: %s %v\n", parseErr.RelativePath, parseErr.Err); err != nil {
			return fmt.Errorf("write index rebuild result: %w", err)
		}
	}

	return nil
}

func runIndexStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index status does not accept positional arguments")
	}

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

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	status, err := chc.NewTicketIndexStatusChecker(index).CheckProjectTicketIndex(ctx, chc.CheckProjectTicketIndexOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if err := printIndexStatusResult(stdout, status); err != nil {
		return err
	}

	if ticketIndexHasDrift(status) {
		return fmt.Errorf("ticket index is not current")
	}

	return nil
}

func printIndexStatusResult(stdout io.Writer, status chc.TicketIndexStatus) error {
	if err := printIndexStatusCounts(stdout, status); err != nil {
		return err
	}

	if err := printIndexStatusDetails(stdout, status); err != nil {
		return err
	}

	return nil
}

func printIndexStatusCounts(stdout io.Writer, status chc.TicketIndexStatus) error {
	if _, err := fmt.Fprintf(stdout, "missing: %d\n", len(status.Missing)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "stale: %d\n", len(status.Stale)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "orphaned: %d\n", len(status.Orphaned)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	if _, err := fmt.Fprintf(stdout, "invalid: %d\n", len(status.Invalid)); err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	return nil
}

func printIndexStatusDetails(stdout io.Writer, status chc.TicketIndexStatus) error {
	for _, missing := range status.Missing {
		if _, err := fmt.Fprintf(stdout, "missing_ticket: %s\n", missing.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, stale := range status.Stale {
		if _, err := fmt.Fprintf(stdout, "stale_ticket: %s\n", stale.Parsed.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, orphaned := range status.Orphaned {
		if _, err := fmt.Fprintf(stdout, "orphaned_ticket: %s\n", orphaned.RelativePath); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, invalid := range status.Invalid {
		if _, err := fmt.Fprintf(stdout, "invalid_ticket: %s %v\n", invalid.RelativePath, invalid.Err); err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	return nil
}

func ticketIndexHasDrift(status chc.TicketIndexStatus) bool {
	return len(status.Missing) > 0 ||
		len(status.Stale) > 0 ||
		len(status.Orphaned) > 0 ||
		len(status.Invalid) > 0
}

func runTicketNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket new requires exactly one title argument")
	}

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

	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	repository := chc.NewTicketRepository(
		projectContext.ProjectDir,
		chc.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
	)

	result, err := repository.CreateTicket(ctx, chc.CreateTicketOptions{
		Title:       cmd.Args().First(),
		Priority:    cmd.String("priority"),
		Severity:    cmd.String("severity"),
		Type:        cmd.String("type"),
		Tags:        cmd.StringSlice("tag"),
		UseCase:     cmd.String("use-case"),
		ProjectArea: cmd.String("project-area"),
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "ticket: %s\n", result.Path); err != nil {
		return fmt.Errorf("write ticket result: %w", err)
	}

	return nil
}

func printInitResult(stdout io.Writer, result chc.InitResult) error {
	if _, err := fmt.Fprintf(stdout, "workspace: %s\n", result.Root); err != nil {
		return fmt.Errorf("write workspace result: %w", err)
	}

	for _, created := range result.Created {
		if _, err := fmt.Fprintf(stdout, "create %s\n", created); err != nil {
			return fmt.Errorf("write created path: %w", err)
		}
	}

	for _, skipped := range result.Skipped {
		if _, err := fmt.Fprintf(stdout, "skip existing %s\n", skipped); err != nil {
			return fmt.Errorf("write skipped path: %w", err)
		}
	}

	if _, err := fmt.Fprintln(stdout, "done"); err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}

	return nil
}
