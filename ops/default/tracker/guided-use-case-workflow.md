# Guided Use-Case Workflow Tracker

Status: In progress; Slice 1 reviewing
Date: 2026-10-22
Delivery set: guided-use-case-workflow
Planning authorization: maintainer, 2026-10-22.
Approved: maintainer, 2026-10-22; slice map, tracker and implementation accepted.
Merge authority: maintainer, 2026-10-22; Looper may merge the six recorded slice PRs into `dev` with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 1, Guided capture and Git checkpoints; T1.1-T1.3 implemented; PR #430 reviewing, awaiting controller validation and integration.
Progress: 0 of 6 slices delivered; 3 of 18 tasks implemented. Slice 1 is in PR #430 and awaits controller validation and integration.
Controller: Looper; startup authorized, live state belongs to its checkpoint.

Spec: [Approved workflow](../spec/stable/guided-use-case-workflow.md)
Model: [State and checkpoint ownership](../spec/stable/guided-use-case-workflow-model.md)
Plan: [Scope, tasks and gates](../plan/guided-use-case-workflow.md)
Specification commit and planning baseline: `afd228595faac22c9008c84b05ad21cdc5a5d353`.

## Canonical slice map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Guided capture and Git checkpoints | feat/slice-1-guided-capture-checkpoints | feat(slice-1): capture guided use cases with Git checkpoints | ops/default/report/slices/guided-use-case-workflow/slice-1-capture-checkpoints.md |
| Slice 2 | Guided quiz and approved specification | feat/slice-2-guided-quiz-spec | feat(slice-2): guide clarification and approve generated specifications | ops/default/report/slices/guided-use-case-workflow/slice-2-quiz-spec.md |
| Slice 3 | Automatic work preparation and execution | feat/slice-3-guided-work-execution | feat(slice-3): prepare and start exact guided work | ops/default/report/slices/guided-use-case-workflow/slice-3-work-execution.md |
| Slice 4 | Web recovery, controls and results | feat/slice-4-guided-web-recovery | feat(slice-4): complete the resumable Web use-case journey | ops/default/report/slices/guided-use-case-workflow/slice-4-web-recovery.md |
| Slice 5 | Guided Telegram journey | feat/slice-5-guided-telegram | feat(slice-5): guide use-case development through Telegram | ops/default/report/slices/guided-use-case-workflow/slice-5-telegram.md |
| Slice 6 | Guided TUI journey | feat/slice-6-guided-tui | feat(slice-6): guide use-case development through the TUI | ops/default/report/slices/guided-use-case-workflow/slice-6-tui.md |

These are planned identities. Reports, branches and pull requests do not exist merely because the map records their paths or names. Read the plan for each task's full contract.

## Tasks

| Task | Required outcome | State | Commit | Validation |
| --- | --- | --- | --- | --- |
| T1.1 | Retained guided run, scope/intent, revisions and compatible authoring-store extension. | implemented | `d142d186c0c3775df57841b32e27021594576ffa` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T1.2 | Scoped Git checkpoints, immutable input reads, commit retention and lost-ack recovery. | implemented | `e9ef91a70a2c13b247edeba1748eee0906974b1e` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T1.3 | Web start/name/dump capture, explicit completion, readiness and resume. | implemented | `ad6db48d366beb9a7a4d04ea252a47175e6ff68b` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T2.1 | Committed-input quiz generation, bound answers, sufficient completion and sealing. | pending | pending | pending |
| T2.2 | Exact generated-specification approval, provenance, checkpoint and recovery. | pending | pending | pending |
| T2.3 | Web clarification/specification progression and complete artifact inspection. | pending | pending | pending |
| T3.1 | Committed-source planning, native materialization, contracts/tracker and work checkpoint. | pending | pending | pending |
| T3.2 | Exact policy-bound native preview/admission and automatic start with retained limits. | pending | pending | pending |
| T3.3 | Web preparation/execution state and real prerequisite/manual-operation regressions. | pending | pending | pending |
| T4.1 | Host recovery, client independence, stale-action safety and native pause/resume/cancel. | pending | pending | pending |
| T4.2 | Later-idea retention and truthful waiting/results from native delivery receipts. | pending | pending | pending |
| T4.3 | Complete Web journey/fault/browser evidence and user-facing documentation. | pending | pending | pending |
| T5.1 | Natural-language Telegram start and shared current-step routing. | pending | pending | pending |
| T5.2 | Concise Telegram progress/artifacts/controls/results and cross-interface continuation. | pending | pending | pending |
| T5.3 | Telegram transport/journey regressions, declared client evidence and documentation. | pending | pending | pending |
| T6.1 | Shared-service TUI journey input, navigation, artifacts and resume. | pending | pending | pending |
| T6.2 | TUI controls/progress/results and cross-adapter revision/authority parity. | pending | pending | pending |
| T6.3 | TUI/parity acceptance, documentation, cycle report and exact-candidate closure gate. | pending | pending | pending |

## Slice delivery

| Slice | State | PR | Integrated commit | Report evidence |
| --- | --- | --- | --- | --- |
| 1 | reviewing | #430 | pending | [Slice 1 report](../report/slices/guided-use-case-workflow/slice-1-capture-checkpoints.md) |
| 2 | pending | pending | pending | pending |
| 3 | pending | pending | pending | pending |
| 4 | pending | pending | pending | pending |
| 5 | pending | pending | pending | pending |
| 6 | pending | pending | pending | pending |

## Acceptance evidence

Criterion numbers correspond to the approved specification. The plan defines the complete ownership mapping; all evidence below remains pending until implementation and verification.

| Criterion | Owning tasks | State and evidence |
| --- | --- | --- |
| GUC-01 | T1.3, T2.1, T2.3, T5.1, T6.1 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |
| GUC-02 | T2.2, T3.1, T3.2 | pending |
| GUC-03 | T2.2 | pending |
| GUC-04 | T3.1, T3.2 | pending |
| GUC-05 | T1.1, T2.2, T3.3 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |
| GUC-06 | T1.1, T4.1, T5.1, T6.2, T6.3 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |
| GUC-07 | T2.3, T4.2, T5.2, T6.2 | pending |
| GUC-08 | T1.3, T2.1 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |
| GUC-09 | T2.2, T3.1, T4.1 | pending |
| GUC-10 | T3.2, T4.1, T4.3 | pending |
| GUC-11 | T4.1, T5.3, T6.2 | pending |
| GUC-12 | T3.3, T4.1, T6.2 | pending |
| GUC-13 | T1.2, T2.3, T3.1, T4.3 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |
| GUC-14 | T4.2, T4.3, T5.2, T6.2, T6.3 | pending |
| GUC-15 | T4.2, T5.2, T6.2 | pending |
| GUC-16 | T1.2, T3.1, T4.3 | partial: Slice 1 capture/checkpoint regressions passed; remaining owning tasks are pending. |

## Prerequisites and checkpoints

1. Approved specification/model: committed at the baseline above; available for planning.
2. Slice-map, implementation and Looper startup: approved on 2026-10-22. The maintainer separately delegated Rebase + Fast-forward merges for the six recorded slice PRs into `dev` after their required checks. No additional delivery scope or final-gate repair grant is included.
3. Shared operation/native owners: present at the planning baseline; runtime readiness is checked per guided run using effective project policy.
4. Initial storage and Git mechanism: bounded T1.1/T1.2 design work, including compatibility, checkpoint reachability, artifact-versus-code-base identity and preservation of unrelated index/worktree changes.
5. Web checkpoint: Slices 1-4 must be delivered and the Web journey verified before Slice 5 begins. Telegram must consume that implementation before TUI integration in Slice 6.
6. Actual client and live-provider evidence: not collected under planning. Local fixture results, browser interaction, user-operated Telegram acceptance and exhaustive scheduled certification remain distinct evidence classes.

## Validation ledger

Planning validation: links, approved source references, canonical slice-map equality, task/criterion coverage, privacy, terminology and staged whitespace are required before the planning commit.

Implementation validation: use the plan's focused selectors and the repository development gate. Proposed selectors must gain actual test coverage; zero matches or skipped required fixtures are not evidence. Record exact commands, candidates and outcomes in the task and slice report after running them.

Delivery-set gate: `make verify`, bounded normal guided-workflow regressions, focused race coverage for changed ownership, then `git diff --check`, on the exact final integrated `dev` candidate. The scheduled nightly retains exhaustive certification ownership.

| Evidence | Candidate | Outcome |
| --- | --- | --- |
| Slice development gates | PR #430 head (controller evidence) | Task regressions and scoped strict lint passed; final PR-head evidence is owned by the controller. |
| Web browser journey | pending | not run |
| Telegram actual-client journey | pending | not observed |
| TUI interaction and cross-adapter parity | pending | not run |
| Final delivery-set gate | pending | not run |
| Scheduled exhaustive certification | pending | not observed |

## Completion boundary

Planning artifacts do not establish implementation or delivery. Advance a task only with its actual commit and validation evidence; close a slice only after its canonical PR is merged and the integrated state is verified. Preserve native receipts and original limits during any recovery.

Review later ideas after the current guided development finishes. They never become extra work in this delivery set or in a user's active invocation merely because they were retained.
