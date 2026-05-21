# Markdown-First Authoring Sessions Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify Markdown-first authoring need.
- [x] Read `AGENTS.md` and `.agents/` delivery rules before planning.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Active Slice

Slice 2: Authoring Session Storage

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Authoring Session Model | Delivered | `feat/slice-1-authoring-session-model` | `feat(slice-1): authoring session model` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-1-authoring-session-model.md` | Repository PR #91. |
| Slice 2: Authoring Session Storage | Planned | `feat/slice-2-authoring-session-storage` | `feat(slice-2): authoring session storage` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-2-authoring-session-storage.md` | Depends on Slice 1. |
| Slice 3: Authoring Artifact Layout | Planned | `feat/slice-3-authoring-artifact-layout` | `feat(slice-3): authoring artifact layout` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-3-authoring-artifact-layout.md` | Depends on Slice 2. |
| Slice 4: Static Quiz Generation | Planned | `feat/slice-4-static-quiz-generation` | `feat(slice-4): static quiz generation` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-4-static-quiz-generation.md` | Depends on Slice 3. |
| Slice 5: Authoring Session CLI | Planned | `feat/slice-5-authoring-session-cli` | `feat(slice-5): authoring session cli` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-5-authoring-session-cli.md` | Depends on Slices 1-4. |
| Slice 6: Manual Promotion and Provenance | Planned | `feat/slice-6-manual-promotion-provenance` | `feat(slice-6): manual promotion provenance` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-6-manual-promotion-provenance.md` | Depends on Slice 5. |
| Slice 7: Authoring Docs Reference | Planned | `feat/slice-7-authoring-docs-reference` | `feat(slice-7): authoring docs reference` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-7-authoring-docs-reference.md` | Final docs pass for the slicing. |

## Slice 1: Authoring Session Model

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-authoring-session-model`
- Expected PR title: `feat(slice-1): authoring session model`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-1-authoring-session-model.md`

Candidate tasks:

- [x] Add authoring session ID, status, mode, target, topic, and timestamps.
- [x] Define input and output artifact references.
- [x] Define provenance links.
- [x] Define validation errors for invalid session inputs.
- [x] Add focused model tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: feat(slice-1): authoring session model
- T1.2: feat(slice-1): authoring session model
- T1.3: feat(slice-1): authoring session model
- T1.4: feat(slice-1): authoring session model
- T1.5: feat(slice-1): authoring session model
- T1.6: feat(slice-1): authoring session model
- T1.7: feat(slice-1): authoring session model

## Slice 2: Authoring Session Storage

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-authoring-session-storage`
- Expected PR title: `feat(slice-2): authoring session storage`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-2-authoring-session-storage.md`

Candidate tasks:

- [ ] Define the session directory layout.
- [ ] Render and parse `session.md`.
- [ ] Record transcript and artifact index paths.
- [ ] Preserve hand-edited Markdown where possible.
- [ ] Add repository tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1:
- T2.2:
- T2.3:
- T2.4:
- T2.5:
- T2.6:
- T2.7:

## Slice 3: Authoring Artifact Layout

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-authoring-artifact-layout`
- Expected PR title: `feat(slice-3): authoring artifact layout`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-3-authoring-artifact-layout.md`

Candidate tasks:

- [ ] Model artifact kinds for dumps, quizzes, drafts, specs, plans, slice sets, slices, trackers, and tickets.
- [ ] Resolve project-local output paths.
- [ ] Detect existing artifacts and collision behavior.
- [ ] Render artifact inventory entries.
- [ ] Add path and inventory tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1:
- T3.2:
- T3.3:
- T3.4:
- T3.5:
- T3.6:
- T3.7:

## Slice 4: Static Quiz Generation

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-static-quiz-generation`
- Expected PR title: `feat(slice-4): static quiz generation`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-4-static-quiz-generation.md`

Candidate tasks:

- [ ] Define target-specific question templates.
- [ ] Generate editable Markdown quiz artifacts.
- [ ] Link the generated quiz to its authoring session.
- [ ] Support dry-run rendering.
- [ ] Add deterministic rendering tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1:
- T4.2:
- T4.3:
- T4.4:
- T4.5:
- T4.6:
- T4.7:

## Slice 5: Authoring Session CLI

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-authoring-session-cli`
- Expected PR title: `feat(slice-5): authoring session cli`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-5-authoring-session-cli.md`

Candidate tasks:

- [ ] Add `author session new`.
- [ ] Add `author session list`.
- [ ] Add `author session show`.
- [ ] Add `author quiz new` or equivalent target-specific entrypoint.
- [ ] Wire commands through the existing `internal/chcapp` command package.
- [ ] Add CLI tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1:
- T5.2:
- T5.3:
- T5.4:
- T5.5:
- T5.6:
- T5.7:
- T5.8:

## Slice 6: Manual Promotion and Provenance

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-manual-promotion-provenance`
- Expected PR title: `feat(slice-6): manual promotion provenance`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-6-manual-promotion-provenance.md`

Candidate tasks:

- [ ] Support one conservative promotion path.
- [ ] Record source session and input artifact paths.
- [ ] Generate an editable target artifact.
- [ ] Avoid overwriting existing artifacts without an explicit flag.
- [ ] Add promotion tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1:
- T6.2:
- T6.3:
- T6.4:
- T6.5:
- T6.6:
- T6.7:

## Slice 7: Authoring Docs Reference

Pre-flight:

- Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
- Slicing: `Markdown-First Authoring Sessions`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-authoring-docs-reference`
- Expected PR title: `feat(slice-7): authoring docs reference`
- Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-7-authoring-docs-reference.md`

Candidate tasks:

- [ ] Add reference documentation for session fields and artifact layout.
- [ ] Add user-guide coverage for static quiz generation and promotion.
- [ ] Document hand-edited Markdown tolerance.
- [ ] Document future assisted-session and TUI compatibility boundaries.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1:
- T7.2:
- T7.3:
- T7.4:
- T7.5:
- T7.6:

## Validation

Per-slice validation target:

- `make verify`
