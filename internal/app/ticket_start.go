package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) ticketStartCommand() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Prepare one ticket execution",
		Flags:  ticketStartFlags(),
		Action: factory.runTicketStart,
	}
}

func ticketStartFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the execution preflight without creating an execution record",
		},
	)
}

func (factory commandFactory) runTicketStart(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket start requires exactly one ticket key argument")
	}

	if handled, err := factory.runTicketStartAuthority(ctx, cmd); handled || err != nil {
		return err
	}

	preflight, err := factory.buildTicketStartPreflight(ctx, cmd)
	if err != nil {
		return core.NewExecutionFailure(core.ExecutionFailurePreflight, err)
	}

	if !core.RuntimeValidationPassed(preflight.Runtime.Checks) {
		preflight.Record.Status = core.ExecutionStatusBlocked
		preflight.Record.FailureReason = core.NewExecutionFailure(
			core.ExecutionFailureRuntime,
			fmt.Errorf("runtime validation failed"),
		).Error()

		err = printTicketStartResult(factory.io.Stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
		if err != nil {
			return err
		}

		return core.NewExecutionFailure(core.ExecutionFailureRuntime, fmt.Errorf("runtime validation failed"))
	}

	if !cmd.Bool("dry-run") {
		record, err := runOwnedTicketStart(ctx, factory.io.Stdout, preflight)
		if err != nil {
			return err
		}

		preflight.Record = record
	}

	return printTicketStartResult(factory.io.Stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
}

func runAndPrintTicketStartExecution(ctx context.Context, stdout io.Writer, preflight ticketStartPreflight) (core.ExecutionRecord, error) {
	record, err := newService(ctx, preflight.Env, preflight.ProjectContext, preflight.AssetsFS).runTicketStartExecution(ctx, ticketExecutionInput{Record: preflight.Record, Runtime: preflight.Runtime, Events: ticketExecutionEvents{
		Sink: agentWorkExecutionEventSink(preflight.Record),
	}})
	if err == nil {
		return record, nil
	}

	printErr := printTicketStartResult(stdout, preflight.Ticket, record, preflight.Runtime, false)
	if printErr != nil {
		return core.ExecutionRecord{}, printErr
	}

	return record, err
}

func printTicketStartResult(stdout io.Writer, ticket core.IndexedTicket, record core.ExecutionRecord, result core.RuntimeValidationResult, dryRun bool) error {
	lines := []string{
		fmt.Sprintf("execution: %s", record.ID),
		fmt.Sprintf("ticket: %s", ticket.Key),
		fmt.Sprintf("ticket_path: %s", ticket.Path),
		fmt.Sprintf("ticket_relative_path: %s", ticket.RelativePath),
		fmt.Sprintf("status: %s", record.Status),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket start result: %w", err)
		}
	}

	err := printTicketStartFailure(stdout, record)
	if err != nil {
		return err
	}

	return printRuntimeValidateResult(stdout, result, dryRun)
}

func printTicketStartFailure(stdout io.Writer, record core.ExecutionRecord) error {
	if record.FailureReason == "" {
		return nil
	}

	kind := core.ExecutionFailureKindFromReason(record.FailureReason)
	if kind != "" {
		_, err := fmt.Fprintf(stdout, "failure_kind: %s\n", kind)
		if err != nil {
			return fmt.Errorf("write ticket start failure kind: %w", err)
		}
	}

	_, err := fmt.Fprintf(stdout, "failure_reason: %s\n", record.FailureReason)
	if err != nil {
		return fmt.Errorf("write ticket start failure reason: %w", err)
	}

	return printTicketStartDiagnostics(stdout, record)
}

func printTicketStartDiagnostics(stdout io.Writer, record core.ExecutionRecord) error {
	diagnostics := []struct {
		name string
		path string
	}{
		{name: "diagnostic_log_path", path: record.Artifacts.LogPath},
		{name: "diagnostic_result_path", path: record.Artifacts.ResultPath},
		{name: "diagnostic_validation_path", path: record.Artifacts.ValidationPath},
	}

	for _, diagnostic := range diagnostics {
		if diagnostic.path == "" {
			continue
		}

		_, err := fmt.Fprintf(stdout, "%s: %s\n", diagnostic.name, diagnostic.path)
		if err != nil {
			return fmt.Errorf("write ticket start diagnostic path: %w", err)
		}
	}

	return nil
}

func runOwnedTicketStart(ctx context.Context, stdout io.Writer, preflight ticketStartPreflight) (core.ExecutionRecord, error) {
	lock, err := acquireAgentOperationLock(ctx, preflight.ProjectContext, core.AgentOperationDevelopment)
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	defer releaseAgentOperationLock(lock)()

	if err := core.CheckLegacyWorkTicket(preflight.Ticket.Path); err != nil {
		return core.ExecutionRecord{}, err
	}

	if err := newService(ctx, preflight.Env, preflight.ProjectContext, preflight.AssetsFS).checkLegacyAgentWorkKey(preflight.Ticket.Key); err != nil {
		return core.ExecutionRecord{}, err
	}

	return runAndPrintTicketStartExecution(ctx, stdout, preflight)
}
