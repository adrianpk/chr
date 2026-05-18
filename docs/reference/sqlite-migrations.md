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
