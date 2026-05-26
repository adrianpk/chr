# CLI TUI Surface Parity Slicing Plan

Status: active

## Purpose

Align the CLI and TUI as first-class operation surfaces over the same
application service logic.

The goal is to make every real TUI operation auditable against an equivalent
CLI operation, or an explicit documented gap. The CLI is not a legacy or
secondary compatibility surface. It must be usable by operators and agents as
part of normal ChrysoCoord workflows.

## Source

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- Existing `internal/chctui` keybindings and command palette actions
- Existing `internal/chcapp` CLI commands and TUI adapters
- Existing service files under `internal/chcapp/service_*.go`
- Conversation decisions about CLI and TUI parity

Conversation decisions:

- CLI commands are active product surface, not legacy.
- Each TUI function should be operable from CLI when the action is meaningful
  outside an interactive terminal.
- CLI and TUI should share the same service logic.
- A granular research pass must identify all parity gaps before broad
  implementation.
- Documentation must distinguish current behavior from desired parity.

## Current State

The TUI has grown into a broad operational console:

- project metadata editing;
- use-case metadata editing;
- dump creation and editing;
- quiz lifecycle actions;
- spec generation;
- plan generation, regeneration, refresh, tracking, and ticket creation;
- ticket board actions;
- work start and ready-loop actions;
- command palette dispatch.

The CLI already exposes many durable workflows, but parity has not been
systematically verified. Some TUI actions may be missing CLI equivalents, and
some CLI commands may not route through the same service logic as TUI adapters.

## Target Model

The target model treats CLI and TUI as equivalent operation surfaces:

- TUI emits interactive intent.
- CLI emits scripted/operator intent.
- Both route through shared application services.
- Documentation lists the current parity status for every action.
- Missing parity is explicit and actionable through planned tickets or slices.

## Slice Map

Slicing name: CLI TUI Surface Parity

Active slice number: Slice 2

Active slice short name: service boundaries

Expected branch name: `feat/slice-2-service-boundaries`

Expected PR title: `feat(slice-2): service boundaries`

Expected report path:
`.prj/reports/slices/cli-tui-surface-parity/slice-2-service-boundaries.md`

Prerequisite PRs or delivered work:

- TUI operational console.
- TUI dump, quiz, spec, plan, board, and work action surfaces.
- Plan tracking and ticket materialization from the Plans tab.
- Current documentation refresh for TUI and command surfaces.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | inventory | `feat/slice-1-cli-tui-inventory` | `feat(slice-1): inventory` | `.prj/reports/slices/cli-tui-surface-parity/slice-1-inventory.md` |
| Slice 2 | service boundaries | `feat/slice-2-service-boundaries` | `feat(slice-2): service boundaries` | `.prj/reports/slices/cli-tui-surface-parity/slice-2-service-boundaries.md` |
| Slice 3 | missing cli actions | `feat/slice-3-missing-cli-actions` | `feat(slice-3): missing cli actions` | `.prj/reports/slices/cli-tui-surface-parity/slice-3-missing-cli-actions.md` |
| Slice 4 | parity tests | `feat/slice-4-parity-tests` | `feat(slice-4): parity tests` | `.prj/reports/slices/cli-tui-surface-parity/slice-4-parity-tests.md` |
| Slice 5 | parity docs | `feat/slice-5-parity-docs` | `feat(slice-5): parity docs` | `.prj/reports/slices/cli-tui-surface-parity/slice-5-parity-docs.md` |

## Slice 1: Inventory

Create the parity matrix before changing runtime behavior.

Expected behavior:

- The repository contains a documented matrix of TUI actions, CLI equivalents,
  service paths, and parity status.
- Each TUI direct keybinding and command palette command is represented.
- Each current CLI command relevant to the same domain is represented.
- Each action has one status:
  - `matched`;
  - `missing_cli`;
  - `missing_tui`;
  - `service_divergence`;
  - `intentionally_interactive`;
  - `not_implemented`.

Out of scope:

- Adding new CLI commands.
- Refactoring service boundaries.
- Changing TUI behavior.

## Slice 2: Service Boundaries

Align service routing for already matched actions where CLI and TUI diverge.

Expected behavior:

- Matched operations call shared application services.
- Duplicated command logic is moved behind service methods where needed.
- Existing CLI and TUI behavior remains unchanged.

Out of scope:

- New user-facing commands unless required to preserve existing behavior.
- Documentation-only parity claims without service evidence.

## Slice 3: Missing CLI Actions

Add CLI commands for operational TUI actions that are meaningful from scripts or
agent workflows.

Expected behavior:

- Project/use-case metadata operations have CLI entrypoints where missing.
- Dump, quiz, spec, plan, tracking, ticket, and work actions have CLI
  equivalents where appropriate.
- New commands call shared services.

Out of scope:

- Purely interactive layout operations such as pane focus or mouse resizing.
- Cosmetic TUI-only behavior.

## Slice 4: Parity Tests

Lock parity behavior with tests.

Expected behavior:

- New CLI commands have CLI tests.
- Shared service paths have tests where behavior is non-trivial.
- TUI adapters have targeted tests for dispatch into shared service behavior.
- `make verify` passes.

Out of scope:

- Broad visual snapshot testing.

## Slice 5: Parity Docs

Publish the final parity contract.

Expected behavior:

- User guide, command reference, TUI reference, and architecture explanation
  describe CLI and TUI as first-class operation surfaces.
- The parity matrix reflects implemented and intentionally deferred gaps.
- No active CLI command is documented as legacy.

Out of scope:

- Implementing remaining gaps discovered during documentation.

## Delivery Order

Implement slices sequentially:

1. Slice 1 creates the inventory and gap map.
2. Slice 2 aligns service boundaries for existing behavior.
3. Slice 3 adds missing CLI entrypoints.
4. Slice 4 locks parity with tests.
5. Slice 5 updates the product documentation contract.

No slice should start before the previous slice is reviewed, merged to `dev`,
and validated.
