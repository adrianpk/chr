package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func newWorkTrackerCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	command := &cli.Command{Name: "tracker", Usage: "Export, inspect or repair a derived workflow view"}
	for _, action := range []string{"export", "status", "repair"} {
		flags := append(workProjectFlags(), &cli.BoolFlag{Name: "json", Usage: "print projection provenance and staleness"})
		if action != "status" {
			flags = append(flags, &cli.StringFlag{Name: "replace-hash", Usage: "explicit permission to replace exactly these observed output bytes"})
		}

		command.Commands = append(command.Commands, &cli.Command{Name: action, ArgsUsage: "<slice-set-id>", Flags: flags, Action: func(ctx context.Context, cmd *cli.Command) error {
			return workTrackerAction(ctx, cmd, stdout, deps)
		}})
	}

	return command
}

func workTrackerAction(ctx context.Context, cmd *cli.Command, stdout io.Writer, deps appDependencies) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("work tracker %s requires exactly one adopted slice-set ID", cmd.Name)
	}

	operations, _, err := workServiceFromCommand(ctx, cmd, deps)
	if err != nil {
		return err
	}

	var result core.WorkflowTrackerStatus
	if cmd.Name == "status" {
		result, err = operations.InspectWorkflowTracker(ctx, cmd.Args().First())
	} else {
		result, err = operations.ExportWorkflowTracker(ctx, cmd.Args().First(), core.WorkflowTrackerOptions{ReplaceHash: cmd.String("replace-hash")})
	}

	if result.SetID == "" {
		return err
	}

	return errors.Join(err, printWorkflowTracker(stdout, result, cmd.Bool("json")))
}

func printWorkflowTracker(stdout io.Writer, result core.WorkflowTrackerStatus, asJSON bool) error {
	if asJSON {
		return json.NewEncoder(stdout).Encode(result)
	}

	_, err := fmt.Fprintf(stdout, "Tracker: %s\nProjection: %s\nReason: %s\nObserved output hash: %s\n%s\n", result.Path, result.Status, result.Reason, result.ObservedHash, result.Detail)

	return err
}
