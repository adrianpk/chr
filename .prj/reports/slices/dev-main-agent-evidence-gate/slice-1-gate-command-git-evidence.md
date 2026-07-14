# Slice 1: Gate Command and Git Evidence

Status: delivered
Slicing: Dev to Main Agent Evidence Gate
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-1-gate-command-git-evidence`
PR: #264

## Purpose

Introduce the first executable surface for the approved dev-to-main gate spec:
`chr gate dev-main --base <ref> --head <ref>`.

This slice focuses only on command routing and deterministic Git range
evidence. Report persistence, validation execution, policy, agent evidence
classification, Forgejo, and Woodpecker behavior remain later slices.

## Delivered Behavior

- Added the `gate` root command group.
- Added the `gate dev-main` subcommand.
- Added required `--base` and `--head` flags plus the normal project `--dir`
  resolution behavior.
- Added Git evidence collection for:
  - resolved base SHA;
  - resolved head SHA;
  - merge base;
  - commits in `base..head`;
  - changed files and file status;
  - short diff stat.
- Added CLI output for the collected evidence.

## Implementation Notes

The Git collection logic lives in `internal/chr` so it can be reused by later
service/reporting slices. The CLI adapter in `internal/chrapp` resolves the
project context and delegates to that package.

The command currently prints evidence directly. Later slices will add JSON and
Markdown report writing under `.chr/default/reports/gates/`.

## Contracts Added or Changed

- New CLI command: `chr gate dev-main`.
- New required flags:
  - `--base`
  - `--head`
- New package-level evidence types:
  - `GateGitEvidence`
  - `GateGitRef`
  - `GateGitCommit`
  - `GateGitChangedFile`

## Files of Interest

- `internal/chrapp/gate.go`
- `internal/chr/gate_git_evidence.go`
- `internal/chr/gate_git_evidence_test.go`
- `internal/chrapp/cli_test.go`

## Validation

- `go test ./internal/chr ./internal/chrapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Report persistence is intentionally deferred to Slice 2.
- Validation command execution is intentionally deferred to Slice 2.
- Agent-managed evidence classification is intentionally deferred to Slice 3.
- Forgejo/Woodpecker behavior is intentionally deferred to Slice 5.
- End-to-end gate execution against this repository is intentionally deferred
  until the later QA slice because the repository is not initialized as a
  Chryso-managed project at `.chr/`.
