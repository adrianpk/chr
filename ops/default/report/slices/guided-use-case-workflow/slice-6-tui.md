# Slice 6: Guided TUI Journey

Status: reviewing
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-6-guided-tui`
PR: [#435](https://forge.adrianpk.com/adrianpk/chr/pulls/435)
Acceptance/documentation commit: `f5c51e8a5ea9754f996bf4967f9581da8ff1fbb7`
Controls/results commit: `da313792a409f1e19145a9c33137b0b854600d0d`
Input/navigation commit: `5e15bddd36e03d7adbb02ad3856d1ae3a22d4519`

## Purpose

Expose the existing retained guided workflow through terminal interaction.
Input, navigation, native controls, truthful results and shared-run continuation
are implemented. Independent terminal interaction acceptance and criterion
reconciliation are recorded; final PR-reference validation remains pending.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Tasks and evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Cases `g` and the Guided use cases command palette entry open a named run
  catalog. Explicit development starts one canonical guided session in the
  selected project/namespace; reopening an existing run reads its original state.
- The current name, capture or clarification input uses a terminal editor.
  Submission freezes its run, question/input binding, revision, operation and
  full text before application. A refresh cannot rebind an unsent answer.
- Failed input retains the draft and original request. Retry preserves the
  original command. Responses from an earlier closed screen cannot replace a
  reopened screen, and mouse events cannot activate controls beneath it.
- Finish capture, quiz answers and default specification approval delegate to
  the shared service. Native prerequisites remain real retained blockers.
- Original dump, quiz, specification, plan and tracker reads use committed
  images. Complete documents preserve line breaks and can scroll beyond the
  viewport; working-copy replacement cannot substitute their contents.
- Explicit TUI startup installs the existing bounded managed recovery/worker
  host. Terminal refresh and render paths remain queries. Leaving the guided
  screen keeps the host available; process shutdown closes its owned host.

- Pause/resume/cancel carry the displayed native revision to the shared owner.
  Stop requests distinguish dispatch prevention from acknowledged quiescence;
  resume retains the original invocation, policy and consumed attempts.
- Native approval opens the existing exact-candidate inspection and confirmation
  flow. Required human review remains visible. Continuing an uncertain operation
  resubmits its retained binding rather than allocating another operation.
- Separate later ideas retain attribution without extending admitted work.
  Verified results and their complete export include delivery, validation and
  closing receipts with real repository, branch, commit and PR locations.

## Implementation Notes

`internal/tui` owns terminal state, editing, scrolling, sequence fences and
asynchronous command results. `internal/app` maps those requests to the existing
guided commands and original reader. No durable workflow state is added to the
terminal adapter. Original run IDs remain internal selection values; catalog
entries use their human names.

The current editor binding is separate from the latest observation. A successful
input clears its editor; a rejection preserves its complete text. Polling keeps
an open committed original available instead of replacing it with the normal
step view. Lifecycle recovery belongs to startup through `guidedWorkHost.start`,
not to model update, render or inspection.

The shared presentation helpers preserve existing Web/Telegram control and
completion semantics while mapping them into TUI views. Native completion alone
cannot become terminal delivery: missing, partial or invalid receipts remain
unconfirmed. Managed shutdown cancels child request contexts before joining owned
host jobs.

## Contracts Added or Changed

- `tui.GuidedExecutor`, `GuidedRequest`, `GuidedView` and `GuidedDocument` carry
  terminal requests and observations over the application adapter boundary.
- Start and input requests explicitly bind the selected project; the service
  remains the owner of namespace resolution and canonical revision checks.
- The command palette now returns asynchronous commands while preserving the
  existing manual command ordering and actions.

## Files of Interest

- `internal/tui/guided*.go`: current-step editing, frozen operations, named run
  selection and scrollable originals.
- `internal/app/tui_adapters_guided.go`: thin shared-service mapping and pure
  inspection/original reads.
- `internal/app/tui_guided_host.go`: explicit startup integration with the shared
  managed workflow host.
- `internal/app/tui_guided_terminal_linux_test.go`: owned production event-loop
  process and retained pause/reopen acceptance.
- `internal/app/tui_guided_parity_test.go`: exact cross-adapter native continuation
  and receipt-backed results with immutable prepared originals.
- `internal/app/tui_guided_test.go`: terminal model keys through real canonical
  and Git owners, generated approval, committed originals and reopen.

## Validation

T6.1 task checks passed:

- `go test ./internal/tui -count=1 -timeout=3m` — passed, 0.064s.
- `go test ./internal/app ./internal/tui -run '^Test(GuidedUseCaseTUI|TUIWorkflow|TUIHarnessSelectionRead|RootCommandDefaultsToTUIAction)' -count=1 -timeout=3m` — passed, app 6.746s, TUI 0.010s.
- `go test -race ./internal/app ./internal/tui -run '^TestGuidedUseCaseTUI' -count=1 -timeout=3m` — passed, app 1.719s, TUI 1.034s.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/app/... ./internal/tui/...` — zero issues.
- `git diff --check` — passed before the task commit.

The viewport regression first exposed collapsed original line breaks; literal
document wrapping corrected the failure. Existing function bounds required a
small modal routing helper, with no new lint suppressions. Terminal model events
exercise actual shared operations, temporary Git repositories and canonical
stores with deterministic content-provider responses. They do not establish an
independently observed terminal process or live-provider acceptance.

T6.2 task checks passed:

- `go test -race ./internal/app ./internal/tui -run '^TestGuidedUseCase(TUI|Parity|Telegram(Journey|ResultGuard))' -count=1 -timeout=3m` — passed, app 49.555s, TUI 1.035s.

- `go test ./internal/app ./internal/tui -run '^TestGuidedUseCase(TUI|Parity|Telegram(Journey|ResultGuard|Clarification|Buttons)|Control)' -count=1 -timeout=3m` — passed, app 30.944s, TUI 0.010s.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/app/... ./internal/tui/...` — zero issues after extracting bounded projection/routing helpers; no suppressions.
- `git diff --check` — passed before the task commit.

The parity fixture starts with one prepared Web run, attaches Telegram, pauses
through terminal model keys, resumes through Telegram and retains a later idea
through terminal input. Stale revision, foreign project and namespace commands
cannot change the shared stop state. Native review, exact membership, actor,
policy and one consumed worker attempt survive interface changes. The same
fixture completes actual temporary Git publication with a lost response and
reconciles original evidence before all three interfaces report completion.
Prepared plan/tracker originals remain bound after later working-copy edits.

T6.3 task checks passed:

- `go test -race ./internal/app ./internal/tui -run '^TestGuidedUseCase(TUITerminal|TUICatalogResult|TUIResultGuard|TelegramResultGuard|Parity)' -count=1 -timeout=3m` — passed after correcting the fixture lock, app 38.620s, TUI 1.029s.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/app/... ./internal/tui/...` — zero issues.
- `git diff --check` — passed after the metadata follow-up removed a trailing
  space in the cycle report. Its initial staged whitespace check failed; the
  task commit had continued, and this follow-up corrects the recorded outcome.

- `go test ./internal/app ./internal/tui -run '^TestGuidedUseCase(TUITerminal|TUICatalogResult|TUIResultGuard|TelegramResultGuard|Parity)' -v -count=1 -timeout=3m` — passed, app 24.729s, TUI 0.007s; PTY parent executed, 14.32s.
- `go test ./internal/app -run '^TestGuidedUseCaseTUITerminal$' -count=1 -timeout=1m` — passed, 14.320s.

The Linux PTY child assembles production adapters and `runManagedTUI`, then uses
actual terminal input/output through Bubble Tea. Its graceful exit joins owned
host work. A second child reopens the same paused run without changing its
revision or later idea. Early fixture runs sent keys before a save/editor
finished; consuming the corresponding rendered output corrected synchronization.
The first race probe found that an embedded bytes buffer promoted `ReadFrom`,
allowing `io.Copy` to bypass the fixture output lock. A private buffer field keeps
all writes behind the same lock. The corrected race result is recorded below.
These are automated local terminal observations with deterministic generation,
not human or live-provider acceptance. Required parent fixtures execute; child
entry tests skip only when not launched by their owning parent.

The [cycle report](../../cycles/guided-use-case-workflow.md) records concrete input
checkpoint and native delivery/validation/closing receipt identities from the
successful parity fixture. Catalog entries request result inspection instead of trusting raw terminal state;
result exports use the same verified status as the visible detail.
The tracker reconciles every approved acceptance
criterion. The User Guide, TUI/application references and Unreleased capability
now describe the completed terminal journey.

Final configured checks run through Looper on the complete clean PR-reference
HEAD. The canonical PR Validation and controller ledger retain the exact ordered
commands, candidate and successful outcomes; earlier task checks are separate.

## Risks and Follow-ups

Final integrated-candidate gate and documentation closure remain controller-owned.
User-operated Telegram client and live-provider acceptance have not been observed.
