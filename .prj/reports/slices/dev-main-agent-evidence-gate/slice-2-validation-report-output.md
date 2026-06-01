# Slice 2: Validation and Report Output

Status: reviewing
Slicing: Dev to Main Agent Evidence Gate
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-2-validation-report-output`
PR: pending

## Purpose

Run the default promotion validation and persist gate evidence as local project
artifacts under `.chc/user/reports/gates/`.

This slice turns `chc gate dev-main` from evidence-only output into a command
that records JSON and Markdown reports and returns failure when validation does
not pass.

## Delivered Behavior

- Added a validation runner with the default command `go test ./...`.
- Added timestamped JSON gate reports.
- Added timestamped Markdown gate reports.
- Wrote reports for successful validation, failed validation, and Git evidence
  failures.
- Made `chc gate dev-main` return an error when validation fails.
- Printed generated report paths in command output.

## Implementation Notes

Gate reports use `.chc/user/reports/gates/dev-main-<timestamp>.json` and the
matching `.md` file. The report model records Git evidence, validation result,
failures, warnings, informational messages, and artifact paths.

The command still uses the default validation command directly. Configurable
validation commands remain part of Slice 4.

## Contracts Added or Changed

- New default gate validation command: `go test ./...`.
- New report artifacts:
  - `.chc/user/reports/gates/dev-main-<timestamp>.json`
  - `.chc/user/reports/gates/dev-main-<timestamp>.md`
- `chc gate dev-main` now fails when validation fails.

## Files of Interest

- `internal/chc/gate_validation.go`
- `internal/chc/gate_report.go`
- `internal/chcapp/gate.go`
- `internal/chcapp/cli_test.go`

## Validation

- `go test ./internal/chcapp -run GateDevMain` (pass)
- `go test ./internal/chc ./internal/chcapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Configurable validation commands are intentionally deferred to Slice 4.
- Agent-managed evidence classification is intentionally deferred to Slice 3.
- Forgejo/Woodpecker behavior is intentionally deferred to Slice 5.
