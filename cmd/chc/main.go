package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/workspace"
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
		},
	}
}

func newInitCommand(stdout io.Writer) *cli.Command {
	initializer := workspace.NewInitializer(assetsFS)

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
				Value: workspace.DefaultProfile,
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

			result, err := initializer.Init(ctx, workspace.InitOptions{
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

func printInitResult(stdout io.Writer, result workspace.InitResult) error {
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
