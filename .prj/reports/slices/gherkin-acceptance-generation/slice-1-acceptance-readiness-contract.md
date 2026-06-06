# Slice 1: Acceptance Readiness Contract

Status: reviewing
Slicing: Gherkin Acceptance Generation
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`
Branch: `feat/slice-1-acceptance-readiness-contract`
PR: pending

## Purpose

Establish the internal acceptance readiness contract before connecting quiz or
spec generation behavior. This slice defines the stable decisions, slots, and
prompt intent shape that later slices can reuse.

## Delivered Behavior

ChrysoCoord now has package-level types for acceptance readiness decisions:
`generate`, `needs_more_info`, and `not_applicable`.

Readiness validation now normalizes reasons, inferred elements, missing slots,
candidate scenarios, and blocking open questions. It rejects invalid decisions,
out-of-range confidence, missing reasons, invalid slots, contradictory
`needs_more_info` payloads, and unsafe `generate` decisions.

The package can also construct semantic prompt intents for acceptance readiness
slots. These intents use `locale: auto`, stable slot names, question guidance,
expected answer kind, and forbidden-slot metadata so later quiz prompts can ask
focused questions without hardcoded user-facing wording.

## Implementation Notes

The contract lives in `internal/chc` because readiness is shared domain behavior
for quiz and spec generation. No CLI, TUI, model-provider, workspace persistence,
or feature-file generation behavior was added in this slice.

The `generate` decision currently requires inferred actor, trigger, initial
state, observable outcome, success path, and at least one candidate scenario.
Failure path and boundaries are captured as stable slots, but they are not
required for every generated happy-path scenario.

## Contracts Added or Changed

- Added `AcceptanceReadinessDecision`.
- Added `AcceptanceReadinessSlot`.
- Added `AcceptanceReadiness`.
- Added `AcceptanceReadinessPromptIntent`.
- Added `NewAcceptanceReadinessPromptIntent`.
- Added `ValidateAcceptanceReadiness`.
- Added stable readiness slot order through `AcceptanceReadinessSlots`.

## Files of Interest

- `internal/chc/acceptance_readiness.go`
- `internal/chc/acceptance_readiness_test.go`

## Validation

- `go test ./internal/chc -run AcceptanceReadiness` (pass)
- `go test ./internal/chc` (pass)

## Risks and Follow-ups

- Later slices still need to connect readiness to quiz requests, persistence,
  spec generation, feature artifact writing, and static feature checks.
- The current static contract does not execute `godog`; that remains deferred by
  the slicing plan.
