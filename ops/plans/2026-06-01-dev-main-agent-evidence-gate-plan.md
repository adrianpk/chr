# Dev to Main Agent Evidence Gate Plan

Status: completed

Active Slice: none

Spec: `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`

## Purpose

Implement the approved dev-to-main gate spec as a sequence of behavior-first
slices. The gate must produce objective promotion evidence for Forgejo and
Woodpecker workflows without treating manual work outside the Chryso board
as an error.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice must update its implementation report under
  `.prj/reports/slices/dev-main-agent-evidence-gate/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before continuing.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | gate command and git evidence | `feat/slice-1-gate-command-git-evidence` | `feat(slice-1): gate command and git evidence` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-1-gate-command-git-evidence.md` |
| Slice 2 | validation and report output | `feat/slice-2-validation-report-output` | `feat(slice-2): validation and report output` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-2-validation-report-output.md` |
| Slice 3 | agent evidence classification | `feat/slice-3-agent-evidence-classification` | `feat(slice-3): agent evidence classification` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-3-agent-evidence-classification.md` |
| Slice 4 | gate policy configuration | `feat/slice-4-gate-policy-configuration` | `feat(slice-4): gate policy configuration` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-4-gate-policy-configuration.md` |
| Slice 5 | forgejo woodpecker integration | `feat/slice-5-forgejo-woodpecker-integration` | `feat(slice-5): forgejo woodpecker integration` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-integration.md` |
| Slice 6 | docs qa and end to end validation | `feat/slice-6-docs-qa-e2e-validation` | `feat(slice-6): docs qa and end to end validation` | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-6-docs-qa-e2e-validation.md` |

## Slice 1: Gate Command and Git Evidence

Goal: introduce the `chr gate dev-main` CLI path and collect deterministic Git
range evidence.

Behavior:

- `chr gate dev-main --base <ref> --head <ref>` resolves base, head, and merge
  base.
- The command collects commits, changed files, file statuses, and diff stats.
- The command returns stable errors for invalid refs and invalid usage.
- No report persistence is required beyond enough internal data structures and
  CLI output to validate the behavior.

Tasks:

- T1.1 Add command routing for `gate dev-main`.
- T1.2 Add git evidence collection service.
- T1.3 Add tests for ref resolution, invalid refs, changed files, and diff
  collection.
- T1.4 Add slice report and tracker updates.

## Slice 2: Validation and Report Output

Goal: run validation and write Markdown/JSON reports under
`.chr/default/reports/gates/`.

Behavior:

- Default validation is `go test ./...` for Go projects.
- Reports are written whenever possible, including failure paths.
- Reports include base/head SHAs, merge base, commits, changed files,
  validation status, failures, warnings, and info.
- Exit codes follow the spec.

Tasks:

- T2.1 Add validation command runner with default `go test ./...`.
- T2.2 Add gate JSON report model and writer.
- T2.3 Add gate Markdown report renderer.
- T2.4 Add exit code handling.
- T2.5 Add tests for pass, validation fail, report write behavior, and stale or
  invalid ref failures.
- T2.6 Add slice report and tracker updates.

## Slice 3: Agent Evidence Classification

Goal: classify Chryso-managed evidence separately from manual or external
changes without producing false positives.

Behavior:

- The gate discovers available Chryso run/report artifacts.
- Run reports are the minimum evidence for v1 agent-managed work.
- Manual or external changes are listed as informational by default.
- Missing evidence for Chryso-recorded work is a failure.

Tasks:

- T3.1 Discover gate-relevant reports under `.chr/default/reports`.
- T3.2 Classify agent-managed evidence from available run reports.
- T3.3 Classify unmatched changed files as manual or external info.
- T3.4 Add failure checks for missing Chryso-owned run evidence.
- T3.5 Add tests for agent-managed, manual/external, and missing-evidence cases.
- T3.6 Add slice report and tracker updates.

## Slice 4: Gate Policy Configuration

Goal: support `.chr/default/config/gates.yml` for project-specific validation and
stricter traceability options while preserving safe defaults.

Behavior:

- The gate reads `.chr/default/config/gates.yml` when present.
- Projects can override validation command.
- Manual/external changes remain `info` by default.
- Warnings and stricter behavior only appear when policy enables them.

Tasks:

- T4.1 Add gate config parser.
- T4.2 Support validation command override.
- T4.3 Support traceability mode and manual/external warning settings.
- T4.4 Support high-risk path configuration as warning-only when enabled.
- T4.5 Add tests for default policy and configured policy.
- T4.6 Add slice report and tracker updates.

## Slice 5: Forgejo Woodpecker Integration

Goal: make Forgejo and Woodpecker the first-class promotion path and add the
Woodpecker setup helper.

Behavior:

- `chr gate woodpecker init` produces a runnable Woodpecker snippet.
- `--dry-run` never writes `.woodpecker.yml`.
- If no `.woodpecker.yml` exists, the command may create a minimal config.
- Existing config is not modified without `--write`.
- Complex existing configs fail safely with instructions.
- The gate can expose Forgejo/Woodpecker feedback hooks according to the
  implementation mechanism selected after API/capability investigation.

Tasks:

- T5.1 Investigate Forgejo/Woodpecker status, comments, and artifact link
  mechanisms.
- T5.2 Add `chr gate woodpecker init`.
- T5.3 Generate Woodpecker snippet and minimal config.
- T5.4 Add non-destructive existing-config behavior.
- T5.5 Add Forgejo/Woodpecker feedback integration selected by investigation.
- T5.6 Add tests for helper behavior and feedback failure handling.
- T5.7 Add slice report and tracker updates.

## Slice 6: Docs QA and End to End Validation

Goal: document, validate, and harden the completed feature.

Behavior:

- Reference docs describe commands, config, report schema, exit codes, and
  Woodpecker setup.
- How-to docs show local and Woodpecker usage.
- QA coverage includes a fixture or scenario that exercises local gate behavior.
- End-to-end validation confirms the feature can be used before `dev -> main`.

Tasks:

- T6.1 Add command reference docs.
- T6.2 Add gate config and report schema docs.
- T6.3 Add Woodpecker how-to docs.
- T6.4 Add QA fixture or scripted scenario.
- T6.5 Run `make verify` and the new gate against this repository.
- T6.6 Add final cycle report and tracker closure.

## Open Implementation Notes

- Forgejo/Woodpecker feedback mechanism must be selected during Slice 5 after
  verifying available APIs and CI capabilities.
- Report schema should be versioned from the first implementation.
- Manual or external changes are informational by default and should not become
  warnings without explicit policy.
- The first implementation should avoid GitHub/GitLab assumptions.
