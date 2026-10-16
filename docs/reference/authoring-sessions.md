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
quiz: created ops/default/quiz/<id>/artifact.mdauthor-20260522-073505-billing-sync.md
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
ops/default/dump/<id>/artifact.md
ops/default/quiz/<id>/artifact.md
ops/default/spec/<id>/artifact.md
ops/default/plan/<id>/artifact.md
.chr/default/slice-sets/
.chr/default/slices/
ops/default/tracker/<id>/artifact.md
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

## Bound Provider Continuations

The application operation service exposes `ObserveAuthoring` and
`ContinueAuthoring` for caller-neutral creation, questions, answers, generation
and refinement. These operations use the configured structured model provider;
they do not require the legacy fake question or summary flags shown above.

`AuthoringContinuationRequest` contains an opaque request/operator binding,
project, frozen namespace, session identity, action and target. Supported targets
are dump, quiz, spec and plan. Creation retains its original capture in a
canonical raw dump, preserving leading spaces and trailing blank lines. An
optional original instruction is retained separately, with a 16 KiB bound;
provider proposals cannot replace it. Subsequent operations require both the observed session
revision and a source hash covering the session, transcript, artifact bodies and
metadata. Refinement may replace only that session's observed output of the
requested kind. A new generation cannot silently replace an existing output.

Managed session frontmatter adds a positive `revision` and frozen `namespace`.
Their canonical artifacts use `ops/<namespace>/<kind>/<id>/artifact.md` and the
normal envelope metadata, files, evidence and producer-extension directories.
Request-derived session and artifact identities are reserved before provider
invocation. Specs remain drafts; preparation does not create work invocations or
change execution authority. Existing sessions remain readable through the
original repository and CLI surfaces; the bound continuation API does not adopt
legacy sessions implicitly.

A question binds the canonical session, turn, question ID, revision and content
digest. The owner assigns those identities. Answers must match the current
binding and observed session; another session's prompt, an answered prompt or a
prompt superseded by a session edit cannot accept the answer. Managed transcript
text uses quoted escaped values so multiline text and Markdown delimiters survive
reload without becoming transcript records. Answer text also preserves its
original whitespace. `answer` records an answer alone; `respond` records the
answer and requested generation or next clarification in one journal and receipt.

`ContinueAuthoringGuarded` accepts a host-owned `AuthoringCommitGuard`. The owner
calls it after generation and before committing canonical changes, while still
holding its writer lock. This lets an adapter recheck current authorization and
bindings after a provider call. Reconciliation of an existing receipt does not
invoke the provider or acquire new write authority.

The authoring repository serializes ordinary and bound writers through an
OS-backed owner lock. A bound operation journals its expected before images,
intended after images and immutable result receipt together. Recovery checks all
images before replay; divergent files preserve the journal and block mutation.
Replaying a confirmed request returns its original receipt even after later
independent edits. Reusing an ID with changed arguments conflicts, including
after provider failure.

Private recovery state is stored under `.chr/authoring/`:

```text
owner.lock
pending.json
requests/<opaque-request-hash>.json
receipts/<opaque-request-hash>.json
```

Each provider attempt is retained before invocation. Five attempts exhaust that
request's budget; retry deadlines and exhaustion survive reopening. The shared
structured provider limits calls to 60 seconds and two concurrent invocations.
Authoring also bounds captures to 64 KiB, instructions to 16 KiB, questions to
eight per session, canonical artifact references to 16, transcript/provenance
entries to 64, and retained request bindings to 4,096 per project. File and
journal reads are bounded to 1 MiB. Capacity and provider failures preserve the
previous canonical session and artifacts.

## Telegram Artifact Continuations

The receiver maps these commands to the same bound continuation service:

| Command | Effect |
| --- | --- |
| `/done [conversation]` | Generate a canonical dump from the retained capture. |
| `/quiz [conversation]` | Ask one provider-generated clarification question. |
| `/spec [conversation]` | Generate a draft specification, or ask for missing information. |
| `/plan [conversation]` | Generate a plan from observed canonical context. |
| `/refine <dump\|spec\|plan> <conversation> <instructions>` | Explicitly replace the selected session's observed output. |

Ordinary-language proposals use the same command registry and typed handlers.
Slash commands and replies to a retained authoring question bypass intent
interpretation. `/dump` remains a view of the captured conversation; it does not
invoke generation.

The receiver resolves an exact conversation, an eligible reply/focus selection,
or the sole eligible conversation. It freezes project, root, namespace and
conversation revision with the canonical session/source image. The complete
owner request is retained atomically with the transport reservation before
generation. Reopening dispatches that request rather than resolving the message
or preparing its source images again.

An actual provider question binds the canonical session, turn, question ID,
revision and content digest to the outgoing prompt. Reply to that prompt or send
`/reply <conversation> <answer>`. A clarification needed for a spec or plan
continues that generation after the answer. An explicit quiz answer is retained
without generating another artifact. Superseded prompts cannot answer the new
question.

After the shared owner confirms its receipt, `FinishTelegramAuthoring` commits
the conversation's session/question binding, revision, source acknowledgement
and outbox reply in one transaction. A lost acknowledgement reconciles the
original canonical receipt without regenerating artifacts. If the conversation
has advanced meanwhile, reconciliation reports the confirmed canonical session
without replacing its current binding. Current operator, project and root
authorization remain required. Neither artifact preparation nor its reply
creates a native work invocation.

## Explanation and Prepared Work

`ExplainAuthoring` answers from a canonical session's observed artifacts, or the
bound project's `README.md` when no session is selected. Context excerpts retain
their source references and hashes. A project overview must be a regular file
of at most 32 KiB; symlinks are rejected. Questions are at most 16 KiB and accepted
explanations at most 12,000 bytes. The shared provider enforces its normal
deadline, concurrency and payload bounds. This operation neither writes a new
authoring artifact nor prepares or admits work.

`PrepareAuthoringExplanation` freezes that bounded canonical context and the
original question without calling the provider. `ExplainPreparedAuthoring` uses
the retained image. `ExplainAuthoring` combines these operations for callers
that do not need a checkpoint between them.

Telegram exposes `/explain <question>`, with optional `--project <id>` or
`--conversation <id>` before the question. An exact conversation or valid reply
uses its canonical authoring session when available. Otherwise the selected or
sole allowed project supplies its overview. Ambiguous or conflicting targets
require correction. Ordinary-language explanation proposals use the same
registry and typed handler, with the original message as the question. A model
proposal cannot rewrite the user's question or turn it into preparation.

The receiver retains the complete prepared context before invoking the provider.
Explanation attempts and results belong to the transport operation separately
from the source's intent-resolution decision. Five retained attempts, backoff and
sanitized failure codes survive reopening. The owner uses OS exclusion, bounds
accepted result JSON to 64 KiB and includes it in logical transport capacity.
No database transaction spans the provider call.

A successful explanation is retained before source acknowledgement and outbox
creation. Restart recovers that result without another provider call or replacing
its source image with later edits. Current operator, root and conversation
bindings are checked before dispatch and before queuing the result. The project
service also checks its resolved root before reading context or entering the
provider. Explanations leave captures, authoring artifacts, native records and
execution authority unchanged.

`PrepareAuthoringWork` builds a preview of one native backlog ticket from an
observed generated dump or draft specification. It uses the existing record
preparation operation and reserves that ticket's identity. Callers must retain
the serialized `PreparedAuthoringWork` before dispatch rather than prepare it
again on retry. A simple ticket can use a generated dump directly.

The preview's `AuthoringWorkScope` retains the original user instruction,
preparation/development intent, project and namespace, canonical session revision,
complete source guard, description reference/hash, native contract hash and exact
prepared record image. Instructions are at most 8 KiB and descriptions 16 KiB.
An explanation cannot produce this preview. A pending canonical question or
stale source image blocks preparation.

`AuthorizeAuthoringWork` accepts a host-validated explicit decision for the
displayed scope hash and configured operator. The provider cannot supply this
authority. Preparation needs no development decision; development requires one.
Changing requirements, the prepared record or operator cannot reuse an old
decision. Internal artifact steps do not acquire separate execution authority.

`ApplyPreparedAuthoringWork` validates that preview under the authoring owner,
reserves immutable arguments with the native record owner and creates the
backlog ticket through that owner. The ticket's envelope retains `authoring_scope`
and `authoring_authorization` metadata. Its normal mutation receipt reconciles
lost acknowledgement before inspecting later edits. A failed dispatch cannot
exchange its request ID for a fresh description. This operation creates no
invocation, claims no work and changes no execution or integration policy.

Telegram conversation reads expose the canonical authoring session and retained
authorization references. A resolver proposal alone cannot create development
authority. The explicit `/develop` source retains an opaque instruction reference;
the adapter must bind that instruction or a new explicit decision to the
displayed scope before dispatching shared development preparation.

`ReadTelegramConversationInstruction` recovers the original retained source for
the exact authorized peer and project binding. Reply invalidation and focus do
not replace that source. An operator remapping clears its usable development
reference; an unavailable source requires a new explicit instruction.
