# Inspect and Recover Local State

Use this guide when local `chc` state looks inconsistent or when you need a
single bundle for handoff.

Maintenance commands are conservative. Status and preview commands are
read-only. Repair commands require explicit project names or paths.

## Write a Diagnostic Bundle

Start with a bundle when you need a compact local snapshot:

```sh
chc maintenance bundle
```

For a specific project:

```sh
chc maintenance bundle --dir /path/to/project
```

The bundle is written to:

```text
.chc/maintenance/diagnostic-bundle.md
```

It includes database status, registry diagnostics, execution artifact drift,
runtime cleanup preview, and report metadata. It omits workspace config
contents, environment values, credential files, and token values.

## Inspect Database Migration State

Check the local SQLite migration state:

```sh
chc maintenance db status
```

Use this when a command fails while opening local user state or when the local
database may have been created by a different binary.

Review these fields first:

- `database_exists`
- `schema_migrations_exists`
- `pending`
- `unknown`

Unknown migrations usually mean the database was migrated by a newer binary.

## Inspect Project Registry Drift

List registered projects and their local path state:

```sh
chc maintenance registry status
```

Use this when project selection fails, a project moved, or a registered path was
deleted.

Important statuses:

- `ok`: the project is usable;
- `disabled`: the registry row is disabled;
- `missing_path`: the registered directory is gone;
- `missing_workspace`: the directory exists but `.chc/` is missing.

## Repair Registry Entries

Disable a stale project registration:

```sh
chc maintenance registry disable example
```

Preview the same action:

```sh
chc maintenance registry disable --dry-run example
```

Update a moved project path:

```sh
chc maintenance registry set-path example /new/project/path
```

Preview the path update:

```sh
chc maintenance registry set-path --dry-run example /new/project/path
```

Registry repair commands update only the registry row. They do not edit tickets,
execution records, runtime artifacts, or active project selection.

## Inspect Execution Artifact Drift

Check whether execution records still point to existing runtime evidence:

```sh
chc maintenance execution artifacts --dir /path/to/project
```

The command reports missing run directories, result artifacts, runtime logs,
validation logs, and execution reports.

Use this before relying on historical execution evidence or before preparing a
handoff for a damaged core.

## Preview Runtime Artifact Cleanup

Preview unreferenced runtime artifact directories:

```sh
chc maintenance execution cleanup-preview --dir /path/to/project
```

This command does not delete anything. It reports leaf directories under the
runtime artifact directory that are not referenced by current execution records.

Review the output before any future destructive cleanup workflow.

## Find Existing Reports

List execution and loop reports:

```sh
chc report list --dir /path/to/project
```

Filter by kind when needed:

```sh
chc report list --dir /path/to/project --kind execution
chc report list --dir /path/to/project --kind loop
```

Show a specific report:

```sh
chc report show --dir /path/to/project --execution <execution-id>
chc report show --dir /path/to/project --loop <loop-id>
```

## Suggested Order

For general local recovery:

1. Run `chc maintenance bundle`.
2. Run `chc maintenance registry status`.
3. Repair missing or moved registry entries if needed.
4. Run `chc maintenance execution artifacts`.
5. Run `chc maintenance execution cleanup-preview`.
6. Inspect reports with `chc report list` and `chc report show`.

Use `--dev-root` with maintenance commands when inspecting isolated development
state.
