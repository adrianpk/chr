# Slice 4: Conversational Authoring

Status: reviewing

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-4-telegram-authoring`

PR: [#393](https://forge.adrianpk.com/adrianpk/chr/pulls/393)

## Purpose

Deliver bounded conversational intent and configured-provider preparation through
the existing shared authoring and artifact owners.

## Delivered Behavior

T4.1 routes ordinary messages through a typed resolver whose available operations
come from the same registries as deterministic controls. Leading slash commands
bypass interpretation. Conversation and record dispatch use the existing typed
handlers, current authorization and shared mutation receipts. Unknown operations,
malformed decisions, missing/null fields, duplicate fields, unknown arguments and
ambiguous targets produce correction prompts without guessed effects.

The resolver receives bounded allowed-project names/IDs and peer-local session
references, excluding transport identities and hidden transcripts. It retains
its proposed operation, configured operator and frozen target before dispatch.
Root, namespace, conversation revision and current authorization are rechecked.
Reopening cannot select a second operation or redirect the frozen namespace.

T4.2 adds caller-neutral continuation operations for dump, quiz, spec and plan
creation, questions, answers, generation and refinement. The configured provider
receives canonical sources and retained answers. Its real questions are bound to
the session, turn, question ID, revision and content digest. Multiline questions
and answers survive transcript reload. Stale answers, changed source images and
replacement without an observed output fail without changing prior artifacts.

Generated documents use the existing renderers and canonical ops envelopes, with
session/source provenance and a frozen namespace. Confirmed requests replay their
original receipts after independent edits. Preparation creates no work invocation.

T4.3 separates explanation, preparation and development through shared operation
contracts. Explanations use bounded canonical project/session sources without
changing authoring state. Work preparation previews one native backlog ticket
from a real generated dump or spec. Its original instruction, source images,
namespace, native description and contract are retained as an exact scope digest.
Development requires an explicit configured-operator decision for that digest;
changed scope or operator cannot reuse the prior decision. Preparation requires
no development decision and never admits an invocation.

The completed artifact portion of T4.4 connects `/done`, `/quiz`, `/spec`, `/plan`
and explicit `/refine` controls to the shared continuation owner. Natural-language
proposals use those same handlers. Actual provider prompts bind outgoing replies
to their canonical session, turn, revision and digest; a clarification answer
continues its requested generation. Captures, original instructions and answer
whitespace remain in the canonical sources. The receiver returns actual artifact
references and retains the conversation's frozen namespace.

Transport reservations retain the complete immutable owner input before a
provider call. Successful owner receipts reconcile lost acknowledgements without
another generation, including after independent canonical edits. Conversation
session/question updates, source acknowledgement and outbox creation commit
together. Changed roots or operators during generation block new canonical
writes; reconciliation cannot restore a conversation binding that advanced.

The explanation portion of T4.4 connects `/explain` and ordinary-language
explanation proposals to the same configured-provider operation. Exact project,
conversation and reply routing use canonical context without creating artifacts
or work. The original question and source image are retained before invocation.
Provider budgets/results are independent of intent resolution; restart cannot
reclassify that request, reset five attempts or regenerate an acknowledged result.

`/prepare-work` and ordinary-language preparation use actual generated dump/spec
content to prepare a native backlog ticket. The preview retains original intent,
instruction, canonical sources and contract before dispatch. Preparation starts
no invocation. A literal `/develop` instruction can bind only the first exact
work reservation; a model proposal supplies no grant. Missing or changed
authority uses one retained scope preview, with equivalent button/reply/text
decisions. Unchanged preparation returns its existing receipt or prompt.

Confirmed native receipts recover lost acknowledgement before later source
checks, preserving independent source and ticket edits. Changed roots/operators
block new effects, and canonical refinement invalidates the prior scope prompt.
The User Guide, receiver reference and Unreleased notes describe the delivered
commands and their preparation behavior.

## Implementation Notes

T4.1 implementation and report introduction: `b54d526fd0471eff7c4df2ef769ba0abe66476ce`.

T4.2 implementation commit: `f7408ea2fa51d4a66ce5be156d2aa192160ad37f`.

T4.3 implementation commit: `4d1932fedf6ac79e659ce2b475b4261b822f50e6`.

T4.4 artifact-bridge implementation commit: `e4aa7bf44bc36ba8a5f1ca0edb82978f79bb0923`.

T4.4 explanation-bridge implementation commit: `37a1a0aad92404cf086bf202f812bde81d39f333`.

T4.4 native preparation implementation commit: `392ce4b1f8bb4180ce6f227ac3340aacaec29b77`.

The shared structured provider uses the existing Codex agent runner or OpenAI
Responses boundary. `CHR_CONVERSATION_GENERATION_BACKEND=openai` selects OpenAI;
otherwise Codex is used. Existing provider model, command and credential settings
retain their meaning. One provider instance permits two in-flight calls, each
with a 60-second deadline, at most 128 KiB of input and 64 KiB of accepted JSON
output. HTTP envelopes are read within 1 MiB. No SQLite transaction spans a
provider call.

Migration 19 adds durable interpretation attempt, retry, sanitized failure and
decision fields. An OS-backed bot-scoped interpretation lock cannot be stolen by
a wall-clock jump. Each attempt is reserved before invocation. Five failures
exhaust that source's budget; restart cannot reset it. Capacity, timeout and
cancellation have separate sanitized failure classifications. Retained decisions
count toward the existing logical transport capacity limit.

The private retained decision format is `chr.telegram.intent.v1`. Record payloads
are decoded to the same strict concrete requests used by slash commands; model
output cannot select an actor, command, executable or shared owner. Existing
transport tests use a deterministic capture provider, while new intent tests
exercise the actual configured HTTP boundary.

The existing authoring repository owns continuation requests, provider budgets,
question bindings and artifact writes. Its OS-backed project lock also guards
ordinary session, dump, quiz, spec and plan writers. A checksummed journal retains
all before/after images and the receipt; recovery verifies every image before
replay. Divergent files preserve the journal and block mutation. Provider calls
hold this owner without opening a SQLite transaction. Five retained attempts and
retry deadlines survive reopening. Captures, source context, transcript entries,
artifact references and private recovery reads have explicit size/count limits.
Legacy sessions remain readable; continuations require a managed revision and
do not adopt them implicitly.

Prepared native descriptions use the existing record preparation and mutation
owners. Callers retain the serialized preview before dispatch, including its
reserved ticket identity. Canonical envelope metadata retains scope and explicit
authority; the normal native mutation receipt reconciles lost acknowledgement
before later source checks. Reservation binds failed dispatch arguments too.
The authoring lock covers source validation and record dispatch. The receiver's
conversation projection now exposes existing authoring/authorization references;
only an original explicit `/develop` control creates an instruction reference.
A model's proposed intent cannot create that authority.
The original retained instruction can be recovered independently of focus/reply
invalidation. Its development reference is rechecked against the current operator;
remapping the user cannot inherit the earlier operator's decision.

The receiver adds a host-owned post-generation guard under the existing authoring
writer lock. Answer-and-continue is one owner journal/receipt rather than two
loosely coupled requests. Compact transport facts retain references and question
bindings without copying the full session into the outbox facts limit. The
configured provider checks cancellation before and after invocation, so a valid
response racing cancellation cannot be accepted.

Migration 20 adds per-operation model attempts, retry deadlines, sanitized failure
codes and bounded result JSON. The transport's existing logical capacity includes
these results. An OS-backed operation-model lock protects active calls; SQLite
transactions end before invocation. A successful result is persisted before its
source/outbox acknowledgement. Explanation source preparation and provider
dispatch recheck the resolved registry-root binding, and queuing rechecks current
operator/root/conversation authority.

Migration 21 pins the conversation's first work reservation with a foreign-key
reference. Reservation and that marker commit together; later failures, reopening
or clock changes cannot transfer the original instruction to another scope.
Prepared native input and scope digests use the existing request owner. Missing
authority uses the existing expiring action owner; its first answer identity is
retained as the development decision. Reoffering a consumed prompt preserves its
token, expiry and prepared projection. Phase projection keeps that request's
immutable conversation revision; subsequent authoring edits advance the revision.
The shared authoring owner rechecks host authority immediately before new native
effects. No separate transport scheduler or work-admission path is added.

## Contracts Added or Changed

- `StructuredModelRequest`, `StructuredModelProvider` and
  `ConfiguredStructuredModel`: bounded configured-provider invocation.
- `ConversationIntentContext` and `ConversationIntent`: a typed proposal with
  registry-owned operation IDs and independently validated arguments.
- `ResolveTelegramInterpretation`: durable decisions and retained retry budgets.
- `TelegramIntentContext` and `ResolveTelegramConversationTarget`: exact-peer
  routing snapshots through the existing conversation owner.
- Strict operation JSON decoding rejects duplicate and unknown fields, excessive
  nesting, trailing input and non-object payloads.
- `AuthoringContinuationRequest`, `AuthoringQuestionBinding` and
  `AuthoringSnapshot`: transport-neutral actions with exact canonical guards.
- `ObserveAuthoring` and `ContinueAuthoring`: the central service's shared
  configured-provider continuation boundary.
- `WithAuthoringMutation` and immutable continuation receipts: shared writer
  exclusion, guarded journal recovery and original-result reconciliation.
- `ExplainAuthoring`: bounded canonical explanation without artifact/work writes.
- `PrepareAuthoringWork` and `ApplyPreparedAuthoringWork`: retained native backlog
  descriptions through the shared record owner.
- `AuthoringWorkScope` and `AuthorizeAuthoringWork`: explicit decisions bound to
  the displayed scope, original instruction and configured operator.
- `ReserveRecordMutation`: immutable failed-dispatch arguments and reconciliation
  through the existing native receipt owner.
- `ReadTelegramConversationInstruction`: original-source recovery and current
  operator checks without using a model summary as user intent.
- `ContinueAuthoringGuarded`, `AuthoringCommitGuard` and the `respond` continuation:
  current host authorization before writes and atomic answer/generation recovery.
- `ResolveTelegramAuthoringTarget`, `ReservePreparedTelegramOperation` and
  `FinishTelegramAuthoring`: frozen routing/input and atomic conversation/outbox
  acknowledgement through the existing transport request owner.
- `PrepareAuthoringExplanation`, `PreparedAuthoringExplanation` and
  `ExplainPreparedAuthoring`: frozen canonical context for retryable explanation.
- `ResolveTelegramOperationModel` and `TelegramOperationModel`: retained provider
  results and five-attempt budgets independent of intent interpretation.
- `ValidateTelegramProjectTarget`: resolved-root validation at the adapter's
  entry to a project-scoped operation.
- `ApplyPreparedAuthoringWorkGuarded`: caller authorization under the shared
  authoring lock before new native effects, after original receipt reconciliation.
- `IsInitialTelegramAuthoringWork` and `LatestTelegramAuthoringWork`: immutable
  first-reservation identity and exact-peer retained work previews/results.
- Retained action offers/results can project the current bound authoring phase;
  replay does not extend expiry or restore a consumed confirmation state.

## Files of Interest

- `internal/app/service_conversation_intent.go`: adapter-neutral resolution.
- `internal/app/telegram_intent*.go`: registry mapping, frozen guards and dispatch.
- `internal/core/structured_model*.go`: configured model boundary and capacity.
- `internal/core/telegram_interpretation.go`: retained model attempts and results.
- `internal/app/service_authoring_continuation.go`: shared provider-backed
  generation and existing artifact renderers.
- `internal/core/authoring_continue.go`, `authoring_snapshot.go` and
  `authoring_journal.go`: canonical observations, question bindings and recovery.
- `internal/app/service_authoring_work.go` and `service_authoring_explanation.go`:
  shared prepared-work and explanation operations.
- `internal/core/authoring_work_scope.go`: immutable scope/decision validation.
- `internal/app/telegram_authoring_{command,input,dispatch,apply}.go`: shared
  command mapping, retained input, guarded dispatch and receipt reconciliation.
- `internal/core/telegram_authoring_finish.go`: atomic conversation/question,
  source and outbox transition.
- `internal/app/telegram_explanation_*.go`: deterministic/NL explanation routing,
  retained context and guarded result delivery.
- `internal/core/telegram_operation_model.go`: provider-result retention and
  bounded independent attempts.
- `internal/app/telegram_prepared_work_*.go`: shared native preparation,
  first-scope authority and retained action dispatch.
- `internal/core/telegram_authoring_work_request.go`: atomic first-scope binding
  and retained work lookup through the existing request owner.
- `internal/core/telegram_authoring*_test.go` and
  `internal/app/telegram_authoring*_test.go`: real SQLite, HTTP, native owner,
  replay and authority regressions.

## Validation

Successful T4.1 task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Fixtures prove slash/NL start, capture, completion and record parity; correction
for malformed or unsupported proposals; unchanged conversations after provider
failure; deterministic status during failure; revoked authorization and changed
registry roots during model calls; frozen namespace and decision after reopening;
five retained attempts; cross-owner interpretation exclusion; provider capacity
and cancellation; and no native work admission.

Successful T4.2 task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramAuthoring|AuthoringSession|ValidateAuthoring|Assisted|CurateAssisted|SynthesizeAssisted|WriteContextualQuiz|WriteQuizToDraft|ArchiveGeneratedPlan|Service.*(Dump|Quiz|Spec|Plan)|DumpNewAndShow)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramAuthoring|AuthoringSession|ValidateAuthoring|Assisted|CurateAssisted|SynthesizeAssisted|WriteContextualQuiz|WriteQuizToDraft|ArchiveGeneratedPlan|Service.*(Dump|Quiz|Spec|Plan)|DumpNewAndShow)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The real HTTP provider fixture generates dump, quiz, draft spec, refined spec and
plan artifacts, including generation that first requires clarification. Owner
fixtures interrupt every journal image replacement, reopen and replay once;
retain divergent files; reject changed request arguments and stale source images;
isolate questions across sessions; retain provider budgets; and serialize
independent owners. Existing CLI authoring regressions pass. These fixtures do
not yet prove the receiver's complete preparation journey.

Successful T4.3 task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test ./internal/core -run '^TestTelegramAuthoringInstructionReference$' -count=1 -timeout=3m`
- `go test -race ./internal/core -run '^TestTelegramAuthoringInstructionReference$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Real HTTP/SQLite/native-record fixtures prove explanation context, strict output,
unchanged sessions, backlog preparation without invocations, exact native scope
metadata, missing/changed authority rejection, stale source guards, immutable
failed bindings and original receipt replay after independent source/ticket
edits. The additional original-instruction fixture passed after source-recovery
support was added; it proves reopening and operator-remapping checks.

Successful T4.4 adapter task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Real HTTP/SQLite fixtures cover the receiver's dump/quiz/spec/refine/plan journey,
slash/NL parity, actual prompt replies, retained original instructions and raw
capture/answer whitespace, frozen namespace, stale prompts, root/operator changes
during generation and a lost owner acknowledgement followed by reopening and
independent artifact edits. Owner fixtures cover atomic answer/generation journal
recovery and failed conversation-finish rollback. A new whitespace assertion
first reproduced trimming loss; the corrected writer passes. The combined
regression also exposed acceptance of a model response racing cancellation; the
provider now rejects that result and the regression passes.

Explanation fixtures cover the real HTTP path for slash/NL project explanations
and canonical-session context without changing conversation state or admitting
work. Reopening after retained success but before acknowledgement delivers the
original result while preserving independent edits. Malformed HTTP output
exhausts five retained attempts without replacing intent or source context;
deterministic status remains usable. Core fixtures prove forged source/input
rejection, provider-call exclusion after clock jumps, independent database writes
during invocation and separate interpretation/result retention. Root/operator
changes during generation block result delivery, and a root rebind before source
preparation fails before reading the new project.

Real Git/SQLite/HTTP native preparation fixtures prove preparation/development
intent, actual backlog descriptions and scope metadata, unchanged-scope reuse,
slash/NL parity, missing model-proposed authority, equivalent text/reply/button
decisions, expanded-scope confirmation and stale prompt rejection. Both initial
instruction and explicit-decision paths recover a lost native acknowledgement
after independent source and ticket edits. Guard fixtures reject operator/root
changes before native effects. Cheap core fixtures prove atomic first-reservation
rollback and preservation after failure/reopening. A focused offer-replay test
first reproduced a consumed prompt regressing its conversation phase; reoffering
now leaves its consumed state, original expiry and projection intact.

Final validation is recorded by the controller after canonical PR references are
committed. Its receipt must match the clean PR HEAD and these ordered commands:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramAuthoring' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramAuthoring' -count=1 -timeout=3m`
- `git diff --check`

Task fixtures and controller validation are distinct from real-provider,
real-bot and exhaustive scheduled certification.

## Risks and Follow-ups

None recorded.
