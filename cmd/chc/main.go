package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
			newInitCommand(stdout),
			newProjectCommand(stdout),
			newTicketCommand(stdout),
		},
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
			newTicketNewCommand(stdout),
		},
	}
}

func newProjectCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage registered projects",
		Commands: []*cli.Command{
			newProjectAddCommand(stdout),
			newProjectListCommand(stdout),
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
		Name:   "add",
		Usage:  "Register a local project",
		Flags:  projectRegistryFlags(),
		Action: projectAddAction(stdout),
	}
}

func projectAddAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if _, err := fmt.Fprintln(stdout, "project add pending"); err != nil {
			return fmt.Errorf("write project add result: %w", err)
		}

		return nil
	}
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
		if _, err := fmt.Fprintln(stdout, "project list pending"); err != nil {
			return fmt.Errorf("write project list result: %w", err)
		}

		return nil
	}
}

func newTicketNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a ticket in the current workspace",
		Flags:  ticketNewFlags(),
		Action: ticketNewAction(stdout),
	}
}

func ticketNewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
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
	}
}

func ticketNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketNew(ctx, cmd, stdout)
	}
}

func runTicketNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket new requires exactly one title argument")
	}

	projectDir, err := filepath.Abs(cmd.String("dir"))
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}

	config, err := chc.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return err
	}

	repository := chc.NewTicketRepository(
		projectDir,
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
