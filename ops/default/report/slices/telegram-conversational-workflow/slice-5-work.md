# Slice 5: Native Work Continuation

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-5-telegram-work`

PR: pending

## Purpose

Admit exact native work from a conversation and continue it through the shared
workflow owner, retaining canonical invocation and delivery evidence.

## Delivered Behavior

T5.1 connects `/work` and the same typed natural-language operation to shared
preview and bound admission for ticket, slice, slice-set, use-case and loose
extents. The preview retains its target commit, membership snapshot, optional
selected-ticket revision, finite execution limits and publication policy.
A changed source, policy, target or authority rejects the displayed admission.
A preview creates no native invocation; admission launches no worker.

A retained action supplies the exact `start` decision. An explicit session keeps
its project and namespace; otherwise the source allocates one independent,
retry-stable conversation. The original request ID becomes the common owner's
RequestKey. Reopening after a lost native acknowledgement observes the original
invocation before rechecking mutable source images. Changed arguments under the
same key remain conflicts. Conversation binding, both confirmed requests, action
consumption, source acknowledgement and result outbox commit atomically.

Native status exposes the same invocation through the central operation service.
The admission result distinguishes accepted work from validated or integrated
work. A new extent cannot replace a conversation's bound invocation. Root and
operator changes before a new effect reject without creating work.

## Implementation Notes

T5.1 implementation: `4456f3f1bbef9b0b8368294f12e40c9cc9ca15ff`.

`StartBoundWorkExtent` adds an optional stable admission binding. Its zero-step
mode calls the existing native admission owner and leaves continuation separate.
Existing callers retain their bounded start-and-continue behavior. The native
request hash retains the displayed admission guard, while original RequestKey
replay takes precedence over changed source images. New admission rechecks the
policy, selected revision and caller's final authority/target check under the
existing workflow lock. Target observation has a ten-second deadline.

The service retains an explicit workflow namespace instead of reading a changed
`OPS_USER` during replay. Read-only execution-policy inspection shares admission's
configuration resolver. Telegram stores only adapter bindings and projections;
it does not allocate a second scheduler or workflow owner.

The earlier receiver fixture used `/work` as an unavailable command. It now uses
an unknown command, preserving its unknown-command and no-work assertions while
the new fixtures cover the delivered admission operation.

## Contracts Added or Changed

- `WorkflowBoundAdmission` carries the RequestKey, displayed execution/publication
  policies and an optional caller-owned final effect check.
- Optional native `admission_guard` retains the displayed scope and participates
  in request hashing; existing invocation images retain their original hashes.
- `StartWorkflowGuarded` reconciles original admission before a new effect check.
- The existing conversation invocation and observed-revision columns now project
  canonical admission; no migration is required.
- `/work` accepts exact project, conversation and namespace scope plus one native
  selector. Only loose work omits its reference. Semantic proposals cannot skip
  its bound confirmation.

## Files of Interest

- `internal/app/service_workflow_admission.go`: common stable admission mode.
- `internal/core/workflow_admission_guard.go`: native displayed-scope guards.
- `internal/app/telegram_work_*.go`: thin command, routing and decision mappings.
- `internal/core/telegram_work_binding.go`: atomic canonical invocation projection.
- `internal/core/telegram_work_target.go`: source-stable conversation allocation.
- `internal/app/telegram_work_*test.go` and
  `internal/core/telegram_work_admission_test.go`: real-store admission fixtures.

## Validation

Activation documentation passed `git diff --check`.

T5.1 task checks passed:

- `go test ./internal/core ./internal/app -run '^(TestTelegram(Work|Authoring|Conversation|Operation|Receiver)|TestWorkflowInvocation|TestServiceWorkNative|TestWorkflowConfig)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^(TestTelegramWork|TestWorkflowInvocation|TestServiceWorkNative|TestWorkflowConfig)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^TestTelegramWork' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramWork' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app` — zero issues.
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The final focused normal/race checks passed after the last helper extraction.
Final clean-HEAD controller validation is pending the
complete slice and its canonical PR references. These are deterministic HTTP,
SQLite and temporary Git fixtures; live and scheduled exhaustive evidence remain
separate.

## Risks and Follow-ups

T5.2 through T5.4 remain pending: bounded asynchronous continuation, unattended
policy-driven progress, coalesced notifications and evidence-bound delivery.
