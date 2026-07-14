# Slice 1: Evidence Strategy Config

Status: delivered
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-1-evidence-strategy-config`
PR: `#320`

## Purpose

Introduce the configuration and core vocabulary for evidence strategy without
changing ticket execution, worker prompts, or review gating behavior yet.

This slice creates the stable defaults and validation surface that later slices
will use for ticket checklist writing, worker context injection, review
transition enforcement, and user-facing visibility.

## Delivered Behavior

- Workspace config now exposes an `Evidence` section loaded from
  `.chr/default/config.yml`.
- `evidence.strategy` defaults are available without requiring generated config
  to include every field.
- Evidence strategy mode, ticket checklist policy, and Gherkin policy values are
  validated through core helpers.
- Core now has evidence strategy status and diagnostic code types for later
  readiness and blocker work.
- Reference documentation describes the new config fields, defaults, and valid
  enum values.

## Implementation Notes

- The first implementation keeps the behavior inert: loading config does not yet
  change worker execution or board transitions.
- The parser accepts the intended nested shape:
  `evidence.strategy.<field>`.
- Defaults match the spec: inline strategy, required ticket checklist, active
  Gherkin when applicable, strict review gating, user edits enabled, and report
  copy enabled.
- The config setter is split by field type to stay under the repository
  complexity gate.

## Contracts Added or Changed

- Added `EvidenceConfig` and `EvidenceStrategyConfig`.
- Added `EvidenceStrategyMode`, `EvidenceTicketChecklistPolicy`, and
  `EvidenceGherkinPolicy`.
- Added evidence strategy validation helpers.
- Added evidence strategy status and diagnostic code constants.
- Added `docs/reference/evidence-strategy-config.md`.

## Files of Interest

- `internal/core/config.go`
- `internal/core/evidence_strategy_config.go`
- `internal/core/evidence_strategy.go`
- `internal/core/config_test.go`
- `docs/reference/evidence-strategy-config.md`
- `docs/index.md`

## Validation

- `go test ./internal/core -run 'WorkspaceConfig|EvidenceStrategy|Validate.*Config'` (pass)
- `golangci-lint run --timeout=5m ./internal/core` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The config is intentionally inert until later slices connect checklist
  writing, worker context, and review gating.
- `separate_pass` is accepted as a future-compatible enum value but not
  implemented as a separate worker mode in this slice.
