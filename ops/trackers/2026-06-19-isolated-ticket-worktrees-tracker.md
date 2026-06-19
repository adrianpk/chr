# Isolated Ticket Worktrees Tracker

Status: in_progress

Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`

Spec: `ops/specs/2026-06-19-isolated-ticket-worktrees.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | workspace execution config | delivered | `feat/slice-1-workspace-execution-config` | `#328` | `.prj/reports/slices/isolated-ticket-worktrees/slice-1-workspace-execution-config.md` |
| Slice 2 | worktree state service | delivered | `feat/slice-2-worktree-state-service` | `#329` | `.prj/reports/slices/isolated-ticket-worktrees/slice-2-worktree-state-service.md` |
| Slice 3 | readiness workspace split | delivered | `feat/slice-3-readiness-workspace-split` | `#330` | `.prj/reports/slices/isolated-ticket-worktrees/slice-3-readiness-workspace-split.md` |
| Slice 4 | runtime workspace mount | delivered | `feat/slice-4-runtime-workspace-mount` | `#331` | `.prj/reports/slices/isolated-ticket-worktrees/slice-4-runtime-workspace-mount.md` |
| Slice 5 | execution workspace results | delivered | `feat/slice-5-execution-workspace-results` | `#332` | `.prj/reports/slices/isolated-ticket-worktrees/slice-5-execution-workspace-results.md` |
| Slice 6 | branch and PR lifecycle | reviewing | `feat/slice-6-branch-and-pr-lifecycle` | `#333` | `.prj/reports/slices/isolated-ticket-worktrees/slice-6-branch-and-pr-lifecycle.md` |
| Slice 7 | review repair and cleanup | planned | `feat/slice-7-review-repair-and-cleanup` | pending | `.prj/reports/slices/isolated-ticket-worktrees/slice-7-review-repair-and-cleanup.md` |
| Slice 8 | docs and QA proving | planned | `feat/slice-8-docs-and-qa-proving` | pending | `.prj/reports/slices/isolated-ticket-worktrees/slice-8-docs-and-qa-proving.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add workspace execution config fields under the existing work/runtime configuration boundary. | done | `770ccd7` | Placed execution workspace config under `work:`. |
| T1.2 | Define controlled values for workspace mode, worktree base policy, worktree cleanup policy, and worktree root. | done | `0c887c8` | Added explicit enum values and workspace validation helpers. |
| T1.3 | Set defaults that preserve current project-mounted execution behavior in existing workspaces. | done | `f902837` | Default is `workspace_mode: project`; no behavior switch in Slice 1. |
| T1.4 | Add config parser and validation coverage for omitted defaults, explicit worktree settings, and invalid enum values. | done | `1c7bfdc` | Added table-driven parser and validation coverage. |
| T1.5 | Add runtime plan value fields for project checkout vs execution workspace without changing harness mount behavior. | done | `df7f790` | Runtime plan records execution workspace metadata while mounts remain project-based. |
| T1.6 | Document the config fields and migration posture in reference docs. | done | `3ce20bc` | Reference docs mark worktree behavior as configured/planned. |
| T1.7 | Add the slice implementation report and validation evidence. | done | `76449ea` | Report delivered through PR `#328`. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add a Git worktree planning model with ticket key, branch, base ref, base SHA, root, and target path. | done | `f90a9b4` | Core package boundary. |
| T2.2 | Add a command abstraction for `git fetch`, `git rev-parse`, `git branch`, `git worktree add`, `git worktree list`, and `git status`. | done | `dea4881` | Deterministic OS adapter and porcelain parser. |
| T2.3 | Implement deterministic worktree path selection using the configured root and ticket key. | done | `5519b44` | Ticket keys are sanitized into stable path components. |
| T2.4 | Implement safe creation from the fetched remote integration ref. | done | `ac410bf` | Preparation fetches the canonical remote base and creates missing ticket branches. |
| T2.5 | Implement inspection and diagnostics for missing, clean, dirty, mismatched, branch-in-use, and orphaned worktree states. | done | `878d40e` | Diagnostics use stable state and diagnostic codes. |
| T2.6 | Add optional metadata read/write for ticket worktree identity and last execution hints. | done | `896cf63` | Metadata is advisory; Git remains source of truth. |
| T2.7 | Add table-driven service tests with fake Git command output and temp repositories where useful. | done | `0674231` | Covers planning, path sanitization, porcelain parsing, preparation, inspection, and metadata round trip. |
| T2.8 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Extend work readiness input with workspace mode and optional ticket key. | done | `011d555` | Legacy callers remain compatible through defaults. |
| T3.2 | Split dirty-state diagnostics into user checkout and ticket worktree diagnostics. | done | `3f387c6` | Worktree mode has explicit `dirty_user_checkout` diagnostics. |
| T3.3 | Make user checkout dirtiness informational in worktree mode. | done | `b02ce7b` | Work selection no longer blocks on the user's dirty or stale checkout in worktree mode. |
| T3.4 | Make ticket worktree dirtiness blocking in worktree mode. | done | `df57f9b` | Dirty, mismatched, branch-in-use, and orphaned ticket worktrees block. |
| T3.5 | Add remote-base diagnostics for failed fetch and missing `<remote>/<integration_branch>`. | done | `c0286ba` | Worktree mode validates the fetched remote base instead of local branch containment. |
| T3.6 | Update `chc work doctor` output and service models to expose both workspace states. | done | `c28b9e4` | Doctor and Dash expose workspace mode, user checkout state, and ticket worktree state. |
| T3.7 | Add safe repair coverage for fetch, root creation, and non-destructive worktree metadata pruning. | done | `88c4cdb` | Safe repair can create roots and remove stale metadata without deleting worktrees. |
| T3.8 | Add CLI/service/TUI tests for legacy mode and worktree mode readiness. | done | `ec8cc69` | Added CLI, service, TUI, and core coverage for the new readiness surfaces. |
| T3.9 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Extend `RuntimePlan` with execution workspace host path and workspace mode metadata. | done | `8e3ee70` | Added ticket branch and workspace base metadata to runtime planning. |
| T4.2 | Update runtime mount planning to use `ExecutionWorkspaceDir` as the writable project workspace mount source. | done | `be3518b` | Worktree mode mounts the execution workspace at the container workdir. |
| T4.3 | Preserve existing project-mode mount output exactly where possible. | done | `21932e9` | Project mode ignores alternate execution workspace input and mounts the project checkout. |
| T4.4 | Update context artifact link rendering so container paths are derived from the execution workspace. | done | `20d57fd` | Runtime artifacts and container links are rooted at the mounted host workspace. |
| T4.5 | Add worker context fields for project checkout, execution workspace, base ref, base SHA, and ticket branch. | done | `47c6e9d` | Worker context includes a dedicated workspace section. |
| T4.6 | Update runtime validation output to show the execution workspace. | done | `b930cd7` | Runtime validation prints and checks the execution workspace. |
| T4.7 | Add harness invocation and runtime plan tests for project and worktree modes. | done | `4503487` | Added Docker validation command coverage for worktree mode. |
| T4.8 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Extend execution record/runtime report models with workspace mode, execution workspace path, base ref, base SHA, and ticket branch. | done | `c5dab13` | Execution records and reports include durable workspace metadata. |
| T5.2 | Add persistence migration or compatible report-only fallback for the new execution workspace metadata. | done | `6a3c84f` | Added additive SQLite migration and repository persistence coverage. |
| T5.3 | Update changed-file discovery to accept an explicit workspace root. | done | `f73945f` | Discovery can run against the execution workspace. |
| T5.4 | Ensure discovered paths are normalized relative to the project root semantics, not absolute worktree paths. | done | `dc82d01` | Discovery rejects absolute and parent-relative paths. |
| T5.5 | Update result application to load execution workspace metadata and inspect the right Git state. | done | `2dc5f31` | `ticket apply-result` discovers changed files from worktree execution records. |
| T5.6 | Add tests proving user checkout changes are ignored during result discovery in worktree mode. | done | `3d2fad2` | CLI test covers dirty checkout ignored while worktree changes are reported. |
| T5.7 | Add tests proving validation and evidence gates still use the same result contract. | done | `ad5b654` | Core test covers evidence blockers with workspace metadata. |
| T5.8 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Define branch selection rules for tickets with and without existing `branch` metadata. | done | `54ce330` | Existing ticket branch metadata is authoritative; missing branches derive from ticket key and title. |
| T6.2 | Integrate worktree preparation into ticket start/work execution before runtime context materialization. | done | `26a4184` | Ticket execution prepares the ticket worktree before runtime validation and context materialization. |
| T6.3 | Add branch creation from `origin/<integration_branch>` with conflict diagnostics. | done | `3e9c2d1` | Existing branches must contain the configured base ref; unrelated branches block instead of being reset. |
| T6.4 | Ensure worktree checkout and branch state are recorded in ticket, execution, and event evidence. | done | `26a4184` | Execution records, worktree metadata, ticket branch metadata, and agent events expose branch/worktree state. |
| T6.5 | Confirm host-side commit/push/PR creation reads from the ticket worktree in worktree mode. | done | `26a4184` | Host-side post-worker state uses execution workspace metadata; the worker container still does not own push or PR operations. |
| T6.6 | Add run-all tests proving sequential worktree preparation without pre-creating the full queue. | done | `26a4184` | Service test proves a two-ticket queue with `MaxIterations: 1` prepares only the selected ticket worktree. |
| T6.7 | Add conflict tests for branch already exists with unrelated history or unsafe remote state. | done | `3e9c2d1` | Core test covers an existing ticket branch that does not contain the selected base ref. |
| T6.8 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 7 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T7.1 | Add review-repair worktree lookup using ticket branch and PR metadata. | pending | pending | Repair should reuse ticket workspace. |
| T7.2 | Add safe recreation of missing repair worktrees from valid ticket branches. | pending | pending | Recreate only when safe. |
| T7.3 | Add dirty repair worktree diagnostics and stop conditions. | pending | pending | No destructive cleanup. |
| T7.4 | Add cleanup operation for clean merged ticket worktrees. | pending | pending | `after_merge` policy. |
| T7.5 | Add cleanup blockers for dirty, mismatched, or unknown worktrees. | pending | pending | Explicit manual state. |
| T7.6 | Add doctor reporting and safe repair hooks for orphaned clean worktrees. | pending | pending | Keep hygiene local-first. |
| T7.7 | Add tests for repair reuse, missing-worktree recreation, dirty blockers, idempotent cleanup, and no dirty deletion. | pending | pending | Safety-critical coverage. |
| T7.8 | Add the slice implementation report and validation evidence. | pending | pending | Report status should be `reviewing` before PR. |

## Slice 8 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T8.1 | Update runtime config, work readiness, ticket execution, and user guide docs. | pending | pending | Explain default and migration posture. |
| T8.2 | Update Dash/TUI reference docs for workspace visibility. | pending | pending | User should know where agent writes. |
| T8.3 | Add or update QA reference fixtures to exercise isolated worktree mode. | pending | pending | Resettable scenarios required. |
| T8.4 | Add smoke scripts or make targets for the worktree QA scenarios. | pending | pending | Prefer reusable commands. |
| T8.5 | Run end-to-end QA proving with a dirty user checkout and a clean ticket worktree. | pending | pending | Proves core user benefit. |
| T8.6 | Record the default-switch decision in tracker notes with validation evidence. | pending | pending | Do not flip default without evidence. |
| T8.7 | Add the slice implementation report and final cycle report. | pending | pending | Final report summarizes all slices. |

## Validation Log

- Planning: `git diff --check` (pass)
- Slice 1: `go test ./internal/core -run 'TestLoadWorkspaceConfig.*Work|TestValidateWorkConfigRejectsInvalid.*Config|TestPlanRuntime|TestCodexCredentialMountsUseReadOnlyRootCodexPaths|TestValidateRuntime'` (pass)
- Slice 1: `make verify` (pass)
- Slice 2: `go test ./internal/core -run 'Worktree'` (pass)
- Slice 2: `make lint-strict` (pass)
- Slice 2: `make verify` (pass)
- Slice 3: `go test ./internal/core -run 'WorkReadiness|WorkSelectionPreflight|Worktree' && go test ./internal/app -run 'WorkDoctor|PreviewAgentWork|StartAgentWork' && go test ./internal/tui -run 'WorkReadiness|Dash|Model'` (pass)
- Slice 3: `make lint-strict` (pass)
- Slice 3: `make verify` (pass)
- Slice 4: `go test ./internal/core -run 'TestPlanRuntime|TestNewHarnessInvocation|TestMaterializeCodexWorkerContextForExecution|TestRenderMainContextPackage|TestValidateRuntime' && go test ./internal/app -run 'TestRuntimeValidatePrintsRuntimePlan'` (pass)
- Slice 4: `make verify` (pass)
- Slice 5: `go test ./internal/core -run 'ExecutionRecord|ExecutionRepository|ExecutionReport|Migrator|Migration'` (pass)
- Slice 5: `go test ./internal/core -run 'ChangedFileDiscoverer|ParseGitStatusChangedFiles'` (pass)
- Slice 5: `go test ./internal/core -run 'TestNewResultApplication|TestChangedFileDiscoverer|TestParseGitStatusChangedFiles' && go test ./internal/app -run 'TestTicketApplyResultDiscoversWorktreeChangedFiles|TestTicketApplyResultAppliesExecutionResult'` (pass)
- Slice 5: `make lint-strict` (pass)
- Slice 5: `make verify` (pass)
- Slice 6: `go test ./internal/core -run 'TestSelectTicketBranch|TestWorktreeService|TestChangedFileDiscoverer|TestNewResultApplication'` (pass)
- Slice 6: `go test ./internal/app -run 'TestServiceStartAgentWorkTicketPreparesWorktreeWorkspace|TestServiceStartAgentWorkLoopPreparesOnlySelectedWorktree|TestServiceStartAgentWorkTicketReusesTicketExecution|TestServiceStartAllReadyAgentWorkEmitsLifecycleEvents|TestTicketApplyResultDiscoversWorktreeChangedFiles|TestTicketApplyResultAppliesExecutionResult'` (pass)
- Slice 6: `make lint-strict` (pass)
- Slice 6: `make verify` (pass)

## Open Decisions

- Final config placement: resolved as `work.workspace_mode` with related
  `work.worktree_*` fields.
- Whether execution workspace metadata requires a SQLite migration in Slice 5
  or can remain report-only until the behavior is proven.
- Whether worktree metadata should be central-only, file-based under the
  worktree root, or both.
- Whether `worktree` becomes the default for new projects during Slice 8 or
  remains opt-in until another proving cycle.

## Notes

- This slicing intentionally starts with config and planning models before
  introducing Git worktree mutation.
- Worktree execution must not copy uncommitted user checkout changes into agent
  workspaces.
- The user checkout can be stale or dirty in worktree mode; the fetched remote
  integration ref is the default operational base.
- The first implementation should prefer clear diagnostics over automatic
  repair whenever Git state is ambiguous.
- Slice 8 must update the changelog for the isolated ticket worktrees cycle.
