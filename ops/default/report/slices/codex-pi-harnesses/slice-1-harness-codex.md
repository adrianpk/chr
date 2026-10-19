# Slice 1: Neutral contracts and Codex extraction

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `refactor/slice-1-harness-codex`
PR: pending

## Purpose

Extract the existing Codex generation and worker transport behind neutral
contracts while keeping workflow authority and semantic validation in core.

## Delivered Behavior

T1.1 supplies neutral technical contracts and bounded process supervision.
T1.2 adds the Codex generation and worker adapter, with private run bundles,
legacy executable/model/timeout resolution and unchanged command flags.
Caller migration remains in progress.

## Implementation Notes

The process owner drains stdout and stderr to private bounded spools. Stream
overflow cancels the owned process group; result overflow rejects partial data.
Diagnostic excerpts are sanitized and capped. Unknown usage remains unknown.

## Contracts Added or Changed

`internal/harness` owns technical contracts and byte ceilings.
`internal/harness/process` owns subprocess lifetime and stream bounds.

## Files of Interest

- `internal/harness/contracts.go`
- `internal/harness/process/executor.go`
- `internal/harness/process/diagnostics.go`

## Validation

Passed during T1.1:

- `go test ./internal/harness/... -run '^TestAgentRunProcess' -count=1 -timeout=30s`

- `go test ./internal/harness/... -run '^(TestHarnessCodex|TestAgentRunProcess)' -count=1 -timeout=30s`

The configured slice checks remain pending.

## Risks and Follow-ups

T1.3 caller migration remains in progress. No PR is ready yet.
