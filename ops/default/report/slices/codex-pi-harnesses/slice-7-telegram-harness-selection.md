# Slice 7: Telegram Harness Selection

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-7-telegram-harness-selection`
PR: pending

## Purpose

Inspect and confirm project worker and generation defaults through Telegram's
existing deterministic commands, mapped actor and retained action boundary.

## Delivered Behavior

- `/harness` observes the explicitly selected enabled registered project,
  readable role/profile/model names, role-specific local readiness and separate
  retained worker bindings. Reads allocate no conversation and invoke no model.
- `/harness worker=PROFILE generation=PROFILE` previews an exact change.
  Omitted roles stay unchanged; `-` restores legacy dispatch. Unsupported,
  duplicate and private setup fields are refused before reservation.
- A ten-minute confirmation binds the mapped actor, selected project/root,
  choices, whole-document revision and original request. Existing buttons,
  replies and `/answer` confirm or cancel the same retained action.
- Confirm publishes both roles atomically through the shared configuration
  owner. Changed project/root, actor mapping, access, revision or local readiness
  rejects the old confirmation. Cancel leaves defaults unchanged.
- Exact redelivery and lost acknowledgements preserve the original request and
  receipt. Uncertain Telegram delivery repeats a labelled retained notice;
  completed actions cannot exchange choices or publish again.
- Web, TUI and independent processes see saved defaults on their next read.
  New admissions use the new defaults; claimed workers, nested generation and
  generation resumed from its artifact retain their original binding.

## Implementation Notes

T7.1 uses existing peer selection and registry binding, with no default-project
fallback or quoted reply override. The public read shows at most eight profiles
and four retained workers. Local readiness is distinct from remote authentication;
private executables, credentials and model configuration references stay local.

T7.2 reserves the prepared envelope and source-owned confirmation conversation
in one SQLite transaction. It does not change the focused authoring context.
The existing action owner authenticates the answering peer and consumes the
exact answer. Publication rechecks current mapped authority inside the shared
configuration owner's cross-process lock. Terminal prompts release conversation
capacity. Definitive preparation rejection preserves its cause; uncertain owner
failures retain the applying state for receipt recovery.

T7.3 exercises real receiver HTTP decoding, SQLite, shared configuration writes,
owned Git workflow claims and persisted generation bindings. A lost owner
acknowledgement recovers the original result even after private profile setup
becomes unreadable. The result notice includes its request ID. Synthetic model
responses cannot authorize a configuration command or private setup.

## Contracts Added or Changed

- Deterministic `/harness` grammar and canonical help entry.
- Explicit selected-project observation without conversation allocation.
- Immutable role-selection confirmation using existing action/receipt owners.
- Typed prepublication preparation rejection with underlying error identity.
- User Guide, Telegram/project references and the existing Unreleased capability.

## Files of Interest

- `internal/app/telegram_harness_command.go`, `telegram_harness_dispatch.go`
- `internal/app/telegram_harness_service.go`, `telegram_harness_prepare.go`
- `internal/app/telegram_harness_action.go`, `telegram_action_dispatch.go`
- `internal/core/telegram_harness_target.go`, `telegram_harness_action.go`
- `internal/core/harness_selection_change.go`
- `internal/app/telegram_harness_*test.go`
- `internal/core/telegram_harness_*test.go`

## Validation

Passed task checks:

- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramHelp)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramOperationAction|TestHarnessSelectionChange)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramOperationAction|TestTelegramOperationUncertain|TestHarnessSelectionChange|TestTelegramHelp)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/core ./internal/app`
- `git diff --check`

The receiver fixtures cover confirmation, clear/cancel, stale Web changes,
project/root changes, remapped and foreign actors, revocation, unavailable local
setup, changed envelopes, locked authority, source redelivery, lost owner/send
acknowledgements and rejection of model-only authority. Core fixtures cover
reservation rollback, expiry and rejected preparation without image or receipt.
The active-work fixture checks a live claim, nested/resumed generation artifacts,
new admissions and existing Web/TUI readers plus an independent process.

Before integration, the controller records the following ordered checks against
the final clean PR head; these entries are planned until its evidence is recorded:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^(TestTelegramOperationAction|TestTelegramOperationUncertain|TestHarnessSelectionChange)' -count=1 -timeout=3m`
- `git diff --check`

Fixtures use owned temporary HTTP/SQLite/Git state and local profile resources;
no message is sent to the live bot. Generation checks retain real production
bindings and do not invoke a remote model or worker driver.

## Risks and Follow-ups

None recorded.
