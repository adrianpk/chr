---
title: "chr Project-wide Planning and Ticket Continuity Gap Plan"
project_id: "project-f14593e169cfa218"
case_slug: ""
status: "planned"
source_request_id: "plan-generation-project"
source_response_id: "response-plan-generation-project"
created_at: "2026-10-17T19:55:52Z"
---

# chr Project-wide Planning and Ticket Continuity Gap Plan

## Current Implementation Baseline

Plan kind: `project_gap_plan`. Readiness: `ready`. Blocking slots: none. This is an implementation-hardening plan for the existing project.

1. The binding baseline reports an existing workspace, Go module `forge.adrianpk.com/chr`, Go version `1.25.1`, Makefile, README, implementation, and one observed test file. Its 80 listed files include historical delivery material and a nested proving-ground application. These observations do not establish that unlisted production packages or tests are absent.
2. The supplied project context establishes Markdown-first authoring, CLI and TUI operations, Docker-bounded execution, derived coordination state, and execution evidence. Extend these facilities rather than recreate commands, storage, exports, or infrastructure.
3. The supplied review tracker records eight delivered slices, the merger tracker records four delivered slices, and the promotion-gate plan records completed delivery. Their older `planned` plan labels do not justify scheduling those implementations again. These are supplied completion records, not fresh remote verification.
4. Several supplied artifacts contain migration pointers instead of requirements. They provide provenance but insufficient implementation detail by themselves. Preserve historical `.prj/`, `.chr/`, and `ops/` material.
5. Workspace inspection identifies the existing implementation owners: baseline collection in `internal/app/planning_baseline.go`, generation-context assembly in `internal/app/service_spec.go`, plan operations in `internal/app/service_plan.go`, and plan validation and workflow materialization in `internal/core`.

### Binding decisions discovered in the workspace

No dedicated ADR directory was found. The following active architecture rules and accepted specifications serve as decision documents.

| ID | Decision source | Constraint on this plan |
| --- | --- | --- |
| D1 | `.agents/ARCHITECTURE.md` | Keep domain validation in `internal/core`, shared operations in the central `internal/app` service, and adapters thin. Markdown and owned workflow records remain authoritative; indexes remain rebuildable. |
| D2 | `ops/default/spec/stable/ticket-native-workflow.md` and its model companion | New ticket-native work has one ticket per task, stable hierarchy and order, and one branch, canonical PR, integration receipt, and merge cycle per slice. Tracker output is a projection. Legacy adoption is explicit and prospective. |
| D3 | `ops/default/spec/stable/native-integration-queue.md` and its model companion | Preserve ordered receipt barriers, exact-candidate evidence, existing coordinator ownership, and one publication owner per target. Serial execution remains the default. Status inspection cannot resume or repair work. |

The document-generation explanation still describes one generated ticket per slice. D2 takes precedence for new ticket-native work. The older review draft's dedicated branch requirement applies to the slice integration unit when tasks belong to a slice; it must not create a PR per task. The older mechanical merger description must not erase D3's accepted, bounded conflict-resolution contract. This plan changes generation and materialization continuity, not merger behavior.

## Assumptions and Low-Confidence Items

1. `slicing_target` is inferred as reliable project-wide context → plan → approved task-ticket materialization. The request supplies no narrower feature objective. This bounds the delivery set; it does not reopen the whole product backlog.
2. `known_existing_code` is incomplete in the supplied file sample. Implementation must inspect the owning functions and existing regressions before each change. Absence from the sample is never evidence that a feature needs creation.
3. `binary_or_library` and `entrypoints` are resolved from the project description and workspace architecture as an existing application with CLI and TUI surfaces. No entrypoint work is planned.
4. Storage and runtime constraints come from the project document and D1–D3: durable project files, derived indexes, existing workflow records, and the established sandbox boundary. No new database, scheduler, runtime, or queue is planned.
5. Validation is resolved from `.agents/LINTING.md`: focused regressions and `make verify` for delivery. Complete suites and live certification belong to the existing scheduled nightly process.
6. Delivery granularity is resolved by D2. The plan uses six sequential slices and twelve task tickets. Native ticket keys and durable record identities are allocated by the existing materializer; the identifiers below are stable planning inputs.
7. Migration pointers may resolve to missing, ambiguous, or duplicated historical content. Such cases must produce explicit diagnostics rather than invented requirements.
8. Accepted specifications define required behavior, not proof that every runtime or provider capability is delivered or certified. This plan does not claim or require new live-provider certification.

## High-level implementation strategy

Delivery set: `project-gap-closure`. Scope: project-wide, with no use-case association. Integration branch: `dev`, as required by repository policy. Namespace: `default`.

Slice strategy: behavior-first. Each slice closes one observable continuity failure through the existing service and domain boundaries, with its regression coverage included in the same change.

Separate bounded file samples from implementation facts; recover explicitly referenced migrated source content; include applicable accepted decisions in generation; validate generated plans before replacement; and connect approved plans to the existing task-ticket materializer. Finish with a deterministic acceptance fixture and corrections to current operator documentation.

Preserve editable Markdown, request/response provenance, existing artifact identities, and explicit approval before executable materialization. Reuse delivered review, merger, execution, and gate capabilities. Exclude product scaffolding, historical batch conversion, new execution services, unrelated interface redesign, and promotion or release work.

## Slice decomposition

### Slice 1: Reliable baseline coverage

1. **Identifier:** `pg-slice-1`; display label `Slice 1`.
2. **Objective:** Prevent a bounded file sample from presenting an incomplete view as complete implementation evidence.
3. **Implementation rationale:** The supplied snapshot contains 80 files and fixture packages without identifying the main application entrypoint. Existing baseline collection derives package and test facts from its capped file list. Planning must distinguish observed samples from repository coverage.
4. **Dependencies:** Existing baseline collector and readiness gate.
5. **Implementation notes:** Extend the current collector. Keep deterministic sampling, collect implementation facts independently where practical, distinguish nested fixture modules, and expose incomplete coverage or traversal failures. Retain bounded resource use and existing setup/gap classifications.
6. **Acceptance criteria:** Saturating the sample with historical files does not hide root application evidence; nested modules remain distinguishable; incomplete scans are explicit; identical inputs produce stable ordering; an existing module and Makefile never trigger setup creation.
7. **Estimated execution complexity:** Medium.
8. **Associated tasks:** T1.1, T1.2.
9. **Ticket metadata:** Baseline and readiness payloads are primary evidence; D1 governs ownership. Task input IDs are `pg-task-1-1` and `pg-task-1-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-1-baseline-coverage`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(planning): report reliable baseline coverage`; include the saturated-sample regression and compatibility results.

### Slice 2: Migrated context resolution

1. **Identifier:** `pg-slice-2`; display label `Slice 2`.
2. **Objective:** Supply substantive, traceable context when an artifact contains only migration provenance.
3. **Implementation rationale:** Multiple supplied dumps, quizzes, specs, and plans contain only a legacy-source pointer. Treating those bodies as complete requirements loses the continuity described by the document-generation dump.
4. **Dependencies:** Slice 1 integrated.
5. **Implementation notes:** Extend lazy context loading to resolve explicitly recorded migration provenance and owned legacy payloads. Bound reads to allowed project paths, reject traversal and symlink escape, retain envelope identity, and preserve source path and content hash. Do not follow arbitrary Markdown links or import raw captures as regular curated context.
6. **Acceptance criteria:** A resolvable pointer contributes its substantive source with provenance; repeated copies are deduplicated deterministically; missing or ambiguous sources produce diagnostics; archived generated plans and raw siblings remain excluded from regular generation context; source files remain unchanged.
7. **Estimated execution complexity:** Medium.
8. **Associated tasks:** T2.1, T2.2.
9. **Ticket metadata:** Cite the supplied migration envelopes and document-generation dump; apply D1's lazy-body and source-authority rules. Task input IDs are `pg-task-2-1` and `pg-task-2-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-2-migrated-context`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(context): resolve migrated planning sources`; demonstrate resolved, missing, ambiguous, duplicate, and unsafe-path cases.

### Slice 3: Decision authority and scope

1. **Identifier:** `pg-slice-3`; display label `Slice 3`.
2. **Objective:** Give generation backends the applicable accepted decisions and explicit conflict information.
3. **Implementation rationale:** The supplied context combines older drafts and planned mirrors with delivered trackers. Workspace discovery also finds accepted contracts that change task granularity and workflow authority. Correct output must not depend on a backend independently finding these differences.
4. **Dependencies:** Slice 2 integrated.
5. **Implementation notes:** Discover decision documents from workspace structure and content, without requiring configured paths. Feed relevant active or accepted decisions through the existing context-section model for single-pass and staged generation. Preserve document status, scope, provenance, and conflicts; do not infer acceptance from a filename alone.
6. **Acceptance criteria:** Both generation paths receive D1–D3 when applicable; accepted decisions outrank conflicting drafts and dumps; project-wide decisions remain available to relevant use-case requests; unrelated case material is excluded; supplied delivered trackers prevent stale plans from being treated as undelivered requirements.
7. **Estimated execution complexity:** Medium.
8. **Associated tasks:** T3.1, T3.2.
9. **Ticket metadata:** Reference D1–D3 and the specific conflicting supplied artifacts. Task input IDs are `pg-task-3-1` and `pg-task-3-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-3-decision-authority`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(planning): apply accepted decision authority`; include precedence, scope, and backend-parity evidence.

### Slice 4: Validated and recoverable plan generation

1. **Identifier:** `pg-slice-4`; display label `Slice 4`.
2. **Objective:** Reject structurally incomplete or contradictory plans before publishing or replacing the canonical artifact.
3. **Implementation rationale:** The required output contract includes executable slices, tasks, dependencies, and delivery metadata. Existing generic plan validation checks basic content presence; the stronger contract must be enforced consistently across generation paths.
4. **Dependencies:** Slice 3 integrated.
5. **Implementation notes:** Extend existing outline, detail, and plan validation. Normalize generated content through existing models rather than introduce another durable workflow authority. Validate required sections, unique identities, task ownership, dependency references and cycles, branch/PR consistency, and readiness obligations. Prepare and validate replacement content before archiving and replacing the observed canonical plan.
6. **Acceptance criteria:** Valid single-pass and staged outputs satisfy the same contract; project-wide plans include the baseline; partial readiness produces a provisional plan with unresolved slots; explicit creation of already-existing scaffolding is rejected when represented by the implementation contract; malformed output, cancellation, or concurrent user edits cannot overwrite the previous plan.
7. **Estimated execution complexity:** High.
8. **Associated tasks:** T4.1, T4.2.
9. **Ticket metadata:** Reference the request's plan/output contracts, deterministic readiness evidence, and D2. Task input IDs are `pg-task-4-1` and `pg-task-4-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-4-plan-validation`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(planning): validate plans before replacement`; show negative cases and canonical-artifact preservation.

### Slice 5: Approved task-ticket materialization

1. **Identifier:** `pg-slice-5`; display label `Slice 5`.
2. **Objective:** Carry approved generated tasks into the existing ticket-native materialization path without collapsing them into slice tickets.
3. **Implementation rationale:** D2 requires one ticket per task and one PR per owning slice. The existing operation supports explicit `GeneratedWorkflowPlan` input, while compatibility tracking remains available. Generation and approval must connect to the correct authority deliberately.
4. **Dependencies:** Slice 4 integrated; existing native materializer and journal/replay mechanisms available.
5. **Implementation notes:** Build approved native input from the validated plan, observed plan hash, stable input IDs, ordered tasks, dependencies, acceptance, and validation commands. Reuse allocation, transactions, receipts, and recovery. Preserve explicit legacy compatibility mode. Map results through the shared service to existing CLI and TUI operations.
6. **Acceptance criteria:** A two-slice, three-task-per-slice fixture creates six tickets and two slice descriptors; matching retries return the original identities and receipt; changed approved content requires reconciliation; task dependencies resolve to allocated identities; all tasks in a slice share one branch and PR expectation; tracker projection failure preserves successful materialization and reports `projection_stale`.
7. **Estimated execution complexity:** High.
8. **Associated tasks:** T5.1, T5.2.
9. **Ticket metadata:** Reference D2, D3, and `docs/reference/plan-workflow-materialization.md`. Task input IDs are `pg-task-5-1` and `pg-task-5-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-5-task-materialization`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(planning): materialize approved task tickets`; include retry, source-edit, recovery, compatibility, and projection-failure evidence. No new storage or export subsystem.

### Slice 6: Continuity acceptance and operator contract

1. **Identifier:** `pg-slice-6`; display label `Slice 6`.
2. **Objective:** Prove the complete planning-to-ticket path and align current documentation with its actual authority boundaries.
3. **Implementation rationale:** Individual artifact tests cannot establish continuity across migration resolution, decision selection, generation, approval, and materialization. Current one-ticket-per-slice wording also conflicts with D2.
4. **Dependencies:** Slice 5 integrated.
5. **Implementation notes:** Add one deterministic acceptance fixture using a controlled model response and temporary project files. Exercise both generation paths and existing materialization recovery. Correct active explanation/reference documentation while leaving historical plans and reports intact.
6. **Acceptance criteria:** The fixture preserves provenance and accepted constraints through ticket creation; retries create no duplicates; a source edit blocks stale approval; failed generation preserves the editable plan; tracker failure does not undo native state; reading status causes no mutation. Documentation distinguishes legacy tracking, task authority, slice PR ownership, and local versus nightly evidence.
7. **Estimated execution complexity:** Medium.
8. **Associated tasks:** T6.1, T6.2.
9. **Ticket metadata:** Reference D1–D3, the supplied document-generation context, and the completed review/merger/gate records as preserved prerequisites. Task input IDs are `pg-task-6-1` and `pg-task-6-2`.
10. **Branch metadata:** `feature/project-gap-closure/slice-6-continuity-validation`, created from updated `dev`.
11. **PR expectations:** One PR titled `fix(planning): verify planning-to-ticket continuity`; include acceptance evidence and the corrected operator contract. No Docker worker or live forge is needed for this fixture.

## Slice ordering

1. Slice 1 establishes trustworthy baseline coverage.
2. Slice 2 restores substantive migrated inputs.
3. Slice 3 establishes decision authority over those inputs.
4. Slice 4 validates the resulting implementation contract and preserves editable artifacts.
5. Slice 5 materializes approved task obligations through existing native operations.
6. Slice 6 verifies the combined path and closes documentation inconsistencies.

Each slice starts only after its predecessor is reviewed, merged into `dev`, and verified in updated local `dev`.

## Slice dependency analysis

| Slice | Direct predecessor | Dependency reason |
| --- | --- | --- |
| Slice 1 | Existing baseline/readiness implementation | Hardens existing evidence collection. |
| Slice 2 | Slice 1 | Context completeness diagnostics must agree with baseline coverage semantics. |
| Slice 3 | Slice 2 | Authority selection needs resolved source content and provenance. |
| Slice 4 | Slice 3 | Contract validation needs the resolved readiness and decision constraints. |
| Slice 5 | Slice 4 | Executable tickets must derive from a validated, observed approved plan. |
| Slice 6 | Slice 5 | Acceptance covers the complete materialization path. |

The dependency graph is a single chain. Generation-context and plan-service files are shared across slices, so parallel execution would add conflicting edits and stale assumptions. No parallel slice is planned. Provider rollout and runtime scheduling are outside this graph.

## Tasks inside slices

Each task produces one ticket and should map to one Conventional Commit. Regression coverage accompanies the implementation it protects unless the task explicitly owns combined acceptance coverage.

| Task | Implementation work | Expected Conventional Commit | Task completion evidence |
| --- | --- | --- | --- |
| T1.1 | Separate baseline implementation facts from the capped file sample; add explicit coverage diagnostics and root/nested-module distinctions. | `fix(planning): separate baseline facts from file samples` | Historical-file saturation preserves root application evidence and deterministic output. |
| T1.2 | Update readiness consumers and context rendering to preserve incomplete-coverage information without falsely scheduling setup work. | `fix(planning): preserve baseline coverage in readiness context` | Existing-project, empty-project, traversal-failure, and incomplete-scan regressions pass. |
| T2.1 | Resolve explicit migration provenance through a bounded, safe source reader integrated with lazy generation-context loading. | `fix(context): resolve migrated artifact source content` | Resolved, missing, ambiguous, traversal, and symlink cases behave as specified. |
| T2.2 | Carry source hashes and envelope provenance; deduplicate equivalent inputs and preserve raw/archive exclusions. | `fix(context): preserve migrated source provenance` | Equivalent copies do not multiply context; distinct conflicting sources remain visible. |
| T3.1 | Add structural discovery of relevant accepted decisions and inject them into existing generation context sections. | `fix(planning): include applicable workspace decisions` | Decisions are found outside a fixed ADR path and reach both generation paths. |
| T3.2 | Apply explicit authority and scope rules, including conflict diagnostics and completed-history classification. | `fix(planning): resolve context authority conflicts` | Accepted task granularity wins over older drafts; unrelated cases remain excluded. |
| T4.1 | Extend existing plan validators to enforce the executable section, identity, dependency, task, and delivery contracts. | `fix(planning): enforce executable plan contracts` | Missing fields, cycles, unknown references, inconsistent branches, and readiness violations reject before publication. |
| T4.2 | Validate replacement plans before archive/replacement and guard the observed canonical source against concurrent edits. | `fix(planning): preserve canonical plans on generation failure` | Invalid output, cancellation, and stale observations preserve the canonical plan and recovery evidence. |
| T5.1 | Translate approved validated plans into existing native materialization input with stable task IDs and observed source hashes. | `fix(planning): bind approved plans to native task inputs` | Six-task fixture yields six tickets; matching retry preserves identities; changed source blocks. |
| T5.2 | Connect existing CLI/TUI tracking operations to explicit native or compatibility authority and render receipts, mappings, and projection failures. | `fix(planning): expose task materialization through shared operations` | Surfaces return the same domain result; legacy work is not silently adopted; export failure does not repeat allocation. |
| T6.1 | Add the deterministic cross-boundary acceptance fixture and interrupted-materialization continuation cases. | `test(planning): cover planning-to-ticket continuity` | Both backends preserve authority, provenance, exact approval, retry safety, and read-only status. |
| T6.2 | Correct active generation/materialization documentation and record the final contract and validation evidence. | `docs(planning): document task ownership and approval boundaries` | Current docs agree with D1–D3; historical artifacts remain intact; diff validation passes. |

## Ticket generation data

Use the existing allocator for `CHR` ticket keys and artifact envelopes. Do not use task labels as durable native identities.

| Slice input ID | Slice order | Task input ID | Display task | Task order | Direct task prerequisite |
| --- | --- | --- | --- | --- | --- |
| `pg-slice-1` | 1 | `pg-task-1-1` | T1.1 | 1 | None |
| `pg-slice-1` | 1 | `pg-task-1-2` | T1.2 | 2 | `pg-task-1-1` |
| `pg-slice-2` | 2 | `pg-task-2-1` | T2.1 | 1 | Slice 1 integration receipt |
| `pg-slice-2` | 2 | `pg-task-2-2` | T2.2 | 2 | `pg-task-2-1` |
| `pg-slice-3` | 3 | `pg-task-3-1` | T3.1 | 1 | Slice 2 integration receipt |
| `pg-slice-3` | 3 | `pg-task-3-2` | T3.2 | 2 | `pg-task-3-1` |
| `pg-slice-4` | 4 | `pg-task-4-1` | T4.1 | 1 | Slice 3 integration receipt |
| `pg-slice-4` | 4 | `pg-task-4-2` | T4.2 | 2 | `pg-task-4-1` |
| `pg-slice-5` | 5 | `pg-task-5-1` | T5.1 | 1 | Slice 4 integration receipt |
| `pg-slice-5` | 5 | `pg-task-5-2` | T5.2 | 2 | `pg-task-5-1` |
| `pg-slice-6` | 6 | `pg-task-6-1` | T6.1 | 1 | Slice 5 integration receipt |
| `pg-slice-6` | 6 | `pg-task-6-2` | T6.2 | 2 | `pg-task-6-1` |

Every ticket carries project ID `project-f14593e169cfa218`, namespace `default`, no use-case association, delivery-set input ID `project-gap-closure`, its slice/task input IDs and order, the task outcome and scope, referenced decisions, acceptance criteria, and validation commands. Resolve input references to allocated native identities during materialization.

The source plan reference and observed hash are captured at explicit approval. Preserve existing classification defaults unless the approved task requires a supported override. Each ticket inherits its slice's branch and PR expectation; no task owns a separate merge cycle. The tracker maps task IDs to allocated keys, accepted commits, evidence, and the shared slice receipt.

## Branch strategy

Use the exact `feature/project-gap-closure/slice-N-short-name` branches recorded in the slice definitions. `project-gap-closure` is a deterministic branch namespace, not an implicitly created use case.

1. Create each branch from updated canonical `dev` after its predecessor's integration is verified.
2. Keep all tasks of the slice on that branch. Repairs retain the same branch and canonical PR.
3. Preserve unrelated working-tree changes. A dirty or conflicting workspace blocks branch preparation until isolated or reconciled.
4. Record one report per slice at `ops/default/report/slices/project-gap-closure/slice-<n>-<short-name>.md`, using the branch's short name.
5. This generated plan authorizes no branch creation, commit, push, or execution. Approved plan/tracker state and implementation activation remain prerequisites.

## Pull request strategy

Open exactly one canonical PR per slice against `dev`, using the recorded title. All associated task tickets refer to that PR.

Use the repository PR structure: `Summary`, `Changes`, and `Validation`. Describe the changed behavior, cite affected decisions and compatibility boundaries, include the immutable slice-report link, and list only checks actually run. A slice report enters `reviewing` before PR review and `delivered` only after verified integration.

Keep generation hardening separate from runtime or provider feature work. Do not create task-level PRs, replacement PRs during ordinary retries, or a promotion PR as part of this delivery set.

## Merge sequencing

1. Materialize and approve the plan/tracker execution map before implementation starts.
2. Complete each slice's ordered task tickets and recorded checks on its shared branch.
3. Complete its review cycle and integrate only under the existing configured review policy and merge authority. Preserve the project's effective policy; this plan does not enable automatic merging.
4. Verify the canonical PR merged into `dev`, confirm the integrated commit is contained in updated local `dev`, and record the shared slice receipt before finalizing its tickets.
5. Start the next slice only after that receipt barrier. A passed review alone is insufficient.
6. After Slice 6 integrates, run the selected closing gate against the exact frozen `dev` candidate and record closure evidence. Leave `main`, tags, releases, and scheduled workflow activation outside this plan.

## Review requirements

1. Confirm every change extends an existing owner and does not introduce another service, database, index authority, scheduler, or export mechanism.
2. Check source containment, bounded reads, deterministic ordering, provenance, archive exclusion, and explicit diagnostics for missing context.
3. Check D1–D3 precedence, project/use-case scope, legacy compatibility, and the absence of inferred historical adoption.
4. Check one ticket per task, stable identities, one branch/PR per slice, exact approval hashes, and retry/recovery behavior.
5. Check that invalid generation and concurrent manual edits preserve user-authored artifacts and that projection failure cannot undo authoritative state.
6. Review behavioral regressions and failure boundaries. Do not expand the review into unrelated architecture cleanup or re-review delivered runtime features.

## Completion validation criteria

The following are required execution checks, not results claimed by this generated plan.

| Slice | Focused validation expectation |
| --- | --- |
| Slice 1 | Baseline/readiness regressions covering saturated samples, nested fixtures, incomplete coverage, and existing/setup classification. |
| Slice 2 | Context regressions covering migration resolution, containment, ambiguity, deduplication, provenance, and archive/raw exclusions. |
| Slice 3 | Context-request regressions covering decision discovery, precedence, relevance, and single-pass/staged parity. |
| Slice 4 | Plan-contract and replacement regressions covering dependency errors, readiness obligations, invalid output, cancellation, and stale source observations. |
| Slice 5 | Application and core materialization regressions covering task cardinality, stable retry identities, dependency mapping, recovery, legacy mode, and projection failure. |
| Slice 6 | Combined planning-continuity acceptance fixture and read-only status verification. |

1. Format changed Go files and run the smallest affected test selection before each task commit. Record exact selectors and outcomes in the ticket evidence and slice report.
2. Run `make verify` before each slice merge. Explicitly select affected workflow/filesystem integration regressions without `-short`; add focused `-race` execution only where ownership or concurrent mutation changes.
3. Run `git diff --check`. Documentation-only task validation also checks references and agreement with the accepted contracts.
4. Run the closing `make verify` gate against the exact integrated `dev` candidate with a frozen plan/member snapshot. Confirm all twelve task tickets share the correct six slice receipts and no unresolved blocker remains.
5. Verify the acceptance fixture creates six task tickets for two three-task slices, preserves identities on retry, and cannot turn a stale approval or tracker edit into execution authority.
6. Preserve supplied delivered review, merger, and promotion-gate capabilities. Do not claim live-provider certification or exhaustive validation from local fixture results.
7. Leave exhaustive normal/race/browser/live checks to the existing scheduled nightly process. Do not dispatch or emulate that process during ordinary slice delivery.
8. Close the delivery set only after its required exact-candidate gate passes, reports and authoritative records agree, and the tracker projection is current or explicitly marked stale without being used as execution authority.

## Provenance

- source: project_context
- source: request_context
- source: quiz_context
- source: request_context
- source: AGENTS.md
- source: .chr/default/project.md
- source: ops/default/dump/00001-curated/artifact.md
- source: ops/default/dump/00002-raw/artifact.md
- source: ops/default/dump/00037-2026-06-02-merger-agent-role/artifact.md
- source: ops/default/dump/00038-dev-main-gate-context/artifact.md
- source: ops/default/dump/00039-document-generation-context/artifact.md
- source: ops/default/dump/00040-project-context/artifact.md
- source: ops/default/dump/00041-runtime-execution-context/artifact.md
- source: ops/default/dump/00042-2026-06-02-merger-agent-role-raw/artifact.md
- source: ops/default/plan/00008-dev-main-gate-plan-md/artifact.md
- source: ops/default/plan/00009-merger-agent-workflow-plan-md/artifact.md
- source: ops/default/plan/00010-ticket-review-integration-plan-md/artifact.md
- source: ops/default/plan/00026-2026-06-01-dev-main-agent-evidence-gate-plan-md/artifact.md
- source: ops/default/plan/00027-2026-06-06-gherkin-acceptance-generation-plan-md/artifact.md
- source: ops/default/plan/00028-2026-06-06-separate-acceptance-feature-workflow-plan-md/artifact.md
- source: ops/default/plan/00029-2026-06-18-agent-work-observability-plan-md/artifact.md
- source: ops/default/plan/00030-2026-06-18-configurable-work-readiness-plan-md/artifact.md
- source: ops/default/plan/00031-2026-06-18-tictactoe-qa-reference-plan-md/artifact.md
- source: ops/default/plan/00032-2026-06-19-agent-test-evidence-strategy-pass-plan-md/artifact.md
- source: ops/default/plan/00033-2026-06-19-isolated-ticket-worktrees-plan-md/artifact.md
- source: ops/default/plan/00034-2026-06-20-serialized-integration-loop-plan-md/artifact.md
- source: ops/default/plan/00035-2026-06-29-optional-local-web-ui-plan-md/artifact.md
- source: ops/default/plan/00048-dev-main-gate-plan/artifact.md
- source: ops/default/plan/00049-merger-agent-workflow-plan/artifact.md
- source: ops/default/plan/00050-ticket-review-integration-plan/artifact.md
- source: ops/default/tracker/00051-merger-agent-workflow-tracker/artifact.md
- source: ops/default/tracker/00052-ticket-review-integration-tracker/artifact.md
- source: ops/default/quiz/00003-curated/artifact.md
- source: ops/default/quiz/00004-raw/artifact.md
- source: ops/default/quiz/00043-ticket-review-integration-quiz/artifact.md
- source: ops/default/quiz/00044-ticket-review-integration-quiz-raw/artifact.md
- source: ops/default/spec/00005-dev-main-gate-spec-md/artifact.md
- source: ops/default/spec/00006-merger-agent-workflow-md/artifact.md
- source: ops/default/spec/00007-ticket-review-integration-draft-md/artifact.md
- source: ops/default/spec/00011-2026-06-01-dev-main-agent-evidence-gate-md/artifact.md
- source: ops/default/spec/00012-2026-06-06-gherkin-acceptance-generation-md/artifact.md
- source: ops/default/spec/00013-2026-06-06-separate-acceptance-feature-workflow-md/artifact.md
- source: ops/default/spec/00014-2026-06-18-agent-work-observability-model-md/artifact.md
- source: ops/default/spec/00015-2026-06-18-agent-work-observability-md/artifact.md
- source: ops/default/spec/00016-2026-06-18-agentic-loop-review-merge-md/artifact.md
- source: ops/default/spec/00017-2026-06-18-configurable-work-readiness-model-md/artifact.md
- source: ops/default/spec/00018-2026-06-18-configurable-work-readiness-md/artifact.md
- source: ops/default/spec/00019-2026-06-19-agent-test-evidence-strategy-pass-md/artifact.md
- source: ops/default/spec/00020-2026-06-19-isolated-ticket-worktrees-md/artifact.md
- source: ops/default/spec/00021-2026-06-20-serialized-integration-loop-model-md/artifact.md
- source: ops/default/spec/00022-2026-06-20-serialized-integration-loop-md/artifact.md
- source: ops/default/spec/00023-2026-06-29-execution-unit-runtime-refinements-md/artifact.md
- source: ops/default/spec/00024-2026-06-29-optional-local-web-ui-md/artifact.md
- source: ops/default/spec/00025-drafts/artifact.md
- source: ops/default/spec/00045-dev-main-gate-spec/artifact.md
- source: ops/default/spec/00046-merger-agent-workflow/artifact.md
- source: ops/default/spec/00047-ticket-review-integration-draft/artifact.md
- source: deterministic_baseline_scan
- source: deterministic_readiness_gate
- source: semantic_slot_catalog
