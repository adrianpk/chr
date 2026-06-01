# Slice 3: Agent Evidence Classification

Status: reviewing
Slicing: Dev to Main Agent Evidence Gate
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-3-agent-evidence-classification`
PR: pending

## Purpose

Classify gate changed files into ChrysoCoord-managed work and manual or
external changes without requiring humans to create ChrysoCoord tickets for
their own work.

## Delivered Behavior

- Discovers execution reports through the existing project report lister.
- Resolves execution `result.json` paths from report metadata.
- Maps Git changed files to agent-managed work when execution result evidence
  names the exact file path.
- Classifies unmatched Git changed files as manual or external context.
- Records missing ChrysoCoord execution evidence as a gate failure.
- Adds classification sections to JSON and Markdown gate reports.

## Implementation Notes

The classifier is intentionally strict about agent-managed evidence and neutral
about unmatched changes. A file is agent-managed only when ChrysoCoord execution
evidence names it. Everything else is reported as manual or external unless
later policy changes ask for different behavior.

Missing evidence is limited to broken ChrysoCoord-owned execution records, such
as an execution report that points to a missing or unreadable `result.json`.

## Contracts Added or Changed

- Gate reports now include:
  - `agent_managed_work`
  - `manual_or_external_changes`
  - `missing_agent_evidence`
- `chc gate dev-main` fails when ChrysoCoord-owned execution evidence is
  present but incomplete.
- Manual or external changes remain informational by default.

## Files of Interest

- `internal/chc/gate_evidence_classification.go`
- `internal/chc/gate_report.go`
- `internal/chc/report_listing.go`
- `internal/chcapp/gate.go`
- `internal/chcapp/cli_test.go`

## Validation

- `go test ./internal/chc -run 'GateEvidence|ReportLister'` (pass)
- `go test ./internal/chc ./internal/chcapp -run 'GateReport|GateEvidence|GateDevMain'` (pass)
- `go test ./internal/chcapp -run 'GateDevMain'` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Matching is exact path-based in this slice.
- Policy configuration remains deferred to Slice 4.
- Forgejo/Woodpecker feedback remains deferred to Slice 5.
