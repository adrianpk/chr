# Authoring Sessions

Authoring sessions are project-local Markdown workflows for turning rough
knowledge into durable artifacts such as quizzes, specs, plans, slice sets, and
tickets.

Markdown remains the canonical editing surface. The CLI creates, lists, and
shows sessions and artifacts. It also keeps a legacy quiz-to-draft command for
older sessions, but users may edit generated Markdown directly.

## Command Surface

Create a session:

```sh
chr author session new --target spec --topic "Billing sync"
```

List sessions:

```sh
chr author session list
```

The list output includes headers:

```text
id                       status  target  updated_at            topic
author-20260522-073505   draft   spec    2026-05-22T07:35:05Z  Billing sync
```

Show a session:

```sh
chr author session show <session-handle>
```

The session handle may be the exact session id, the exact topic, or a unique
topic slug such as `billing-sync`. Existing `auth-...` session ids remain valid.

Create a context-aware quiz:

```sh
chr author quiz new --target spec --topic "Billing sync"
```

Successful creation prints the quiz path relative to the managed project:

```text
quiz: created .chr/default/quizzes/author-20260522-073505-billing-sync.md
session: author-20260522-073505
```

Preview a context-aware quiz without writing files:

```sh
chr author quiz new --target spec --topic "Billing sync" --dry-run
```

All commands accept `--dir` to target another managed project.

Start an assisted session:

```sh
chr author assisted start --target spec --topic "Billing sync"
```

Ask the next assisted question with the current fake adapter path:

```sh
chr author assisted resume \
  --session <session-handle> \
  --fake-question "What should the spec include?"
```

Submit one answer:

```sh
chr author assisted answer \
  --session <session-handle> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id question-20260521-120000 \
  --question "What should the spec include?" \
  --answer "It should include retry behavior."
```

Show assisted transcript status:

```sh
chr author assisted status --session <session-handle>
```

Finalize into a curated summary and synthesized artifact:

```sh
chr author assisted finalize \
  --session <session-handle> \
  --fake-summary "Billing sync needs retry behavior and idempotent writes." \
  --target-kind spec \
  --target-name "billing sync spec"
```

## Session Layout

Sessions are stored under:

```text
.chr/authoring/sessions/<session-id>/
  session.md
  transcript.md
  artifacts.md
  curated-summary.md
  model/<request-id>/
    request.md
    response.md
```

`session.md` stores session metadata, input artifacts, output artifacts, and
provenance links. `transcript.md` stores assisted authoring question, answer,
result, and state events as Markdown blocks with frontmatter. `artifacts.md`
remains the session artifact registry. `curated-summary.md` is written during
assisted finalization.

The `model/<request-id>/` directory records model-boundary evidence. Each
invocation stores the request context and the response that drove the next
authoring step.

## Artifact Paths

Generated authoring artifacts use project-local `.chr/` paths:

```text
.chr/default/dumps/
.chr/default/quizzes/
.chr/default/specs/
.chr/default/plans/
.chr/default/slice-sets/
.chr/default/slices/
.chr/default/trackers/
.chr/default/tickets/
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
- `mode`: `markdown_direct`, `contextual`, `assisted`, or `tui`.
- `target`: `dump`, `quiz`, `spec`, `plan`, `slice_set`, `slice`,
  `tracker`, or `ticket`.
- `topic`: human topic for the authoring work.
- `created_at`, `started_at`, `finished_at`, `updated_at`: session times.
- `description`: optional session description.

Quiz CLI sessions use contextual mode.
Assisted CLI sessions use `active` status and `assisted` mode.
New CLI-created sessions use the `author-<timestamp>` id prefix. Older
`auth-<timestamp>` sessions remain readable and usable by exact id.

## Provenance

Provenance links record how one artifact produced another.

Each link records:

- relation: `seeded_by`, `generated`, `promoted`, or `refined`;
- source artifact;
- target artifact;
- decision text;
- recorded time.

Context-aware quiz generation records a `generated` link. Assisted synthesis
records `generated` links. The legacy quiz-to-draft command records a
`promoted` link for compatibility with older sessions.

Assisted finalization records generated links from the transcript to the
curated summary and from the curated summary to the synthesized artifact.

## Context-Aware Quizzes

Context-aware quizzes are generated from a structured request that includes the
authoring session, project context sections, the target artifact, the topic,
and an output contract.

The model response must be structured and validated before Chryso writes
the Markdown quiz artifact. Generic template questions are not used as a
fallback.

The request sent to the quiz provider includes:

- action: `authoring_quiz_generation`;
- strategy: `deep`;
- session id, target, topic, and requested question count;
- context sections for the project, authoring session, agent protocol,
  `.chr/` artifacts, docs, project tree, and Go inventory;
- an output contract for question objects.

The response must contain a `questions` array. Each question must include:

- `id`;
- `question`;
- `tags`;
- `follow_up_depth`;
- `completion_recommended`.

If the provider response is missing required structure, asks for sensitive
material, or yields no valid questions, the command fails instead of writing a
generic fallback quiz.

## Legacy Quiz-To-Draft

The current product model does not use `draft` as a first-class artifact
family. Roughness is represented by a spec with `status: draft`.

The intended manual document flow is:

```text
dump
quiz
dump + quiz -> spec(status=draft)
spec(status=draft) -> spec(status=accepted/final)
```

Legacy `quiz-to-draft` remains a compatibility path for older sessions. New
flows should generate or edit specs instead of creating draft artifacts.

## Assisted Authoring

Assisted authoring is a CLI-driven host-side workflow over the same Markdown
session model. It prepares documents. It does not execute tickets, edit source
code, or run the Docker worker.

The current command path is intentionally fake-adapter backed. `resume` accepts
`--fake-question`, and `finalize` accepts `--fake-summary`. This keeps the
workflow deterministic while preserving the model-boundary contracts and
artifacts that a real provider adapter will use later.

The assisted workflow has four durable stages:

1. `start` creates an active assisted session.
2. `resume` builds explicit authoring context, asks the adapter for the next
   question, records request and response artifacts, and appends transcript
   events.
3. `answer` appends the user's answer as transcript events.
4. `finalize` curates answers into `curated-summary.md`, synthesizes a target
   Markdown artifact, and updates session outputs and provenance.

Assisted context includes:

- default authoring analyst identity;
- project and session metadata;
- applicable `AGENTS.md` files;
- selected `.chr/` protocol files;
- selected session documents;
- transcript summary and unanswered gaps;
- source inventory.

Implicit conversation history is not included. Relevant prior discussion must be
captured in Markdown files or transcript events.

## Assisted Artifact Paths

Assisted sessions write these files:

```text
.chr/authoring/sessions/<session-id>/transcript.md
.chr/authoring/sessions/<session-id>/model/<request-id>/request.md
.chr/authoring/sessions/<session-id>/model/<request-id>/response.md
.chr/authoring/sessions/<session-id>/curated-summary.md
.chr/<target-kind>/<session-id>-<target-name>.md
```

The synthesized artifact is editable Markdown. Existing target artifacts are not
replaced unless `--allow-replace` is supplied to `finalize`.

## Model Boundary

Provider details are behind the internal assisted model adapter boundary. The
CLI currently exposes deterministic fake inputs, but the persisted artifacts are
the same shape expected by a real adapter:

- request ID, session ID, turn ID, intent, target, prompt, and context;
- response ID, provider, status, question or blocked result, raw text, and
  timestamp.

Secrets and credentials are not written into request, response, transcript, or
summary artifacts.

## Boundaries

Authoring sessions prepare context and documents. They do not execute tickets
and do not run the Docker worker.

The TUI remains a future surface over the same session model.
