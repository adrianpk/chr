# Ticket Execution Context Hardening Slicing Tracker

Status: in progress

Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify context hardening need.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Read `AGENTS.md` and `.agents/` delivery rules before planning.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Context Composer Contract | Delivered | `feat/slice-1-context-composer-contract` | `feat(slice-1): context composer contract` | `.prj/reports/slices/ticket-execution-context-hardening/slice-1-context-composer-contract.md` | Repository PR #78 merged. |
| Slice 2: Agent Identity and Protocol Context | Delivered | `feat/slice-2-context-agent-protocols` | `feat(slice-2): context agent protocols` | `.prj/reports/slices/ticket-execution-context-hardening/slice-2-context-agent-protocols.md` | Repository PR #79 merged. |
| Slice 3: Project and Use-Case Document Context | Delivered | `feat/slice-3-context-project-documents` | `feat(slice-3): context project documents` | `.prj/reports/slices/ticket-execution-context-hardening/slice-3-context-project-documents.md` | Repository PR #80 merged. |
| Slice 4: Tree and Search Context | Delivered | `feat/slice-4-context-tree-search` | `feat(slice-4): context tree search` | `.prj/reports/slices/ticket-execution-context-hardening/slice-4-context-tree-search.md` | Repository PR #81 merged. |
| Slice 5: Go Graph Context Provider | Delivered | `feat/slice-5-context-gograph-provider` | `feat(slice-5): context gograph provider` | `.prj/reports/slices/ticket-execution-context-hardening/slice-5-context-gograph-provider.md` | Repository PR #82 merged. |
| Slice 6: Ralph Attempt Context | Reviewing | `feat/slice-6-context-ralph-attempts` | `feat(slice-6): context ralph attempts` | `.prj/reports/slices/ticket-execution-context-hardening/slice-6-context-ralph-attempts.md` | Repository PR #83. |
| Slice 7: Docker-Compatible Context Materialization | Planned | `feat/slice-7-context-docker-materialization` | `feat(slice-7): context docker materialization` | `.prj/reports/slices/ticket-execution-context-hardening/slice-7-context-docker-materialization.md` | Runtime artifact layout. |
| Slice 8: Execution Readiness and Handoff | Planned | `feat/slice-8-context-readiness-handoff` | `feat(slice-8): context readiness handoff` | `.prj/reports/slices/ticket-execution-context-hardening/slice-8-context-readiness-handoff.md` | Readiness and evidence links. |

## Slice 1: Context Composer Contract

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-context-composer-contract`
- Expected PR title: `feat(slice-1): context composer contract`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-1-context-composer-contract.md`

Candidate tasks:

- [x] Add composer input and output models.
- [x] Define source inventory entries.
- [x] Define materialized artifact paths.
- [x] Define readiness status and warnings.
- [x] Define invariant versus fresh source classification.
- [x] Add focused tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `95b1186`
- T1.2: `95b1186`
- T1.3: `95b1186`
- T1.4: `95b1186`
- T1.5: `95b1186`
- T1.6: `95b1186`
- T1.7: `docs(slice-1): record context composer report`
- T1.8: `95b1186`

## Slice 2: Agent Identity and Protocol Context

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-context-agent-protocols`
- Expected PR title: `feat(slice-2): context agent protocols`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-2-context-agent-protocols.md`

Candidate tasks:

- [x] Define default Go/software/infrastructure agent identity.
- [x] Load applicable root and nested `AGENTS.md` files.
- [x] Load selected `.chc/` workflow and runtime context.
- [x] Render `agent.md`.
- [x] Render `protocols.md`.
- [x] Record source provenance.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `e46fe56`
- T2.2: `e46fe56`
- T2.3: `e46fe56`
- T2.4: `e46fe56`
- T2.5: `e46fe56`
- T2.6: `e46fe56`
- T2.7: `docs(slice-2): record agent protocol report`
- T2.8: `e46fe56`

## Slice 3: Project and Use-Case Document Context

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-context-project-documents`
- Expected PR title: `feat(slice-3): context project documents`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-3-context-project-documents.md`

Candidate tasks:

- [x] Summarize project identity and validation commands.
- [x] Follow ticket references to related documents.
- [x] Render `project.md`.
- [x] Render `documents.md`.
- [x] Record included sources in `source-inventory.md`.
- [x] Avoid implicit conversation history.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `b29ac1d`
- T3.2: `b29ac1d`
- T3.3: `b29ac1d`
- T3.4: `b29ac1d`
- T3.5: `b29ac1d`
- T3.6: `b29ac1d`
- T3.7: `docs(slice-3): record project document report`
- T3.8: `b29ac1d`

## Slice 4: Tree and Search Context

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-context-tree-search`
- Expected PR title: `feat(slice-4): context tree search`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-4-context-tree-search.md`

Candidate tasks:

- [x] Capture bounded filesystem tree output.
- [x] Derive text-search terms from ticket and validation context.
- [x] Run targeted text searches.
- [x] Render tree and search sections into `code-context.md`.
- [x] Include relevant source, test, doc, and `.chc/` paths.
- [x] Keep output deterministic enough for tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `49381ac`
- T4.2: `49381ac`
- T4.3: `49381ac`
- T4.4: `49381ac`
- T4.5: `49381ac`
- T4.6: `49381ac`
- T4.7: `docs(slice-4): record tree search report`
- T4.8: `49381ac`

## Slice 5: Go Graph Context Provider

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-context-gograph-provider`
- Expected PR title: `feat(slice-5): context gograph provider`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-5-context-gograph-provider.md`

Candidate tasks:

- [x] Detect whether `gograph` is available.
- [x] Run `gograph build .` by default.
- [x] Support optional `gograph build . --precise`.
- [x] Query selected source, callers, callees, tests, deps, changes, and hotspots.
- [x] Render selected output into `code-context.md`.
- [x] Degrade gracefully when `gograph` is unavailable.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `5ef69cb`
- T5.2: `5ef69cb`
- T5.3: `5ef69cb`
- T5.4: `5ef69cb`
- T5.5: `5ef69cb`
- T5.6: `5ef69cb`
- T5.7: `docs(slice-5): record gograph provider report`
- T5.8: `5ef69cb`

## Slice 6: Ralph Attempt Context

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-context-ralph-attempts`
- Expected PR title: `feat(slice-6): context ralph attempts`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-6-context-ralph-attempts.md`

Candidate tasks:

- [x] Load previous validation artifact.
- [x] Summarize failed commands and repeated failure signatures.
- [x] Include previous changed files and result status.
- [x] Render `previous-attempt.md`.
- [x] Link previous attempt context from the next `context.md`.
- [x] Keep conversation history out unless materialized as an artifact.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `8328d87`
- T6.2: `8328d87`
- T6.3: `8328d87`
- T6.4: `8328d87`
- T6.5: `8328d87`
- T6.6: `8328d87`
- T6.7: `docs(slice-6): record ralph attempt report`
- T6.8: `8328d87`

## Slice 7: Docker-Compatible Context Materialization

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-context-docker-materialization`
- Expected PR title: `feat(slice-7): context docker materialization`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-7-context-docker-materialization.md`

Candidate tasks:

- [ ] Write `context.md` under `run_artifact_dir`.
- [ ] Write sibling context artifacts under `run_artifact_dir/context/`.
- [ ] Ensure paths are visible through the runtime mount.
- [ ] Keep `CHC_WORKER_CONTEXT_PATH` as the main Codex input.
- [ ] Add dry-run visibility for included sources.
- [ ] Test path rendering and artifact layout.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending
- T7.7: pending
- T7.8: pending

## Slice 8: Execution Readiness and Handoff

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
- Slicing: `Ticket Execution Context Hardening`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-context-readiness-handoff`
- Expected PR title: `feat(slice-8): context readiness handoff`
- Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-8-context-readiness-handoff.md`

Candidate tasks:

- [ ] Define required context fields.
- [ ] Report missing ticket, protocol, validation, or runtime context.
- [ ] Record context completeness and warnings.
- [ ] Link context artifacts from execution records or reports.
- [ ] Document remaining worktree and PAT-backed push requirements.
- [ ] Add focused tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending
- T8.2: pending
- T8.3: pending
- T8.4: pending
- T8.5: pending
- T8.6: pending
- T8.7: pending
- T8.8: pending

## Validation Record

- Planning validation: `make verify` pass.

## Notes

- This slicing is compatible with the existing Docker runtime boundary.
- Context composition stays outside Docker.
- Codex consumes the materialized `context.md` inside Docker.
- The mounted workspace model remains the initial runtime mode.
- Ephemeral worktree or clone execution remains a later runtime-workspace concern.
