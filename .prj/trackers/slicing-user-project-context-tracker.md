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
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.

## Slice Status

| Slice | Status | Notes |
| --- | --- | --- |
| Slice 1: Project Registry Persistence | Delivered | Repository PR #9 title: `feat(slice-1): project registry repository`. |
| Slice 2: Project Registry CLI | Delivered | Repository PR #10 title: `feat(slice-2): project registry cli`. |
| Slice 3: Active Project Persistence | Delivered | Store and load selected project state. Repository PR #13 title: `feat(slice-3): active project persistence`. |
| Slice 4: Project Select and Current CLI | Delivered | Add `project select` and `project current`. Repository PR #14 title: `feat(slice-4): project select current cli`. |
| Slice 5: Project Context Resolver | Delivered | Reusable project resolution for later commands. Repository PR #15 title: `feat(slice-5): project context resolver`. |
| Slice 6: Project Validation and Recovery | Active | Clear stale/missing/disabled project behavior. Repository PR #16 title: `feat(slice-6): project validation recovery`. |

## Slice 3: Active Project Persistence

Candidate tasks:

- [ ] Decide persistence location from the existing user environment model.
- [ ] Add a core model for active project state.
- [ ] Add a repository or config service for storing active project state.
- [ ] Add tests for empty, set, replace, and missing-project cases.
- [ ] Keep development-mode state under `_workspace`.
- [ ] Run `make verify`.

## Slice 4: Project Select and Current CLI

Candidate tasks:

- [ ] Add `chc project select <name-or-path>`.
- [ ] Add `chc project current`.
- [ ] Resolve registered project by name.
- [ ] Resolve registered project by path when unambiguous.
- [ ] Print stable current-project output.
- [ ] Add focused CLI tests.
- [ ] Run `make verify`.

## Slice 5: Project Context Resolver

Candidate tasks:

- [ ] Add reusable core command context resolution.
- [ ] Support explicit project input where command semantics require it.
- [ ] Fall back to active project where appropriate.
- [ ] Preserve current-directory behavior for project-local commands that should keep it.
- [ ] Return clear errors when no project can be resolved.
- [ ] Add tests for explicit, active, current-directory, and missing-context cases.
- [ ] Run `make verify`.

## Slice 6: Project Validation and Recovery

Candidate tasks:

- [ ] Detect registered paths that no longer exist.
- [ ] Detect registered paths without `.chc/`.
- [ ] Distinguish disabled projects from missing projects.
- [ ] Decide whether disable/remove commands belong here or in maintenance.
- [ ] Add recovery-oriented error messages.
- [ ] Run `make verify`.

## Validation Record

- Active Project Persistence: `make verify` pass.
- Project Select and Current CLI: `make verify` pass.
- Project Context Resolver: `make verify` pass.

## Notes

- This tracker is for the full `User Project Context` slicing.
- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- Implementation branches should be created one planned slice at a time.
- Ticket visibility starts in the next slicing, after project context is reliable.
