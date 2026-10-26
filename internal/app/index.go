package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) indexCommand() *cli.Command {
	return &cli.Command{
		Name:  "index",
		Usage: "Manage derived project indexes",
		Commands: []*cli.Command{
			factory.indexRebuildCommand(),
			factory.indexStatusCommand(),
		},
	}
}

func (factory commandFactory) indexRebuildCommand() *cli.Command {
	return &cli.Command{
		Name:   "rebuild",
		Usage:  "Rebuild the resolved project's ticket index",
		Flags:  ticketProjectFlags(),
		Action: factory.runIndexRebuild,
	}
}

func (factory commandFactory) indexStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report the resolved project's ticket index status",
		Flags:  ticketProjectFlags(),
		Action: factory.runIndexStatus,
	}
}

func (factory commandFactory) runIndexRebuild(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index rebuild does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := core.NewTicketIndex(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := core.NewTicketIndexRebuilder(index).RebuildProjectTickets(ctx, core.RebuildProjectTicketsOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return printIndexRebuildResult(factory.io.Stdout, result)
}

func printIndexRebuildResult(stdout io.Writer, result core.RebuildProjectTicketsResult) error {
	_, err := fmt.Fprintf(stdout, "indexed: %d\n", result.IndexedTickets)
	if err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "parse_errors: %d\n", len(result.ParseErrors))
	if err != nil {
		return fmt.Errorf("write index rebuild result: %w", err)
	}

	for _, parseErr := range result.ParseErrors {
		_, err = fmt.Fprintf(stdout, "parse_error: %s %v\n", parseErr.RelativePath, parseErr.Err)
		if err != nil {
			return fmt.Errorf("write index rebuild result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) runIndexStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("index status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := core.NewTicketIndex(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	status, err := core.NewTicketIndexStatusChecker(index).CheckProjectTicketIndex(ctx, core.CheckProjectTicketIndexOptions{
		ProjectID:  projectContext.ProjectID,
		ProjectDir: projectContext.ProjectDir,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	err = printIndexStatusResult(factory.io.Stdout, status)
	if err != nil {
		return err
	}

	if ticketIndexHasDrift(status) {
		return fmt.Errorf("ticket index is not current")
	}

	return nil
}

func printIndexStatusResult(stdout io.Writer, status core.TicketIndexStatus) error {
	err := printIndexStatusCounts(stdout, status)
	if err != nil {
		return err
	}

	err = printIndexStatusDetails(stdout, status)
	if err != nil {
		return err
	}

	return nil
}

func printIndexStatusCounts(stdout io.Writer, status core.TicketIndexStatus) error {
	_, err := fmt.Fprintf(stdout, "missing: %d\n", len(status.Missing))
	if err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "stale: %d\n", len(status.Stale))
	if err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "orphaned: %d\n", len(status.Orphaned))
	if err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "invalid: %d\n", len(status.Invalid))
	if err != nil {
		return fmt.Errorf("write index status result: %w", err)
	}

	return nil
}

func printIndexStatusDetails(stdout io.Writer, status core.TicketIndexStatus) error {
	for _, missing := range status.Missing {
		_, err := fmt.Fprintf(stdout, "missing_ticket: %s\n", missing.RelativePath)
		if err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, stale := range status.Stale {
		_, err := fmt.Fprintf(stdout, "stale_ticket: %s\n", stale.Parsed.RelativePath)
		if err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, orphaned := range status.Orphaned {
		_, err := fmt.Fprintf(stdout, "orphaned_ticket: %s\n", orphaned.RelativePath)
		if err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	for _, invalid := range status.Invalid {
		_, err := fmt.Fprintf(stdout, "invalid_ticket: %s %v\n", invalid.RelativePath, invalid.Err)
		if err != nil {
			return fmt.Errorf("write index status result: %w", err)
		}
	}

	return nil
}

func ticketIndexHasDrift(status core.TicketIndexStatus) bool {
	return len(status.Missing) > 0 ||
		len(status.Stale) > 0 ||
		len(status.Orphaned) > 0 ||
		len(status.Invalid) > 0
}
