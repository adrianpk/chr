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

Slice 6: Manual Promotion and Provenance

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Authoring Session Model | Delivered | `feat/slice-1-authoring-session-model` | `feat(slice-1): authoring session model` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-1-authoring-session-model.md` | Repository PR #91. |
| Slice 2: Authoring Session Storage | Delivered | `feat/slice-2-authoring-session-storage` | `feat(slice-2): authoring session storage` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-2-authoring-session-storage.md` | Repository PR #92. |
| Slice 3: Authoring Artifact Layout | Delivered | `feat/slice-3-authoring-artifact-layout` | `feat(slice-3): authoring artifact layout` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-3-authoring-artifact-layout.md` | Repository PR #93. |
| Slice 4: Static Quiz Generation | Delivered | `feat/slice-4-static-quiz-generation` | `feat(slice-4): static quiz generation` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-4-static-quiz-generation.md` | Repository PR #94. |
| Slice 5: Authoring Session CLI | Delivered | `feat/slice-5-authoring-session-cli` | `feat(slice-5): authoring session cli` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-5-authoring-session-cli.md` | Repository PR #95. |
| Slice 6: Manual Promotion and Provenance | Delivered | `feat/slice-6-manual-promotion-provenance` | `feat(slice-6): manual promotion provenance` | `.prj/reports/slices/markdown-first-authoring-sessions/slice-6-manual-promotion-provenance.md` | Repository PR #96. |
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

- [x] Define the session directory layout.
- [x] Render and parse `session.md`.
- [x] Record transcript and artifact index paths.
- [x] Preserve hand-edited Markdown where possible.
- [x] Add repository tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: feat(slice-2): authoring session storage
- T2.2: feat(slice-2): authoring session storage
- T2.3: feat(slice-2): authoring session storage
- T2.4: feat(slice-2): authoring session storage
- T2.5: feat(slice-2): authoring session storage
- T2.6: feat(slice-2): authoring session storage
- T2.7: feat(slice-2): authoring session storage

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

- [x] Model artifact kinds for dumps, quizzes, drafts, specs, plans, slice sets, slices, trackers, and tickets.
- [x] Resolve project-local output paths.
- [x] Detect existing artifacts and collision behavior.
- [x] Render artifact inventory entries.
- [x] Add path and inventory tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: feat(slice-3): authoring artifact layout
- T3.2: feat(slice-3): authoring artifact layout
- T3.3: feat(slice-3): authoring artifact layout
- T3.4: feat(slice-3): authoring artifact layout
- T3.5: feat(slice-3): authoring artifact layout
- T3.6: feat(slice-3): authoring artifact layout
- T3.7: feat(slice-3): authoring artifact layout

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

- [x] Define target-specific question templates.
- [x] Generate editable Markdown quiz artifacts.
- [x] Link the generated quiz to its authoring session.
- [x] Support dry-run rendering.
- [x] Add deterministic rendering tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: feat(slice-4): static quiz generation
- T4.2: feat(slice-4): static quiz generation
- T4.3: feat(slice-4): static quiz generation
- T4.4: feat(slice-4): static quiz generation
- T4.5: feat(slice-4): static quiz generation
- T4.6: feat(slice-4): static quiz generation
- T4.7: feat(slice-4): static quiz generation

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

- [x] Add `author session new`.
- [x] Add `author session list`.
- [x] Add `author session show`.
- [x] Add `author quiz new` or equivalent target-specific entrypoint.
- [x] Wire commands through the existing `internal/chcapp` command package.
- [x] Add CLI tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: feat(slice-5): authoring session cli
- T5.2: feat(slice-5): authoring session cli
- T5.3: feat(slice-5): authoring session cli
- T5.4: feat(slice-5): authoring session cli
- T5.5: feat(slice-5): authoring session cli
- T5.6: feat(slice-5): authoring session cli
- T5.7: feat(slice-5): authoring session cli
- T5.8: feat(slice-5): authoring session cli

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

- [x] Support one conservative promotion path.
- [x] Record source session and input artifact paths.
- [x] Generate an editable target artifact.
- [x] Avoid overwriting existing artifacts without an explicit flag.
- [x] Add promotion tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: feat(slice-6): manual promotion provenance
- T6.2: feat(slice-6): manual promotion provenance
- T6.3: feat(slice-6): manual promotion provenance
- T6.4: feat(slice-6): manual promotion provenance
- T6.5: feat(slice-6): manual promotion provenance
- T6.6: feat(slice-6): manual promotion provenance
- T6.7: feat(slice-6): manual promotion provenance

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
