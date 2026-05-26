# SQLite Migrations

## Purpose

ChrysoCoord uses embedded SQLite migrations for user-level database schema setup.

The global user database is derived local state. It supports project registration, ticket indexing, daemon queries, and messaging workflows.

## Layout

SQLite migrations live under the embedded asset tree:

```text
cmd/chc/assets/migration/sqlite/
```

Migration filenames must sort in application order:

```text
<version>-<name>.sql
```

Example:

```text
20260518000001-create_projects.sql
```

## Sections

Migration files use explicit up/down markers:

```sql
-- +migrate Up
CREATE TABLE projects (...);

-- +migrate Down
DROP TABLE IF EXISTS projects;
```

The migrator applies only the `Up` section.

## Tracking

The migrator owns the `schema_migrations` table and creates it before applying embedded migrations.

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version text PRIMARY KEY,
    name text NOT NULL,
    applied_at text NOT NULL
);
```

Applied versions are skipped on later runs.

## Board Ticket Metadata

Board ticket metadata extends the existing `indexed_tickets` table. It does not
introduce a separate Board read model.

The Board migration adds:

- `synopsis`;
- `board_column`;
- `source_kind`;
- `source_ref`;
- `slice`;
- Board query indexes for column, use case, project area, updated timestamp,
  and source kind;
- `ticket_key_sequences` for transaction-safe ticket key allocation foundation.

Markdown ticket files remain the durable ticket body. SQLite stores searchable
and selectable metadata for operational views.
