# Maintenance

Maintenance commands inspect and recover local `chc` state.

They are intended for local diagnostics. They should report exact state before
any repair or cleanup command mutates files or database rows.

## Database Migration Status

Inspect local SQLite migration state:

```sh
chc maintenance db status
```

Use an isolated development environment:

```sh
chc maintenance db status --dev-root /tmp/chc-dev
```

The command compares embedded migration assets with the local
`schema_migrations` table.

It does not apply migrations. It does not create the database when the database
file is missing.

Output fields:

- `database`: local SQLite database path;
- `database_exists`: whether the database file exists;
- `schema_migrations_exists`: whether the migration table exists;
- `available`: migration assets available to this binary;
- `applied`: migrations recorded in the database and known to this binary;
- `pending`: available migrations not recorded in the database;
- `unknown`: database migrations not known to this binary.

Detailed rows:

- `applied_migration: <version> <name> <applied_at>`;
- `pending_migration: <version> <name>`;
- `unknown_migration: <version> <name> <applied_at>`.

Unknown migrations usually mean the database was created or migrated by a newer
binary than the one currently running.
