package chcapp

import (
	"context"
	"io"
	"io/fs"

	"github.com/urfave/cli/v3"
)

var assetsFS fs.FS

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, commandAssetsFS fs.FS) error {
	assetsFS = commandAssetsFS
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
			newDaemonCommand(stdout),
			newIndexCommand(stdout),
			newInitCommand(stdout),
			newProjectCommand(stdout),
			newProvingCommand(stdout),
			newRuntimeCommand(stdout),
			newTicketCommand(stdout),
		},
	}
}
