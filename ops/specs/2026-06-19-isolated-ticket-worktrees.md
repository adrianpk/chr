# Isolated Ticket Worktrees

Status: draft

Date: 2026-06-19

Related:

- `docs/reference/ticket-execution.md`
- `docs/reference/runtime-config.md`
- `docs/reference/work-readiness.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/ticket-result-application.md`
- `ops/specs/2026-06-18-configurable-work-readiness.md`
- `ops/specs/2026-06-18-agentic-loop-review-merge.md`
- `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

## Summary

ChrysoCoord should move ticket execution from the user's active checkout into an
isolated Git worktree created per ticket execution or per ticket branch.

The system remains sequential: one active agent, one active ticket, one active
runtime worker. The change is not about parallelism. It is about filesystem
ownership and operational safety.

Today, the Docker runtime mounts the managed project directory as the writable
workspace inside the container. That means the containerized Codex worker edits
the same checkout the user sees and may want to use. The current readiness model
therefore protects the project by requiring a clean local worktree before agent
work can start.

That is safe, but it makes the user's checkout part of the agent execution
surface. The desired pattern separates those concerns:

```text
user checkout
  owned by the human; may be dirty or stale

origin/<integration_branch>
  published operational base for agent work

ticket worktree
  created by ChrysoCoord from the published operational base
  mounted into Docker as the writable workspace
  kept until ticket branch / PR lifecycle is complete
```

The default base for ticket worktrees should be the configured canonical remote
and integration branch, normally `origin/dev`. The worker must not silently use
the user's local checkout state as the starting point for managed work.

## Goals

- Keep the current sequential orchestration model.
- Keep Docker as the runtime boundary for Codex execution.
- Stop requiring the user's active checkout to be clean for ordinary ticket
  execution.
- Start each ticket from an explicit, fetched, published ref such as
  `origin/dev`.
- Mount the ticket worktree, not the user's checkout, into the runtime
  container.
- Let the user keep editing, browsing, or experimenting in the main checkout
  while ChrysoCoord runs an agent task.
- Preserve deterministic runtime artifacts, event streams, execution reports,
  validation logs, and result application behavior.
- Tie each ticket worktree to the ticket branch and PR lifecycle.
- Reuse a ticket worktree for review repair when the ticket branch remains open.
- Clean up ticket worktrees after merge when safe.
- Give `doctor` commands enough visibility to detect, repair, or explain stale,
  dirty, missing, or orphaned worktrees.
- Keep the design local-first and forge-agnostic.

## Non-Goals

- Do not introduce parallel ticket execution.
- Do not allow multiple active agents by default.
- Do not require the user to adopt ChrysoCoord's own branch names.
- Do not copy uncommitted user changes into ticket worktrees automatically.
- Do not treat the user's local integration branch as the source of truth when
  the canonical remote tracking ref is available.
- Do not make worktrees a reason to skip validation, evidence strategy, review,
  PR creation, or merge rules.
- Do not require the worker container to push branches directly in the first
  implementation.
- Do not require PAT-backed clone or remote operations from inside the
  container.
- Do not delete dirty worktrees automatically.
- Do not change the user-owned `dev -> main` release alignment protocol.

## Current Runtime Shape

The current runtime plan resolves:

- `ProjectDir`: the managed project directory.
- `DockerfilePath`: usually `.chc/user/runtime/agent/Dockerfile`.
- `ContainerWorkdir`: usually `/workspace`.
- `ArtifactDir`: usually `.chc/user/runtime/runs`.
- `WorkerContextPath`, `ResultPath`, `LogPath`, and `ValidationPath`.

The Docker harness mounts the project directory as a writable volume:

```text
host projectDir -> container /workspace
```

The Codex worker then runs inside the Docker container:

```sh
codex exec --skip-git-repo-check --dangerously-bypass-approvals-and-sandbox < "$CHC_WORKER_CONTEXT_PATH"
```

This confirms the important distinction:

- Codex is containerized.
- The writable filesystem is still the managed project checkout.

The current context readiness notes already mark ephemeral worktree or clone
execution as a future runtime-workspace concern. This spec makes that future
concern explicit.

## Desired Runtime Shape

For ticket execution, the runtime plan should resolve two separate host
directories:

```text
ProjectDir
  the user-facing managed project checkout

ExecutionWorkspaceDir
  the ChrysoCoord-owned ticket worktree mounted into the runtime container
```

The Docker harness should mount:

```text
host ExecutionWorkspaceDir -> container /workspace
```

Runtime artifacts may remain under the user's `.chc/user/runtime/runs`
directory or move to a workspace-owned runtime directory, but the first version
should preserve existing artifact paths when practical to avoid breaking event
views and reports.

The worker context should describe both boundaries:

```text
project_dir: <user checkout>
execution_workspace_dir: <ticket worktree>
container_workdir: /workspace
base_ref: origin/dev
ticket_branch: <configured/generated ticket branch>
```

The worker should be told clearly that `/workspace` is the ticket execution
workspace, not necessarily the user's active checkout.

## Configuration

Introduce a workspace execution mode under `work:` or `runtime:`. The naming
should make ownership clear.

Proposed shape:

```yml
work:
  workspace_mode: worktree
  worktree_root: .chc/user/runtime/worktrees
  worktree_base: remote_integration_branch
  worktree_cleanup: after_merge
```

Possible values:

`workspace_mode`

: `project` or `worktree`.

`project`

: Legacy mode. The runtime mounts the managed project checkout. Existing clean
  worktree preflight rules apply.

`worktree`

: Ticket execution happens in a ChrysoCoord-owned Git worktree created from an
  explicit base ref.

`worktree_root`

: Project-relative directory where ChrysoCoord stores ticket worktrees.
  Default: `.chc/user/runtime/worktrees`.

`worktree_base`

: Policy for selecting the worktree base ref. Default:
  `remote_integration_branch`.

`remote_integration_branch`

: Use `<canonical_remote>/<integration_branch>`, normally `origin/dev`.

`local_integration_branch`

: Use the local integration branch. This should not be the default because it
  can include stale or unpublished user state.

`current_branch`

: Use the current branch. This should be explicit and mostly useful for
  advanced/manual workflows.

`worktree_cleanup`

: `after_merge`, `manual`, or `never`.

`after_merge`

: Remove clean ticket worktrees after the associated PR or ticket branch is
  merged.

`manual`

: Leave worktrees until the user runs an explicit cleanup command.

`never`

: Keep worktrees indefinitely. Useful for debugging but not recommended as a
  default.

## Default Policy

The recommended default should become:

```yml
work:
  mode: integration_branch
  integration_branch: dev
  canonical_remote: origin
  require_remote_tracking: true
  workspace_mode: worktree
  worktree_base: remote_integration_branch
  worktree_cleanup: after_merge
```

The meaning of `require_clean_worktree` changes depending on
`workspace_mode`:

- in `project` mode, it applies to the user's managed project checkout;
- in `worktree` mode, it applies to the ticket worktree only.

The user's checkout may be dirty in `worktree` mode. That dirtiness should be
reported as informational, not blocking, unless the command explicitly needs to
use the user's checkout.

## Worktree Identity

Each ticket worktree should have a deterministic identity.

Recommended path shape:

```text
.chc/user/runtime/worktrees/<ticket-key>/
```

For repeated attempts against the same ticket branch, reuse the ticket worktree
when safe.

For multiple repair attempts, do not create a new worktree per execution by
default. The ticket branch is the meaningful development line; execution
attempts are attempts within that line.

Execution-specific artifacts still live under:

```text
.chc/user/runtime/runs/<ticket-key>/<execution-id>/
```

This separates:

- ticket workspace state;
- execution attempt evidence.

## Branch Identity

Every ticket should have exactly one implementation branch during active work.

The existing ticket metadata already includes a `branch` field. Worktree mode
should make that field operational:

```text
ticket.branch -> git branch checked out by ticket worktree
```

If the ticket branch is missing, ChrysoCoord creates it from the selected base
ref.

If the ticket branch exists locally and remotely, ChrysoCoord must verify that
the relationship is safe before reusing it.

If the branch exists with unrelated history, ChrysoCoord must block and report
the conflict.

## Base Ref Selection

For the default configuration:

```text
base_ref = refs/remotes/<canonical_remote>/<integration_branch>
```

Usually:

```text
refs/remotes/origin/dev
```

Before creating or updating a ticket worktree, ChrysoCoord should fetch the
canonical remote.

The operation must not use local `dev` merely because it exists. Local `dev`
may be:

- stale;
- dirty;
- ahead of origin with user commits;
- behind origin;
- temporarily checked out for unrelated human work.

Those states should not contaminate agent work.

## Initial Ticket Execution Flow

Default happy path:

```text
select ticket
-> fetch canonical remote
-> resolve base_ref = origin/<integration_branch>
-> create or reuse ticket branch
-> create or reuse ticket worktree
-> ensure ticket worktree is clean
-> materialize worker context and artifacts
-> run Docker with ticket worktree mounted as /workspace
-> run Codex worker
-> run validation in same worktree
-> capture result
-> discover changed files in ticket worktree
-> commit or leave changes according to existing branch/PR flow
-> push branch / open PR according to existing ticket workflow
-> update ticket evidence and state
```

The exact commit/PR steps may remain controlled by existing ticket execution
and review/merge features. Worktree mode only changes where the filesystem work
happens.

## Review Repair Flow

When a ticket is in review repair:

```text
ticket has branch + PR metadata
-> fetch canonical remote
-> verify PR branch still exists
-> prepare same ticket worktree
-> checkout ticket branch
-> fast-forward or block if unsafe
-> run Codex with review repair handoff context
-> validate
-> update same branch
-> update PR and ticket metadata
```

Review repair should not create a fresh branch unless the existing ticket branch
is missing and policy explicitly allows reconstruction.

The worktree should survive across review rounds until the PR merges or the
ticket is closed.

## Merge Cleanup Flow

After a ticket PR is merged:

```text
detect merge success
-> verify ticket worktree exists
-> verify ticket worktree is clean
-> remove worktree
-> prune stale Git worktree metadata
-> mark ticket worktree cleanup complete
```

If the worktree is dirty:

```text
do not delete
-> mark cleanup blocked
-> report path and dirty files
-> require manual cleanup or explicit safe repair
```

If the worktree is missing:

```text
record cleanup as already absent
```

Cleanup must be idempotent.

## Dirty State Rules

There are two different dirty states.

`user checkout dirty`

: The user's active project checkout has uncommitted changes.

`ticket worktree dirty`

: The ChrysoCoord-owned ticket worktree has uncommitted changes.

In `worktree` mode:

- user checkout dirty is informational;
- ticket worktree dirty is blocking unless it belongs to the current unfinished
  execution and can be resumed or applied safely.

In `project` mode:

- user checkout dirty remains blocking when `require_clean_worktree: true`.

Diagnostic messages should use explicit names:

```text
user checkout has local changes; ignored for isolated worktree execution
ticket worktree has uncommitted changes; manual repair required
```

Avoid generic `dirty_worktree` wording when more precise language is available.

## Stale State Rules

The user checkout may be stale without blocking isolated execution.

The remote tracking ref may not be stale. Before execution, ChrysoCoord should
fetch the canonical remote and verify that the selected base ref exists.

If fetch fails:

```text
block execution
reason: cannot refresh canonical remote
```

If `origin/dev` is missing:

```text
block execution
reason: missing remote integration branch
```

If local `dev` is behind `origin/dev`, this is informational in worktree mode.

If local `dev` is ahead of `origin/dev`, this is informational unless the user
has configured `worktree_base: local_integration_branch`.

## Runtime Artifact Placement

The first implementation should keep runtime artifacts in the managed project
artifact directory:

```text
.chc/user/runtime/runs/<ticket-key>/<execution-id>/
```

Reasons:

- existing Dash and report views already know this shape;
- artifacts remain visible from the user's checkout;
- cleanup of worktrees does not destroy execution evidence.

The ticket worktree should not be the only place where execution evidence lives.

Worktree-specific metadata may be stored under:

```text
.chc/user/runtime/worktrees/<ticket-key>/.chc-worktree.json
```

or in the central runtime record. Prefer central records if a durable execution
repository already owns the required fields.

## Execution Record Additions

Execution records should eventually capture:

- `workspace_mode`;
- `workspace_host_path`;
- `workspace_git_branch`;
- `workspace_base_ref`;
- `workspace_base_sha`;
- `workspace_cleanup_policy`.

These values make reports and result application explain exactly where the
worker ran.

The first implementation may avoid a schema migration if the information can be
derived from runtime paths and ticket metadata, but durable reporting should not
depend on inference forever.

## Worktree Metadata

A ticket worktree should be inspectable by ChrysoCoord without relying only on
Git's internal worktree list.

Minimum metadata:

```json
{
  "version": 1,
  "ticket_key": "CHC-123",
  "branch": "feature/use-case/s01-slice",
  "base_ref": "origin/dev",
  "base_sha": "<sha>",
  "created_at": "<rfc3339>",
  "updated_at": "<rfc3339>",
  "last_execution_id": "<execution-id>",
  "cleanup_policy": "after_merge"
}
```

Metadata must be advisory. Git remains the source of truth for branch and
worktree state. If metadata and Git disagree, `doctor` should report the
mismatch and prefer safe manual repair.

## Doctor Behavior

`chc work doctor` should report both user checkout and ticket worktree state.

In worktree mode, a healthy report should make the separation obvious:

```text
workspace_mode: worktree
base_ref: origin/dev
user_checkout: dirty informational
ticket_worktree: ready
```

Doctor should detect:

- missing worktree root;
- missing ticket worktree;
- dirty ticket worktree;
- missing ticket branch;
- branch checked out in another worktree;
- missing remote integration branch;
- failed fetch;
- ticket branch diverged from remote;
- orphaned worktrees for closed or merged tickets;
- stale metadata that no longer matches Git.

Safe repair may:

- create missing worktree root;
- fetch canonical remote;
- prune Git worktree metadata;
- remove clean worktrees for merged tickets;
- recreate a missing clean worktree from a valid ticket branch;
- fast-forward a clean ticket worktree when safe.

Safe repair must not:

- discard uncommitted changes;
- delete dirty worktrees;
- force-reset ticket branches;
- rewrite user branches;
- copy user checkout changes into ticket worktrees.

## TUI Behavior

Dash should show the execution workspace explicitly.

Useful fields:

```text
workspace: worktree
base: origin/dev @ <sha>
ticket branch: <branch>
worktree: .chc/user/runtime/worktrees/CHC-123
user checkout: ignored / dirty / clean
```

When work starts, event stream entries should include:

- worktree path;
- base ref;
- base SHA;
- ticket branch;
- cleanup policy.

This lets the user understand that their current checkout is not being edited
by the agent.

## CLI Behavior

`chc work doctor`

: Shows project checkout state and worktree workspace state.

`chc work ticket <key>`

: Uses worktree execution when configured.

`chc work run-all`

: Still processes tickets sequentially. Each ticket receives or reuses its own
  worktree.

Possible future commands:

```sh
chc work worktree list
chc work worktree show <ticket-key>
chc work worktree clean <ticket-key>
chc work worktree clean --merged
```

These commands are optional. The first implementation can expose worktree state
through `doctor` and reports only.

## Result Application

Result application should inspect changed files in the ticket worktree, not the
user checkout.

The execution record or runtime plan must tell changed-file discovery which
workspace path to inspect.

Existing result application semantics remain:

- successful worker result may move a ticket toward `ready_for_review`;
- evidence strategy gates still apply;
- failed worker result blocks or records failure;
- validation evidence remains required.

The file paths recorded in result JSON and reports should remain
project-relative, even though discovery happens in the ticket worktree.

## PR and Merge Lifecycle

Worktree mode should align with the existing one-ticket, one-branch, one-PR
model.

Expected lifecycle:

```text
ticket ready_for_dev
-> ticket branch created from origin/dev
-> ticket worktree created
-> worker modifies ticket worktree
-> changes committed on ticket branch
-> branch pushed
-> PR opened
-> ticket ready_for_review
-> review/repair if needed, using same worktree
-> PR merged
-> ticket merged
-> worktree cleaned if clean and cleanup policy allows
```

The PR remains the durable evidence that a human can inspect in the forge.

Automerge mode may still merge after successful validation and review policy.
Formal review mode may still run reviewer and merger hats. Worktree mode does
not decide that policy.

## Interaction With User Local Work

The user's active checkout should be treated as a control surface and
inspection surface, not the agent workspace.

The user may:

- keep local uncommitted edits;
- inspect docs;
- run commands;
- switch branches;
- review generated artifacts;
- reset example projects;
- open files in an editor.

CHC should avoid surprising the user by:

- not writing implementation changes into the active checkout;
- reporting where work is actually happening;
- keeping runtime evidence in the main `.chc/user/runtime/runs` tree;
- making worktree cleanup explicit and recoverable.

## Failure Modes

### Fetch Fails

Block before creating or modifying a ticket worktree.

Diagnostic:

```text
cannot refresh canonical remote origin
```

### Remote Integration Branch Missing

Block before branch creation.

Diagnostic:

```text
missing remote integration branch origin/dev
```

### Ticket Branch Already Exists With Unexpected Base

Block unless policy allows explicit repair.

Diagnostic:

```text
ticket branch exists but is not based on configured integration ref
```

### Ticket Worktree Dirty Before Execution

Block unless it is the current resumable attempt.

Diagnostic:

```text
ticket worktree has uncommitted changes
```

### User Checkout Dirty

Do not block in worktree mode.

Diagnostic:

```text
user checkout has local changes; ignored for isolated execution
```

### Worktree Missing During Repair

Recreate from ticket branch if the branch exists and is safe.

Diagnostic:

```text
ticket worktree missing; safe to recreate from ticket branch
```

### Cleanup Blocked By Dirty Worktree

Do not remove.

Diagnostic:

```text
ticket worktree cleanup blocked by uncommitted changes
```

## Security and Safety

Worktree mode does not make Codex more sandboxed by itself. Docker remains the
runtime boundary.

The safety improvement is filesystem ownership:

- Codex writes to a ChrysoCoord-owned worktree;
- user checkout changes are not modified;
- destructive repair can be scoped to the ticket worktree and still remain
  non-destructive by default.

Credential mounts remain read-only. The runtime should continue mounting Codex
credentials as read-only host files when present.

## Migration Strategy

Recommended rollout:

1. Add runtime/work config fields and defaults while keeping `project` mode as
   the default.
2. Add worktree planning and reporting without changing execution.
3. Add `doctor` visibility for prospective worktree execution.
4. Enable worktree mode for QA reference projects.
5. Switch ticket execution to mount the ticket worktree when
   `workspace_mode: worktree`.
6. Add cleanup after merge.
7. Consider making `worktree` the default for new projects.

The default should not flip until:

- result application reads from the execution workspace correctly;
- review repair reuses ticket worktrees safely;
- cleanup is idempotent;
- doctor can explain common failures clearly;
- QA reference projects prove the flow end to end.

## Test Strategy

Unit and integration tests should cover:

- runtime plan resolves both project dir and execution workspace dir;
- Docker mount source switches from project dir to worktree dir in worktree
  mode;
- container paths remain stable under `/workspace`;
- changed-file discovery inspects the execution workspace;
- user checkout dirty state is informational in worktree mode;
- ticket worktree dirty state blocks execution;
- base ref uses `origin/<integration_branch>` by default;
- local stale `dev` does not affect worktree base selection;
- ticket branch creation from remote base;
- safe reuse of existing ticket worktree;
- safe recreation of missing worktree from ticket branch;
- cleanup removes clean merged worktrees;
- cleanup refuses dirty worktrees;
- doctor reports user checkout and ticket worktree state separately;
- Dash/event streams include workspace mode, worktree path, base ref, and branch.

QA scenarios should include:

- a dirty user checkout while a ticket executes successfully in a worktree;
- a stale local `dev` while execution starts from fresh `origin/dev`;
- a dirty ticket worktree blocking execution;
- a review repair run reusing the same ticket worktree;
- a merged ticket cleaning up its worktree;
- a missing worktree recreated safely from an existing ticket branch.

## Acceptance Criteria

- A project can configure ticket execution to use isolated worktrees.
- In worktree mode, the Docker runtime mounts the ticket worktree rather than
  the user checkout.
- In worktree mode, uncommitted changes in the user checkout do not block ticket
  execution.
- In worktree mode, uncommitted changes in the ticket worktree block ticket
  execution unless explicitly resumable.
- Ticket worktrees are created from a fetched canonical remote integration ref
  by default.
- Local stale or dirty integration branches are not used as ticket bases unless
  explicitly configured.
- Runtime reports identify the execution workspace path and base ref.
- Result application discovers changed files from the execution workspace and
  records project-relative paths.
- Review repair reuses the ticket worktree for the same ticket branch.
- Clean worktrees are removed after successful merge when cleanup policy is
  `after_merge`.
- Dirty worktrees are never deleted automatically.
- `chc work doctor` reports user checkout state separately from ticket
  worktree state.
- The existing sequential execution model remains unchanged.

## Open Questions

- Should `workspace_mode: worktree` live under `work:` or `runtime:`?
- Should runtime artifacts remain in the main project `.chc/user/runtime/runs`
  directory forever, or should each worktree also receive local attempt
  artifacts?
- Should a ticket worktree be keyed by ticket key, branch name, or execution ID?
  This spec recommends ticket key.
- Should the first implementation create commits and PRs from the host process
  only, or allow the worker to commit from inside the container later?
- Should cleanup run immediately after merge or as an explicit periodic doctor
  repair?
- How much worktree metadata belongs in SQLite execution records versus
  file-based metadata under `.chc/user/runtime/worktrees`?
- Should old `project` mode remain permanently supported, or become an escape
  hatch after worktree mode is proven?
- Should `run-all` pre-create worktrees for the full queue or create one
  worktree at a time? This spec recommends one at a time to preserve sequential
  semantics and reduce cleanup surface.

