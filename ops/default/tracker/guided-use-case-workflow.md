# Guided Use-Case Workflow Tracker

Status: In progress; Slice 4 active
Date: 2026-10-22
Delivery set: guided-use-case-workflow
Planning authorization: maintainer, 2026-10-22.
Approved: maintainer, 2026-10-22; slice map, tracker and implementation accepted.
Merge authority: maintainer, 2026-10-22; Looper may merge the six recorded slice PRs into `dev` with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 4, Web recovery, controls and results; all three tasks are complete, canonical PR #433 awaits controller validation and integration.
Progress: 3 of 6 slices delivered; 12 of 18 tasks completed. Slice 4 implementation and focused Web acceptance are complete; canonical PR integration remains pending.
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
| T1.1 | Retained guided run, scope/intent, revisions and compatible authoring-store extension. | done | `d142d186c0c3775df57841b32e27021594576ffa` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T1.2 | Scoped Git checkpoints, immutable input reads, commit retention and lost-ack recovery. | done | `e9ef91a70a2c13b247edeba1748eee0906974b1e` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T1.3 | Web start/name/dump capture, explicit completion, readiness and resume. | done | `ad6db48d366beb9a7a4d04ea252a47175e6ff68b` | Focused capture/checkpoint regressions passed; see Slice 1 report and canonical PR for exact-HEAD controller evidence. |
| T2.1 | Committed-input quiz generation, bound answers, sufficient completion and sealing. | done | `a255f0be01f0a55916f95b965613eccc38e04131` | Focused quiz regressions and race checks passed; committed inputs, duplicate/stale replies, sufficiency and owner recovery. |
| T2.2 | Exact generated-specification approval, provenance, checkpoint and recovery. | done | `5d67648370079e7a373d0def7664015259d94905` | Focused specification regressions passed; exact approval, lost draft/approval acknowledgements, retained seal and manual mutation compatibility. |
| T2.3 | Web clarification/specification progression and complete artifact inspection. | done | `5ca7c06550c4e44ec6a2fd3ae10ab93309e9a18b` | Focused capture/quiz/specification HTTP and view regressions passed, including race mode and committed-original inspection. |
| T3.1 | Committed-source planning, native materialization, contracts/tracker and work checkpoint. | done | `cf45d5255edd6834df83ce7c82a3ed7f3425fd88` | Focused preparation, native journal recovery and standalone plan regressions passed; focused preparation race checks and strict scoped lint passed. |
| T3.2 | Exact policy-bound native preview/admission and automatic start with retained limits. | done | `a740689accce48f067b66db400a6b6f37698da41` | Focused preparation/admission, native manual-work and Telegram admission regressions passed; focused admission race checks and strict scoped lint passed. |
| T3.3 | Web preparation/execution state and real prerequisite/manual-operation regressions. | done | `4eaf8bd2db4126d8d08596daa795fec52acb50a9` | Focused HTTP/native progress, missing prerequisites, changed-policy and manual plan regressions passed; focused race checks and strict scoped lint passed. |
| T4.1 | Host recovery, client independence, stale-action safety and native pause/resume/cancel. | done | `0ea3ff90e3d4bc5e35b36b7ad188cde3c8cb0274` | Focused recovery/control, native stop/closing and Web regressions passed in normal and race modes; scoped strict lint passed. See the Slice 4 report. |
| T4.2 | Later-idea retention and truthful waiting/results from native delivery receipts. | done | `c6fd84c0197842011a2ae6cdcececaf739894c56` | Focused idea/result normal and race checks and scoped strict lint passed; native delivery/closing receipts, corruption and read-only inspection covered. |
| T4.3 | Complete Web journey/fault/browser evidence and user-facing documentation. | done | `0fa7cec16dc3e6b02f3db1d593e3cf4c0081da69` | HTTP journey, actual lost Git commit responses, focused race, owned Chromium journey and scoped strict lint passed. User Guide and changelog updated. |
| T5.1 | Natural-language Telegram start and shared current-step routing. | pending | pending | pending |
| T5.2 | Concise Telegram progress/artifacts/controls/results and cross-interface continuation. | pending | pending | pending |
| T5.3 | Telegram transport/journey regressions, declared client evidence and documentation. | pending | pending | pending |
| T6.1 | Shared-service TUI journey input, navigation, artifacts and resume. | pending | pending | pending |
| T6.2 | TUI controls/progress/results and cross-adapter revision/authority parity. | pending | pending | pending |
| T6.3 | TUI/parity acceptance, documentation, cycle report and exact-candidate closure gate. | pending | pending | pending |

## Slice delivery

| Slice | State | PR | Integrated commit | Report evidence |
| --- | --- | --- | --- | --- |
| 1 | delivered | #430 | `245cb6f306aabb5e7e9877491aa2ea26174299e6` | [Slice 1 report](../report/slices/guided-use-case-workflow/slice-1-capture-checkpoints.md) |
| 2 | delivered | #431 | `2246651001dd6002f783fd86d0dcc13f5bef02e6` | [Slice 2 report](../report/slices/guided-use-case-workflow/slice-2-quiz-spec.md) |
| 3 | delivered | #432 | `dfcfae6e787742d976f99d40f292a7e047f5252a` | [Slice 3 report](../report/slices/guided-use-case-workflow/slice-3-work-execution.md) |
| 4 | reviewing | [#433](https://forge.adrianpk.com/adrianpk/chr/pulls/433) | pending | [Slice 4 report](../report/slices/guided-use-case-workflow/slice-4-web-recovery.md) |
| 5 | pending | pending | pending | pending |
| 6 | pending | pending | pending | pending |

## Acceptance evidence

Criterion numbers correspond to the approved specification. The plan defines the complete ownership mapping; all evidence below remains pending until implementation and verification.

| Criterion | Owning tasks | State and evidence |
| --- | --- | --- |
| GUC-01 | T1.3, T2.1, T2.3, T5.1, T6.1 | partial: Slice 2 quiz/specification and HTTP evidence passed; later owning tasks remain pending. |
| GUC-02 | T2.2, T3.1, T3.2 | implemented: an eligible HTTP answer automatically approves, prepares, seals and activates exactly its native set; a real worker reaches the native review boundary without another start prompt. |
| GUC-03 | T2.2 | implemented: exact generated revision/hash approval with initiating actor, run and policy; focused mutation and recovery regressions passed. |
| GUC-04 | T3.1, T3.2 | implemented: sealed native membership and contracts bound to the canonical preview; older/later unrelated tickets excluded and later-ticket adoption disabled for the invocation. |
| GUC-05 | T1.1, T2.2, T3.3 | implemented: manual authoring/approval and standalone Web plan publication regressions pass; standalone plan generation does not activate a worker. |
| GUC-06 | T1.1, T4.1, T5.1, T6.2, T6.3 | partial: shared run/revision and concurrent control regressions passed; Telegram/TUI integration and parity remain pending. |
| GUC-07 | T2.3, T4.2, T5.2, T6.2 | partial: Slice 2 quiz/specification and HTTP evidence passed; later owning tasks remain pending. |
| GUC-08 | T1.3, T2.1 | implemented: nonempty capture and valid, sufficient clarification are required; focused invalid/empty/budget/foreign completion regressions passed. |
| GUC-09 | T2.2, T3.1, T4.1 | implemented: original authoring/publication/materialization operations reconcile before further progression; host startup recovers a retained quiz owner receipt without regeneration. |
| GUC-10 | T3.2, T4.1, T4.3 | partial: lost admission/start/control responses and host startup retain the original invocation and consumed attempts; complete publication-response variants remain with T4.3. |
| GUC-11 | T4.1, T5.3, T6.2 | partial: request cancellation, real dead child owners, graceful host interruption and durable operator stops passed local normal/race checks; remaining adapter acceptance is pending. |
| GUC-12 | T3.3, T4.1, T6.2 | partial: Web controls delegate to native revision-bound receipts and distinguish stop intent from quiescence; original policy, spent attempts and human review remain enforced. TUI controls remain pending. |
| GUC-13 | T1.2, T2.3, T3.1, T4.3 | partial: Slice 2 quiz/specification and HTTP evidence passed; later owning tasks remain pending. |
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

Slice 3 T3.1 task checks passed on the implemented preparation content:

- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCasePreparation|WorkflowMaterializationRecovery|PlanPublicationRecovery|WebPlanDelivery|ServicePlanOrderPublication)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCasePreparation' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/...` — zero issues.
- `git diff --check`

The first lost-Git-acknowledgement fixture missed the managed Git argument
prefix and failed to reach its fault. The corrected fixture passed, including
recovery after later native working-copy edits. These task checks do not
establish the final Slice 3 PR gate; the final PR head requires controller validation.

Slice 3 T3.2 task checks passed for implementation commit
`a740689accce48f067b66db400a6b6f37698da41`:

- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCase(Preparation|Admission)|ServiceWorkNative|TelegramWorkAdmission)' -count=1 -timeout=3m` — passed (core 0.689s, app 49.879s).
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCaseAdmission' -count=1 -timeout=3m` — passed (core 1.031s, app 41.298s).
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/...` — zero issues.
- `git diff --check`

The first host fixture lacked a coordinator marker on its canonical target.
Guided admission now rejects that prerequisite before allocating native work;
the corrected fixture commits only its coordinator registration before pushing
the code target. A separate proof fixture initially omitted its required
checkpoint artifact; adding the original artifact binding made its positive
case valid. Corrected regressions pass, including one actual worker attempt,
request-context cancellation, duplicate activation and a lost start response
after reaching native human review. Delivery remains unconfirmed at that wait.
The managed host is implemented through shared native continuation; wiring its
lifetime and native status into Web is completed by T3.3 below. Final PR-head
validation has not run.

Slice 3 T3.3 task checks passed for implementation commit
`4eaf8bd2db4126d8d08596daa795fec52acb50a9`:

- `go test ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(AdmissionWeb|AdmissionPrerequisites|AdmissionPolicy|AdmissionProvider|SpecWeb)|WebPlanDelivery|ServicePlanOrderPublication)' -count=1 -timeout=3m` — passed (core 0.011s, app 23.381s; Web rendering exercised through app HTTP fixtures).
- `go test -race ./internal/app -run '^TestGuidedUseCaseAdmission(Web|Policy|Prerequisites)$' -count=1 -timeout=3m` — passed (app 45.028s).
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.
- `git diff --check`

Development probes found an obsolete unavailable-preparation assertion and an
invalid nested provider fixture key. Assertions now require the actual missing
workflow prerequisite, and the provider fixture uses the existing flat
configuration key. Corrected checks pass. HTTP progression prepares and starts
one exact task; duplicate submission and repeated GETs do not repeat work.
Native human review still reports zero confirmed deliveries. All owned local
host fixtures finish and join through their cleanup. Final controller checks
remain bound to the PR head after its reference follow-up commit.

Slice 4 T4.1 task checks passed for implementation commit
`0ea3ff90e3d4bc5e35b36b7ad188cde3c8cb0274`:

- `go test ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(Recovery|Control|Admission(Web|Host|StartedAck)|QuizWeb|SpecWeb)|Workflow(Stop|ClosingRecovery|ClosingStop|ControlCAS))' -count=1 -timeout=3m` — passed (core 13.670s, app 34.983s; Web rendering exercised through app HTTP fixtures).
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(Recovery|Control|Admission(Web|Host|StartedAck)|QuizWeb|SpecWeb)|Workflow(Stop|ClosingRecovery|ClosingStop|ControlCAS))' -count=1 -timeout=3m` — passed (core 28.059s, app 85.131s).
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.
- `git diff --check`

Real child-process crashes, graceful worker interruption, uncertain native
control/admission results, concurrent controls and rendered Web revision guards
are covered. The graceful shutdown fixture exposed a retained native
`interrupted` marker that prevented a resumed task from continuing; native
resume now clears that task/validation marker while preserving set-closure
recovery. Corrected fixtures distinguish native stop intent from quiescence,
immutable admission from mutable member images, and the existing HTTP 400 stale
response from a hypothetical HTTP 409. All focused checks pass. Owned host and
worker fixtures finish through their cleanup. Subsequent T4.2/T4.3 evidence follows;
the final exact-head controller gate belongs to the PR-reference head.

Slice 4 T4.2 is implemented in `c6fd84c0197842011a2ae6cdcececaf739894c56`.
Focused checks passed:

- `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Ideas|Result)' -count=1 -timeout=3m` — passed (core 0.241s, app 11.708s).
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Ideas|Result)' -count=1 -timeout=3m` — passed (core 1.536s, app 37.456s).
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.

Separate ideas preserve sealed inputs and stop state; duplicate/stale/foreign
requests and interrupted journals are covered. Real native Git delivery and set
closure are required for final results. Changed receipt bytes prevent a completion
claim; inspection does not mutate authoring or native owners.

Slice 4 T4.3 is implemented in `0fa7cec16dc3e6b02f3db1d593e3cf4c0081da69`. Focused checks passed:

- `go test ./internal/core ./internal/app -run '^TestGuidedUseCase(CheckpointRecovery|PreparationCommit|WebJourney)$' -count=1 -timeout=3m` — passed (core 0.095s, app 13.966s).
- `go test -race ./internal/app -run '^TestGuidedUseCaseWebJourney$' -count=1 -timeout=3m` — passed (app 16.675s).
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestGuidedUseCaseWebBrowser$' -count=1 -timeout=5m` — passed, not skipped (app 14.683s).
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.

HTTP fixtures cover the complete retained journey after later working-copy edits,
one lost admission response and one applied Git publication with a lost response.
Only observation confirms that original publication; no second submission is made.
The initial retry test incorrectly allocated a new operation and was corrected to
submit the rendered original binding. Result validation stopped exposing internal
invocation paths in the normal view, preserving the earlier Web presentation guard.
Lost real Git ref-update acknowledgement regressions retain their original commits.

Separately, actual owned Chromium submitted rendered start, name, capture, answer,
pause/resume, later-idea and exact native-approval forms. It observed receipt-backed
completion and deferred ideas at desktop (1440x1000) and narrow (390x844) widths.
These observations use temporary local Git and deterministic provider fixtures;
no live external-forge browser acceptance is claimed. Owned host jobs, workers,
servers and browsers finish through their cleanup.

The final required ordered commands are executed through Looper on the clean
PR-reference HEAD. Their exact candidate-bound evidence belongs to the canonical
PR Validation and the controller checks ledger.

Delivery-set gate: `make verify`, bounded normal guided-workflow regressions, focused race coverage for changed ownership, then `git diff --check`, on the exact final integrated `dev` candidate. The scheduled nightly retains exhaustive certification ownership.

| Evidence | Candidate | Outcome |
| --- | --- | --- |
| Initial Slice 1 gate | `c12f62a9a1ba0f6879b2a295e0e7bb2b0a54b8cd` | failed: static Web navigation returned 503; focused correction passed and the final controller gate is regenerated. |
| Slice 1 development gate | `245cb6f306aabb5e7e9877491aa2ea26174299e6` | All four configured checks passed through Looper; PR #430 merged by the controller at that exact commit. |
| Slice 2 task checks | `5ca7c06550c4e44ec6a2fd3ae10ab93309e9a18b` | Focused capture/quiz/specification integration and race checks passed. Canonical PR and Looper retain the exact final-head gate evidence. |
| Slice 2 development gate | `2246651001dd6002f783fd86d0dcc13f5bef02e6` | All four configured checks passed through Looper; PR #431 merged by the controller at that exact commit. |
| Slice 3 development gate | `dfcfae6e787742d976f99d40f292a7e047f5252a` | All four configured checks passed through Looper; PR #432 merged by the controller at that exact commit. |
| Initial Slice 4 gate | `70608f4ced207bcaad5d0eae32dde7909d9aed95` | `make verify` failed because missing development intent returned the stopped-state error. Validate admission authority before the lifecycle guard; retain both authority and explicit-stop regressions. Final evidence is regenerated for the corrected PR head. |
| Web browser journey | `0fa7cec16dc3e6b02f3db1d593e3cf4c0081da69` | Actual owned Chromium journey passed at desktop and narrow widths against temporary Git and deterministic providers. |
| Telegram actual-client journey | pending | not observed |
| TUI interaction and cross-adapter parity | pending | not run |
| Final delivery-set gate | pending | not run |
| Scheduled exhaustive certification | pending | not observed |

## Completion boundary

Planning artifacts do not establish implementation or delivery. Advance a task only with its actual commit and validation evidence; close a slice only after its canonical PR is merged and the integrated state is verified. Preserve native receipts and original limits during any recovery.

Review later ideas after the current guided development finishes. They never become extra work in this delivery set or in a user's active invocation merely because they were retained.
