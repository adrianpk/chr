package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) workUpgradeCommand() *cli.Command {
	flags := append(workProjectFlags(),
		&cli.StringFlag{Name: "config-revision", Required: true, Usage: "unchanged revision observed by work config"},
		&cli.StringFlag{Name: "state-authority", Required: true, Usage: "explicit tickets or tracker_compat authority"},
		&cli.StringFlag{Name: "review-mode", Usage: "explicit new review choice; omission preserves prior policy"},
		&cli.StringFlag{Name: "adopt-file", Usage: "JSON envelope selecting one approved legacy plan and its exact task map"},
		&cli.BoolFlag{Name: "json", Usage: "print the prior-policy snapshot and upgrade result as JSON"},
	)

	return &cli.Command{Name: "upgrade", Usage: "Explicitly upgrade one quiescent legacy project to workflow version one", Flags: flags, Action: func(ctx context.Context, cmd *cli.Command) error {
		return factory.workUpgradeAction(ctx, cmd)
	}}
}

func (factory commandFactory) workUpgradeAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("work upgrade does not accept positional arguments")
	}

	adoption, err := readWorkflowUpgradeSelection(cmd.String("adopt-file"))
	if err != nil {
		return err
	}

	operations, _, err := factory.workServiceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	opts := core.WorkflowUpgradeOptions{ExpectedRevision: cmd.String("config-revision"), StateAuthority: core.WorkStateAuthority(cmd.String("state-authority")), ReviewMode: core.ReviewWorkflowMode(cmd.String("review-mode")), Adoption: adoption}

	result, err := operations.UpgradeWorkflowConfiguration(ctx, opts)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return json.NewEncoder(factory.io.Stdout).Encode(result)
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"Workflow version one enabled. Prior policy: %s\nConfiguration revision: %s\n", result.Snapshot, result.Current.Revision)

	return err
}

func readWorkflowUpgradeSelection(path string) (*core.WorkflowPlanAdoptionOptions, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(file, core.WorkflowFileLimit+1))

	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, err
	}

	if len(data) > core.WorkflowFileLimit {
		return nil, fmt.Errorf("workflow adoption envelope exceeds 1 MiB")
	}

	return decodeWorkflowUpgradeSelection(data)
}

func decodeWorkflowUpgradeSelection(data []byte) (*core.WorkflowPlanAdoptionOptions, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	selection := &core.WorkflowPlanAdoptionOptions{}
	if err := decoder.Decode(selection); err != nil {
		return nil, err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("workflow adoption envelope must contain one JSON object")
	}

	if err := core.ValidateGeneratedWorkflowPlan(selection.Plan); err != nil {
		return nil, err
	}

	return selection, nil
}
