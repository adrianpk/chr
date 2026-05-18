# User Daemon and Global Index Model

Status: initial

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`

## Workspace

Represents a project-local `.chc/` directory.

Fields:

| Field | Semantics |
| --- | --- |
| `project_dir` | Absolute path to the project root. |
| `workspace_dir` | Absolute path to `<project>/.chc`. |
| `config_path` | Absolute path to `<project>/.chc/config.yml`. |
| `tickets_dir` | Absolute path to `<project>/.chc/tickets`. |

## User Environment

Resolves user-level paths for config, data, cache, and state.

Fields:

| Field | Semantics |
| --- | --- |
| `config_dir` | User-level configuration directory. |
| `data_dir` | User-level durable application data directory. |
| `cache_dir` | User-level cache directory. |
| `state_dir` | User-level runtime or state directory. |
| `dev_mode` | Whether paths are resolved into a local development sandbox. |

## Registered Project

Represents a local project known by the user daemon.

Fields:

| Field | Semantics |
| --- | --- |
| `id` | Stable local project identity. |
| `name` | Human-readable project name. |
| `path` | Absolute project path. |
| `enabled` | Whether the daemon should observe the project. |
| `registered_at` | Timestamp when the project was registered locally. |
| `last_indexed_at` | Timestamp of last successful index refresh. |

Relational shape:

```sql
CREATE TABLE projects (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE,
    path text NOT NULL UNIQUE,
    enabled integer NOT NULL,
    registered_at text NOT NULL,
    last_indexed_at text
);
```

Rules:

- `id` is stable on the local machine.
- `name` is the human-facing identifier used by CLI and messaging surfaces.
- `path` points to the project root, not the `.chc/` directory.
- `enabled` controls whether the daemon should observe the project.
- `last_indexed_at` is operational metadata for the derived index.
- The row does not replace project-local `.chc/config.yml`.

## Migration

Represents one embedded database schema migration.

Fields:

| Field | Semantics |
| --- | --- |
| `version` | Ordered migration identifier derived from the filename prefix. |
| `name` | Human-readable migration name derived from the filename suffix. |
| `driver` | Database driver namespace, initially `sqlite`. |
| `up_sql` | SQL statements in the `-- +migrate Up` section. |
| `down_sql` | SQL statements in the `-- +migrate Down` section. |

Migration files use deterministic filenames:

```text
assets/migration/sqlite/<version>-<name>.sql
```

Migration files use explicit sections:

```sql
-- +migrate Up
...

-- +migrate Down
...
```

Applied migration tracking table:

```sql
CREATE TABLE schema_migrations (
    version text PRIMARY KEY,
    name text NOT NULL,
    applied_at text NOT NULL
);
```

Rules:

- Migrations are embedded assets.
- Migrations are applied in filename order.
- Already applied versions are skipped.
- Failed migrations stop execution and return an error.
- The migrator is a core `chc` primitive and must not require an external framework dependency.

## Indexed Ticket

Represents a ticket row derived from a Markdown file.

Fields:

| Field | Semantics |
| --- | --- |
| `project_id` | Registered project identity. |
| `path` | Absolute ticket file path. |
| `relative_path` | Ticket path relative to project root or `.chc/tickets`. |
| `key` | Ticket key from Markdown metadata. |
| `title` | Ticket title from Markdown metadata. |
| `status` | Ticket status from Markdown metadata. |
| `priority` | Ticket priority from Markdown metadata. |
| `severity` | Ticket severity from Markdown metadata. |
| `type` | Ticket type from Markdown metadata. |
| `use_case` | Ticket use case from Markdown metadata. |
| `project_area` | Ticket project area from Markdown metadata. |
| `file_mtime` | Modification time observed when indexed. |
| `file_size` | File size observed when indexed. |
| `content_hash` | Content hash observed when indexed. |
| `indexed_at` | Timestamp when the row was indexed. |

Tags may be modeled as a separate indexed table.

## Invariants

- Markdown ticket files are canonical.
- The global SQLite database is derived and rebuildable.
- The global SQLite database must not contain canonical ticket data that cannot be reconstructed from Markdown.
- A single `chc daemon` process may observe multiple registered projects.
- Project-local `.chc/` state and user-level `chc` state are separate concepts.
- Development mode must be able to redirect user-level paths away from real user config and data directories.
- The global database schema must be created through embedded migrations.
- The `projects` table is the relational source for registered project records.

## Open Questions

- Exact project id generation is not yet defined.
- Exact content hash algorithm is not yet defined.
- Exact SQLite Go driver is not yet selected.
