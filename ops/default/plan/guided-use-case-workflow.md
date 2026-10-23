# Guided Use-Case Workflow Delivery

Status: Approved
Date: 2026-10-22
Delivery set: guided-use-case-workflow
Planning authorization: maintainer, 2026-10-22.
Approved: maintainer, 2026-10-22; slice map, tracker and implementation accepted.
Merge authority: maintainer, 2026-10-22; Looper may merge the six recorded slice PRs into `dev` with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 5, Guided Telegram journey; T5.1 is next after the verified Web checkpoint.

Spec: [Guided workflow](../spec/stable/guided-use-case-workflow.md)
Model: [Retained state and ownership](../spec/stable/guided-use-case-workflow-model.md)
Tracker: [Tasks and acceptance](../tracker/guided-use-case-workflow.md)
Specification commit and planning baseline: `afd228595faac22c9008c84b05ad21cdc5a5d353` on `dev`.

## Outcome and strategy

Deliver one persistent guided journey: name, dump, quiz, approved specification, plan/tracker/native tickets, automatic execution and receipt-backed results. Completed authoring stages are committed; later edits and new ideas remain available for review after the admitted work finishes.

Slice strategy: behavior-first.

Reason: each increment includes its required domain, persistence, shared service, adapter and regression work. The first four slices progressively deliver and verify Web behavior; Telegram and TUI then consume the same application workflow. Separate skeleton, intentionally failing test or duplicate per-interface orchestration slices are unnecessary.

The set has six ordered slices and 18 tasks. Each slice uses one branch and one PR targeting `dev`. The branches, PR titles and reports below are planned identities; actual execution and delivery evidence belong to the tracker and Looper checkpoint.

## Existing prerequisites and design boundaries

1. The approved specification and model are committed on `dev`. Existing authoring sessions, quiz/specification operations, revision-bound plan generation, native materialization and workflow controls are available at the planning baseline.
2. Reuse the central service in `internal/app`, domain/persistence owners in `internal/core`, existing Web handlers/views, Telegram operation dispatch and TUI adapters. Queries remain read-only. The workspace index remains a rebuildable projection.
3. `GenerateSpec`, `AnswerQuiz`, `CompleteQuiz`, `GeneratePlan`, `GeneratePlanDelivery`, `MaterializeWorkflowPlan`, `PreviewWorkExtent` and `StartWorkExtent` are existing integration points to evaluate. The guided coordinator composes their owners; standalone generation retains its existing semantics.
4. Settle the guided-state encoding and compatible session-store extension in T1.1. Settle the scoped Git checkpoint mechanism in T1.2: preserve the user's branch/index/dirty work, retain reachable input commits, and keep artifact source commits distinct from the native code target/base. These are bounded implementation decisions within the approved model.
5. Reuse existing revision/hash guards, operation identities, journals and native receipts. A Git commit fixes input bytes; it does not replace idempotency, admission authority, task state or completion evidence.
6. Existing project provider, host, finite budgets, review and publication settings remain prerequisites for a particular run. A guided start cannot enable a disabled publication profile or grant missing integration authority.
7. New cases and resumption of guided runs are in scope. Adopting manually started cases is deferred. Do not create a second execution engine or a new batch of development tickets for these plan tasks.

## Canonical slice map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Guided capture and Git checkpoints | feat/slice-1-guided-capture-checkpoints | feat(slice-1): capture guided use cases with Git checkpoints | ops/default/report/slices/guided-use-case-workflow/slice-1-capture-checkpoints.md |
| Slice 2 | Guided quiz and approved specification | feat/slice-2-guided-quiz-spec | feat(slice-2): guide clarification and approve generated specifications | ops/default/report/slices/guided-use-case-workflow/slice-2-quiz-spec.md |
| Slice 3 | Automatic work preparation and execution | feat/slice-3-guided-work-execution | feat(slice-3): prepare and start exact guided work | ops/default/report/slices/guided-use-case-workflow/slice-3-work-execution.md |
| Slice 4 | Web recovery, controls and results | feat/slice-4-guided-web-recovery | feat(slice-4): complete the resumable Web use-case journey | ops/default/report/slices/guided-use-case-workflow/slice-4-web-recovery.md |
| Slice 5 | Guided Telegram journey | feat/slice-5-guided-telegram | feat(slice-5): guide use-case development through Telegram | ops/default/report/slices/guided-use-case-workflow/slice-5-telegram.md |
| Slice 6 | Guided TUI journey | feat/slice-6-guided-tui | feat(slice-6): guide use-case development through the TUI | ops/default/report/slices/guided-use-case-workflow/slice-6-tui.md |

## Slice tasks

### Slice 1: Guided capture and Git checkpoints

1. **T1.1:** Extend the canonical authoring owner with versioned guided-run state, selected project/namespace/case, progression intent, revisions, pending operations and read-only inspection. Preserve old/manual sessions. Cover duplicate starts, scope collisions and concurrent/stale writes.
2. **T1.2:** Implement stage-owned Git checkpoint creation and committed-input reads through the existing managed Git boundary. Retain repository/commit/path/content bindings, dependency references and commit-operation recovery. Preserve unrelated staged/unstaged/untracked work and retain referenced commits. Cover a successful commit with a lost acknowledgement and missing/corrupt source objects.
3. **T1.3:** Add the Web start/name/dump journey through shared commands. Support multiple inputs, explicit meaningful capture completion, readiness explanations, reload/resume and committed dump completion. Include focused service/HTTP/view regressions; do not expose later stages as completed before their implementations exist.

Exit: a user can start a new use case, finish its dump and inspect the retained committed version through Web. The repository remains valid even while later capabilities are pending.

### Slice 2: Guided quiz and approved specification

1. **T2.1:** Generate and conduct the quiz from the committed name/dump. Bind answers to the current run/question and source snapshot; retain progress across reconnects. Require valid generation and sufficient answers before sealing the answered quiz. Cover cross-run replies, duplicate answers and incomplete/invalid results.
2. **T2.2:** Generate the specification from the committed dump/quiz and save default approval through the canonical mutation owner. Commit the approved revision with actor/run/policy provenance. Reconcile interrupted generation, approval and commit without replacement artifacts; preserve standalone approval behavior.
3. **T2.3:** Extend the Web journey to ask only the current question, advance through generation without routine confirmations, and expose complete original artifacts. Verify committed-source reads after working-copy edits and update the relevant reference documentation.

Exit: Web can progress from a committed dump through an answered quiz to an inspectable, committed approved specification, without a separate specification approval prompt.

### Slice 3: Automatic work preparation and execution

1. **T3.1:** Generate the plan from the committed approved specification using existing revision-bound authoring. Materialize ordered native tickets/contracts and the tracker through their canonical owner, then seal the prepared work. Preserve exact plan and membership bindings through partial writes and interrupted stage commits.
2. **T3.2:** Bind the initiating actor's progression intent to the exact native preview, effective policy and materialized membership. Admit and start that extent through the existing host lifecycle, with original finite limits and no extra routine start prompt. Reconcile lost admission/start acknowledgements; exclude older/later/unrelated tickets.
3. **T3.3:** Extend Web with preparation and execution state backed by native queries. Cover changed policy, missing host/provider/publication prerequisites, failed contracts and duplicate starts. Preserve manual plan/ticket/work operations and document the guided-only automatic behavior.

Exit: an eligible Web journey automatically prepares and starts exactly its generated native work. A real prerequisite failure produces an actionable state rather than a false execution claim.

### Slice 4: Web recovery, controls and results

1. **T4.1:** Complete shared progression recovery across process restarts, client disconnects and concurrent interfaces. Reconcile every pending owner operation before advancing. Delegate pause/resume/cancel to the native control owner, retain stop intent and original budgets, and keep query/render/poll paths free of mutations.
2. **T4.2:** Retain later ideas separately from the frozen input chain and admitted work. Continue the current scope, acknowledge the deferred review, and surface the ideas with results. Project actual waits, validations, publication/integration receipts, remaining work and code locations; never infer delivery from worker exit or an open PR.
3. **T4.3:** Verify the complete Web journey and fault variants with deterministic local providers, real temporary Git/filesystem stores and the existing browser-test harness. Cover later working-copy edits, lost commit/admission/publication responses and explicit controls. Update the Web User Guide and reader-facing changelog for the now-usable journey.

Exit: the complete Web workflow is resumable and evidence-backed, including controls and deferred ideas. Fixture evidence and actual browser observations are recorded separately. Close this Web checkpoint before starting the Telegram slice.

### Slice 5: Guided Telegram journey

1. **T5.1:** Map a natural-language guided-workflow start and current-step replies/buttons to the shared commands. Use recognizable names and existing reply/focus bindings; request selection only for genuine ambiguity. Do not require opaque IDs or slash commands for the normal journey.
2. **T5.2:** Present current input, stage, real waits and final results through existing concise message/original-document delivery. Coalesce unchanged notifications; map later ideas and explicit controls without changing the committed scope. Reuse the same run when switching between Web and Telegram.
3. **T5.3:** Verify the journey through existing receiver/transport fixtures, including duplicate callbacks, stale/foreign replies, lost acknowledgements and full original attachments. Record any separately authorized user-operated Telegram acceptance honestly. Update Telegram guidance and the changelog for the available capability.

Exit: Telegram guides the same retained journey without manual artifact choreography or a second coordinator. Transport fixtures do not stand in for unobserved live-client acceptance.

### Slice 6: Guided TUI journey

1. **T6.1:** Add TUI entry, current-step input, artifact inspection and resumable run navigation through the shared service. Keep domain and progression logic out of terminal callbacks.
2. **T6.2:** Present shared execution progress, explicit controls, deferred ideas and receipt-backed code/results. Verify continuation of the same run across TUI, Web and Telegram, including conflicting/stale actions.
3. **T6.3:** Verify TUI interaction and cross-adapter parity, reconcile every acceptance criterion with candidate-bound evidence, and complete the User Guide/reference/changelog updates. Produce the delivery-set cycle report after implementation and run the selected final gate on the exact integrated `dev` candidate.

Exit: all three interfaces expose the same workflow and authority boundaries. All planned tasks and criteria have actual evidence before delivery closure.

## Acceptance traceability

`GUC-01` through `GUC-16` refer to numbered acceptance criteria 1 through 16 in the approved specification. These aliases do not add requirements.

| Criterion | Contract | Owning tasks |
| --- | --- | --- |
| GUC-01 | Canonical name/dump/quiz journey | T1.3, T2.1, T2.3, T5.1, T6.1 |
| GUC-02 | Automatic approval, preparation and execution | T2.2, T3.1, T3.2 |
| GUC-03 | Exact generated-specification approval boundary | T2.2 |
| GUC-04 | Exactly the materialized native membership | T3.1, T3.2 |
| GUC-05 | Preserved independent/manual operations | T1.1, T2.2, T3.3 |
| GUC-06 | Shared run and current revision across interfaces | T1.1, T4.1, T5.1, T6.2, T6.3 |
| GUC-07 | Useful progress and complete inspectable documents | T2.3, T4.2, T5.2, T6.2 |
| GUC-08 | Empty/invalid clarification cannot advance | T1.3, T2.1 |
| GUC-09 | Recover partial authoring/materialization | T2.2, T3.1, T4.1 |
| GUC-10 | Reconcile lost admission/publication acknowledgements | T3.2, T4.1, T4.3 |
| GUC-11 | Connection/process independence and retained limits | T4.1, T5.3, T6.2 |
| GUC-12 | Native policy, review, budgets and explicit controls | T3.3, T4.1, T6.2 |
| GUC-13 | Read committed inputs despite later edits | T1.2, T2.3, T3.1, T4.3 |
| GUC-14 | Completion from receipts with actual code locations | T4.2, T4.3, T5.2, T6.2, T6.3 |
| GUC-15 | Later ideas deferred without scope mutation | T4.2, T5.2, T6.2 |
| GUC-16 | Recover an acknowledged or uncertain stage commit | T1.2, T3.1, T4.3 |

## Validation and delivery gates

For documentation planning, validate links, terminology, slice/task/acceptance consistency, scoped staged content and whitespace. No runtime check is claimed by this package.

During implementation, add behavior regressions with the following proposed selectors. They are test obligations, not symbols already present at the planning baseline. Zero matching tests or a skipped required fixture does not establish a pass.

| Slice | Packages | Proposed focused selector |
| --- | --- | --- |
| 1 | `./internal/core ./internal/app ./internal/webui` | `^TestGuidedUseCase(Capture\|Checkpoint)` |
| 2 | `./internal/core ./internal/app ./internal/webui` | `^TestGuidedUseCase(Quiz\|Spec)` |
| 3 | `./internal/core ./internal/app ./internal/webui` | `^TestGuidedUseCase(Preparation\|Admission)` |
| 4 | `./internal/core ./internal/app ./internal/webui` | `^TestGuidedUseCase(Recovery\|Control\|Ideas\|Result\|WebJourney)` |
| 5 | `./internal/core ./internal/app` | `^TestGuidedUseCaseTelegram` |
| 6 | `./internal/core ./internal/app ./internal/tui` | `^TestGuidedUseCase(TUI\|Parity)` |

Use each table's packages/selector with `go test <packages> -run '<selector>' -count=1 -timeout=3m`. The escaped table separators represent literal regex alternatives. Run the smallest affected group during tasks, `make verify` for each slice integration, the focused integration group without `-short`, affected concurrency cases with `-race`, and `git diff --check` last. Add existing affected regressions to the same bounded invocation where needed, recording their names and purpose.

For Web acceptance in Slice 4, extend the existing browser harness with `TestGuidedUseCaseWebBrowser` and run `go test ./internal/app -run '^TestGuidedUseCaseWebBrowser$' -count=1 -timeout=5m` with its documented Chromium prerequisite. This is a targeted check; unavailable browser evidence remains explicitly unverified. Do not substitute HTTP fixtures for observed browser interaction.

Expected local cost is the existing short development gate plus the selected fixture group, using three-minute package timeouts; the targeted browser has a five-minute timeout. Measure actual durations before expanding the budget. Complete repository/race/live campaigns remain with the scheduled nightly; this plan does not authorize a local exhaustive run or nightly dispatch.

For final delivery closure, run these commands on the exact integrated `dev` candidate:

```sh
make verify
go test ./internal/core ./internal/app ./internal/webui ./internal/tui -run '^TestGuidedUseCase(Capture|Checkpoint|Quiz|Spec|Preparation|Admission|Recovery|Control|Ideas|Result|WebJourney|Telegram|TUI|Parity)' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app ./internal/tui -run '^TestGuidedUseCase(Capture|Checkpoint|Quiz|Spec|Preparation|Admission|Recovery|Control|Ideas|Telegram|TUI|Parity)' -count=1 -timeout=3m
git diff --check
```

The combined selection excludes the separate browser test. Retain the targeted browser evidence from the Web checkpoint; repeat it on a later candidate when changes affect that journey. Reuse candidate-bound evidence only for its actual commit and command, and record scheduled exhaustive certification separately from this development gate.

## Handoff and completion

The tracker owns task state, actual commits, PRs, validations and acceptance evidence. Reports are created when their slices are implemented, not as empty evidence during planning. One conventional implementation commit per task is expected where practical.

The maintainer approved this slice map/tracker, implementation and Looper startup on 2026-10-22, then explicitly delegated Rebase + Fast-forward merges for all six recorded slice PRs into `dev` after their required checks. Looper owns the execution and merge loop for this delivery set. This delegation does not include additional slices, final-gate repair grants, `main` promotion, mirrors or release publication.

After implementation, record the final end-to-end outcome in `ops/default/report/cycles/guided-use-case-workflow.md`, including observed Web/Telegram/TUI scope, input checkpoint identities, actual native receipts, validation candidate and remaining external prerequisites. This future path is not current acceptance evidence.
