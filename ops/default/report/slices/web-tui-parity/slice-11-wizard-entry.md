# Slice 11: Wizard entry and Dump in shared tabs

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-11-tui-wizard-entry`
PR: [#465](https://forge.adrianpk.com/adrianpk/chr/pulls/465)

## Purpose

Resolve G-005, G-006, G-007 and runtime G-008 through named wizard entry, retained
case/run navigation and exact capture continuation in the existing terminal shell.

## Delivered Behavior

- Cases exposes Wizard and Open workflow. One Name/Create calls
  StartNamedGuidedUseCase; there is no second name step or namespace/ID entry.
  Re-entry reads the selected case, preferring its bound run and otherwise the
  latest retained run with a stable identity tie-break. Ambiguous project-wide
  re-entry requires selecting a case. Unsaved case metadata is preserved.
- The normal tabs remain visible with case and stage progress. Domain stages map
  to Cases, Dumps, Quizzes, Specs, Plans and Board. Shared forms replace the
  detached editor/menu. Passive refresh preserves the tab and captured draft;
  explicit progression follows the current stage.
- Dump uses the existing body/draft/external-editor component. Continue to quiz
  submits the complete current content in one retained finish command. It does
  not perform a separate append or manual Save. Failed acknowledgement preserves
  the body and operation; unchanged retry reconciles the already committed input.
- Earlier and future stages are read-only. Continue returns to the current tab
  without submitting. Keyboard, palette and canonical-source editing routes
  cannot bypass wizard/manual separation. Current inputs retain their run,
  revision and question binding; a foreign captured run is rejected before IO.
  Explicit context clearing restores manual operation.
- Committed originals remain complete and scrollable after working-copy changes.
  Their bytes, reference and commit are checked against the canonical reader.
  Existing lifecycle, native approval, later ideas and result inspection retain
  their application-owned behavior and guards.

## Implementation Notes

T11.1: `d0516dc70ac86ac9491a1ff5076650d40d7c9297`.
T11.2: `a1ed8105573e40b269362b2559e902cf895e465e`.
T11.3: `7524663d82fafeb04658281cc8e14319e977f977`.
T11.4: `06d019e5d3117982539da0c14977ddd0c10d3e1c`.

T11.5 report introduction: `f51e7e3921876794d162350b36c22b16494d9188`.

The terminal adapter uses the existing named start and guided command owners.
Stage mapping and mode restrictions are presentation concerns; the terminal does
not own progression, checkpoint sealing, approvals, claims or attempt accounting.
The shared Dump editor carries a captured guided finish request and delegates
Ctrl+S to that owner. Other shared form submissions freeze the same command
identity and input instead of rebinding to a later observation.

The G-008 runtime check wraps the real owner only at acknowledgement delivery.
It verifies final edited capture, identical repeated request, unchanged retained
revision and run, one checkpoint, and one exact content occurrence in the
committed original. A separate real-owner read-only test compares original bytes,
reference and commit after working-copy replacement, then proves navigation and
blocked writers leave the canonical image unchanged. Provenance alone is not
used as evidence of content equivalence.

The production PTY uses the managed host and existing 45-second fixture /
40-second child limits. It waits for the Create button and the editor-specific
Continue shortcut before typing, avoiding matches in older stage text. When
background progression advances while Quizzes remains open, it explicitly returns
to the current stage before Pause. It retains the final-capture, approval,
checkpoint, later-idea, original and paused restart assertions. Only its owned
child is stopped after checks; no assertion or resource bound was weakened.

## Contracts Added or Changed

Terminal guided observations include verified originals for shared tab rendering.
Existing forms and Dump editing can carry a captured guided request. Presentation
mode guards prevent manual writes from forking wizard progression. No domain
command, persistence schema, progression semantics or durable authority changes.

## Files of Interest

- `internal/tui/wizard_entry.go`, `case_metadata_actions.go`: named entry.
- `internal/tui/guided.go`, `guided_input.go`, `guided_view.go`, shell files:
  retained binding, existing tabs, shared forms and stage presentation.
- `internal/tui/wizard_dump.go`, `dump_actions.go`: shared body editor and finish.
- `internal/tui/wizard_mode.go`, `command_palette.go`: read-only/manual guards.
- `internal/app/tui_adapters_guided.go`, `tui_guided_entry.go`: owner calls,
  deterministic case re-entry and verified original mapping.
- `internal/app/tui_wizard_*test.go`, `tui_guided_*test.go`,
  `internal/tui/wizard_*test.go`: real owner, terminal, retry and mode coverage.
- User Guide, TUI reference, action matrix and changelog: delivered interactions.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed for shared tabs, stale input,
  captured retry, mode guards and existing catalog frames.
- `go test ./internal/app -run '^Test(TUIWizard.*|GuidedUseCaseTUI.*|GuidedUseCaseParityNative)$' -count=1 -timeout=3m` — passed for named entry, real acknowledgement loss, exact committed bytes, read-only navigation, production PTY and retained native parity.
- `go test ./internal/core/authoring -run '^TestGuidedDumpContent$' -count=1` — passed for canonical finish sealing and replay.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.
- `git diff --check` — passed before task commits.

Final controller validation `COHC2KP7RZY5ONJTIC4UAQFWZJ` passed on clean
HEAD `6ee42aac701f6b2e6f479cbd639601965b29602b` in this order:

- `make verify`
- `go test -race ./internal/tui -count=1`
- `go test -race ./internal/app -run '^Test(TUIWizard.*|GuidedUseCaseTUI.*|GuidedUseCaseParityNative)$' -count=1 -timeout=3m`
- `git diff --check`

PR #465 merged into dev at that same commit. Acta merge event:
`01M64A1TKMCZ7A560E0HWSJQGM`. No checks were repeated for documentation closure.

## Risks and Follow-ups

None recorded.
