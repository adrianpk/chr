# Slice 1: QA Workspace Reset Scaffold

Status: reviewing
Slicing: QA Agent Work Scenario
Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`
Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`
Branch: `feat/slice-1-qa-workspace-reset-scaffold`
PR: pending

## Purpose

Create the reusable reset and validation mechanics for QA scenarios before the
full blog reference project is added.

## Delivered Behavior

- `qa/ref` can hold versioned scenario references.
- `qa/project` and `qa/tmp` are ignored generated workspaces.
- `make qa-reset-blog` creates `qa/project/blog` from `qa/ref/001/blog`.
- Reset moves any previous active project to `qa/tmp/blog`.
- Reset initializes `qa/project/blog` as an independent Git repository with a
  clean base commit on `main`.
- `make qa-test-blog` dispatches validation for the active blog project.
- The placeholder blog reference proves reset mechanics without introducing the
  full Go app yet.

## Implementation Notes

The reset script uses explicit repository-relative paths and `git -C` for all
Git operations inside the generated active project. This keeps the nested
project repository separate from the parent `chc` repository.

`qa-test-blog` currently exits successfully with a clear message when the active
project has no Go module. Slice 2 replaces the placeholder reference with the
Go blog module, at which point the same target will run `go test ./...`.

## Contracts Added or Changed

- New generated workspace ignore convention:
  - `qa/project/`
  - `qa/tmp/`
- New Make targets:
  - `qa-reset`
  - `qa-test`
  - `qa-reset-blog`
  - `qa-test-blog`
- New script contracts:
  - `scripts/qa/reset.sh <case> [ref]`
  - `scripts/qa/test.sh <case>`

## Files of Interest

- `.gitignore`
- `Makefile`
- `scripts/qa/reset.sh`
- `scripts/qa/test.sh`
- `qa/ref/001/blog/README.md`
- `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

## Validation

- `make qa-reset-blog`
- `make qa-test-blog`
- `make verify`
- `git diff --check`

## Risks and Follow-ups

- The active blog project has no Go module until Slice 2.
- The reset script intentionally replaces the previous `qa/tmp/blog` copy
  instead of retaining timestamped history.
