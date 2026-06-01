# Slice 4: Gate Policy Configuration

Status: delivered
Slicing: Dev to Main Agent Evidence Gate
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-4-gate-policy-configuration`
PR: #267

## Purpose

Add project-local policy configuration for `chc gate dev-main` under
`.chc/user/config/gates.yml`.

## Delivered Behavior

- Added a gate config parser with defaults.
- Added `validation_command` override support.
- Added traceability policy for missing agent evidence and manual change
  warnings.
- Added `manual_changes.list` to hide manual change details while preserving
  warning behavior.
- Added high-risk manual change matching with warning-only behavior when
  enabled.

## Implementation Notes

The config file is intentionally separate from `.chc/user/config.yml` so gate
policy can evolve without overloading the existing workspace/runtime config.

High-risk path matching supports exact paths, shell-style path matches, and
directory prefix patterns ending in `/**`.

## Contracts Added or Changed

- New config path: `.chc/user/config/gates.yml`.
- New default gate policy:
  - validation command: `go test ./...`
  - traceability mode: `agent_only`
  - fail on missing agent evidence: true
  - warn on manual changes: false
  - list manual changes: true
  - warn on high-risk manual paths: false
- `chc gate dev-main` now honors configured validation and traceability policy.

## Files of Interest

- `internal/chc/gate_config.go`
- `internal/chc/gate_validation.go`
- `internal/chc/gate_evidence_classification.go`
- `internal/chc/gate_report.go`
- `internal/chcapp/gate.go`
- `internal/chcapp/cli_test.go`

## Validation

- `go test ./internal/chc -run GateConfig` (pass)
- `go test ./internal/chc ./internal/chcapp -run 'GateValidation|GateDevMain'` (pass)
- `go test ./internal/chc ./internal/chcapp -run 'GateEvidence|GateReport|GateDevMain'` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The config parser intentionally supports the documented `gates.dev_main`
  subset needed by this slice.
- Forgejo/Woodpecker feedback remains deferred to Slice 5.
