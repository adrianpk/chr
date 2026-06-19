# Slice 2: Evidence Checklist Writer

Status: reviewing
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-2-evidence-checklist-writer`
PR: pending

## Purpose

Add the ticket Markdown foundation for evidence strategy so later slices can
inject worker context, enforce review gates, and surface evidence state without
rewriting user-authored ticket content.

This slice keeps behavior explicit and local: the evidence strategy block is a
normal Markdown checklist bounded by stable machine anchors.

## Delivered Behavior

- Core can parse the anchored `chc:evidence-strategy:v1` block.
- Core validates required labels, checked state, primary status values, missing
  labels, duplicate labels, unknown labels, and malformed anchors.
- Core can render a deterministic human-readable v1 checklist block.
- Core can insert a missing evidence strategy block into a ticket body.
- Core can update only provided checklist labels while preserving existing valid
  user-authored values.
- App service code can write the checklist to a ticket by key and reload ticket
  state.
- Service tests prove ticket body sections such as Description, Comments,
  Blocker, and Validation are preserved.

## Implementation Notes

- The parser intentionally rejects malformed existing blocks instead of silently
  replacing user content.
- Inserts happen after ticket frontmatter when frontmatter exists.
- Partial updates require valid canonical labels and statuses.
- The updater does not infer evidence strategy. It writes caller-provided
  strategy state, preserving non-updated checklist items.

## Contracts Added or Changed

- Added `EvidenceStrategyChecklist`.
- Added `EvidenceStrategyChecklistItem`.
- Added `ParseEvidenceStrategyChecklist`.
- Added `RenderEvidenceStrategyChecklist`.
- Added `UpsertEvidenceStrategyChecklist`.
- Added `TicketEvidenceStrategyChecklistWriter`.
- Added `service.UpsertTicketEvidenceStrategyChecklist`.

## Files of Interest

- `internal/core/evidence_strategy_checklist.go`
- `internal/core/evidence_strategy_checklist_test.go`
- `internal/core/ticket_evidence_strategy_checklist.go`
- `internal/app/service_artifact_ticket.go`
- `internal/app/service_test.go`

## Validation

- `go test ./internal/core ./internal/app -run 'EvidenceStrategyChecklist|ServiceUpsertTicketEvidenceStrategyChecklist'` (pass)
- `golangci-lint run --timeout=5m ./internal/core ./internal/app` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The writer is available through core/app service paths only; worker prompt
  integration is delivered in a later slice.
- Review gating is not enforced by this slice.
- Gherkin applicability decisions are not inferred by this slice.
