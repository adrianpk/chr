# Execution Unit Runtime Refinements

Status: draft

Date: 2026-06-29

Related:

- `docs/reference/ticket-execution.md`
- `docs/reference/project-workspace.md`
- `docs/reference/work-readiness.md`
- `docs/reference/execution-reports.md`
- `ops/specs/2026-06-19-isolated-ticket-worktrees.md`
- `ops/specs/2026-06-18-configurable-work-readiness.md`

## Summary

This spec collects small runtime refinements for the unit that the agent
resolves inside the Docker boundary.

The purpose is not to change the surrounding workflow. The purpose is to make
each execution unit safer, clearer, and easier to inspect:

- prefer isolated Git worktrees for new managed execution;
- keep the worker context compact and directly actionable;
- expose a clearer one-command execution path for one unit;
- separate worker-facing resolution context from orchestration and audit
  context.

The existing result envelope and validation behavior are intentionally kept.

## Current Shape

The runtime plan resolves:

- the managed project checkout;
- the effective execution workspace;
- the runtime Dockerfile;
- the runtime image;
- the container workdir;
- the worker context path;
- the result path;
- the runtime log path;
- the validation log path;
- the configured validation command.

The Docker harness runs the worker inside the configured runtime container. The
worker receives paths through environment variables and reads the worker context
from the context path.

The workspace can currently be:

- `project`: the user-facing managed checkout;
- `worktree`: an isolated Git worktree prepared for the execution unit.

The existing model is sound. The refinements below should reduce accidental
workspace coupling and reduce prompt noise without weakening evidence.

## Goals

- Make isolated worktree execution the recommended default for new managed
  projects.
- Preserve compatibility for existing projects that explicitly use `project`
  mode.
- Keep Docker as the runtime boundary.
- Keep the result envelope contract unchanged.
- Keep validation and evidence behavior unchanged except for clearer paths and
  summaries.
- Make the worker context easier for an implementation agent to follow.
- Keep orchestration metadata available for reports and diagnostics without
  pushing all of it into the worker's main instruction surface.
- Add a direct execution entrypoint only if it composes existing operations.

## Non-Goals

- Do not remove `project` workspace mode.
- Do not rewrite the result envelope.
- Do not add parallel execution.
- Do not change merge or release behavior.
- Do not move runtime artifacts without a separate migration spec.
- Do not hide evidence required for debugging or review.
- Do not reduce validation requirements.
- Do not require a remote provider for local execution.

## Proposal 1: Worktree-First Defaults

New managed projects should default to isolated execution worktrees.

The effective default for new generated project config should be:

```yml
work:
  workspace_mode: worktree
  worktree_root: .chr/user/runtime/worktrees
  worktree_base: remote_integration_branch
  worktree_cleanup: after_merge
```

Existing projects without this setting should not silently change behavior in a
way that surprises a user. The implementation should distinguish:

- config generation defaults for new projects;
- runtime fallback defaults for old projects;
- explicit user config.

Recommended rule:

- new `chr init` output writes `workspace_mode: worktree`;
- existing config without `workspace_mode` continues using the current fallback
  until an explicit migration or doctor repair is accepted;
- `doctor` and runtime validation recommend worktree mode when project mode is
  detected for agent execution.

### Why This Matters

In `project` mode, the container edits the same checkout the user sees.

In `worktree` mode, the container edits an isolated checkout tied to the
execution branch. That makes the resolved unit easier to inspect and keeps
human local edits outside the worker's writable workspace.

### Acceptance Criteria

- New generated config prefers `workspace_mode: worktree`.
- Existing explicit `workspace_mode: project` continues to work.
- Existing config with no workspace mode is handled through a documented
  compatibility rule.
- Runtime validation prints the effective workspace mode.
- Runtime validation prints the execution workspace path.
- `doctor` explains how to migrate to worktree mode without mutating config by
  default.

## Proposal 2: Compact Worker Context

The worker context should be split into two conceptual layers:

```text
resolution context
  short, direct instructions needed by the worker

supporting context
  detailed metadata, evidence policy, previous attempts, references, and
  diagnostics
```

The worker can still receive both layers in one `context.md` file at first. The
important change is structure and priority, not storage.

The top of `context.md` should answer:

- what unit is being resolved;
- what the worker should change;
- where the writable workspace is;
- what files or docs should be read first;
- what output must be written;
- what validation matters;
- when to stop as blocked instead of guessing.

Detailed operational metadata should move below that main section or into
separate linked sections.

### Proposed Context Shape

```md
# Worker Context

## Resolution Brief

- Execution id:
- Unit key:
- Writable workspace:
- Result path:
- Runtime log path:
- Validation path:

## Objective

One concise description of the current unit.

## Scope

- Allowed changes:
- Out of scope:
- Files or areas likely relevant:

## Required Output

Write the result envelope to the configured result path.

## Validation

- Command:
- Required:
- What to report if validation cannot run:

## Stop Conditions

Stop as blocked when:

- required product behavior is ambiguous;
- required credentials or services are missing;
- the requested change conflicts with repository instructions;
- the implementation would require broad changes outside the current unit.

## Supporting Context

Detailed metadata, references, evidence policy, previous attempts, and
diagnostics.
```

### Safe Analysis Path

Because the current context works, this should not be changed blindly.

First implementation should add a render-only comparison mode:

```sh
chr work context-preview <unit-key>
```

or an equivalent dry-run output from the existing execution command.

The preview should write or print:

- current context size;
- proposed context size;
- section list;
- missing required fields;
- links to generated preview files.

The worker runtime should not switch to the compact layout until tests and at
least one manual comparison show no loss of necessary information.

### Acceptance Criteria

- The compact layout is generated without running the worker.
- The preview includes enough information to compare old and new context.
- The existing context renderer remains available during the transition.
- No execution behavior changes until the compact context is explicitly enabled.
- Tests cover required sections and missing required fields.

## Proposal 3: Direct Single-Unit Execution Path

This proposal does not change runtime semantics.

It only clarifies the operator path for one unit:

```text
resolve unit -> preflight -> prepare workspace -> run worker -> run validation -> show evidence paths
```

If current commands already do this, the improvement is documentation and output
clarity rather than new behavior.

The command or TUI action should surface one concise result summary:

```text
status: succeeded
execution: <execution-id>
workspace: <execution-workspace-dir>
result: <result-path>
runtime log: <log-path>
validation: <validation-path>
report: <report-path>
```

### Where This Impacts The System

This impacts only the operation boundary around the existing runtime:

- CLI command composition;
- TUI action labeling;
- stdout summary;
- docs and examples.

It should not change:

- runtime planning;
- Docker invocation;
- result envelope schema;
- validation execution;
- persistence schema.

### Acceptance Criteria

- One documented command or TUI action runs one eligible unit end to end.
- The command composes existing service operations.
- The command prints the effective workspace mode before running.
- The command prints evidence paths after running or failing.
- Failure output points to the diagnostic log, result path, and validation path
  when available.

## Proposal 4: Separate Resolution Context From Audit Context

The worker and the system do not need the same view of the execution.

The worker needs direct resolution instructions. The system needs durable audit
metadata. Mixing both too heavily in the worker prompt can make the actual
implementation objective harder to see.

This proposal keeps all existing audit metadata, but classifies it:

| Context class | Primary consumer | Examples |
| --- | --- | --- |
| Resolution | worker | objective, scope, writable workspace, required output, validation |
| Repository | worker | project docs, agent instructions, relevant files |
| Evidence | worker and system | result path, validation path, evidence policy |
| Audit | system | execution id, timestamps, previous records, report references |
| Diagnostics | system, sometimes worker | failed prior attempt details, repair handoff |

Worker context should lead with Resolution and Repository context.

Audit and Diagnostics context should remain available, but should not bury the
objective.

### Acceptance Criteria

- Context renderer has named sections for the context classes above.
- Tests assert section order.
- Worker-facing context starts with the resolution objective.
- Audit fields remain available in reports and execution records.
- No audit field is removed from durable storage.

## Risk Management

The main risk is changing prompt shape for a runtime that already works.

Mitigations:

- make worktree default changes only for new generated config first;
- keep existing config behavior stable;
- add compact context preview before enabling compact context for execution;
- compare old and new context outputs on a real execution unit;
- keep result envelope and validation unchanged;
- ship changes in small slices.

## Suggested Implementation Slices

| Slice | Name | Purpose |
| --- | --- | --- |
| Slice 1 | Worktree default posture | Update new config generation and diagnostics so new projects prefer worktree mode while existing projects remain stable. |
| Slice 2 | Context preview | Add render-only compact worker context preview and tests without changing execution. |
| Slice 3 | Direct execution summary | Improve one-unit command or TUI output so it clearly shows workspace mode and evidence paths. |
| Slice 4 | Context classification | Introduce named context classes and section-order tests, then gate compact context for execution behind explicit config. |

## Open Questions

- Should old projects with no `workspace_mode` continue using `project`, or
  should runtime validation prompt for explicit selection?
- Should compact context be enabled by config, environment variable, or command
  flag during its trial period?
- Should the direct single-unit execution path be a new command, a clearer alias,
  or only improved output from the existing command?
- Which real execution scenario should be used as the comparison case for old
  and compact context?
