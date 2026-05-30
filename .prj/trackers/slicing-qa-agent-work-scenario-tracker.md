# QA Agent Work Scenario Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`

## Active Slice

Slicing name: QA Agent Work Scenario

Active slice number: Slice 5

Active slice short name: qa runbook and smoke validation

Expected branch name: `feat/slice-5-qa-runbook-smoke-validation`

Expected PR title: `feat(slice-5): qa runbook smoke validation`

Expected report path:
`.prj/reports/slices/qa-agent-work-scenario/slice-5-qa-runbook-smoke-validation.md`

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
| Slice 1: QA Workspace Reset Scaffold | Completed | `feat/slice-1-qa-workspace-reset-scaffold` | `feat(slice-1): qa workspace reset scaffold` | `.prj/reports/slices/qa-agent-work-scenario/slice-1-qa-workspace-reset-scaffold.md` | Merged into `dev`. |
| Slice 2: Blog Reference Project | Completed | `feat/slice-2-blog-reference-project` | `feat(slice-2): blog reference project` | `.prj/reports/slices/qa-agent-work-scenario/slice-2-blog-reference-project.md` | Merged into `dev`. |
| Slice 3: Seeded CHC Context And ADR Docs | Completed | `feat/slice-3-seeded-chc-context-adr-docs` | `feat(slice-3): seeded chc context adr docs` | `.prj/reports/slices/qa-agent-work-scenario/slice-3-seeded-chc-context-adr-docs.md` | Merged into `dev`. |
| Slice 4: Test Based Agent Guidance | Completed | `feat/slice-4-test-based-agent-guidance` | `feat(slice-4): test based agent guidance` | `.prj/reports/slices/qa-agent-work-scenario/slice-4-test-based-agent-guidance.md` | Merged into `dev`. |
| Slice 5: QA Runbook And Smoke Validation | Reviewing | `feat/slice-5-qa-runbook-smoke-validation` | `feat(slice-5): qa runbook smoke validation` | `.prj/reports/slices/qa-agent-work-scenario/slice-5-qa-runbook-smoke-validation.md` | PR pending. |

## Slice 1: QA Workspace Reset Scaffold

Candidate tasks:

- [x] T1.1 Add `qa/` ignore rules for generated project and temporary copies.
- [x] T1.2 Add generic and blog-specific Make targets.
- [x] T1.3 Add reset script under `scripts/qa/`.
- [x] T1.4 Add test script under `scripts/qa/`.
- [x] T1.5 Add minimal placeholder reference content required to prove reset.
- [x] T1.6 Add tests or shell validation for reset behavior.
- [x] T1.7 Add slice implementation report and run quality gates.

Task commits:

- `1e9fe77 chore(qa): ignore generated qa workspaces`
- `8f13091 build(qa): add scenario make targets`
- `71d1cd6 feat(qa): add scenario reset script`
- `6a04d33 feat(qa): add scenario test script`
- `a8c8147 docs(qa): seed blog reference placeholder`

Validation:

- `make qa-reset-blog`
- `make qa-test-blog`
- `make verify`
- `git diff --check`

## Slice 2: Blog Reference Project

Candidate tasks:

- [x] T2.1 Add the blog Go module and package structure.
- [x] T2.2 Add post listing and single-post HTTP behavior.
- [x] T2.3 Add the intentional unknown-slug bug.
- [x] T2.4 Add a failing test that documents the expected `404` behavior.
- [x] T2.5 Add a concise blog project README.
- [x] T2.6 Validate reset and expected failing test behavior.
- [x] T2.7 Add slice implementation report and run quality gates.

Task commits:

- `11e4605 feat(qa): add blog go module structure`
- `2025c4b feat(qa): add blog post endpoints`
- `cbc4837 test(qa): seed unknown post status bug`
- `e390c97 test(qa): cover blog post endpoints`
- `506badf docs(qa): document blog baseline`

Validation:

- `make qa-reset-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test expects `404`
- `make verify`
- `git diff --check`

## Slice 3: Seeded CHC Context And ADR Docs

Candidate tasks:

- [x] T3.1 Add project and use-case metadata for `comments`.
- [x] T3.2 Add comments dumps and curated quiz notes under `.chc/user/...`.
- [x] T3.3 Add a ready bug ticket for the unknown-slug behavior.
- [x] T3.4 Add ADR documentation under `docs/explanation/adr/`.
- [x] T3.5 Ensure no generated comments spec/plan/tracking/tickets are seeded.
- [x] T3.6 Validate CHC can open the active project after reset.
- [x] T3.7 Add slice implementation report and run quality gates.

Task commits:

- `629ac44 docs(qa): seed blog project metadata`
- `e0858ec docs(qa): seed comments discovery context`
- `fcc1262 docs(qa): seed unknown post bug ticket`
- `501c908 docs(qa): seed blog api adr`

Validation:

- `make qa-reset-blog`
- `go run ./cmd/chc ticket list --dir qa/project/blog`
- `go run ./cmd/chc spec list --dir qa/project/blog`
- `find qa/project/blog/.chc/user/specs qa/project/blog/.chc/user/plans -maxdepth 2 -type f 2>/dev/null | sort`
- `make qa-test-blog` expected failure: unknown post returns `500`, test expects `404`
- `make verify`
- `git diff --check`

## Slice 4: Test Based Agent Guidance

Candidate tasks:

- [x] T4.1 Update CHC agent conventions with test-based guidance.
- [x] T4.2 Update generated tracking/ticket prompt instructions with the same
      preference.
- [x] T4.3 Add tests that assert prompt guidance is present where relevant.
- [x] T4.4 Seed equivalent guidance in the blog reference project.
- [x] T4.5 Add slice implementation report and run quality gates.

Task commits:

- `b57a362 docs(agents): prefer test based bug fixes`
- `540bf43 feat(generation): add test based ticket guidance`
- `5757357 docs(qa): seed blog agent conventions`

Validation:

- `go test ./internal/chc`
- `make qa-reset-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test expects `404`
- `make verify`
- `git diff --check`

## Slice 5: QA Runbook And Smoke Validation

Candidate tasks:

- [x] T5.1 Add a concise QA runbook in repository docs.
- [x] T5.2 Update reference docs for QA Make targets.
- [x] T5.3 Add smoke validation for reset, Git initialization, and expected
      failing blog tests.
- [x] T5.4 Verify active project directories remain ignored.
- [x] T5.5 Add slice implementation report and run quality gates.

Task commits:

- Pending.

Validation:

- `make qa-smoke-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test expects `404`
- `make verify`
- `git diff --check`

## Completion

The slicing is not complete.

Tuning needed: undecided.
