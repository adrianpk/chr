# Board Ticket Operations Model

Status: initial

Spec: `.prj/specs/specs/board-ticket-operations.md`
Source quiz: `.prj/quizzes/curated/dashboard-board-ticket-operations-quiz.md`

## Existing Models

The current implementation already has these relevant concepts:

- `Ticket` renders Markdown frontmatter and body.
- `IndexedTicket` stores searchable ticket metadata in SQLite.
- `TicketRepository` creates Markdown tickets.
- `TicketIndex` persists indexed ticket metadata in `chr.db`.
- `TicketReconciler` refreshes SQLite rows from Markdown tickets.

The Board MVP should evolve these models instead of introducing an unrelated
ticket store.

## Ticket Metadata

The ticket Markdown frontmatter should support:

| Field | Required | Semantics |
| --- | --- | --- |
| `key` | yes | Human-facing ticket id such as `CHR-123`. |
| `title` | yes | Short ticket title. |
| `synopsis` | yes for new chr-created tickets | One short operational summary for Board rows and detail. |
| `status` | yes | Board lifecycle status. |
| `priority` | yes for Board-created tickets | Scheduling importance. |
| `severity` | yes for Board-created tickets | Impact level. |
| `type` | yes for Board-created tickets | Work category. |
| `tags` | no | Free-form grouping values. |
| `use_case` | no | Use-case grouping; empty means project-wide. |
| `project_area` | no | Technical or product area. |
| `source_kind` | no | Origin such as `manual`, `plan_tracking`, `spec`, or `sync`. |
| `source_ref` | no | Primary source path or identifier. |
| `slice` | no | Slice identifier when generated from a plan. |
| `created_at` | yes | Creation timestamp. |
| `updated_at` | yes | Last metadata/body mutation timestamp. |

Markdown remains the durable body store for description, checklist,
acceptance criteria, comments, blockers, validation, and execution notes.

## Board Status

MVP Board statuses:

| Status | UI Label | Semantics |
| --- | --- | --- |
| `backlog` | Backlog | Captured but not yet ready for implementation. |
| `ready_for_dev` | Ready for dev | Ready for an agent or operator to implement. |
| `in_progress` | In progress | Implementation is active. |
| `ready_for_review` | Ready for review | Implementation exists and needs review. |
| `ready_for_merge` | Ready for merge | Review accepted and merge is the next step. |
| `done` | Done | Integrated or otherwise complete. |

Legacy status aliases should be explicit:

| Legacy Status | Preferred Board Status |
| --- | --- |
| `open` | `backlog` |
| `ready` | `ready_for_dev` |
| `reviewing` | `ready_for_review` |
| `solved` | `ready_for_merge` |
| `merged` | `done` |
| `closed` | `done` when completed, otherwise attention-worthy until clarified |
| `blocked` | attention-worthy; normal column can remain the current stored status or map to backlog with attention |

Unknown statuses are not silently folded into backlog. They must appear through
attention.

## Attention Reason

An attention reason is derived metadata. It should be stored or returned in a
plain shape that can be rendered as text:

| Field | Semantics |
| --- | --- |
| `kind` | Stable reason code such as `invalid_status` or `validation_failed`. |
| `message` | Human-readable reason. |
| `source_path` | Optional Markdown or evidence path. |
| `detected_at` | Optional detection timestamp. |

MVP reason kinds:

- `blocked`
- `agent_comment`
- `mention`
- `validation_failed`
- `invalid_status`
- `invalid_category`

The first implementation may compute reasons during Board query rather than
persisting a separate table if the SQLite row already has enough metadata. The
schema should not block adding a dedicated attention table later.

## Indexed Ticket Row

The current `indexed_tickets` table should be extended or replaced through
migrations to support Board metadata:

| Column | Semantics |
| --- | --- |
| `project_id` | Project id. |
| `path` | Absolute Markdown ticket path. |
| `relative_path` | Path relative to the ticket root. |
| `key` | Human-facing key. |
| `title` | Ticket title. |
| `synopsis` | Short Board summary. |
| `status` | Stored status from Markdown. |
| `board_column` | Normalized Board column when valid. |
| `priority` | Priority. |
| `severity` | Severity. |
| `type` | Ticket type. |
| `use_case` | Use-case grouping. |
| `project_area` | Project area grouping. |
| `source_kind` | Origin category. |
| `source_ref` | Primary source reference. |
| `slice` | Generated slice identifier when present. |
| `created_at` | Ticket creation timestamp. |
| `updated_at` | Ticket update timestamp. |
| `file_mtime` | Source file modification time. |
| `file_size` | Source file size. |
| `content_hash` | Source file hash. |
| `indexed_at` | Index refresh timestamp. |

Recommended indexes:

- unique `(project_id, key)`;
- query `(project_id, board_column, priority, severity)`;
- query `(project_id, status)`;
- query `(project_id, use_case)`;
- query `(project_id, project_area)`;
- query `(project_id, type)`;
- query `(project_id, updated_at)`;
- query `(project_id, source_kind)`.

## Ticket Key Allocation

`CHR-XXX` remains the default human-facing key shape.

The MVP should move allocation toward SQLite-backed monotonic numbering so new
ticket creation does not depend only on recursive file scans. Allocation must
be transaction-safe inside one local process.

The existing configurable key prefix should remain supported.

## Userspace

MVP active userspace:

```text
.chr/default/
```

The active operator id is resolved from configuration or environment, but
artifact writes use the single active userspace path for now.

Future multi-operator shape:

```text
.chr/operators/<operator-id>/
```

This future model is not part of the first Board implementation, but code that
resolves userspace paths should avoid hard-coding assumptions that make the
future split expensive.

## Persistence Ownership

The current database lives in the user environment data directory. Board MVP
needs a project-local metadata index decision because the product direction is
to treat SQLite as operational project state, not only disposable user cache.

The implementation should resolve this deliberately:

- preserve existing user-level DB behavior where required by existing commands;
- add or adapt a project-local DB under `.chr/default/` if needed for Board state;
- keep Markdown as the authoritative body source;
- ensure index rebuild can recover Board metadata from Markdown when possible.

## Validation Rules

- New chr-created tickets require non-empty title and synopsis.
- New Board-created tickets default to `backlog`.
- New generated tickets must include source metadata when generated from plan
  tracking.
- Unknown statuses produce attention.
- Empty synopsis on externally supplied Markdown is not an attention condition
  by itself.
- Non-empty synopsis text is never overwritten by sync enrichment.

## Open Questions

- Should the first migration extend `indexed_tickets` directly or introduce a
  dedicated `board_tickets` read model?
- Should attention reasons be persisted immediately or computed from ticket
  metadata and Markdown summaries?
- Should `blocked` become a normal status later or remain an attention overlay?
- Should project-local DB state replace user-level DB state for all ticket
  indexes or only Board-specific indexes?
