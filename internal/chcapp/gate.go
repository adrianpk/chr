package chcapp

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

func newGateCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "gate",
		Usage: "Run project promotion gates",
		Commands: []*cli.Command{
			newGateDevMainCommand(stdout, deps),
		},
	}
}

func newGateDevMainCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "dev-main",
		Usage: "Collect dev to main promotion evidence",
		Flags: gateDevMainFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 0 {
				return fmt.Errorf("gate dev-main does not accept positional arguments")
			}

			return fmt.Errorf("gate dev-main is not implemented")
		},
	}
}

func gateDevMainFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
		&cli.StringFlag{
			Name:     "base",
			Usage:    "base ref for promotion evidence",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "head",
			Usage:    "head ref for promotion evidence",
			Required: true,
		},
	)
}
