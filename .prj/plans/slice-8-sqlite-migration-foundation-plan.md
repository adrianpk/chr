# Slice 8 SQLite Migration Foundation Plan

Status: active

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`
Tracker: `.prj/trackers/slice-8-sqlite-migration-foundation-tracker.md`

## Purpose

Add the SQLite migration foundation required by the future project registry and global ticket index.

This slice creates the embedded migration mechanism and the initial relational `projects` table. It does not add project registration commands yet.

## Scope

This slice covers:

- selecting and wiring a SQLite driver,
- embedding SQLite migration assets,
- parsing migration files with `-- +migrate Up` and `-- +migrate Down` sections,
- applying pending migrations in deterministic order,
- recording applied migrations in `schema_migrations`,
- creating the initial `projects` table through migration,
- focused tests for migration parsing and application.

## Task Breakdown

### T8.1 SQLite Driver and Migration Asset

Add the SQLite driver dependency and the first embedded SQLite migration.

Expected behavior:

- use a pure-Go SQLite driver to avoid CGO requirements for the CLI binary,
- add migration asset path `assets/migration/sqlite`,
- create `schema_migrations` and `projects` through embedded SQL.

### T8.2 Migrator Core Primitive

Implement the core migrator in `internal/chc`.

Expected behavior:

- parse embedded migration files,
- split up/down sections,
- apply migrations in filename order,
- skip already applied versions,
- stop and return errors on failed migration application.

### T8.3 Validation and Documentation

Document the migration layout and validate the slice.

Expected behavior:

- document embedded SQLite migration behavior,
- update tracker,
- run `make verify`.

## Out of Scope

This slice does not cover:

- project registration commands,
- project registry repository APIs,
- ticket indexing,
- daemon command,
- filesystem watcher,
- Telegram integration.

## Validation

Required checks:

```sh
make verify
```
