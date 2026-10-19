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

## Implementation Notes

T7.1 resolves project identity through the existing peer selection and registry
binding. Its thin application adapter calls the shared selection/readiness owner.
The response bounds named choices to eight and retained workers to four; local
readiness remains distinct from remote authentication.

T7.2 immutable confirmation and T7.3 redelivery/cross-interface acceptance remain
in progress. No selection is changed by the current inspection path.

## Contracts Added or Changed

- Deterministic `/harness` inspection and role-choice grammar.
- Explicit selected-project target observation without conversation allocation.
- Shared application read projection and canonical help entry.

## Files of Interest

- `internal/app/telegram_harness_command.go`, `telegram_harness_dispatch.go`
- `internal/app/telegram_harness_service.go`, `telegram_help_metadata.go`
- `internal/core/telegram_harness_target.go`
- `internal/app/telegram_harness_test.go`
- `internal/core/telegram_harness_target_test.go`

## Validation

Passed focused T7.1 checks:

- `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramHelp)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/core ./internal/app`
- `git diff --check`

Owned HTTP/SQLite receiver fixtures inspect actual prepared profiles, require
explicit project selection and exclude private references. Core fixtures cover
foreign overrides/peers and repeated inspection without consuming conversation
capacity. Existing help coverage verifies the complete single-message catalogue.

## Risks and Follow-ups

Complete T7.2 and T7.3 before opening the canonical PR. Confirmation and retained
operation/cross-interface acceptance remain pending.
