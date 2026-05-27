# Curated Dump Pairs Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-curated-dump-pairs-plan.md`

## Active Slice

Slicing name: Curated Dump Pairs

Active slice number: Slice 1

Active slice short name: dump pair service contract

Expected branch name: `feat/slice-1-dump-pair-service-contract`

Expected PR title: `feat(slice-1): dump pair service contract`

Expected report path:
`.prj/reports/slices/curated-dump-pairs/slice-1-dump-pair-service-contract.md`

## Slicing Checklist

- [x] Capture product definition without external project references.
- [x] Capture raw and curated sibling artifact behavior.
- [x] Capture LLM curation and deterministic fallback behavior.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [ ] Commit plan and tracker directly to `dev`.
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Dump Pair Service Contract | Pending | `feat/slice-1-dump-pair-service-contract` | `feat(slice-1): dump pair service contract` | `.prj/reports/slices/curated-dump-pairs/slice-1-dump-pair-service-contract.md` | Service, LLM curation, fallback, CLI behavior. |
| Slice 2: Dump Pair Consumption And TUI | Pending | `feat/slice-2-dump-pair-consumption-tui` | `feat(slice-2): dump pair consumption and tui` | `.prj/reports/slices/curated-dump-pairs/slice-2-dump-pair-consumption-tui.md` | Workspace index, context assembly, TUI presentation. |

## Slice 1: Dump Pair Service Contract

Tasks:

- [ ] T1.1 Add a dump curation provider contract for polished body, highlights,
      and warnings.
- [ ] T1.2 Add an LLM-backed provider implementation using the existing LLM
      adapter and trace sink.
- [ ] T1.3 Add deterministic fallback curation when LLM curation fails or is
      disabled.
- [ ] T1.4 Change `CreateDump` to write curated and raw sibling files in one
      operation.
- [ ] T1.5 Preserve title, use-case scope, slug collision handling, and service
      error behavior.
- [ ] T1.6 Update CLI coverage for `chc dump new`.
- [ ] T1.7 Add service tests for successful curation, fallback curation, raw
      sibling creation, frontmatter relationship fields, and filename collision
      handling.
- [ ] T1.8 Add slice implementation report and validation evidence.

Task commits:

- T1.1: pending
- T1.2: pending
- T1.3: pending
- T1.4: pending
- T1.5: pending
- T1.6: pending
- T1.7: pending
- T1.8: pending

Validation:

- Pending.

## Slice 2: Dump Pair Consumption And TUI

Tasks:

- [ ] T2.1 Update workspace indexing or dump read-model mapping so raw dump
      siblings are distinguishable from curated dumps.
- [ ] T2.2 Ensure context assembly uses curated dump content by default.
- [ ] T2.3 Ensure raw dump siblings do not double-count as regular dumps in
      spec, plan, tracking, or ticket generation context.
- [ ] T2.4 Update F3 Dumps to present curated dumps as primary items.
- [ ] T2.5 Expose raw sibling metadata in the detail panel when available.
- [ ] T2.6 Keep `alt+e` editing behavior focused on the selected curated dump.
- [ ] T2.7 Add tests for index classification, context selection, and TUI
      rendering.
- [ ] T2.8 Update user-facing documentation for dump pair behavior.
- [ ] T2.9 Add slice implementation report and validation evidence.

Task commits:

- T2.1: pending
- T2.2: pending
- T2.3: pending
- T2.4: pending
- T2.5: pending
- T2.6: pending
- T2.7: pending
- T2.8: pending
- T2.9: pending

Validation:

- Pending.
