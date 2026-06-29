# Bootstrap Safety Pass

Status: delivered
Slicing: Managed Project Bootstrap
Plan: `.prj/plans/slicing-managed-project-bootstrap-plan.md`
Tracker: `.prj/trackers/slicing-managed-project-bootstrap-tracker.md`
Branch: `feat/bootstrap-safety-pass`
PR: `#116`

## Purpose

Close the remaining bootstrap safety pass by verifying `chr init` from the CLI
boundary against dry-run, existing-file preservation, invalid target handling,
and real embedded profile assets.

## Delivered Behavior

No production behavior changes were required. Existing `chr init` behavior
already met the reviewed bootstrap safety criteria.

This slice adds CLI-boundary tests that prove:

- `chr init --dry-run` prints planned paths without writing `.chr`;
- existing root `AGENTS.md` is preserved and reported as skipped;
- the standard profile installs embedded protocol assets from the real asset
  filesystem;
- file targets are rejected with a clear not-a-directory error.

## Files of Interest

- `internal/chrapp/cli_test.go`
- `.prj/trackers/slicing-managed-project-bootstrap-tracker.md`

## Validation

- `go test ./internal/chrapp -run 'TestInit'`: pass
- `make verify`: pass

## Risks and Follow-ups

- Permission-specific failures are environment-dependent and were not forced
  through brittle filesystem mode tests.
- The slice intentionally avoids expanding bootstrap scope into registry,
  tickets, runtime, daemon, or execution behavior.
