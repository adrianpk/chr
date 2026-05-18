# User Project Context Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-user-project-context-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Specs:

- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Review existing project registry, CLI, and user environment implementation.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [ ] Commit plan and tracker directly to `dev`.
- [ ] Select the first pending slice for implementation.

## Slice Status

| Slice | Status | Notes |
| --- | --- | --- |
| User Environment Paths | Delivered | Implemented by the existing user environment path work. |
| Project Registry Persistence | Delivered | SQLite registry and repository exist. |
| Project Registry CLI | Delivered | `project add` and `project list` exist. |
| Active Project Persistence | Pending | Store and load selected project state. |
| Project Select and Current CLI | Pending | Add `project select` and `project current`. |
| Project Context Resolver | Pending | Reusable project resolution for later commands. |
| Project Validation and Recovery | Pending | Clear stale/missing/disabled project behavior. |

## Pending Slice Candidate: Active Project Persistence

Candidate tasks:

- [ ] Decide persistence location from the existing user environment model.
- [ ] Add a core model for active project state.
- [ ] Add a repository or config service for storing active project state.
- [ ] Add tests for empty, set, replace, and missing-project cases.
- [ ] Keep development-mode state under `_workspace`.
- [ ] Run `make verify`.

## Pending Slice Candidate: Project Select and Current CLI

Candidate tasks:

- [ ] Add `chc project select <name-or-path>`.
- [ ] Add `chc project current`.
- [ ] Resolve registered project by name.
- [ ] Resolve registered project by path when unambiguous.
- [ ] Print stable current-project output.
- [ ] Add focused CLI tests.
- [ ] Run `make verify`.

## Pending Slice Candidate: Project Context Resolver

Candidate tasks:

- [ ] Add reusable core command context resolution.
- [ ] Support explicit project input where command semantics require it.
- [ ] Fall back to active project where appropriate.
- [ ] Preserve current-directory behavior for project-local commands that should keep it.
- [ ] Return clear errors when no project can be resolved.
- [ ] Add tests for explicit, active, current-directory, and missing-context cases.
- [ ] Run `make verify`.

## Pending Slice Candidate: Project Validation and Recovery

Candidate tasks:

- [ ] Detect registered paths that no longer exist.
- [ ] Detect registered paths without `.chc/`.
- [ ] Distinguish disabled projects from missing projects.
- [ ] Decide whether disable/remove commands belong here or in maintenance.
- [ ] Add recovery-oriented error messages.
- [ ] Run `make verify`.

## Validation Record

No implementation validation has been run for this tracker yet.

## Notes

- This tracker is for the full `User Project Context` slicing.
- Implementation branches should be created one pending slice at a time.
- Ticket visibility starts in the next slicing, after project context is reliable.
