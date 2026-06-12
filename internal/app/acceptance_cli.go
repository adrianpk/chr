package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/workspace"
)

func newAcceptanceCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "acceptance",
		Usage: "Manage acceptance artifacts",
		Commands: []*cli.Command{
			newAcceptanceFeatureCommand(stdout),
		},
	}
}

func newAcceptanceFeatureCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "feature",
		Usage: "Manage acceptance feature artifacts",
		Commands: []*cli.Command{
			newAcceptanceFeatureGenerateCommand(stdout),
		},
	}
}

func newAcceptanceFeatureGenerateCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "generate",
		Usage:  "Generate an acceptance feature",
		Flags:  workspaceScopeFlags(),
		Action: acceptanceFeatureGenerateAction(stdout),
	}
}

func acceptanceFeatureGenerateAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runAcceptanceFeatureGenerate(ctx, cmd, stdout)
	}
}

func runAcceptanceFeatureGenerate(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("acceptance feature generate does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.GenerateAcceptanceFeature(ctx, AcceptanceFeatureGenerateRequest{
		Scope: workspaceScopeFromCommand(cmd),
	})
	if err != nil {
		return err
	}

	return writeAcceptanceFeatureGenerateResult(stdout, result)
}

func writeAcceptanceFeatureGenerateResult(stdout io.Writer, result AcceptanceFeatureOperationResult) error {
	_, err := fmt.Fprintf(stdout, "%s\nstatus: %s\n", result.Message, result.Status)
	if err != nil {
		return fmt.Errorf("write acceptance feature generate result: %w", err)
	}

	if result.Status == workspace.AcceptanceFeatureStatusGenerated {
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
