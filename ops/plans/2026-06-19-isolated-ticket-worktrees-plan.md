# Isolated Ticket Worktrees Plan

Status: in_progress

Active Slice: none

Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`

Spec: `ops/specs/2026-06-19-isolated-ticket-worktrees.md`

## Purpose

Move ticket execution from the user's active checkout into a ChrysoCoord-owned
Git worktree while preserving the current sequential agent model.

This delivery must keep one active ticket and one active runtime worker at a
time. The goal is not parallelism. The goal is to stop treating the user's
checkout as the writable agent workspace.

The target runtime shape is:

```text
user checkout
  human-owned; may be dirty or stale

origin/<integration_branch>
  fetched published base for agent work

ticket worktree
  ChrysoCoord-owned execution workspace
  mounted into Docker as /workspace
  kept until the ticket branch / PR lifecycle completes
```

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/isolated-ticket-worktrees/`.
- Each task gets one conventional commit.
- Plan and tracker must be committed before runtime code changes begin.
- After each merged slice, return to `dev`, sync, and validate before starting
  the next slice.
- Do not switch the default for new projects until the QA proving slice passes
  end to end.
- Do not delete dirty worktrees automatically in any slice.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | workspace execution config | `feat/slice-1-workspace-execution-config` | `feat(slice-1): workspace execution config` | `.prj/reports/slices/isolated-ticket-worktrees/slice-1-workspace-execution-config.md` |
| Slice 2 | worktree state service | `feat/slice-2-worktree-state-service` | `feat(slice-2): worktree state service` | `.prj/reports/slices/isolated-ticket-worktrees/slice-2-worktree-state-service.md` |
| Slice 3 | readiness workspace split | `feat/slice-3-readiness-workspace-split` | `feat(slice-3): readiness workspace split` | `.prj/reports/slices/isolated-ticket-worktrees/slice-3-readiness-workspace-split.md` |
| Slice 4 | runtime workspace mount | `feat/slice-4-runtime-workspace-mount` | `feat(slice-4): runtime workspace mount` | `.prj/reports/slices/isolated-ticket-worktrees/slice-4-runtime-workspace-mount.md` |
| Slice 5 | execution workspace results | `feat/slice-5-execution-workspace-results` | `feat(slice-5): execution workspace results` | `.prj/reports/slices/isolated-ticket-worktrees/slice-5-execution-workspace-results.md` |
| Slice 6 | branch and PR lifecycle | `feat/slice-6-branch-and-pr-lifecycle` | `feat(slice-6): branch and PR lifecycle` | `.prj/reports/slices/isolated-ticket-worktrees/slice-6-branch-and-pr-lifecycle.md` |
| Slice 7 | review repair and cleanup | `feat/slice-7-review-repair-and-cleanup` | `feat(slice-7): review repair and cleanup` | `.prj/reports/slices/isolated-ticket-worktrees/slice-7-review-repair-and-cleanup.md` |
| Slice 8 | docs and QA proving | `feat/slice-8-docs-and-qa-proving` | `docs(slice-8): docs and QA proving` | `.prj/reports/slices/isolated-ticket-worktrees/slice-8-docs-and-qa-proving.md` |

## Slice 1: Workspace Execution Config

Goal: introduce the configuration and core value model for execution workspaces
without changing ticket execution behavior.

Behavior:

- Workspace config can express whether execution uses the user checkout or an
  isolated ticket worktree.
- Defaults preserve current behavior until later slices switch execution.
- Config validation rejects unknown workspace mode, base policy, and cleanup
  policy values.
- Runtime/work readiness code can read the new values without changing current
  Docker mounts or preflight outcomes.
- Docs identify the fields as planned/available configuration, not as fully
  enforced worktree execution yet.

Tasks:

- T1.1 Add workspace execution config fields under the existing work/runtime
  configuration boundary.
- T1.2 Define controlled values for workspace mode, worktree base policy,
  worktree cleanup policy, and worktree root.
- T1.3 Set defaults that preserve current project-mounted execution behavior in
  existing workspaces.
- T1.4 Add config parser and validation coverage for omitted defaults, explicit
  worktree settings, and invalid enum values.
- T1.5 Add runtime plan value fields for project checkout vs execution
  workspace without changing harness mount behavior.
- T1.6 Document the config fields and migration posture in reference docs.
- T1.7 Add the slice implementation report and validation evidence.

## Slice 2: Worktree State Service

Goal: add a core service that can plan, inspect, create, reuse, and describe
ticket worktrees without invoking the runtime worker.

Behavior:

- Core can resolve the default base ref as
  `<canonical_remote>/<integration_branch>`, normally `origin/dev`.
- Core fetches or requests fetch before relying on the remote integration ref.
- Core can create a ticket branch from the remote base when missing.
- Core can create a deterministic ticket worktree path under the configured
  worktree root.
- Core can inspect existing ticket worktrees and classify clean, dirty,
  missing, mismatched, or orphaned states.
- Core records enough metadata for humans and later commands to understand the
  worktree, but Git remains the source of truth.

Tasks:

- T2.1 Add a Git worktree planning model with ticket key, branch, base ref,
  base SHA, root, and target path.
- T2.2 Add a command abstraction for `git fetch`, `git rev-parse`,
  `git branch`, `git worktree add`, `git worktree list`, and `git status`.
- T2.3 Implement deterministic worktree path selection using the configured
  root and ticket key.
- T2.4 Implement safe creation from the fetched remote integration ref.
- T2.5 Implement inspection and diagnostics for missing, clean, dirty,
  mismatched, branch-in-use, and orphaned worktree states.
- T2.6 Add optional metadata read/write for ticket worktree identity and last
  execution hints.
- T2.7 Add table-driven service tests with fake Git command output and temp
  repositories where useful.
- T2.8 Add the slice implementation report and validation evidence.

## Slice 3: Readiness Workspace Split

Goal: make work readiness and doctor distinguish the user's checkout from the
ticket execution workspace.

Behavior:

- In legacy `project` mode, current readiness behavior remains unchanged.
- In `worktree` mode, dirty user checkout state is informational, not blocking.
- In `worktree` mode, dirty ticket worktree state is blocking unless explicitly
  resumable in a later slice.
- Missing or stale remote integration refs block execution before worktree
  creation.
- Doctor reports `workspace_mode`, base ref, user checkout state, and ticket
  worktree state separately.
- Safe repair fetches the canonical remote and creates missing safe worktree
  roots, but does not delete or reset user work.

Tasks:

- T3.1 Extend work readiness input with workspace mode and optional ticket key.
- T3.2 Split dirty-state diagnostics into user checkout and ticket worktree
  diagnostics.
- T3.3 Make user checkout dirtiness informational in worktree mode.
- T3.4 Make ticket worktree dirtiness blocking in worktree mode.
- T3.5 Add remote-base diagnostics for failed fetch and missing
  `<remote>/<integration_branch>`.
- T3.6 Update `chc work doctor` output and service models to expose both
  workspace states.
- T3.7 Add safe repair coverage for fetch, root creation, and non-destructive
  worktree metadata pruning.
- T3.8 Add CLI/service/TUI tests for legacy mode and worktree mode readiness.
- T3.9 Add the slice implementation report and validation evidence.

## Slice 4: Runtime Workspace Mount

Goal: make the Docker harness mount the execution workspace instead of the user
checkout when worktree mode is enabled.

Behavior:

- Runtime planning resolves both `ProjectDir` and `ExecutionWorkspaceDir`.
- In project mode, `ExecutionWorkspaceDir == ProjectDir`.
- In worktree mode, `ExecutionWorkspaceDir` points at the prepared ticket
  worktree.
- Docker build behavior remains compatible with the configured runtime
  Dockerfile.
- Docker run and validation commands mount the execution workspace at the
  configured container workdir.
- Container paths remain stable; the worker still sees `/workspace` by default.
- Worker context identifies both the user checkout and execution workspace.

Tasks:

- T4.1 Extend `RuntimePlan` with execution workspace host path and workspace
  mode metadata.
- T4.2 Update runtime mount planning to use `ExecutionWorkspaceDir` as the
  writable project workspace mount source.
- T4.3 Preserve existing project-mode mount output exactly where possible.
- T4.4 Update context artifact link rendering so container paths are derived
  from the execution workspace.
- T4.5 Add worker context fields for project checkout, execution workspace,
  base ref, base SHA, and ticket branch.
- T4.6 Update runtime validation output to show the execution workspace.
- T4.7 Add harness invocation and runtime plan tests for project and worktree
  modes.
- T4.8 Add the slice implementation report and validation evidence.

## Slice 5: Execution Workspace Results

Goal: make execution records, reports, changed-file discovery, and result
application use the execution workspace while preserving project-relative
outputs.

Behavior:

- Execution records capture the workspace mode and execution workspace path.
- Runtime reports show where the worker actually ran.
- Changed-file discovery inspects the execution workspace, not the user
  checkout.
- Changed file paths remain project-relative in result JSON, reports, gates,
  and ticket comments.
- Validation runs in the same execution workspace as the worker.
- Existing evidence strategy and ready-for-review gates continue to work.

Tasks:

- T5.1 Extend execution record/runtime report models with workspace mode,
  execution workspace path, base ref, base SHA, and ticket branch.
- T5.2 Add persistence migration or compatible report-only fallback for the new
  execution workspace metadata.
- T5.3 Update changed-file discovery to accept an explicit workspace root.
- T5.4 Ensure discovered paths are normalized relative to the project root
  semantics, not absolute worktree paths.
- T5.5 Update result application to load execution workspace metadata and
  inspect the right Git state.
- T5.6 Add tests proving user checkout changes are ignored during result
  discovery in worktree mode.
- T5.7 Add tests proving validation and evidence gates still use the same
  result contract.
- T5.8 Add the slice implementation report and validation evidence.

## Slice 6: Branch and PR Lifecycle

Goal: connect ticket worktrees to the one-ticket, one-branch, one-PR lifecycle.

Behavior:

- A ticket branch is created from the fetched remote integration branch when it
  is missing.
- The ticket worktree checks out that ticket branch.
- Existing branch metadata in the ticket is authoritative when present and
  valid.
- Branch conflicts block instead of forcing reset or overwriting history.
- Commit, push, and PR creation continue to happen through host-side CHC
  orchestration, not from inside the worker container.
- `run-all` still executes one ticket at a time and prepares one worktree at a
  time.

Tasks:

- T6.1 Define branch selection rules for tickets with and without existing
  `branch` metadata.
- T6.2 Integrate worktree preparation into ticket start/work execution before
  runtime context materialization.
- T6.3 Add branch creation from `origin/<integration_branch>` with conflict
  diagnostics.
- T6.4 Ensure worktree checkout and branch state are recorded in ticket,
  execution, and event evidence.
- T6.5 Confirm host-side commit/push/PR creation reads from the ticket worktree
  in worktree mode.
- T6.6 Add run-all tests proving sequential worktree preparation without
  pre-creating the full queue.
- T6.7 Add conflict tests for branch already exists with unrelated history or
  unsafe remote state.
- T6.8 Add the slice implementation report and validation evidence.

## Slice 7: Review Repair and Cleanup

Goal: reuse ticket worktrees for review repair and clean them after successful
merge when safe.

Behavior:

- Review repair reuses the ticket worktree associated with the ticket branch.
- Missing clean worktrees can be recreated from the existing ticket branch.
- Dirty repair worktrees block with a clear diagnostic.
- Merge cleanup removes only clean worktrees for merged tickets.
- Dirty worktrees are never deleted automatically.
- Cleanup is idempotent and safe to run multiple times.
- Doctor can list or explain orphaned worktrees.

Tasks:

- T7.1 Add review-repair worktree lookup using ticket branch and PR metadata.
- T7.2 Add safe recreation of missing repair worktrees from valid ticket
  branches.
- T7.3 Add dirty repair worktree diagnostics and stop conditions.
- T7.4 Add cleanup operation for clean merged ticket worktrees.
- T7.5 Add cleanup blockers for dirty, mismatched, or unknown worktrees.
- T7.6 Add doctor reporting and safe repair hooks for orphaned clean worktrees.
- T7.7 Add tests for repair reuse, missing-worktree recreation, dirty blockers,
  idempotent cleanup, and no dirty deletion.
- T7.8 Add the slice implementation report and validation evidence.

## Slice 8: Docs and QA Proving

Goal: document the worktree execution model and prove it against resettable QA
projects before considering a default switch.

Behavior:

- Reference docs explain project mode vs worktree mode.
- User guide explains that the user's checkout is not edited during worktree
  execution.
- Ticket execution docs identify the Docker mount source as the execution
  workspace.
- Work readiness docs explain the new non-blocking dirty-user-checkout behavior.
- QA scenarios prove dirty user checkout, stale local dev, dirty ticket
  worktree blocking, review repair reuse, and merge cleanup.
- The tracker records whether worktree mode is ready to become the default for
  new projects.

Tasks:

- T8.1 Update runtime config, work readiness, ticket execution, and user guide
  docs.
- T8.2 Update Dash/TUI reference docs for workspace visibility.
- T8.3 Add or update QA reference fixtures to exercise isolated worktree mode.
- T8.4 Add smoke scripts or make targets for the worktree QA scenarios.
- T8.5 Run end-to-end QA proving with a dirty user checkout and a clean ticket
  worktree.
- T8.6 Record the default-switch decision in tracker notes with validation
  evidence.
- T8.7 Add the slice implementation report and final cycle report.

## Cross-Slice Dependencies

- Slice 2 depends on Slice 1 config and value types.
- Slice 3 depends on Slice 2 worktree inspection and diagnostics.
- Slice 4 depends on Slice 2 worktree preparation and Slice 3 readiness model.
- Slice 5 depends on Slice 4 runtime mount behavior.
- Slice 6 depends on Slice 5 result discovery and execution metadata.
- Slice 7 depends on Slice 6 branch/PR lifecycle state.
- Slice 8 depends on all runtime behavior slices.

## Validation Strategy

Each slice should run targeted package tests first, then `make verify` before
opening its PR.

Expected broad gate:

```sh
make verify
```

Additional likely focused commands:

```sh
go test ./internal/core -run 'RuntimePlan|Worktree|WorkReadiness|Harness|Execution'
go test ./internal/app -run 'Work|Ticket|Runtime|ApplyResult'
go test ./internal/tui -run 'Work|Dash|Runtime'
```

QA proving should add project-level smoke commands once the fixtures exist.

## Risks

- Git worktree state can become confusing if branch, worktree, and ticket
  metadata disagree.
- Result application may accidentally inspect the wrong checkout unless
  execution workspace metadata is explicit.
- Cleanup can destroy useful debugging evidence if it is too aggressive.
- Leaving worktrees forever can create disk and branch hygiene problems.
- Existing project-mode assumptions may be spread across tests and docs.
- Forge workflows may differ in branch reuse and PR update behavior.

## Backlog Candidates

- Dedicated `chc work worktree list/show/clean` commands.
- Configurable worktree path templates beyond ticket key.
- Per-execution clone mode for highly isolated or remote-only workflows.
- Container-side commit/push support with explicit credentials.
- Worktree disk usage reporting in Dash.
- Scheduled cleanup for stale merged worktrees.
