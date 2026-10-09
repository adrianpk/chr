package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

type workWorkflowCLI struct {
	stdout io.Writer
}

func newWorkExtentCommand(stdout io.Writer, deps appDependencies, name, extent string) *cli.Command {
	argsUsage := "<identity>"
	if extent == "loose" {
		argsUsage = " "
	}

	return &cli.Command{Name: name, Usage: "Run the explicit " + name + " extent through shared workflow operations", ArgsUsage: argsUsage, Flags: workWorkflowExtentFlags(extent == "loose"), Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := workWorkflowExtentInput(cmd, extent == "loose"); err != nil {
			return err
		}

		operations, _, err := workServiceFromCommand(ctx, cmd, deps)
		if err != nil {
			return err
		}

		return (workWorkflowCLI{stdout: stdout}).extent(ctx, cmd, operations, extent)
	}}
}

func workWorkflowExtentFlags(loose bool) []cli.Flag {
	flags := append(workProjectFlags(),
		&cli.BoolFlag{Name: "preview", Aliases: []string{"dry-run"}, Usage: "observe the selection without admission or execution"},
		&cli.BoolFlag{Name: "json", Usage: "print the bounded operation result as JSON"},
		&cli.IntFlag{Name: "offset", Usage: "preview member offset"},
		&cli.IntFlag{Name: "limit", Value: core.WorkflowPageLimit, Usage: "preview page size, 1..128"},
		&cli.StringFlag{Name: "snapshot", Usage: "exact preview snapshot for pagination or admission"},
		&cli.IntFlag{Name: "max-steps", Usage: "bounded continuation steps, 0 uses the retained limit"},
		&cli.BoolFlag{Name: "human-review", Usage: "require exact human approval for this invocation"},
	)
	if !loose {
		return flags
	}

	for _, name := range []string{"status", "priority", "severity", "project-area", "tag", "type", "use-case"} {
		flags = append(flags, &cli.StringSliceFlag{Name: name, Usage: "loose ticket " + name + " filter"})
	}

	return append(flags, &cli.StringFlag{Name: "order", Usage: "loose ticket rank order"}, &cli.StringFlag{Name: "min-priority", Usage: "minimum classified priority"}, &cli.StringFlag{Name: "min-severity", Usage: "minimum classified severity"})
}

func workWorkflowArguments(cmd *cli.Command, loose bool) error {
	if loose {
		if cmd.NArg() != 0 {
			return fmt.Errorf("work loose does not accept an identity")
		}

		return nil
	}

	identity := cmd.Args().First()
	if cmd.NArg() != 1 || len(identity) < 1 || len(identity) > 128 || strings.TrimSpace(identity) != identity || strings.ContainsAny(identity, "\x00\r\n") {
		return fmt.Errorf("work %s requires one identity of 1..128 plain bytes", cmd.Name)
	}

	return nil
}

func (adapter workWorkflowCLI) extent(ctx context.Context, cmd *cli.Command, operations service, extent string) error {
	if err := workWorkflowExtentInput(cmd, extent == "loose"); err != nil {
		return err
	}

	request := core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: extent, Identity: cmd.Args().First()}, Offset: cmd.Int("offset"), Limit: cmd.Int("limit"), SnapshotHash: cmd.String("snapshot")}
	if extent == "loose" {
		selection := agentWorkSelectionOptionsFromCommand(cmd)
		request.Statuses, request.Filters, request.Order = selection.Statuses, selection.Filters, selection.Order
		request.MinPriority, request.MinSeverity = cmd.String("min-priority"), cmd.String("min-severity")
	}

	if cmd.Bool("preview") {
		preview, err := operations.PreviewWorkExtent(ctx, request)
		if err != nil {
			return err
		}

		return adapter.preview(cmd, preview)
	}

	restriction := core.ReviewWorkflowMode("")
	if cmd.Bool("human-review") {
		restriction = core.ReviewWorkflowModeHumanReview
	}

	result, err := operations.StartWorkExtent(ctx, request, cmd.Int("max-steps"), restriction)
	if result.Workflow != nil {
		err = errors.Join(err, adapter.status(ctx, cmd, operations, result.Workflow.ID, cmd.Name))
	}

	return err
}

func workWorkflowExtentInput(cmd *cli.Command, loose bool) error {
	if err := workWorkflowArguments(cmd, loose); err != nil {
		return err
	}

	if err := workWorkflowPageInput(cmd.Int("offset"), cmd.Int("limit")); err != nil {
		return err
	}

	if !cmd.Bool("preview") && cmd.Int("offset") != 0 {
		return fmt.Errorf("execution requires offset zero")
	}

	if err := workWorkflowStepsInput(cmd.Int("max-steps")); err != nil {
		return err
	}

	if hash := cmd.String("snapshot"); hash != "" {
		return workWorkflowHashInput("snapshot", hash)
	}

	return nil
}

func workWorkflowPageInput(offset, limit int) error {
	if offset < 0 || offset > core.WorkflowRecordLimit || limit < 1 || limit > core.WorkflowPageLimit {
		return fmt.Errorf("workflow page bounds exceeded")
	}

	return nil
}

func workWorkflowStepsInput(steps int) error {
	if steps < 0 || steps > core.WorkflowPageLimit {
		return fmt.Errorf("max-steps must contain 0..128 checkpoint steps")
	}

	return nil
}

func workWorkflowHashInput(kind, hash string) error {
	if len(hash) != 64 {
		return fmt.Errorf("%s requires the exact 64-character binding hash", kind)
	}

	if _, err := hex.DecodeString(hash); err != nil {
		return fmt.Errorf("%s requires a hexadecimal binding hash", kind)
	}

	return nil
}

func newWorkControlCommand(stdout io.Writer, deps appDependencies, action string) *cli.Command {
	return &cli.Command{Name: action, Usage: "Observe or " + action + " one retained workflow invocation", ArgsUsage: "<execution-id>", DisableSliceFlagSeparator: true, Flags: workWorkflowControlFlags(action), Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := workWorkflowControlInput(cmd, action); err != nil {
			return err
		}

		operations, _, err := workServiceFromCommand(ctx, cmd, deps)
		if err != nil {
			return err
		}

		return (workWorkflowCLI{stdout: stdout}).control(ctx, cmd, operations, action)
	}}
}

func workWorkflowControlFlags(action string) []cli.Flag {
	flags := append(workProjectFlags(), &cli.BoolFlag{Name: "json", Usage: "print bounded workflow status as JSON"}, &cli.BoolFlag{Name: "preview", Aliases: []string{"dry-run"}, Usage: "observe retained inputs without changing state"})

	switch action {
	case "status":
		return append(flags, &cli.IntFlag{Name: "offset", Usage: "retained member offset"}, &cli.IntFlag{Name: "limit", Value: core.WorkflowPageLimit, Usage: "status page size, 1..128"}, &cli.IntFlag{Name: "revision", Usage: "observed invocation revision for consistent pagination"})
	case "continue":
		return append(flags, &cli.IntFlag{Name: "max-steps", Usage: "bounded steps within the frozen invocation limit"})
	case "approve":
		return append(flags, &cli.StringFlag{Name: "binding", Usage: "exact displayed approval hash"})
	case "reconcile":
		return append(flags, &cli.StringFlag{Name: "kind", Value: "ownership", Usage: "ownership, membership, submission or manual"}, &cli.StringSliceFlag{Name: "harness", Usage: "manual harness argv; repeat once per argument"})
	default:
		return flags
	}
}

func workWorkflowControlInput(cmd *cli.Command, action string) error {
	if err := workWorkflowArguments(cmd, false); err != nil {
		return err
	}

	switch action {
	case "approve":
		return workWorkflowHashInput("approval", cmd.String("binding"))
	case "continue":
		return workWorkflowStepsInput(cmd.Int("max-steps"))
	case "status":
		if cmd.Int("revision") < 0 {
			return fmt.Errorf("revision must be nonnegative")
		}

		return workWorkflowPageInput(cmd.Int("offset"), cmd.Int("limit"))
	case "reconcile":
		return workWorkflowReconcileInput(cmd)
	}

	return nil
}

func workWorkflowReconcileInput(cmd *cli.Command) error {
	kind, harness := cmd.String("kind"), cmd.StringSlice("harness")
	switch kind {
	case "ownership", "membership", "submission":
		if len(harness) != 0 {
			return fmt.Errorf("harness argv requires manual reconciliation")
		}
	case "manual":
		return validateWorkflowManualHarness(harness)
	default:
		return fmt.Errorf("unsupported reconciliation kind")
	}

	return nil
}

func (adapter workWorkflowCLI) control(ctx context.Context, cmd *cli.Command, operations service, action string) error {
	if err := workWorkflowControlInput(cmd, action); err != nil {
		return err
	}

	id := cmd.Args().First()
	if cmd.Bool("preview") || action == "status" {
		return adapter.status(ctx, cmd, operations, id, action)
	}

	err := performWorkflowControl(ctx, operations, id, action, cmd)

	return errors.Join(err, adapter.status(ctx, cmd, operations, id, action))
}

func performWorkflowControl(ctx context.Context, operations service, id, action string, cmd *cli.Command) error {
	switch action {
	case "continue":
		_, err := operations.ContinueWorkExtent(ctx, id, cmd.Int("max-steps"))

		return err
	case "pause", "cancel":
		_, err := operations.StopWorkflow(ctx, id, action)

		return err
	case "resume":
		_, err := operations.ResumeWorkflow(ctx, id)

		return err
	case "approve":
		_, err := operations.ApproveWorkflowSlice(ctx, id, cmd.String("binding"))

		return err
	case "reconcile":
		return performWorkflowReconciliation(ctx, operations, id, cmd)
	default:
		return fmt.Errorf("unsupported workflow control")
	}
}

func performWorkflowReconciliation(ctx context.Context, operations service, id string, cmd *cli.Command) error {
	switch cmd.String("kind") {
	case "ownership":
		_, err := operations.ReconcileWorkflow(ctx, id)

		return err
	case "membership":
		_, err := operations.ReconcileWorkflowMembership(ctx, id)

		return err
	case "submission":
		_, err := operations.ReconcileWorkflowSubmission(ctx, id)

		return err
	case "manual":
		_, err := operations.RunWorkflowManual(ctx, id, cmd.StringSlice("harness"))

		return err
	default:
		return fmt.Errorf("unsupported reconciliation kind")
	}
}
