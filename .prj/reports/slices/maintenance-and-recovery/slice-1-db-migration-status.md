# Slice 1: Database Migration Status

Status: reviewing
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-1-db-migration-status`
PR: #126

## Purpose

Expose local SQLite migration state through the CLI.

This slice lets users inspect whether the local user database is missing,
up-to-date, pending migrations, or contains migrations unknown to the current
binary.

## Delivered Behavior

- Added a migration status model.
- Compared embedded SQLite migration assets with `schema_migrations`.
- Added `chc maintenance db status`.
- Reported database path, database existence, migration table existence,
  available, applied, pending, and unknown migration counts.
- Added detailed applied, pending, and unknown migration rows.
- Documented the command.

## Implementation Notes

The status checker is read-only. It does not run migrations and does not create
the database when the database file is missing.

When the database is missing, every embedded migration is reported as pending.
When `schema_migrations` is missing in an existing database, every embedded
migration is also reported as pending.

## Contracts Added or Changed

- Added `MigrationStatus`.
- Added `AppliedMigration`.
- Added `MigrationStatusChecker`.
- Added `UserDatabasePath`.
- Added `chc maintenance db status`.
- Added `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chc/migration_status.go`
- `internal/chc/migration_status_test.go`
- `internal/chcapp/maintenance.go`
- `internal/chcapp/app.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chc -run 'TestMigrationStatus|TestMigrator|TestReadSQLiteMigrations'
go test ./internal/chcapp -run 'TestMaintenanceDBStatus'
make verify
```

All commands pass.

## Risks and Follow-ups

- The command reports migration state only. It does not repair or migrate the
  database.
- Project registry diagnostics remain deferred to Slice 2.
