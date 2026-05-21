# Authoring Sessions

Authoring sessions are project-local Markdown workflows for turning rough
knowledge into durable artifacts such as quizzes, drafts, specs, plans, slice
sets, and tickets.

Markdown remains the canonical editing surface. The CLI creates, lists, shows,
and promotes files, but users may edit generated Markdown directly.

## Command Surface

Create a session:

```sh
chc author session new --target spec --topic "Billing sync"
```

List sessions:

```sh
chc author session list
```

Show a session:

```sh
chc author session show <session-id>
```

Create a static quiz:

```sh
chc author quiz new --target spec --topic "Billing sync"
```

Preview a static quiz without writing files:

```sh
chc author quiz new --target spec --topic "Billing sync" --dry-run
```

Promote a quiz to a draft:

```sh
chc author promote quiz-to-draft --session <session-id> --quiz .chc/quizzes/<quiz>.md
```

Preview promotion without writing the draft or updating the session:

```sh
chc author promote quiz-to-draft --session <session-id> --quiz .chc/quizzes/<quiz>.md --dry-run
```

All commands accept `--dir` to target another managed project.

## Session Layout

Sessions are stored under:

```text
.chc/authoring/sessions/<session-id>/
  session.md
  transcript.md
  artifacts.md
```

`session.md` stores session metadata, input artifacts, output artifacts, and
provenance links. `transcript.md` and `artifacts.md` are durable sibling files
reserved for later richer authoring flows.

## Artifact Paths

Generated authoring artifacts use project-local `.chc/` paths:

```text
.chc/dumps/
.chc/quizzes/
.chc/drafts/
.chc/specs/
.chc/plans/
.chc/slice-sets/
.chc/slices/
.chc/trackers/
.chc/tickets/
```

The current generated filename shape is:

```text
<session-id>-<topic-slug>.md
```

Commands refuse to replace an existing generated artifact unless
`--allow-replace` is supplied.

## Session Fields

The session model contains:

- `id`: stable authoring session identifier.
- `project_id`: resolved project identity.
- `status`: `draft`, `active`, `finalized`, or `discarded`.
- `mode`: `markdown_direct`, `static`, `assisted`, or `tui`.
- `target`: `dump`, `quiz`, `draft`, `spec`, `plan`, `slice_set`, `slice`,
  `tracker`, or `ticket`.
- `topic`: human topic for the authoring work.
- `created_at`, `started_at`, `finished_at`, `updated_at`: session times.
- `description`: optional session description.

Current CLI-created sessions use static mode.

## Provenance

Provenance links record how one artifact produced another.

Each link records:

- relation: `seeded_by`, `generated`, `promoted`, or `refined`;
- source artifact;
- target artifact;
- decision text;
- recorded time.

Static quiz generation records a `generated` link. Quiz-to-draft promotion
records a `promoted` link.

## Static Quizzes

Static quizzes are deterministic Markdown questionnaires. They include
frontmatter, a topic, target-specific questions, and editable `Answer:` slots.

Static quizzes do not use a model. They are useful for starting structured
context capture while keeping the result reviewable and hand-editable.

## Manual Promotion

The current promotion path is:

```text
quiz -> draft
```

Promotion reads the selected quiz file, creates an editable draft, and updates
the session with the new output artifact and provenance link.

Promotion does not synthesize or summarize answers with an LLM. It preserves the
quiz content inside the draft so the user can continue editing from a durable
Markdown artifact.

## Boundaries

Authoring sessions prepare context and documents. They do not execute tickets
and do not run the Docker worker.

Agent-assisted authoring and a TUI are future surfaces over the same session
model. They are not required for the current Markdown-first workflow.
