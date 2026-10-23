# Slice 5: Guided Telegram Journey

Status: drafting
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-5-guided-telegram`
PR: pending
Initial routing/capture commit: `77b7f637587a63026ce914368939e7f0ed7158e7`
Current-step routing commit: `5b032e9f6c7ce34b0e0222ca6162ca63fbcc6e53`
Shared lifecycle/presentation commit: `5de6795a8cadcf78fb0dda91a3f67830eebaeb72`

## Purpose

Map Telegram onto the same retained guided use case delivered by the Web
checkpoint. T5.1 and T5.2 are complete. T5.3 transport faults and documentation
remain active; the slice stays drafting until those and the final exact-HEAD
checks are complete.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Tasks and evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Explicit natural-language development starts create one canonical guided
  session through the shared service. Incidental or negated development text
  cannot grant progression authority through this deterministic entry.
- Unique project names resolve genuine project ambiguity. Disabled projects are
  rejected. A peer can select an existing run with Show use case and its unique
  visible name; duplicate names cannot select an arbitrary run.
- Current name, capture and clarification replies bind the displayed authoring
  revision, question and committed input digest. Show current step reads the
  original run without dispatching generation; Retry retains the canonical
  pending operation or deterministic next-step identity.
- Show controls offers sealed buttons for Finish capture and available
  pause/resume/cancel actions. A callback consumes the existing retained action;
  its confirmed result, next prompt binding and source acknowledgement share one
  receiver transaction. Duplicate callbacks and lost application acknowledgements
  do not repeat the canonical mutation.
- Clarification reaches shared generated-specification approval. Missing native
  setup is reported as a retained prerequisite. Later ideas remain separate from
  committed inputs, and explicit controls delegate to the shared guided owner.
- Replaced prompts, foreign peers and changed registered roots cannot replace
  the original binding or canonical input. The additive nullable association
  leaves existing manual rows and authoring eligibility unchanged.

- The production receiver reuses the shared managed host for retained authoring
  progression and exact native work. Association discovery is paged and bounded;
  current peer, operator, registered root and canonical run checks remain effective
  at recovery and native dispatch boundaries.
- A unique name can attach Telegram to an existing guided Web run. Inspection
  does not advance it or create another run; the managed lifecycle continues the
  original actor, admission, native invocation and consumed limits.
- Unchanged human views coalesce. Background projection defers while a receiver
  owner request awaits acknowledgement, preserving its original reply fence.
  Native review and real prerequisites remain visible.
- Show dump/quiz/specification/plan/tracker attaches complete committed originals.
  Long current questions use the same complete-document snapshots. Working-copy
  replacements cannot replace those bytes, and paused runs show their control
  state rather than an active question attachment.
- Results show completion only from verified native delivery and closing evidence,
  with actual repository/branch/commit locations. Show results exports complete
  verified delivery, validation, closing and pending-idea evidence. Terminal
  associations remain inspectable without becoming new execution authority.

## Implementation Notes

The receiver owns transport routing and display references in its existing
conversation/request/outbox stores. Canonical authoring, checkpoints, controls,
native work and results remain in project owners. Conversation `collecting`
denotes transport input eligibility, not execution status.

Each complete owner input is frozen before application. The current registered
project and its retained root binding are rechecked. Ordinary replies retain the
original displayed revision; explicit inspection and control offers read current
state without mutation. The generic action owner now accepts a concise sealed
message presentation as well as its existing original-document presentation.
Guided callbacks compose their display projection into the existing action finish
transaction rather than introducing a second action coordinator.

Explicit new-project starts route before generic action replies because a
project-choice diagnostic intentionally grants no reply/action authority. A named
new start cannot adopt a quoted conversation. Existing capture replies and action
callbacks keep their original guards.

The shared host retains its existing sixteen-job bounds, finite native budgets
and failed-operation fingerprints. The receiver discovers at most sixteen
confirmed associations per sweep and delegates journal recovery through a shared
exact-scope lifecycle operation. Read paths remain effect-free. Native stop and
status controls use reserved application lanes; normal continuation retains its
ordinary execution lane.

A digest of the human view coalesces unchanged notifications without changing
native revisions or receipts. The existing outbox and document admission helpers
retain the exact frozen image, quota, hash, destination, retry and cleanup rules.
No new workflow store or execution engine is introduced.

## Contracts Added or Changed

- Optional `telegram_conversations.guided_run_id` associates a canonical guided
  session without extending manual intent/phase enums.
- Guided display bindings retain authoring, question/input and native revisions.
  Manual capture/authoring commands cannot adopt a guided association.
- Existing action peer, scope, immutable presentation, choice and expiry guards
  cover guided buttons. The original request ID remains the shared command ID.
- Existing reply retention/pruning, peer allocation limits, prepared-request
  bounds and canonical authoring CAS remain authoritative.

## Files of Interest

- `internal/core/telegram_guided_binding.go` and `telegram_guided_finish.go`:
  stable association, peer/root checks and atomic result/reply projection.
- `internal/app/telegram_guided_*.go`: shared-operation mapping, frozen current
  inputs, name selection, concise prompts and retained action dispatch.
- `internal/core/telegram_action*.go`: reuse of sealed message actions and their
  original finish transaction for guided display bindings.
- `cmd/chr/assets/migration/sqlite/20261023000029-add_telegram_guided_binding.sql`:
  additive association preserving existing manual rows.

## Validation

T5.1 focused task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCaseTelegram|TelegramOperationAction|TelegramPresentationAction|TelegramActionPresentation|TelegramConversation)' -count=1 -timeout=3m` — core 5.877s, app 2.853s.
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCaseTelegram' -count=1 -timeout=3m` — core 5.352s, app 15.516s.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/...` — zero issues.
- `git diff --check` — passed before the task commit.

T5.2 focused task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCaseTelegram|TelegramDocument|TelegramPresentationDocument|TelegramControlLane|TelegramConversation)' -count=1 -timeout=3m` — core 5.137s, app 17.559s.
- `go test ./internal/app -run '^TestGuidedUseCaseTelegramResultGuard$' -count=1 -timeout=3m` — app 0.007s; a terminal state cannot replace missing or partial native evidence.
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCaseTelegram' -count=1 -timeout=3m` — core 7.728s, app 43.367s.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/...` — zero issues.
- `git diff --check` — passed before the task commit.

The managed receiver fixture keeps status responsive while generation is blocked,
then pauses/resumes at the shared authoring boundary. It uploads the complete long
question. The cross-interface fixture attaches an existing Web run without a
canonical mutation, executes a real owned worker to native review, preserves
pause/resume and limits, and reconciles a lost response after an actual fixture
Git publication before showing verified completion. Repeated views do not replace
reply bindings. Result exports remain available after terminal projection, and
plan/tracker originals survive replacement of their current files.

These representative additions cost roughly 18s normally and 44s under race
instrumentation in the application package. They remain within the configured
three-minute package timeout; no nightly or exhaustive gate was run.

Focused failures distinguished the shared authoring boundary from immediate
provider interruption and caught paused long-question rendering. The fixture now
checks the actual owner boundary, and automatic question attachments require an
active dispatch state. Asynchronous notice checks await the specific result export
rather than assuming that the next delivered notice is its response.

The physical pre-migration database preserves its manual topic, capture,
revision and eligibility after upgrade. Real HTTP, SQLite and temporary Git
fixtures exercise current clarification, exact default approval, named selection,
project ambiguity, disabled projects, button consumption, duplicate callbacks,
application-response loss, ideas and pause. Provider and lost-response doubles
replace only those unavailable external boundaries. They do not establish a
user-operated Telegram client observation.

Focused failures exposed the existing constrained intent enum, canonical capture
separators, unbound diagnostic reply routing and strict function-complexity limits.
The additive association, canonical-content assertion, explicit-start routing and
cohesive helper boundaries corrected them; focused checks pass.

Initial routing/capture also passed the broader normal and race selection
`go test [-race] ./internal/core ./internal/app -run '^Test(GuidedUseCaseTelegram|TelegramConversation|TelegramAuthoring)' -count=1 -timeout=3m`
with each form executed separately. That race selection cost core 119.881s and
app 100.355s. Subsequent checks stay focused unless another manual owner changes.
The final configured ordered commands remain pending until the complete clean
PR-reference HEAD; Looper will retain that exact evidence.

## Risks and Follow-ups

T5.3 still owns the remaining stale/foreign source, concurrent control, transport
loss/restart regressions, Telegram guidance and changelog, followed by the final
exact-HEAD PR checks. Include an urgent control finishing between an ordinary
guided effect and its receiver acknowledgement, preserving the original operation
and latest display fence. Authoring controls serialize at the shared mutation boundary
and retain provider deadlines; native stop intent remains distinct from quiescence.
No separately authorized user-operated Telegram client observation has been
performed for this slice.
