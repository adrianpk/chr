# SQLite Migrations

## Purpose

Chryso uses embedded SQLite migrations for user-level database schema setup.

The global user database is derived local state. It supports project registration, ticket indexing, daemon queries, and messaging workflows.

## Layout

SQLite migrations live under the embedded asset tree:

```text
cmd/chr/assets/migration/sqlite/
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

The migration file is:

```text
20260526000008-add_board_ticket_metadata.sql
```

The Board uses the existing `indexed_tickets` read model instead of a dedicated
Board table. `ticket_key_sequences` stores monotonic ticket key allocation state
for generated `CHR-<number>` keys.

## Ticket Review Metadata

Ticket review metadata extends the existing `indexed_tickets` table. It does
not introduce provider-specific review tables in the first schema.

The review metadata migration adds:

- `branch`;
- `review_refs`;
- `review_state`;
- `review_round`;
- `review_result`;
- `review_comments_url`;
- `blocking_findings`;
- indexes for review state and branch lookups.

`review_refs` and `blocking_findings` are stored as JSON arrays in SQLite. The
canonical representation remains Markdown ticket frontmatter.

The migration file is:

```text
20260601000009-add_ticket_review_metadata.sql
```
