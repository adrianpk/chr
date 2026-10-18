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
journal v1 reads are bounded to 1 MiB. Capacity and provider failures preserve the
previous canonical session and artifacts.

## Dump Mutation API

`ReadDumpDocument` resolves a retained dump reference and returns its artifact,
full document content and a semantic `Revision`. The revision binds the project
root, resolved reference and exact body/metadata hashes. Inspection follows the
same read-only alias and pending-journal rules as project dump observation.

`MutateDump` accepts `DumpMutationRequest` with `OperationID`, `Action`,
`Namespace`, `Reference`, `ExpectedRevision`, `Title`, `Content` and `CaseSlug`:

| Action | Destination and revision |
| --- | --- |
| `capture`, general | Leave namespace, reference and case empty. Resolve the existing project destination or allocate the first pair in `default`. An optional expected revision binds the complete project dump observation. |
| `capture`, named | Supply an existing case slug and its selected namespace. Allocate a distinct case capture; an optional expected revision binds that case's metadata source hash. |
| `append` | Supply the document reference and revision from `ReadDumpDocument`; extend that observed document. |
| `save` | Supply the document reference and revision from `ReadDumpDocument`; replace the editable body, retaining the original frontmatter. |

Capture requires a title and at most 64 KiB of content. Titles are at most 512
bytes, case/namespace identities 128 bytes and revisions 64 bytes. Append and
Save require nonempty content of at most 1 MiB. The final document and each raw
component remain bounded to 1 MiB, including retained frontmatter. Raw components
cannot be edited as canonical dumps. Scoped edits validate the existing case.
Old alias references target the resolved canonical document; their original
superseded bodies remain unchanged.

An optional `DumpCapturePreparer` receives the resolved body and raw references
and returns the prepared documents. Their scope and raw/curated relationships
must match the retained operation. Without a preparer, the owner retains the
capture text directly. Later general captures append the prepared editable body
to the same canonical document. Every capture retains its raw component and
source hashes; every replacement retains the prior exact body and metadata in
immutable artifact history. Body saves preserve unknown metadata and do not
change frontmatter ownership or provenance.

The existing authoring request directory retains a
`chr-project-dump-request/v1` binding containing `Schema`, `Digest`, `Checksum`,
`Request`, `Destination`, `RawReference`, optional `Prepared` documents and
`Attempts`. Its encoded size is bounded to 8 MiB. The digest binds the project
root and original arguments. The owner reserves arguments and destination
before preparation, retains a successful preparation before publication and
allows at most five preparation attempts across restarts. Each preparation has
a 60-second timeout; the complete mutation, including owner acquisition, has
a 90-second timeout. New bindings and receipts use the existing authoring
capacity limits. Reusing an identity for changed arguments or a different
resolved destination conflicts, including after preparation failure.

Publication uses the same journal v2 and immutable result receipt as
reconciliation. A first general capture publishes its body, raw component,
metadata, project reference and receipt together. A retry recovers the journal
or returns the existing receipt before inspecting later source changes or
calling the preparer. A receipt never overwrites later independent edits.

The central application service routes `CreateDump`, `AppendDump` and `SaveDump`
through this owner. Requests carry `OperationID` and `ExpectedRevision`; edits
require the revision from `ReadDump`. Named captures retain their selected ops
namespace. The service exposes reconciliation preview/apply through the same
boundary. Dump catalogs inventory all namespaces from canonical files, retain
independent scoped documents and hide raw components and superseded aliases.
Catalog inspection has a ten-second deadline. Display update time uses artifact
metadata's update time, falling back to creation time when unavailable.

Successful service writes return `Committed`, the immutable mutation result and
its body/metadata hashes. If a later index/catalog refresh fails, they also return
`ProjectionStale` and a pending-refresh message. The post-commit refresh has a
ten-second deadline. An identical retry reuses the
receipt and retries the projection. Returned content/revision belongs to that
committed image only when it still matches the canonical body/metadata hashes.

CLI writes accept `--operation-id`. `chr dump show --revision <path>` prints the
document revision for `dump save` or `dump append`; retain it with `--revision`
and the original operation ID when retrying an edit. A fresh edit without a
supplied ID observes the document before dispatch. A retained edit ID requires
its original revision. For `dump new`, an optional revision binds the project
dump observation or selected case metadata, as in the table above.

Web details and TUI editors retain the document revision and operation identity
for Save. TUI text changes allocate a new edit identity while keeping the
observed source guard. Web mutations compare revisions in the owning operation
after receipt lookup, without a separate stale-source preflight. Confirmed Web
results expose a committed operation and any pending catalog refresh.

## Bound Authoring Dump Outputs

Bound authoring continuations resolve general dump outputs through the same
project-wide destination as ordinary captures. The first output uses the default
namespace or adopts the sole existing general identity. Later sessions append
contributions to that identity, retaining each session's original raw input in
its own namespace. Explicit refinement replaces the observed editable body and
retains its previous body and metadata in immutable history.

The continuation prepares dump, history, project-reference, session, transcript,
provenance and receipt images before publishing any of them. One v2 authoring
journal owns the complete change set. Other artifact targets retain their v1
journal format. Reopening recovers the complete transaction, and the existing
continuation receipt returns the original result before checking later edits or
invoking a provider again.

Session snapshots resolve retained dump aliases across namespaces. Their source
hash includes the alias images, resolved body and metadata, and project metadata.
Providers and native work descriptions receive the resolved body. The session's
namespace still owns its raw capture, quizzes, specifications, plans and proposed
native records; a project-wide dump reference does not move those records.
A changed shared dump requires a fresh session observation before continuation.
The owner also rechecks the project dump inventory after provider preparation.

A question-only creation may retain raw input before producing its first dump.
Its bound continuation can publish that output only when every orphaned raw
component is one of its exact observed inputs. Unrelated raw-only inventories,
multiple general candidates and conflicting project references still require
explicit reconciliation.

## Project Dump Reconciliation

Project metadata at `.chr/default/project.md` supports an optional
`project_dump_ref` pointing to `ops/<namespace>/dump/<id>/artifact.md`. Its
namespace belongs to the artifact identity, independently of `OPS_USER`.
`ObserveProjectDump` reads all namespaces without adopting a destination or
recovering a pending journal. General dumps, raw components, scoped dumps and
retained aliases have distinct identities. Inspection rejects unsafe paths,
inconsistent envelopes and unresolved aliases. Ordinary project metadata saves
preserve the reference; `AdoptProjectDump` binds a sole existing general dump
under authoring exclusion after comparing the observation revision.

`PreviewProjectDumpReconciliation` returns `Revision`, `Destination`, `Content`
and ordered `Sources` for 2–32 general dumps. It proposes the earliest valid
creation timestamp, using relative path to break ties. Sources with missing or
invalid timestamps follow valid evidence in path order, retaining an explicit
`TimestampIssue`; if all timestamps are invalid, path determines the proposal.
The composed document retains the chosen target's frontmatter and includes each
source body with its reference and source hash. Composition does not call a
provider, deduplicate or summarize material.

`ReconcileProjectDump` requires the preview's `ExpectedRevision`, exact proposed
`Destination` and a stable `OperationID` of 1–128 letters, digits, dots,
underscores or hyphens, beginning with a letter or digit. It rechecks all source
images under the shared authoring owner. Each source's original body and
metadata are retained byte for byte under its envelope at
`files/history/<image-hash>/artifact.md` and `metadata.json`. The image hash binds
the original body and metadata hashes. An existing different history image is a
conflict. Superseded bodies stay unchanged; their metadata gains `canonical_ref`
and `dump_history_ref`. Target metadata retains source hashes, history references,
the observed revision, operation and original target hash in
`dump_reconciliation`. Unknown metadata and target frontmatter survive.

The existing `.chr/authoring/pending.json` journal accepts
`chr-authoring-journal/v2` for these mutations. Its checksum and before/after
image rules are unchanged. Version 2 permits the project metadata file and the
dump history paths above in addition to v1 targets. It carries at most 128
changes and 24 MiB of encoded data; individual files remain bounded to 1 MiB.
Version 1 retains its original target, 12-change and 1 MiB limits. Recovery reads
both versions and rejects divergent images before replay. No separate dump
journal or recovery owner exists.

The same journal publishes the composed body, preserved images, alias metadata,
project reference and `chr-project-dump-receipt/v1` receipt. Receipts use the
existing authoring receipt directory and retain `Schema`, `Digest`, `Checksum`
and `Result`; the result holds operation ID, destination and committed body and
metadata hashes. The digest binds the project root, operation, revision and
destination. An identical retry returns that immutable result after recovery;
changed arguments conflict. Receipt identity is shared with authoring requests,
so another operation cannot reuse an already occupied identity.
An existing pending authoring request also reserves its identity. New dump
operations require fewer than 4,096 existing entries in the shared receipt
directory; confirmed retries still replay at capacity.

`ResolveProjectDumpReference` follows at most 16 alias links across namespaces,
rejecting cycles, missing targets and aliases into raw or scoped artifacts.
Observation excludes superseded aliases from its general candidate list. Reads
report a pending journal without recovering it. Inventory is bounded to 4,096
entries and 8 MiB of body/metadata images, composed bodies to 1 MiB, inspection to
10 seconds and reconciliation to 30 seconds. These APIs do not apply a migration
to user data merely by reading it.

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
authorization references. `/prepare-work [conversation]` retains one exact native
preview before dispatch. An unchanged request returns its prior receipt or
pending prompt. A resolver proposal alone cannot create development authority.
The explicit `/develop` source retains an opaque instruction reference and can
bind only the conversation's first work reservation. Migration 21 pins that
reservation, including failed attempts, so later requirements cannot silently
acquire another scope from the same instruction.

Missing or changed authority uses the existing retained action owner. The prompt
shows the original instruction, canonical description reference and hashes, with
bounded excerpts when needed. Its `develop` choice accepts a button, exact reply
or `/answer <token> develop`, binding the first retained answer identity to the
prepared scope. Duplicate sources reuse that decision; expiry never extends on
replay. Conversation phases project awaiting confirmation and preparation for
that bound request without changing its immutable conversation revision. A later
canonical refinement advances the revision and invalidates the earlier prompt.

`ApplyPreparedAuthoringWorkGuarded` rechecks current host authority under the
authoring writer lock before a new native effect. Confirmed native receipts
reconcile first, preserving later edits. The record envelope retains actual scope
and authority; no work admission or integration policy is inferred from success.

`ReadTelegramConversationInstruction` recovers the original retained source for
the exact authorized peer and project binding. Reply invalidation and focus do
not replace that source. An operator remapping clears its usable development
reference; an unavailable source requires a new explicit instruction.
