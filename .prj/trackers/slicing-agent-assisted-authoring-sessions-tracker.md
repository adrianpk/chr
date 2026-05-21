# Agent-Assisted Authoring Sessions Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify assisted authoring need.
- [x] Read `AGENTS.md` and `.agents/` delivery rules before planning.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Active Slice

Slice 1: Assisted Session Contract

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Assisted Session Contract | Planned | `feat/slice-1-assisted-session-contract` | `feat(slice-1): assisted session contract` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-1-assisted-session-contract.md` | First implementation slice. |
| Slice 2: Transcript Event Storage | Planned | `feat/slice-2-assisted-transcript-storage` | `feat(slice-2): assisted transcript storage` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-2-assisted-transcript-storage.md` | Depends on Slice 1. |
| Slice 3: Authoring Context Package | Planned | `feat/slice-3-assisted-context-package` | `feat(slice-3): assisted context package` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-3-assisted-context-package.md` | Depends on Slice 2. |
| Slice 4: Model Invocation Adapter | Planned | `feat/slice-4-assisted-model-adapter` | `feat(slice-4): assisted model adapter` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-4-assisted-model-adapter.md` | Depends on Slice 3. |
| Slice 5: Adaptive Quiz Loop | Planned | `feat/slice-5-assisted-quiz-loop` | `feat(slice-5): assisted quiz loop` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-5-assisted-quiz-loop.md` | Depends on Slice 4. |
| Slice 6: Answer Curation | Planned | `feat/slice-6-assisted-answer-curation` | `feat(slice-6): assisted answer curation` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-6-assisted-answer-curation.md` | Depends on Slice 5. |
| Slice 7: Assisted Artifact Synthesis | Planned | `feat/slice-7-assisted-artifact-synthesis` | `feat(slice-7): assisted artifact synthesis` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-7-assisted-artifact-synthesis.md` | Depends on Slice 6. |
| Slice 8: Assisted Authoring CLI | Planned | `feat/slice-8-assisted-authoring-cli` | `feat(slice-8): assisted authoring cli` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-8-assisted-authoring-cli.md` | Depends on Slices 1-7. |
| Slice 9: Assisted Authoring Docs | Planned | `feat/slice-9-assisted-authoring-docs` | `feat(slice-9): assisted authoring docs` | `.prj/reports/slices/agent-assisted-authoring-sessions/slice-9-assisted-authoring-docs.md` | Final docs pass for the slicing. |

## Slice 1: Assisted Session Contract

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-assisted-session-contract`
- Expected PR title: `feat(slice-1): assisted session contract`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-1-assisted-session-contract.md`

Candidate tasks:

- [ ] Add assisted session configuration.
- [ ] Define turn state and turn result models.
- [ ] Define user answer and assistant question models.
- [ ] Define completion and blocked reasons.
- [ ] Add focused model validation tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T1.1:
- T1.2:
- T1.3:
- T1.4:
- T1.5:
- T1.6:
- T1.7:

## Slice 2: Transcript Event Storage

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-assisted-transcript-storage`
- Expected PR title: `feat(slice-2): assisted transcript storage`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-2-assisted-transcript-storage.md`

Candidate tasks:

- [ ] Define transcript event types.
- [ ] Render transcript events to Markdown.
- [ ] Parse stored transcript events.
- [ ] Append events without rewriting unrelated session artifacts.
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

## Slice 3: Authoring Context Package

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-assisted-context-package`
- Expected PR title: `feat(slice-3): assisted context package`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-3-assisted-context-package.md`

Candidate tasks:

- [ ] Include agent identity for authoring analyst behavior.
- [ ] Include project and authoring session metadata.
- [ ] Include relevant `.chc/` artifacts and selected docs.
- [ ] Include transcript summary and unanswered gaps.
- [ ] Materialize deterministic Markdown context.
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

## Slice 4: Model Invocation Adapter

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-assisted-model-adapter`
- Expected PR title: `feat(slice-4): assisted model adapter`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-4-assisted-model-adapter.md`

Candidate tasks:

- [ ] Define model request and response envelopes.
- [ ] Add fake adapter for deterministic tests.
- [ ] Record request and response artifacts.
- [ ] Keep provider details behind an interface.
- [ ] Classify adapter failures.
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

## Slice 5: Adaptive Quiz Loop

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-assisted-quiz-loop`
- Expected PR title: `feat(slice-5): assisted quiz loop`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-5-assisted-quiz-loop.md`

Candidate tasks:

- [ ] Ask the next question.
- [ ] Accept one user answer.
- [ ] Allow limited follow-up depth.
- [ ] Stop on completion, user stop, blocked state, or turn limit.
- [ ] Persist each turn.
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

## Slice 6: Answer Curation

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-assisted-answer-curation`
- Expected PR title: `feat(slice-6): assisted answer curation`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-6-assisted-answer-curation.md`

Candidate tasks:

- [ ] Generate curated answers from transcript events.
- [ ] Preserve explicit decisions and open questions.
- [ ] Write a curated summary artifact.
- [ ] Update session provenance.
- [ ] Add deterministic tests with fake model output.
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

## Slice 7: Assisted Artifact Synthesis

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-assisted-artifact-synthesis`
- Expected PR title: `feat(slice-7): assisted artifact synthesis`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-7-assisted-artifact-synthesis.md`

Candidate tasks:

- [ ] Synthesize one conservative artifact path.
- [ ] Preserve source session and summary provenance.
- [ ] Write an editable target artifact.
- [ ] Avoid overwriting without an explicit flag.
- [ ] Add synthesis tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1:
- T7.2:
- T7.3:
- T7.4:
- T7.5:
- T7.6:
- T7.7:

## Slice 8: Assisted Authoring CLI

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-assisted-authoring-cli`
- Expected PR title: `feat(slice-8): assisted authoring cli`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-8-assisted-authoring-cli.md`

Candidate tasks:

- [ ] Add assisted session start.
- [ ] Add answer submission.
- [ ] Add resume/status display.
- [ ] Add finalize command.
- [ ] Include dry-run and fake-adapter test paths.
- [ ] Add CLI tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1:
- T8.2:
- T8.3:
- T8.4:
- T8.5:
- T8.6:
- T8.7:
- T8.8:

## Slice 9: Assisted Authoring Docs

Pre-flight:

- Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
- Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
- Slicing: `Agent-Assisted Authoring Sessions`
- Slice: `Slice 9`
- Expected branch: `feat/slice-9-assisted-authoring-docs`
- Expected PR title: `feat(slice-9): assisted authoring docs`
- Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-9-assisted-authoring-docs.md`

Candidate tasks:

- [ ] Update authoring reference.
- [ ] Update user guide.
- [ ] Document transcript, context, and generated artifact paths.
- [ ] Document model-boundary and persistence guarantees.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T9.1:
- T9.2:
- T9.3:
- T9.4:
- T9.5:
- T9.6:

## Validation

Per-slice validation target:

- `make verify`
