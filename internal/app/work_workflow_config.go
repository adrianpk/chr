package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) workConfigCommand() *cli.Command {
	flags := append(workProjectFlags(),
		&cli.BoolFlag{Name: "json", Usage: "print effective workflow values and sources as JSON"},
		&cli.StringFlag{Name: "review-restriction", Usage: "observe the human_review invocation restriction"},
	)

	return &cli.Command{Name: "config", Usage: "Observe effective workflow configuration without upgrading the project", Flags: flags, Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() != 0 {
			return fmt.Errorf("work config does not accept positional arguments")
		}

		operations, _, err := factory.workServiceFromCommand(ctx, cmd)
		if err != nil {
			return err
		}

		result, err := operations.WorkflowConfiguration(ctx, core.ReviewWorkflowMode(cmd.String("review-restriction")))
		if err != nil {
			return err
		}

		if cmd.Bool("json") {
			return json.NewEncoder(factory.io.Stdout).Encode(result)
		}

		return renderWorkWorkflowConfig(factory.io.Stdout, result)
	}}
}

func renderWorkWorkflowConfig(stdout io.Writer, result core.WorkflowConfiguration) error {
	for _, value := range result.Values {
		if _, err := fmt.Fprintf(stdout, "%s: %s (source: %s)\n", value.Key, value.Value, value.Source); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(stdout, "Configuration revision: %s\n", result.Revision)

	return err
}
