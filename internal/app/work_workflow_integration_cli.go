package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

func workWorkflowIntegrationFlags(action string) []cli.Flag {
	flags := []cli.Flag{
		&cli.Uint64Flag{Name: "revision", Usage: "exact displayed invocation revision"},
		&cli.StringFlag{Name: "entry-id", Usage: "displayed integration entry identity"},
		&cli.Uint64Flag{Name: "entry-revision", Usage: "displayed integration entry revision"},
		&cli.Uint64Flag{Name: "budget-revision", Usage: "displayed integration budget revision"},
		&cli.StringFlag{Name: "candidate-hash", Usage: "exact displayed integration candidate hash"},
	}
	if action == "approve" {
		return flags
	}

	flags = append(flags, &cli.StringFlag{Name: "record-id", Usage: "stable identity of this explicit operator response"}, &cli.StringFlag{Name: "actor", Usage: "operator identity"}, &cli.StringFlag{Name: "reason", Usage: "explicit reason for the action"})
	if action == "grant" {
		return append(flags, &cli.StringFlag{Name: "kind", Usage: "executor, reviewer, closing, merger, drift, submission or tokens"}, &cli.Int64Flag{Name: "amount", Usage: "positive allowance extension"})
	}

	flags = append(flags, &cli.Uint64Flag{Name: "decision-revision", Usage: "displayed question record revision"})
	for _, name := range []string{"decision-hash", "decision-id", "choice", "answer-action", "authority", "original-commit", "fix-category", "fix-title", "fix-detail"} {
		flags = append(flags, &cli.StringFlag{Name: name, Usage: "explicit decision " + name})
	}

	return flags
}

func workWorkflowIntegrationFields(cmd *cli.Command, action string) map[string]string {
	fields := map[string]string{
		"entry_id":        cmd.String("entry-id"),
		"entry_revision":  strconv.FormatUint(cmd.Uint64("entry-revision"), 10),
		"budget_revision": strconv.FormatUint(cmd.Uint64("budget-revision"), 10),
		"candidate_hash":  cmd.String("candidate-hash"),
	}
	if action == "approve" {
		return fields
	}

	for _, name := range []string{"record-id", "actor", "reason"} {
		fields[strings.ReplaceAll(name, "-", "_")] = cmd.String(name)
	}

	if action == "grant" {
		fields["kind"], fields["amount"] = cmd.String("kind"), strconv.FormatInt(cmd.Int64("amount"), 10)

		return fields
	}

	fields["decision_revision"] = strconv.FormatUint(cmd.Uint64("decision-revision"), 10)
	for _, name := range []string{"decision-hash", "decision-id", "choice", "answer-action", "authority", "original-commit", "fix-category", "fix-title", "fix-detail"} {
		fields[strings.ReplaceAll(name, "-", "_")] = cmd.String(name)
	}

	return fields
}

func workWorkflowIntegrationInput(cmd *cli.Command, action string) error {
	if cmd.Uint64("revision") == 0 {
		return fmt.Errorf("integration action requires the displayed invocation revision")
	}

	_, err := workflowIntegrationControlFields(action, workWorkflowIntegrationFields(cmd, action))

	return err
}

func performWorkflowOperatorControl(ctx context.Context, operations service, id, action string, cmd *cli.Command) error {
	switch action {
	case "approve":
		return performWorkflowApproval(ctx, operations, id, cmd)
	case "grant", "answer-decision":
		return performWorkflowIntegrationControl(ctx, operations, id, action, cmd)
	default:
		return fmt.Errorf("unsupported workflow control")
	}
}

func performWorkflowIntegrationControl(ctx context.Context, operations service, id, action string, cmd *cli.Command) error {
	if err := workWorkflowIntegrationInput(cmd, action); err != nil {
		return err
	}

	request, err := workflowIntegrationControlFields(action, workWorkflowIntegrationFields(cmd, action))
	if err != nil {
		return err
	}

	request.ID, request.ExpectedRevision, request.BindingHash = id, cmd.Uint64("revision"), cmd.String("binding")
	_, err = operations.ControlWorkExtent(ctx, request)

	return err
}

func performWorkflowApproval(ctx context.Context, operations service, id string, cmd *cli.Command) error {
	status, err := operations.WorkflowStatus(ctx, id, WorkflowStatusRequest{Limit: 1})
	if err != nil {
		return err
	}

	if status.Integration != nil {
		return performWorkflowIntegrationControl(ctx, operations, id, "approve", cmd)
	}

	var expected []uint64
	if cmd.Uint64("revision") != 0 {
		expected = append(expected, cmd.Uint64("revision"))
	}

	_, err = operations.ApproveWorkflowSlice(ctx, id, cmd.String("binding"), expected...)

	return err
}
