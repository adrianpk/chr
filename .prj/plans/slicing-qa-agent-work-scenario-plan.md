# QA Agent Work Scenario Slicing Plan

Status: active

## Purpose

Implement the first reproducible QA scenario for exercising Chryso's
agent-work flow end to end: reset a seeded project, run a single bug ticket,
generate a comments spec, generate a plan, materialize tracking and tickets,
and run the agent loop over the generated work.

The scenario should be easy to reset and inspect manually. It must not turn into
a large product sample before the CHR workflow itself is proven.

## Source

Primary sources:

- `.prj/specs/specs/qa-agent-work-scenario.md`
- `.agents/CONVENTIONS.md`
- `.agents/LINTING.md`
- `docs/reference/artifact-generation-backends.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop-execution-policy.md`
- `docs/reference/tui.md`

Conversation decisions:

- The QA workspace lives under `qa/` inside the `chr` repository.
- Only the reference project is versioned.
- Active and temporary QA project copies are ignored.
- Reset is intentionally destructive for the active project copy.
- The active project is a Git repository initialized by the reset flow.
- The first reference project is named `blog`.
- The blog project uses only the Go standard library.
- The initial seeded bug has an existing failing test.
- The comments feature is not implemented in the reference project.
- The comments spec must be generated from seeded project context, dumps,
  quizzes, and ADRs.
- ADRs live as reusable project documentation under `docs/explanation/adr/`.
- Test-based development and table-driven tests are baseline engineering
  preferences for both CHR and the QA reference project.

## Current State

Delivered on `dev`:

- Codex-backed spec, plan, and plan tracking generation.
- Codex prompt discovery for decision documents during artifact generation.
- Ticket execution and loop execution.
- Dash agent-work controls and progress display.
- CLI work commands for one ticket and all selected work.
- Project-local `.chr/default/...` artifact layout.

Remaining gaps:

- No versioned QA reference project exists for end-to-end manual testing.
- No reset/test Make targets exist for QA scenarios.
- No ignored `qa/project` or `qa/tmp` active workspace convention exists.
- No seeded blog project exists with `.chr` context, ADRs, bug ticket, and
  failing test.
- The repo agent conventions do not yet explicitly state the test-first
  preference for bug fixes.
- Generated tracking/ticket prompts do not yet strongly encode the test-based
  approach for implementation work.

## Target Model

The first QA scenario provides:

- a versioned reference under `qa/ref/001/blog`;
- ignored active and temporary workspaces under `qa/project/blog` and
  `qa/tmp/blog`;
- generic Make targets plus blog aliases;
- reset scripts that copy the reference and initialize an independent Git repo;
- a realistic small Go HTTP service using standard library routing;
- a known failing test for the seeded bug ticket;
- seeded `.chr` context for comments spec generation;
- project documentation including ADRs under Diataxis explanation docs;
- CHR and project agent conventions that prefer reproducing bugs with tests and
  table-driven tests for input/output behavior.

## Slice Map

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

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | qa workspace reset scaffold | `feat/slice-1-qa-workspace-reset-scaffold` | `feat(slice-1): qa workspace reset scaffold` | `.prj/reports/slices/qa-agent-work-scenario/slice-1-qa-workspace-reset-scaffold.md` |
| Slice 2 | blog reference project | `feat/slice-2-blog-reference-project` | `feat(slice-2): blog reference project` | `.prj/reports/slices/qa-agent-work-scenario/slice-2-blog-reference-project.md` |
| Slice 3 | seeded chr context and adr docs | `feat/slice-3-seeded-chr-context-adr-docs` | `feat(slice-3): seeded chr context adr docs` | `.prj/reports/slices/qa-agent-work-scenario/slice-3-seeded-chr-context-adr-docs.md` |
| Slice 4 | test based agent guidance | `feat/slice-4-test-based-agent-guidance` | `feat(slice-4): test based agent guidance` | `.prj/reports/slices/qa-agent-work-scenario/slice-4-test-based-agent-guidance.md` |
| Slice 5 | qa runbook and smoke validation | `feat/slice-5-qa-runbook-smoke-validation` | `feat(slice-5): qa runbook smoke validation` | `.prj/reports/slices/qa-agent-work-scenario/slice-5-qa-runbook-smoke-validation.md` |

## Slice 1: QA Workspace Reset Scaffold

Create the reusable QA workspace mechanics without the full blog reference
content.

Expected behavior:

- `qa/ref` is versionable.
- `qa/project` and `qa/tmp` are ignored by the parent repository.
- Generic Make targets exist for reset and test operations.
- Blog-specific aliases call the generic targets with the correct scenario.
- Reset logic moves any existing active project to `qa/tmp/blog`, replacing the
  previous temporary copy.
- Reset logic copies the reference to `qa/project/blog`.
- Reset logic initializes a Git repository and base commit in the active copy.
- Scripts operate on explicit paths and do not rely on current directory state.

Candidate tasks:

- T1.1 Add `qa/` ignore rules for generated project and temporary copies.
- T1.2 Add generic and blog-specific Make targets.
- T1.3 Add reset script under `scripts/qa/`.
- T1.4 Add test script under `scripts/qa/`.
- T1.5 Add minimal placeholder reference content required to prove reset.
- T1.6 Add tests or shell validation for reset behavior.
- T1.7 Add slice implementation report and run quality gates.

Out of scope:

- Full blog app implementation.
- Seeded `.chr` artifacts.
- CHR prompt changes.

## Slice 2: Blog Reference Project

Build the realistic initial Go blog service in the reference project.

Expected behavior:

- `qa/ref/001/blog` is a small Go module named `blog`.
- The app uses standard library HTTP routing.
- The code has a simple internal structure suitable for early real projects.
- Existing endpoints include `GET /posts` and `GET /posts/{slug}`.
- Unknown slug behavior is intentionally wrong and covered by a failing test.
- `make qa-reset-blog` produces an active project whose `go test ./...` fails
  because of the seeded bug.

Candidate tasks:

- T2.1 Add the blog Go module and package structure.
- T2.2 Add post listing and single-post HTTP behavior.
- T2.3 Add the intentional unknown-slug bug.
- T2.4 Add a failing test that documents the expected `404` behavior.
- T2.5 Add a concise blog project README.
- T2.6 Validate reset and expected failing test behavior.
- T2.7 Add slice implementation report and run quality gates.

Out of scope:

- Comments implementation.
- CHR seeded context.
- Bug fix.

## Slice 3: Seeded CHR Context And ADR Docs

Seed the blog reference project with the context required for CHR to generate
the comments spec, plan, tracking, and tickets.

Expected behavior:

- The reference project contains `.chr` root files only where required.
- User-owned CHR artifacts live under `.chr/default/...`.
- The project has metadata and a `comments` use case.
- Dumps and curated quiz notes describe the comments feature.
- No comments spec, plan, tracking, or generated tickets are present.
- A manual bug ticket is present and ready for agent development.
- ADRs live under `docs/explanation/adr/` and influence the comments feature.

Candidate tasks:

- T3.1 Add project and use-case metadata for `comments`.
- T3.2 Add comments dumps and curated quiz notes under `.chr/default/...`.
- T3.3 Add a ready bug ticket for the unknown-slug behavior.
- T3.4 Add ADR documentation under `docs/explanation/adr/`.
- T3.5 Ensure no generated comments spec/plan/tracking/tickets are seeded.
- T3.6 Validate CHR can open the active project after reset.
- T3.7 Add slice implementation report and run quality gates.

Out of scope:

- Implementing comments.
- Running Codex generation.
- Changing CHR prompt policy.

## Slice 4: Test Based Agent Guidance

Make the test-based development preference explicit in CHR and in generated work
units.

Expected behavior:

- `.agents/CONVENTIONS.md` states the baseline preference for reproducing bugs
  with failing tests when feasible.
- `.agents/CONVENTIONS.md` preserves table-driven tests as the default for
  input/output behavior.
- Advanced testing approaches remain use-case dependent.
- Generation prompts for plan tracking and tickets bias validation expectations
  toward test-based implementation when feasible.
- The blog reference project carries equivalent project-local guidance.

Candidate tasks:

- T4.1 Update CHR agent conventions with test-based guidance.
- T4.2 Update generated tracking/ticket prompt instructions with the same
      preference.
- T4.3 Add tests that assert prompt guidance is present where relevant.
- T4.4 Seed equivalent guidance in the blog reference project.
- T4.5 Add slice implementation report and run quality gates.

Out of scope:

- Enforcing TDD mechanically.
- Adding new test frameworks.
- Changing ticket execution runtime behavior.

## Slice 5: QA Runbook And Smoke Validation

Document and verify the manual QA flow.

Expected behavior:

- Documentation explains how to reset, run tests, open CHR, resolve the manual
  bug ticket, generate comments artifacts, and run the generated ticket loop.
- The runbook is concise and operational.
- Smoke validation proves reset creates an independent Git repo and expected
  failing baseline.
- Make targets and scripts are documented.

Candidate tasks:

- T5.1 Add a concise QA runbook in repository docs.
- T5.2 Update reference docs for QA Make targets.
- T5.3 Add smoke validation for reset, Git initialization, and expected failing
      blog tests.
- T5.4 Verify active project directories remain ignored.
- T5.5 Add slice implementation report and run quality gates.

Out of scope:

- Automating full Codex/TUI interaction.
- Building additional reference scenarios.
- Fixing the seeded bug or implementing comments in the reference.

## Completion Criteria

The slicing is complete when:

- `make qa-reset-blog` recreates the active blog project from reference state;
- `make qa-test-blog` runs `go test ./...` in the active project;
- the reset baseline has the expected failing bug test;
- the active project is an independent Git repo ignored by the parent repo;
- the active project contains seeded CHR context for comments generation;
- the active project contains ADRs under `docs/explanation/adr/`;
- CHR conventions and generated ticket guidance express the test-based
  preference;
- docs explain the manual QA flow.
