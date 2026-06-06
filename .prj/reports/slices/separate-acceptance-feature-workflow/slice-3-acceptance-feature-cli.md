# Slice 3: Acceptance Feature CLI

Status: delivered
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-3-acceptance-feature-cli`
PR: `#296`

## Purpose

Expose the dedicated acceptance feature service from the command line before
adding TUI wiring.

## Delivered Behavior

The CLI now registers:

```sh
chc acceptance feature generate --use-case <case>
```

The command calls `GenerateAcceptanceFeature` for the selected scope. It prints
the operation message, status, generated feature path when available, and reason
text for non-generated outcomes.

## Implementation Notes

The command is under a top-level `acceptance` namespace rather than under
`spec`, matching the separated workflow. It uses the same workspace scope flags
as quiz, spec, and plan generation.

## Contracts Added or Changed

- Added `acceptance feature generate` CLI command.
- Added generated, needs-more-info, rejected, and not-applicable CLI output.
- Updated command reference docs to place feature generation outside spec
  generation.

## Files of Interest

- `internal/chcapp/acceptance_cli.go`
- `internal/chcapp/app.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/commands.md`

## Validation

- `go test ./internal/chcapp -run 'RootCommand|AcceptanceFeatureGenerate'` (pass)
- `go test ./internal/chc ./internal/chcapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Feature-round defaults were delivered in Slice 4.
- TUI action wiring was delivered in Slice 5.
