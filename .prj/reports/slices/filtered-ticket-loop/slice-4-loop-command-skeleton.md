# Slice 4: Loop Command Skeleton

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-4-loop-command-skeleton`
PR: `#59`

## Purpose

Add a foreground `ticket loop` command skeleton that previews eligible loop
tickets without executing them.

## Delivered Behavior

The CLI now includes:

```sh
chr ticket loop --dry-run
```

The command:

- rejects positional arguments;
- requires `--dry-run` until execution slices are delivered;
- resolves project context using the existing project resolver;
- refreshes the resolved project's ticket index;
- queries eligible tickets through `TicketLoopQuery`;
- prints the planned loop selection.

Preview flags:

- `--status`;
- `--priority`;
- `--severity`;
- `--project-area`;
- `--tag`;
- `--type`;
- `--use-case`;
- `--order`.

## Implementation Notes

This slice intentionally keeps the command side-effect free beyond the existing
ticket index refresh behavior shared with ticket reads.

The command does not create loop run records, execution records, runtime
artifacts, result applications, or ticket mutations.

## Contracts Added or Changed

New CLI command:

- `ticket loop`

## Files of Interest

- `cmd/chr/main.go`
- `cmd/chr/main_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./cmd/chr -run 'TestTicketLoop|TestTicketList'`: pass
- `make verify`: pass

## Risks and Follow-ups

Non-dry-run execution is intentionally blocked until later slices add loop run
records, repeated execution, stop conditions, and reporting.
