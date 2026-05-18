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

## Open Questions

- Exact SQL schema is not yet defined.
- Exact project id generation is not yet defined.
- Exact content hash algorithm is not yet defined.
