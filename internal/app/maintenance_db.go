package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) maintenanceDBCommand() *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Inspect local database state",
		Commands: []*cli.Command{
			factory.maintenanceDBStatusCommand(),
		},
	}
}

func (factory commandFactory) maintenanceDBStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report local database migration status",
		Flags:  projectRegistryFlags(),
		Action: factory.runMaintenanceDBStatus,
	}
}

func (factory commandFactory) runMaintenanceDBStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance db status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	status, err := core.NewMigrationStatusChecker(factory.deps.assetsFS).Check(ctx, env)
	if err != nil {
		return err
	}

	return printMigrationStatus(factory.io.Stdout, status)
}

func printMigrationStatus(stdout io.Writer, status core.MigrationStatus) error {
	lines := []string{
		fmt.Sprintf("database: %s", status.DatabasePath),
		fmt.Sprintf("database_exists: %t", status.DatabaseExists),
		fmt.Sprintf("schema_migrations_exists: %t", status.SchemaTableExists),
		fmt.Sprintf("available: %d", len(status.Available)),
		fmt.Sprintf("applied: %d", len(status.Applied)),
		fmt.Sprintf("pending: %d", len(status.Pending)),
		fmt.Sprintf("unknown: %d", len(status.Unknown)),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write migration status: %w", err)
		}
	}

	err := printAppliedMigrations(stdout, status.Applied)
	if err != nil {
		return err
	}

	err = printPendingMigrations(stdout, status.Pending)
	if err != nil {
		return err
	}

	return printUnknownMigrations(stdout, status.Unknown)
}

func printAppliedMigrations(stdout io.Writer, migrations []core.AppliedMigration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(
			stdout,
			"applied_migration: %s %s %s\n",
			migration.Version,
			migration.Name,
			migration.AppliedAt,
		)
		if err != nil {
			return fmt.Errorf("write applied migration: %w", err)
		}
	}

	return nil
}

func printPendingMigrations(stdout io.Writer, migrations []core.Migration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(stdout, "pending_migration: %s %s\n", migration.Version, migration.Name)
		if err != nil {
			return fmt.Errorf("write pending migration: %w", err)
		}
	}

	return nil
}

func printUnknownMigrations(stdout io.Writer, migrations []core.AppliedMigration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(
			stdout,
			"unknown_migration: %s %s %s\n",
			migration.Version,
			migration.Name,
			migration.AppliedAt,
		)
		if err != nil {
			return fmt.Errorf("write unknown migration: %w", err)
		}
	}

	return nil
}
