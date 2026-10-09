# Ticket-Native Work Delivery

Status: active

Date: 2026-10-08

Approved: 2026-10-08; maintainer requested implementation through slice-loop.

Delivery set: ticket-native-workflow

Specification: [Ticket-Native Work Execution](../spec/stable/ticket-native-workflow.md)

Model: [Ticket-Native Work Model](../spec/stable/ticket-native-workflow-model.md)

Tracker: [Ticket-Native Work Tracker](../tracker/ticket-native-workflow.md)

Specification approval: accepted on 2026-10-08

Plan approval: approved on 2026-10-08

Implementation authorization: granted on 2026-10-08 for the approved delivery set through slice-loop

Active slice: Slice 10 — Shared operation controls

Next slice: Slice 11 — Defaults and authority transition

Slice strategy: behavior-first

Reason: durable storage, isolated execution, verified publication and ordered
continuation each need their complete failure behavior before the next capability
can use them. Vertical increments let each slice pass real boundary tests and
keep the existing workflow usable. Contracts, mappings, logic and focused tests
are delivered together; preparatory interfaces without working behavior are not
separate slices.

## Purpose and Delivery Boundary

Deliver the approved ticket-native workflow through the central operation
service, with one task ticket per task and one canonical PR per slice. Preserve
existing project behavior while adding explicit ordered slice/set/use-case
execution, automatic integration, bounded independent workers, and shared
CLI/TUI/Web/daemon/manual controls.

The implementation baseline inspected during planning is
`d36dee206c3e4ad3ab6a6230365fd0812b5923a9` on `dev`. Recheck current `dev`,
`origin/dev`, worktrees, relevant tickets and PRs before activation. Do not reset
dirty checkouts or remove unrelated branches, worktrees or operational artifacts.

The maintainer authorized slice-loop to implement this exact delivery set and
to manage Rebase + Fast-forward merges of its twelve canonical slice PRs into
`dev`. The external controller owns merge mutations; implementation workers do
not merge PRs. This delegation does not change the product publication policy. Historical batch conversion, new forge providers,
Telegram workflow controls, releases and promotion to `main` are outside this set.

## Entry and Activation Gates

1. The specification and model are accepted together. This gate is satisfied.
2. The maintainer approves this plan and tracker and explicitly starts the set.
   These gates are satisfied by the 2026-10-08 request to start slice-loop.
3. The approved artifacts are committed to `dev` before a slice branch is created.
   The bootstrap commit versions this approved set before controller execution.
4. Materialize durable task tickets for the approved task map through the available
   repository ticket workflow and record their keys in the tracker before their
   tasks become executable. Existing ticket representations can bootstrap this
   delivery; do not mark them as the new schema before support is delivered.
   All 48 task keys are bound in the tracker using the existing Markdown workflow.
5. Verify current branch, canonical `origin/dev`, tracker frontier and open PRs.
   Slice 1 is active: baseline `d36dee206c3e4ad3ab6a6230365fd0812b5923a9`
   matches canonical `dev`, all task keys are bound, and no open PR was observed.
   A dedicated clean controller checkout preserves unrelated work and the
   original local `dev` commit. Recheck these observations at controller startup.

Every later slice requires its predecessor merged into `dev`, its report/tracker
closed with the verified merge commit, and a refreshed baseline. Future branch
names, PR titles and report paths below are reserved delivery metadata, not
claims that those objects exist.

## Shared Implementation Rules

- Keep one application operation service in `internal/app`; domain contracts and
  persistence stay in `internal/core`. Extend existing adapters rather than
  introducing entity services or adapter-specific schedulers.
- Use project-local shared records as authority. Extend the existing SQLite
  migration/index path only for rebuildable queries and diagnostic links.
- Keep the current coarse operation lock effective through Slice 8. Slice 9 may
  replace it only after repository-wide ownership, capacity and target-exclusion
  tests pass for every work-start entrypoint.
- Add tests for each slice's real behavior and failures in that slice. Do not
  defer basic recovery or race safety to final assurance.
- Keep version-1 defaults inactive for ordinary new projects until Slice 11.
  Earlier slices use explicit opt-in/test configuration and preserve old defaults.
- Provider rollout starts with the existing Forgejo adapter, then the existing
  GitHub adapter. Both use capability checks. A method without an atomic source/base
  guard remains explicitly blocked; no new provider or fallback is inferred.
- Task-ticket status is updated explicitly. The delivery tracker mirrors task
  evidence and PR integration; it cannot silently override task records.

## Validation Policy

During development, run the focused checks appropriate to the current task.
For the final immutable slice candidate, record `make verify` first, then the
listed focused checks in their recorded order, and `git diff --check` last.
This command order is the canonical slice-loop validation contract. Run race detection for touched ownership,
persistence, scheduling and adapter boundaries. Do not describe those focused
race checks as a full repository race gate.

Use real temporary records, subprocesses, linked Git worktrees and bare remotes
where those boundaries determine correctness. Provider doubles validate transport
contracts; they do not establish live provider support. Do not start ordinary
project services or publish test PRs in a production project for validation.

Every slice report records the exact candidate SHA, commands, outcomes, evidence
paths and material limitations. A failed required check prevents slice closure.
Report paths below are expected outputs and are created during their own slices.

## Controller Contract

Use the literal slice identities in the canonical table in both plan and tracker.
The controller runs one slice at a time, verifies its exact PR head, required
checks and canonical Forgejo integration, then continues the same approved set.
Use the exact PR body headings `## Summary`, `## Changes`, `## Validation`, in
that order. Summary is one prose paragraph. Changes are plain bullet items;
the final item is the slice report link at its full introduction commit.
Validation contains only successful backtick-wrapped commands: `make verify`,
the slice's focused checks in their recorded order, any additional focused
checks, then `git diff --check`. Follow this controller contract when an older
repository playbook example differs. Complete PR-reference commits and push
before returning the PR to the controller.

After a verified integration, close the slice report and its task tickets,
record the exact merge, and activate the next slice without another approval.
Update ticket status and containing directory together, and update tracker
links when moving tickets. Preserve task identities and evidence through any
explicit adoption introduced by this delivery set.

The exact final gate command is:

```sh
make verify && go test -race ./... && make qa-ticket-native-workflow && git diff --check
```

The composed QA target is delivered by Slice 12. A failing required gate or
missing live-provider authority blocks the affected claim; it does not authorize
production publication tests, extra slices or a weaker success claim. Keep local
controller configuration and checkpoints outside the repository.

## Ordered Slice Map

| Slice | Name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Guarded workflow records | `feat/slice-1-workflow-records` | `feat(slice-1): persist guarded workflow records` | `ops/default/report/slices/ticket-native-workflow/slice-1-workflow-records.md` |
| Slice 2 | Task-ticket materialization | `feat/slice-2-workflow-materialization` | `feat(slice-2): materialize ordered task tickets` | `ops/default/report/slices/ticket-native-workflow/slice-2-workflow-materialization.md` |
| Slice 3 | Deterministic work selection | `feat/slice-3-workflow-selection` | `feat(slice-3): select ordered workflow extents` | `ops/default/report/slices/ticket-native-workflow/slice-3-workflow-selection.md` |
| Slice 4 | Owned isolated task execution | `feat/slice-4-workflow-execution` | `feat(slice-4): execute owned isolated tasks` | `ops/default/report/slices/ticket-native-workflow/slice-4-workflow-execution.md` |
| Slice 5 | Slice candidates and delivery receipts | `feat/slice-5-workflow-delivery` | `feat(slice-5): verify aggregate slice delivery` | `ops/default/report/slices/ticket-native-workflow/slice-5-workflow-delivery.md` |
| Slice 6 | Guarded automatic publication | `feat/slice-6-workflow-publication` | `feat(slice-6): publish guarded integration candidates` | `ops/default/report/slices/ticket-native-workflow/slice-6-workflow-publication.md` |
| Slice 7 | Ordered set and use-case continuation | `feat/slice-7-workflow-continuation` | `feat(slice-7): continue ordered sets and use cases` | `ops/default/report/slices/ticket-native-workflow/slice-7-workflow-continuation.md` |
| Slice 8 | Loose-ticket integration loop | `feat/slice-8-workflow-loose-tickets` | `feat(slice-8): integrate filtered loose tickets` | `ops/default/report/slices/ticket-native-workflow/slice-8-workflow-loose-tickets.md` |
| Slice 9 | Independent parallel runners | `feat/slice-9-workflow-concurrency` | `feat(slice-9): coordinate independent parallel runners` | `ops/default/report/slices/ticket-native-workflow/slice-9-workflow-concurrency.md` |
| Slice 10 | Shared operation controls | `feat/slice-10-workflow-controls` | `feat(slice-10): expose shared workflow controls` | `ops/default/report/slices/ticket-native-workflow/slice-10-workflow-controls.md` |
| Slice 11 | Defaults and authority transition | `feat/slice-11-workflow-defaults` | `feat(slice-11): enable versioned ticket workflow defaults` | `ops/default/report/slices/ticket-native-workflow/slice-11-workflow-defaults.md` |
| Slice 12 | End-to-end workflow assurance | `test/slice-12-workflow-assurance` | `test(slice-12): verify ticket workflow end to end` | `ops/default/report/slices/ticket-native-workflow/slice-12-workflow-assurance.md` |

## Slice 1: Guarded workflow records

Branch: `feat/slice-1-workflow-records`

PR title: `feat(slice-1): persist guarded workflow records`

Report: `ops/default/report/slices/ticket-native-workflow/slice-1-workflow-records.md`

Outcome: CHR can persist and reload versioned workflow records, reject invalid
hierarchy and stale writers, and recover interrupted shared transitions without
depending on a tracker or SQLite cache.

Likely owners: `internal/core/ticket*.go`, `ops_artifact*.go`, new workflow record
files in `internal/core`, and the existing SQLite migration assets.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T1.1 | Define versioned task/set/slice/receipt identities, fields and lifecycle validation in native core types. | Invalid/partial associations, duplicate scoped identity/order and unsupported schemas reject; independent tickets remain valid. |
| T1.2 | Implement revision-guarded canonical and adopted-Markdown read/write with unknown-field preservation. | Round trips retain bodies, keys, creation times and compatible fields; stale competing writers cannot overwrite records. |
| T1.3 | Implement process-safe mutation locking and journaled multi-file transitions with guarded replay. | Subprocess/crash tests recover matching partial transitions and reject conflicting manual edits. |
| T1.4 | Extend rebuildable SQLite queries/migrations and record/file bounds. | Deleting/rebuilding the index preserves shared state and ordering; uniqueness and size-limit failures are explicit. |

Focused commands: `go test ./internal/core -count=1` and
`go test -race ./internal/core -count=1`.

Acceptance owned: A02, A22; foundations for A24.

## Slice 2: Task-ticket materialization

Branch: `feat/slice-2-workflow-materialization`

PR title: `feat(slice-2): materialize ordered task tickets`

Report: `ops/default/report/slices/ticket-native-workflow/slice-2-workflow-materialization.md`

Outcome: an approved plan creates stable ordered slices and a ticket for each
task, including resumable generation and selected legacy adoption.

Likely owners: `internal/core/plan_slice_detail.go`, `plan_tracking.go`,
`quiz_openai_model.go`, and `internal/app/service_plan.go`.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T2.1 | Extend plan-generation contracts to distinguish set, slice and task identity/order. | Multi-task slices validate as one integration unit; generated labels cannot replace stable identity. |
| T2.2 | Persist one ticket per task and bind task-owned membership to set-owned slice contracts. | A two-slice, three-task-per-slice fixture produces six correctly associated tickets. |
| T2.3 | Add materialization receipts and idempotent recovery after partial generation. | Retry after any created ticket reuses keys/receipts and creates no duplicate task or set. |
| T2.4 | Add explicit quiescent adoption for selected active legacy plans/tickets and update generation guidance. | Original content/history survives; ambiguous identities or live ownership block; ordinary reads never bulk-migrate. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Acceptance owned: A01, A02.

## Slice 3: Deterministic work selection

Branch: `feat/slice-3-workflow-selection`

PR title: `feat(slice-3): select ordered workflow extents`

Report: `ops/default/report/slices/ticket-native-workflow/slice-3-workflow-selection.md`

Outcome: one adapter-neutral preview resolves exact extents, prerequisites,
hierarchy order, exclusions and filtered loose-ticket ranks.

Likely owners: `internal/core/ticket_loop_query.go`, workflow query/dependency
files, `internal/app/service_work.go` and service request/result contracts.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T3.1 | Add exact ticket/slice/set/use-case/loose selectors and bounded preview results. | Missing identities never become backlog requests; unprepared and excluded records have explicit counts/reasons. |
| T3.2 | Resolve set-local/task-local order and homogeneous use-case set order. | Interleaved ticket creation cannot change execution order; mixed explicit/absent set orders reject. |
| T3.3 | Validate implicit and explicit prerequisites, including target delivery proof. | Missing, cyclic, forward or outside-selection dependencies block without admitting extra work. |
| T3.4 | Preserve native classification/rank rules and implement loose ownership exclusion and thresholds. | All four specification examples, rank/time/key ties and unclassified values return deterministic results. |

Focused commands: `go test ./internal/core ./internal/app -count=1`.

Acceptance owned: A04, A05, A07; query bounds in A24.

## Slice 4: Owned isolated task execution

Branch: `feat/slice-4-workflow-execution`

PR title: `feat(slice-4): execute owned isolated tasks`

Report: `ops/default/report/slices/ticket-native-workflow/slice-4-workflow-execution.md`

Outcome: an exact task executes once in an isolated lane, records validated
evidence and supports safe manual handoff, pause, cancellation and recovery.
Same-project execution remains serial in this slice.

Likely owners: `internal/core/execution*.go`, worktree/runtime/context boundaries,
and shared work-operation methods in `internal/app`.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T4.1 | Implement guarded invocation/attempt identity, exclusive claims and fencing generations under the existing serial lock. | Duplicate starts attach; overlapping selectors and stale owners cannot apply results. |
| T4.2 | Execute against a retained isolated cumulative lane branch with frozen inputs and host validation. | A grouped ticket validates alone without a partial PR; unchanged retry reuses its evidence/worktree. |
| T4.3 | Persist stop intent, lease/process ownership, attempt budgets and safe replacement. | Pause/cancel/crash tests retain evidence; a live child remains counted and expired tokens cannot finalize or release newer claims. |
| T4.4 | Expose the shared manual claim/checkpoint/evidence-reconciliation protocol. | Manual results undergo the same revisions, prerequisites and validation; status edits alone do not advance work. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; real worker subprocess
and linked-worktree recovery fixtures are required.

Acceptance owned: A06, A09, A12, A17, A21; execution budgets in A24.

## Slice 5: Slice candidates and delivery receipts

Branch: `feat/slice-5-workflow-delivery`

PR title: `feat(slice-5): verify aggregate slice delivery`

Report: `ops/default/report/slices/ticket-native-workflow/slice-5-workflow-delivery.md`

Outcome: ordered task results form one aggregate slice candidate and one
canonical PR. Verified external delivery can finalize its tasks once; automatic
target publication is added in Slice 6.

Likely owners: shared work service, task-result/status writers, review/merge
metadata and core candidate/receipt/transition boundaries.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T5.1 | Run remaining tasks cumulatively and host-validate aggregate slice acceptance on an immutable tree. | Local task validation does not mark `done`; aggregate failure retains reusable task evidence with affected evidence invalidated. |
| T5.2 | Create/reuse one canonical PR and retain branch/candidate/PR bindings. | Restart reuses the PR; a closed-unmerged or mismatched PR blocks rather than creating replacements. |
| T5.3 | Verify external integration through canonical receipts and method-specific target/tree/history proof. | Real bare-remote merge/rebase/squash fixtures reject worker claims, wrong branches and unverifiable receipts. |
| T5.4 | Journal shared receipt creation and task-status finalization. | Crash during fan-out replays once; every task shares the same delivery receipt and no per-task PR is created. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Acceptance owned: A01, A08, A13, A22.

## Slice 6: Guarded automatic publication

Branch: `feat/slice-6-workflow-publication`

PR title: `feat(slice-6): publish guarded integration candidates`

Report: `ops/default/report/slices/ticket-native-workflow/slice-6-workflow-publication.md`

Outcome: human, agentic and trusted review policies drive a retained candidate
through supported guarded publication or an explicit resumable wait/blocker.

Likely owners: existing review/merge preflight and provider adapters, integration
cycle policy, publication journal and shared work-operation methods.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T6.1 | Bind review decisions, bounded repairs and exact-candidate manual approval to current policy and immutable inputs. | Changed content/policy invalidates approval; the three modes produce their distinct expected stops. |
| T6.2 | Prepare and validate configured merge/rebase/squash results and enforce atomic source/base publication guards. | Supported Forgejo/GitHub paths prove the required result; unsupported guards/methods block without fallback. |
| T6.3 | Journal submission before provider IO and retain target exclusion for uncertain responses. | Lost-response/restart fixtures observe actual delivery before retry; budgets survive candidate revision and continuation. |
| T6.4 | Reconcile target movement, external delivery and stale candidate evidence through the shared operation. | A moved base forces fresh preparation/validation/review; verified delivery finalizes without another worker or submission. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`. Exercise real target
movement and controlled provider responses; live-provider claims require Slice 12
evidence on designated disposable repositories.

Acceptance owned: A11, A13, A14, A15, A21; publication budgets in A24.

## Slice 7: Ordered set and use-case continuation

Branch: `feat/slice-7-workflow-continuation`

PR title: `feat(slice-7): continue ordered sets and use cases`

Report: `ops/default/report/slices/ticket-native-workflow/slice-7-workflow-continuation.md`

Outcome: one invocation completes several ordered slices and PRs, rehydrates
after every integration, closes the set and resumes idempotently.

Likely owners: work-operation traversal, durable invocation cursor, dependency
checks and set-closure records/validation.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T7.1 | Traverse slice/set/use-case extents through the verified integration barrier. | Slice 2 cannot start before Slice 1 finalizes; unsliced use-case tickets remain excluded. |
| T7.2 | Persist equivalent-request identity, confirmed cursor and bounded continuation checkpoints. | Restart after task validation, PR creation or integration skips confirmed work and cannot broaden the extent. |
| T7.3 | Implement frozen/default and explicit later-admission rules, quiescent amendments and set reopening. | Retroactive/reordered work blocks; later admission stays after the frontier; terminal invocations and previous closures remain immutable. |
| T7.4 | Confirm closing validation against all actual members and exact target/plan snapshots. | Changed members/target prevent closure; a failed gate retries closure without re-executing integrated tasks; empty/waived gates remain explicit. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; a multi-slice real-Git
scenario must create separate canonical PR fixture identities and integrate in order.

Acceptance owned: A03, A04, A07, A13, A19, A20; checkpoint bounds in A24.

## Slice 8: Loose-ticket integration loop

Branch: `feat/slice-8-workflow-loose-tickets`

PR title: `feat(slice-8): integrate filtered loose tickets`

Report: `ops/default/report/slices/ticket-native-workflow/slice-8-workflow-loose-tickets.md`

Outcome: independent tickets execute, validate, review and integrate serially
under deterministic filters, with fresh authoritative selection after each merge.

Likely owners: ticket-loop query, shared work-operation traversal and existing
`work ticket` / `work run-all` compatibility mapping.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T8.1 | Compose exact independent-ticket and loose extents with the common candidate/integration lifecycle. | Independent work requires no fabricated use case/set/slice and completes only with verified delivery. |
| T8.2 | Reselect eligible loose work after each integration without claiming slice-owned tickets. | Changed rank/status and delivered prerequisites are observed; all requested filter/order examples pass. |
| T8.3 | Preserve bounded selection, blocker-stop, duplicate-request and resume behavior. | A failed ticket stops its invocation; terminal retry and outside-scope dependencies never admit extra work. |
| T8.4 | Map legacy `run-all` and single-ticket entrypoints according to explicit authority mode. | Ticket mode reports excluded grouped work; transitional scope remains visible and cannot execute the same work twice. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Acceptance owned: A05, A07, A13, A19.

## Slice 9: Independent parallel runners

Branch: `feat/slice-9-workflow-concurrency`

PR title: `feat(slice-9): coordinate independent parallel runners`

Report: `ops/default/report/slices/ticket-native-workflow/slice-9-workflow-concurrency.md`

Outcome: opt-in capacity permits independent lanes to overlap while task order,
slice integration barriers and publication exclusion remain enforced.

Likely owners: repository coordination, work claims, operation preflight,
execution reservations, worktree identity and target-publication ownership.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T9.1 | Introduce repository-wide process reservations and validated capacity/queue admission across worktrees/namespaces. | Capacity one serializes and capacity two overlaps independent lanes; aliases/config conflicts cannot create extra pools. |
| T9.2 | Enforce exclusive task/slice/set/use-case frontier claims and fenced result application on every start path. | CLI/TUI/Web/daemon/manual/legacy starts cannot bypass a live owner or dependency; unsupported cross-host coordination blocks. |
| T9.3 | Serialize publication per canonical remote target, including uncertain outcomes and external target movement. | Competing lanes never submit simultaneously; human waits release worker capacity; stale combined results revalidate before publication. |
| T9.4 | Replace the coarse lock only after subprocess/crash/race acceptance passes and document the ownership boundary. | Expired live workers retain slots, replaced owners cannot release newer claims, and unrelated lanes survive another lane's blocker. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; subprocess barriers must
prove actual overlap/exclusion, not infer it from a race-detector pass.

Acceptance owned: A09, A10, A11, A12; capacity/admission bounds in A24.

## Slice 10: Shared operation controls

Branch: `feat/slice-10-workflow-controls`

PR title: `feat(slice-10): expose shared workflow controls`

Report: `ops/default/report/slices/ticket-native-workflow/slice-10-workflow-controls.md`

Outcome: CLI, TUI, Web and daemon expose the same work operations, progress,
recovery, stop and approval semantics without independent scheduling logic.

Likely owners: `internal/app/work.go`, TUI adapter methods, `internal/tui`,
`internal/webui`, daemon coordination and adapter-neutral service result types.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T10.1 | Expose proposed CLI extents, preview/status/continue and exact-bound approval/reconciliation controls. | Input errors remain bounded; previews do not mutate; reported IDs/evidence distinguish task validation from slice integration. |
| T10.2 | Add TUI controls and progress/wait/blocker rendering over shared operations. | Selected identity and revision reach the service unchanged; stale/conflicting actions are visible. |
| T10.3 | Add Web operation controls, guarded writes and resumable progress with visible failure envelopes. | Real handler/browser checks cover activation, stale ownership/revisions, review wait and continuation. |
| T10.4 | Wire daemon resume, pause/cancel and manual status parity; document Telegram's notification-only boundary. | Startup resumes only retained authorization, never the whole backlog; adapter contract tests prove shared selection and lifecycle. |

Focused commands: `go test ./internal/app ./internal/tui ./internal/webui -count=1`
and `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`;
extend repository-owned browser smoke validation for the new controls.

Acceptance owned: A09, A17, A21, A23.

## Slice 11: Defaults and authority transition

Branch: `feat/slice-11-workflow-defaults`

PR title: `feat(slice-11): enable versioned ticket workflow defaults`

Report: `ops/default/report/slices/ticket-native-workflow/slice-11-workflow-defaults.md`

Outcome: every new-project path agrees on the approved defaults; existing
projects keep effective policy and adopt selected active work deliberately.
Tracker export remains a repairable projection.

Likely owners: config parsing/resolution, embedded workspace profiles and bootstrap,
selected adoption operations, tracker generation and User Guide/reference docs.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T11.1 | Deliver version marker, approved configuration keys, bounds and effective-value/source reporting. | Invalid/conflicting values block; invocation restrictions and current publication policy have documented precedence. |
| T11.2 | Align standard/minimal profiles, bootstrap and fallback defaults for new projects. | Every new path selects tickets/worktrees/trusted auto-merge/capacity one; repeated init does not reset existing policy. |
| T11.3 | Upgrade selected projects with prior-policy snapshots and explicit authority transition. | Legacy absent/explicit policy survives; ambiguous or actively owned work blocks adoption; no historical batch conversion occurs. |
| T11.4 | Export/repair trackers one way and publish the complete operation/configuration/manual guidance. | Export failure reports staleness without losing progress; edited projections are preserved; documentation distinguishes current delivered capabilities from deferred surfaces. |

Focused commands: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; cover every generator
and effective-default path with a new/existing/reinitialization matrix.

Acceptance owned: A16, A18; explicit adoption and configuration bounds in A24.

## Slice 12: End-to-end workflow assurance

Branch: `test/slice-12-workflow-assurance`

PR title: `test(slice-12): verify ticket workflow end to end`

Report: `ops/default/report/slices/ticket-native-workflow/slice-12-workflow-assurance.md`

Outcome: reproducible composed scenarios and exact-commit evidence establish the
approved scope across storage, execution, parallel lanes, adapters and supported
publication methods. This is a planned assurance slice, not unbounded cleanup.

Likely owners: package integration fixtures, `qa/scenarios/ticket-native-workflow/`,
repository-owned QA commands/Makefile, compatibility evidence and User Guide.

| Task | Work | Done when / focused evidence |
| --- | --- | --- |
| T12.1 | Compose real-files/subprocess/Git scenarios for ordered slices, loose queues and parallel independent work. | A01-A24 are linked to deterministic evidence; scopes, PR count and verified integration barriers are observable. |
| T12.2 | Add restart/crash/lost-response/manual-edit fault scenarios and a bounded repository QA entrypoint. | The documented new `make qa-ticket-native-workflow` target reproduces those scenarios without production state or unbounded waits. |
| T12.3 | Verify supported existing provider/method paths in designated disposable live repositories. | Source/base guarding and actual resulting history are proven per claimed path; unavailable access/capabilities remain explicit blockers rather than simulated success. |
| T12.4 | Run final exact-candidate gates, reconcile the acceptance matrix and publish the cycle evidence/report. | Configured full checks, QA and race suite pass; every approved requirement is delivered or a blocking unmet criterion is recorded; no release or main promotion is implied. |

Focused commands: run the composed scenarios via the QA entrypoint delivered by
T12.2, then the standard `make verify`, `go test -race ./...` and
`git diff --check` on one immutable candidate. Future `make qa-ticket-native-workflow`
is not a command available at planning time.

Acceptance owned: composed confirmation of A01-A24 and live-provider limitations.

## Acceptance Ownership Matrix

| Criterion | Primary delivery slices |
| --- | --- |
| A01 | 2, 5 |
| A02 | 1, 2 |
| A03 | 6, 7 |
| A04 | 3, 7 |
| A05 | 3, 8 |
| A06 | 4 |
| A07 | 3, 4, 7, 8 |
| A08 | 5 |
| A09 | 4, 9, 10 |
| A10 | 9 |
| A11 | 6, 9 |
| A12 | 4, 9 |
| A13 | 5, 6, 7, 8 |
| A14 | 6 |
| A15 | 6 |
| A16 | 11 |
| A17 | 4, 10 |
| A18 | 11 |
| A19 | 7, 8 |
| A20 | 7 |
| A21 | 4, 6, 10 |
| A22 | 1, 5 |
| A23 | 10 |
| A24 | 1, 3, 4, 6, 7, 9, 11 |

## Delivery Risks and External Prerequisites

| Concern | Required handling |
| --- | --- |
| Existing dirty work and stale active branches | Preserve it; every slice starts from verified current `dev` in its named worktree. |
| Host-specific SQLite diagnostics | Reconstruct execution decisions from shared records/receipts; cache loss is tested before depending on recovery. |
| Source/base atomic publication support | Audit existing adapter capabilities in Slice 6; validate claimed paths live in Slice 12. Unsupported paths block rather than weakening the accepted contract. |
| Lease expiry while a child survives | Process-held reservations and fenced writes are delivered in Slice 4 before parallel capacity is enabled. |
| Coarse-lock removal | Explicitly deferred until Slice 9's replacement coordination passes all start-path checks. |
| Live-provider validation | A designated disposable repository, permissions and credentials are external prerequisites. Record absence without altering production policy or inventing evidence. |
| New defaults applied too early | Keep explicit opt-in through Slice 10 and change all new-project paths together in Slice 11. |

## Closure

The set closes only after all 48 task tickets are complete, all 12 slice PRs are
verified integrated into `dev`, and every canonical slice report is delivered.
Record one final exact-`dev` validation after the last merge with `make verify`,
the full race suite and the composed QA target. Do not repeat an unchanged green
candidate's gate without a new concern. If integration changes the candidate,
record fresh evidence for the resulting commit.

The final cycle report is
`ops/default/report/ticket-native-workflow-cycle.md`. It links the 24 acceptance
criteria to evidence and distinguishes real provider validation from controlled
adapters. Missing required evidence or behavior prevents closure. Corrective work
must remain within the approved requirement map and be reflected in the tracker;
an added scope requires an explicit plan amendment.

## Open Questions

No product-contract question blocks execution. Plan/tracker approval and
implementation authorization are recorded above; the bootstrap commit versions
the approved set and its 48 task bindings with Slice 1 active. The particular disposable provider repositories/access are resolved
before live validation; they are not selected by this document.
