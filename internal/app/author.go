package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

func (factory commandFactory) authorCommand() *cli.Command {
	return &cli.Command{
		Name:  "author",
		Usage: "Manage Markdown-first authoring sessions",
		Commands: []*cli.Command{
			factory.authorAssistedCommand(),
			factory.authorPromoteCommand(),
			factory.authorQuizCommand(),
			factory.authorSessionCommand(),
		},
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

func (factory commandFactory) authorProjectContext(ctx context.Context, cmd *cli.Command) (commandProjectContext, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return commandProjectContext{}, err
	}

	return newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
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
