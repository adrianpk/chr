# Slice 7: Telegram Harness Selection

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-7-telegram-harness-selection`
PR: pending

## Purpose

Inspect and confirm project worker and generation defaults through Telegram's
existing deterministic commands, mapped actor and retained action boundary.

## Delivered Behavior

- `/harness` observes the explicitly selected enabled registered project.
  Default-project fallback, foreign overrides and quoted reply targets cannot
  replace that selection.
- The reply shows readable worker/generation defaults, named profile readiness
  and separate retained worker bindings. It states that changes affect new
  operations while admitted work keeps its binding.
- The deterministic named-role grammar uses `-` to clear a role and rejects
  duplicate, unsupported or private setup fields. Help uses its canonical syntax.
- Inspection allocates no conversation, invokes no model and exposes no private
  executable, credential or model-configuration references.
- Named role changes require an exact retained confirmation from the mapped
  actor. Confirm publishes both choices once; cancel preserves defaults, and
  `-` restores legacy dispatch for the requested role.
- Stale revisions, changed selected projects/roots, revoked access, changed
  actor mappings and unavailable profiles cannot publish the pending change.

## Implementation Notes

T7.1 resolves project identity through the existing peer selection and registry
binding. Its thin application adapter calls the shared selection/readiness owner.
The response bounds named choices to eight and retained workers to four; local
readiness remains distinct from remote authentication.

T7.2 freezes the mapped actor, exact project/root, role choices, observed full
configuration revision and request ID in the existing action ledger. The
source-owned confirmation context and prepared operation are reserved together.
Confirm/cancel uses the existing reply, callback and `/answer` owner. Publication
rechecks current authority under the shared configuration lock; definitive local
preparation failures supersede the action while uncertain owner failures retain
its applying state for receipt recovery. Terminal prompts release conversation
capacity without disturbing focused authoring or admitted work.

T7.3 redelivery and cross-interface acceptance remains in progress.

## Contracts Added or Changed

- Deterministic `/harness` inspection and role-choice grammar.
- Explicit selected-project target observation without conversation allocation.
- Shared application read projection and canonical help entry.
- Immutable deterministic selection confirmation and mapped actor authorization.
- Prepublication preparation-error identity with the underlying cause retained.

## Files of Interest

- `internal/app/telegram_harness_command.go`, `telegram_harness_dispatch.go`
- `internal/app/telegram_harness_service.go`, `telegram_help_metadata.go`
- `internal/core/telegram_harness_target.go`
- `internal/app/telegram_harness_test.go`
- `internal/core/telegram_harness_target_test.go`
- `internal/core/telegram_harness_action.go`, `telegram_harness_action_test.go`
- `internal/app/telegram_harness_prepare.go`, `telegram_harness_action.go`
- `internal/app/telegram_harness_confirm_test.go`

## Validation

Passed focused task checks:

- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramHelp)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/core ./internal/app`
- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramOperationAction|TestHarnessSelectionChange)' -count=1 -timeout=3m`
- `git diff --check`

Owned HTTP/SQLite receiver fixtures inspect actual prepared profiles, require
explicit project selection and exclude private references. Core fixtures cover
foreign overrides/peers and repeated inspection without consuming conversation
capacity. Existing help coverage verifies the complete single-message catalogue.
Confirmation coverage verifies atomic role updates, the mapped actor, clear and
cancel, stale Web revisions, selected project/root changes, disabled projects,
revoked project access, changed actor mapping, unavailable local profiles,
changed envelopes and more sequential confirmations than the live-context limit.
Core coverage verifies failed reservation rollback and preparation rejection
without a published image or receipt.

## Risks and Follow-ups

Complete T7.3 before opening the canonical PR. Retained redelivery and
cross-interface acceptance remain pending.
