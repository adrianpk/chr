# Slice 5: Go Graph Context Provider

Status: delivered
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-5-context-gograph-provider`
PR: #82

## Purpose

Integrate `gograph` CLI as the first semantic Go code context provider without
introducing Tree-sitter, LSP, or daemon lifecycle dependencies.

## Delivered Behavior

The repository now has a `gograph` context provider that can:

- detect whether `gograph` is available;
- run `gograph build .` by default;
- run `gograph build . --precise` when requested;
- query `changes` and `hotspot --top 20`;
- query selected symbols with `context`, `source`, `callers`, `callees`,
  `tests`, and `deps`;
- render selected output for `code-context.md`;
- record source provenance;
- degrade gracefully when `gograph` is unavailable.

## Implementation Notes

The provider executes through a small command runner interface so tests do not
depend on a local `gograph` installation.

The provider intentionally materializes selected query output, not a complete
graph dump. The repository files, tests, and validation commands remain the
source of truth.

## Contracts Added or Changed

New internal contracts:

- `GoGraphContextOptions`
- `GoGraphContext`
- `GoGraphCommandResult`
- `GoGraphQueryResult`
- `GoGraphCommandRunner`
- `OSGoGraphCommandRunner`
- `LoadGoGraphContext`
- `RenderGoGraphContext`

## Files of Interest

- `internal/chr/context_gograph.go`
- `internal/chr/context_gograph_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chr -run 'TestLoadGoGraphContext'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice does not install `gograph` and does not decide when precise mode
should be selected automatically. It only exposes the provider contract and
bounded query behavior for later materialization and execution readiness slices.
