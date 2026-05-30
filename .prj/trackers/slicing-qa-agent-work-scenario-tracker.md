# QA Agent Work Scenario Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`

## Active Slice

Slicing name: QA Agent Work Scenario

Active slice number: Slice 1

Active slice short name: qa workspace reset scaffold

Expected branch name: `feat/slice-1-qa-workspace-reset-scaffold`

Expected PR title: `feat(slice-1): qa workspace reset scaffold`

Expected report path:
`.prj/reports/slices/qa-agent-work-scenario/slice-1-qa-workspace-reset-scaffold.md`

Prerequisite PRs or delivered work:

- `dev` includes agent-work, Codex generation, and ticket loop support.
- `dev` includes Codex decision-document discovery for generated artifacts.

## Slicing Checklist

- [x] Create slicing spec.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with repository quality gates.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: QA Workspace Reset Scaffold | Planned | `feat/slice-1-qa-workspace-reset-scaffold` | `feat(slice-1): qa workspace reset scaffold` | `.prj/reports/slices/qa-agent-work-scenario/slice-1-qa-workspace-reset-scaffold.md` | First implementation slice. |
| Slice 2: Blog Reference Project | Planned | `feat/slice-2-blog-reference-project` | `feat(slice-2): blog reference project` | `.prj/reports/slices/qa-agent-work-scenario/slice-2-blog-reference-project.md` | Depends on Slice 1. |
| Slice 3: Seeded CHC Context And ADR Docs | Planned | `feat/slice-3-seeded-chc-context-adr-docs` | `feat(slice-3): seeded chc context adr docs` | `.prj/reports/slices/qa-agent-work-scenario/slice-3-seeded-chc-context-adr-docs.md` | Depends on Slice 2. |
| Slice 4: Test Based Agent Guidance | Planned | `feat/slice-4-test-based-agent-guidance` | `feat(slice-4): test based agent guidance` | `.prj/reports/slices/qa-agent-work-scenario/slice-4-test-based-agent-guidance.md` | Can run after Slice 3. |
| Slice 5: QA Runbook And Smoke Validation | Planned | `feat/slice-5-qa-runbook-smoke-validation` | `feat(slice-5): qa runbook smoke validation` | `.prj/reports/slices/qa-agent-work-scenario/slice-5-qa-runbook-smoke-validation.md` | Final documentation and smoke validation. |

## Slice 1: QA Workspace Reset Scaffold

Candidate tasks:

- [ ] T1.1 Add `qa/` ignore rules for generated project and temporary copies.
- [ ] T1.2 Add generic and blog-specific Make targets.
- [ ] T1.3 Add reset script under `scripts/qa/`.
- [ ] T1.4 Add test script under `scripts/qa/`.
- [ ] T1.5 Add minimal placeholder reference content required to prove reset.
- [ ] T1.6 Add tests or shell validation for reset behavior.
- [ ] T1.7 Add slice implementation report and run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 2: Blog Reference Project

Candidate tasks:

- [ ] T2.1 Add the blog Go module and package structure.
- [ ] T2.2 Add post listing and single-post HTTP behavior.
- [ ] T2.3 Add the intentional unknown-slug bug.
- [ ] T2.4 Add a failing test that documents the expected `404` behavior.
- [ ] T2.5 Add a concise blog project README.
- [ ] T2.6 Validate reset and expected failing test behavior.
- [ ] T2.7 Add slice implementation report and run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 3: Seeded CHC Context And ADR Docs

Candidate tasks:

- [ ] T3.1 Add project and use-case metadata for `comments`.
- [ ] T3.2 Add comments dumps and curated quiz notes under `.chc/user/...`.
- [ ] T3.3 Add a ready bug ticket for the unknown-slug behavior.
- [ ] T3.4 Add ADR documentation under `docs/explanation/adr/`.
- [ ] T3.5 Ensure no generated comments spec/plan/tracking/tickets are seeded.
- [ ] T3.6 Validate CHC can open the active project after reset.
- [ ] T3.7 Add slice implementation report and run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 4: Test Based Agent Guidance

Candidate tasks:

- [ ] T4.1 Update CHC agent conventions with test-based guidance.
- [ ] T4.2 Update generated tracking/ticket prompt instructions with the same
      preference.
- [ ] T4.3 Add tests that assert prompt guidance is present where relevant.
- [ ] T4.4 Seed equivalent guidance in the blog reference project.
- [ ] T4.5 Add slice implementation report and run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 5: QA Runbook And Smoke Validation

Candidate tasks:

- [ ] T5.1 Add a concise QA runbook in repository docs.
- [ ] T5.2 Update reference docs for QA Make targets.
- [ ] T5.3 Add smoke validation for reset, Git initialization, and expected
      failing blog tests.
- [ ] T5.4 Verify active project directories remain ignored.
- [ ] T5.5 Add slice implementation report and run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Completion

The slicing is not complete.

Tuning needed: undecided.
