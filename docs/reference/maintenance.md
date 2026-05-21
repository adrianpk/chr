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

## Project Registry Status

Inspect registered project state:

```sh
chc maintenance registry status
```

Use an isolated development environment:

```sh
chc maintenance registry status --dev-root /tmp/chc-dev
```

The command is read-only. It does not select, disable, remove, or repair
projects.

Output starts with:

- `projects`: number of registered project rows.

Each project row includes:

- `project`: registered project name;
- `path`: registered project path;
- `enabled`: whether the registry row is enabled;
- `active`: whether the project is currently selected;
- `path_exists`: whether the registered path exists and is a directory;
- `workspace_exists`: whether `<path>/.chc` exists and is a directory;
- `status`: diagnostic status;
- `id`: registered project id.

Diagnostic statuses:

- `ok`: the project is enabled, the path exists, and `.chc/` exists;
- `disabled`: the project row is disabled;
- `missing_path`: the registered project path is missing or is not a directory;
- `missing_workspace`: the path exists but does not contain `.chc/`.

## Project Registry Repair

Disable a registered project:

```sh
chc maintenance registry disable example
```

Preview the same operation:

```sh
chc maintenance registry disable --dry-run example
```

Update a registered project path:

```sh
chc maintenance registry set-path example /path/to/project
```

Preview the path update:

```sh
chc maintenance registry set-path --dry-run example /path/to/project
```

Repair commands require the project name, not the project path or id.

`set-path` requires the target directory to contain `.chc/`. It does not create
or initialize a workspace.

Repair commands update only the project registry row. They do not mutate ticket
files, runtime artifacts, execution records, or the active project selection.

## Execution Artifact Drift

Inspect execution records for missing runtime evidence:

```sh
chc maintenance execution artifacts
```

Inspect a specific project directory:

```sh
chc maintenance execution artifacts --dir /path/to/project
```

Use an isolated development environment:

```sh
chc maintenance execution artifacts --dev-root /tmp/chc-dev --dir /path/to/project
```

The command is read-only. It does not create, rewrite, delete, or repair
runtime artifacts.

Output starts with:

- `executions`: number of execution records for the resolved project;
- `drifted`: number of execution records with missing artifacts;
- `missing_artifacts`: total missing artifact paths across all inspected
  execution records.

Each execution row includes:

- `execution`: execution id;
- `ticket`: ticket key;
- `status`: execution status;
- `artifact_status`: `ok` or `missing_artifacts`.

Each artifact row includes:

- `artifact`: execution id;
- `name`: artifact name;
- `exists`: whether the expected path exists with the expected file type;
- `path`: expected artifact path.

Inspected artifacts:

- `run_dir`: execution run directory;
- `result`: result JSON artifact;
- `log`: runtime log;
- `validation`: validation log;
- `report`: execution report Markdown derived from the run directory.
