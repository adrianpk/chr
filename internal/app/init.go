package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

func newInitCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	workspaceInitializer := core.NewWorkspace(deps.assetsFS)

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
				Value: core.DefaultProfile,
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

			result, err := workspaceInitializer.Init(ctx, core.InitOptions{
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

func printInitResult(stdout io.Writer, result core.InitResult) error {
	_, err := fmt.Fprintf(stdout, "workspace: %s\n", result.Root)
	if err != nil {
		return fmt.Errorf("write workspace result: %w", err)
	}

	for _, created := range result.Created {
		_, err = fmt.Fprintf(stdout, "create %s\n", created)
		if err != nil {
			return fmt.Errorf("write created path: %w", err)
		}
	}

	for _, skipped := range result.Skipped {
		_, err = fmt.Fprintf(stdout, "skip existing %s\n", skipped)
		if err != nil {
			return fmt.Errorf("write skipped path: %w", err)
		}
	}

	_, err = fmt.Fprintln(stdout, "done")
	if err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}

	return nil
}
