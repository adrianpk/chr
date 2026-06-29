# Curated Dump Pairs Slicing Tracker

Status: delivered

Plan: `.prj/plans/slicing-curated-dump-pairs-plan.md`

## Active Slice

Slicing name: Curated Dump Pairs

Active slice number: none

Active slice short name: none

Expected branch name: none

Expected PR title: none

Expected report path:
none

## Slicing Checklist

- [x] Capture product definition without external project references.
- [x] Capture raw and curated sibling artifact behavior.
- [x] Capture LLM curation and deterministic fallback behavior.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Dump Pair Service Contract | Delivered | `feat/slice-1-dump-pair-service-contract` | `feat(slice-1): dump pair service contract` | `.prj/reports/slices/curated-dump-pairs/slice-1-dump-pair-service-contract.md` | PR #231. |
| Slice 2: Dump Pair Consumption And TUI | Delivered | `feat/slice-2-dump-pair-consumption-tui` | `feat(slice-2): dump pair consumption and tui` | `.prj/reports/slices/curated-dump-pairs/slice-2-dump-pair-consumption-tui.md` | PR #232. |

## Slice 1: Dump Pair Service Contract

Tasks:

- [x] T1.1 Add a dump curation provider contract for polished body, highlights,
      and warnings.
- [x] T1.2 Add an LLM-backed provider implementation using the existing LLM
      trace sink.
- [x] T1.3 Add deterministic fallback curation when LLM curation fails or is
      disabled.
- [x] T1.4 Change `CreateDump` to write curated and raw sibling files in one
      operation.
- [x] T1.5 Preserve title, use-case scope, slug collision handling, and service
      error behavior.
- [x] T1.6 Update CLI coverage for `chr dump new`.
- [x] T1.7 Add service tests for successful curation, fallback curation, raw
      sibling creation, frontmatter relationship fields, and filename collision
      handling.
- [x] T1.8 Add slice implementation report and validation evidence.

Task commits:

- T1.1: `02173b1` adds the dump curation provider contract.
- T1.2: `caeffc6` adds the OpenAI dump curation provider.
- T1.3: `0f6ef72` adds deterministic fallback curation.
- T1.4: `385a86a` creates curated/raw dump pairs from `CreateDump`.
- T1.5: `e74b919` preserves pair slug collision behavior.
- T1.6: `c9fa68f` covers CLI dump pair creation.
- T1.7: `62d5120` and `69438fc` cover curation provider and service paths.
- T1.8: this slice report commit.

Validation:

- `go test ./internal/chrapp` pass.
- `go test ./internal/chr ./internal/chrapp` pass.
- `make verify` fail on pre-existing strict lint issues outside this slice's
  files.

## Slice 2: Dump Pair Consumption And TUI

Tasks:

- [x] T2.1 Update workspace indexing or dump read-model mapping so raw dump
      siblings are distinguishable from curated dumps.
- [x] T2.2 Ensure context assembly uses curated dump content by default.
- [x] T2.3 Ensure raw dump siblings do not double-count as regular dumps in
      spec, plan, tracking, or ticket generation context.
- [x] T2.4 Update F3 Dumps to present curated dumps as primary items.
- [x] T2.5 Expose raw sibling metadata in the detail panel when available.
- [x] T2.6 Keep `alt+e` editing behavior focused on the selected curated dump.
- [x] T2.7 Add tests for index classification, context selection, and TUI
      rendering.
- [x] T2.8 Update user-facing documentation for dump pair behavior.
- [x] T2.9 Add slice implementation report and validation evidence.

Task commits:

- T2.1: workspace graph node metadata and dump classification helpers.
- T2.2: workspace context selection skips raw dump siblings.
- T2.3: spec and quiz supplemental context skip raw dump siblings.
- T2.4: dump list service filters raw siblings from primary views.
- T2.5: TUI dump detail renders raw sibling metadata.
- T2.6: external edit remains bound to selected curated dump paths.
- T2.7: graph, context, service, spec, and TUI tests cover dump pair consumption.
- T2.8: user guide and reference docs describe dump pair behavior.
- T2.9: slice report added.

Validation:

- `go test ./internal/chr ./internal/chrapp ./internal/chrtui` pass.
