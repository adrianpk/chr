# Ticket Template Model

Status: initial

Spec: `.prj/specs/specs/ticket-template.md`
Source quiz: `.prj/quizzes/curated/ticket-template-model-quiz.md`

## Ticket

The ticket is a Markdown document with YAML frontmatter and a compact body.

## Fields

| Field | Required | Source | Semantics |
| --- | --- | --- | --- |
| `key` | yes | generated | Stable ticket identifier. |
| `title` | yes | user or command input | Human-readable ticket title. |
| `status` | yes | default/generated | Current lifecycle state. Defaults to `open`. |
| `priority` | no | user or command input | Scheduling importance. |
| `severity` | no | user or command input | Impact level. |
| `type` | no | user or command input | Work category. |
| `tags` | no | user or command input | Free-form grouping values. |
| `use_case` | no | user or command input | Feature or flow grouping. Empty or absent means project-wise. |
| `project_area` | no | user or command input | Transversal concern or technical area. |
| `created_at` | yes | generated | Creation timestamp. |
| `updated_at` | yes | generated | Last update timestamp. |

## Status Values

| Value | Semantics |
| --- | --- |
| `open` | Created but not yet ready or started. |
| `ready` | Ready to be picked up. |
| `in_progress` | Work is actively happening. |
| `reviewing` | Work is waiting for review. |
| `blocked` | Work cannot continue without external input or action. |
| `solved` | A solution exists. |
| `merged` | The solution was integrated. |
| `closed` | The ticket is closed without necessarily being solved. |

## Priority Values

```text
low
normal
high
critical
```

## Severity Values

```text
low
normal
high
critical
```

## Type Values

```text
task
bug
feature
cleanup
research
docs
```

## Invariants

- Status is stored in metadata, not in the path.
- `labels` is not part of the initial model; use `tags`.
- Empty or absent `use_case` means project-wise work.
- Generated metadata must not create manual ceremony for ticket creation.
- Body sections may be empty.

## Open Questions

- Which fields should the parser reject as missing?
- Should empty strings and absent optional fields be treated identically?
- What exact timestamp format should generated fields use?
