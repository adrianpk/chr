package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) acceptanceCommand() *cli.Command {
	return &cli.Command{
		Name:  "acceptance",
		Usage: "Manage acceptance artifacts",
		Commands: []*cli.Command{
			factory.acceptanceFeatureCommand(),
		},
	}
}

func (factory commandFactory) acceptanceFeatureCommand() *cli.Command {
	return &cli.Command{
		Name:  "feature",
		Usage: "Manage acceptance feature artifacts",
		Commands: []*cli.Command{
			factory.acceptanceFeatureGenerateCommand(),
		},
	}
}

func (factory commandFactory) acceptanceFeatureGenerateCommand() *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate an acceptance feature",
		Flags:  workspaceScopeFlags(),
		Action: factory.runAcceptanceFeatureGenerate,
	}
}

func (factory commandFactory) runAcceptanceFeatureGenerate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("acceptance feature generate does not accept positional arguments")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GenerateAcceptanceFeature(ctx, AcceptanceFeatureGenerateRequest{
		Scope: workspaceScopeFromCommand(cmd),
	})
	if err != nil {
		return err
	}

	return writeAcceptanceFeatureGenerateResult(factory.io.Stdout,
		result)
}

func writeAcceptanceFeatureGenerateResult(stdout io.Writer, result AcceptanceFeatureOperationResult) error {
	_, err := fmt.Fprintf(stdout, "%s\nstatus: %s\n", result.Message, result.Status)
	if err != nil {
		return fmt.Errorf("write acceptance feature generate result: %w", err)
	}

	if result.Status == core.AcceptanceFeatureStatusGenerated {
		_, err = fmt.Fprintf(stdout, "feature: %s\n", result.Feature)
		if err != nil {
			return fmt.Errorf("write acceptance feature generate result: %w", err)
		}
	}

	if result.RejectionReason != "" {
		_, err = fmt.Fprintf(stdout, "reason: %s\n", result.RejectionReason)
		if err != nil {
			return fmt.Errorf("write acceptance feature generate result: %w", err)
		}
	}

	return nil
}
