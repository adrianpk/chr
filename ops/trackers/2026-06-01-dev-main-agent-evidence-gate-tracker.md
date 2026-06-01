# Dev to Main Agent Evidence Gate Tracker

Status: active

Spec: `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | gate command and git evidence | delivered | `feat/slice-1-gate-command-git-evidence` | #264 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-1-gate-command-git-evidence.md` |
| Slice 2 | validation and report output | pending | `feat/slice-2-validation-report-output` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-2-validation-report-output.md` |
| Slice 3 | agent evidence classification | pending | `feat/slice-3-agent-evidence-classification` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-3-agent-evidence-classification.md` |
| Slice 4 | gate policy configuration | pending | `feat/slice-4-gate-policy-configuration` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-4-gate-policy-configuration.md` |
| Slice 5 | forgejo woodpecker integration | pending | `feat/slice-5-forgejo-woodpecker-integration` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-integration.md` |
| Slice 6 | docs qa and end to end validation | pending | `feat/slice-6-docs-qa-e2e-validation` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-6-docs-qa-e2e-validation.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add command routing for `gate dev-main`. | done | `347fc7f` | Registered the root `gate` command and `dev-main` subcommand. |
| T1.2 | Add git evidence collection service. | done | `33b87ae` | Added Git evidence collection and CLI output. |
| T1.3 | Add tests for ref resolution, invalid refs, changed files, and diff collection. | done | `080b71e` | Added package and CLI command coverage. |
| T1.4 | Add slice report and tracker updates. | done | this commit | This tracker update and slice report. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add validation command runner with default `go test ./...`. | pending | pending |  |
| T2.2 | Add gate JSON report model and writer. | pending | pending |  |
| T2.3 | Add gate Markdown report renderer. | pending | pending |  |
| T2.4 | Add exit code handling. | pending | pending |  |
| T2.5 | Add tests for pass, validation fail, report write behavior, and stale or invalid ref failures. | pending | pending |  |
| T2.6 | Add slice report and tracker updates. | pending | pending |  |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Discover gate-relevant reports under `.chc/user/reports`. | pending | pending |  |
| T3.2 | Classify agent-managed evidence from available run reports. | pending | pending |  |
| T3.3 | Classify unmatched changed files as manual or external info. | pending | pending |  |
| T3.4 | Add failure checks for missing ChrysoCoord-owned run evidence. | pending | pending |  |
| T3.5 | Add tests for agent-managed, manual/external, and missing-evidence cases. | pending | pending |  |
| T3.6 | Add slice report and tracker updates. | pending | pending |  |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add gate config parser. | pending | pending |  |
| T4.2 | Support validation command override. | pending | pending |  |
| T4.3 | Support traceability mode and manual/external warning settings. | pending | pending |  |
| T4.4 | Support high-risk path configuration as warning-only when enabled. | pending | pending |  |
| T4.5 | Add tests for default policy and configured policy. | pending | pending |  |
| T4.6 | Add slice report and tracker updates. | pending | pending |  |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Investigate Forgejo/Woodpecker status, comments, and artifact link mechanisms. | pending | pending |  |
| T5.2 | Add `chc gate woodpecker init`. | pending | pending |  |
| T5.3 | Generate Woodpecker snippet and minimal config. | pending | pending |  |
| T5.4 | Add non-destructive existing-config behavior. | pending | pending |  |
| T5.5 | Add Forgejo/Woodpecker feedback integration selected by investigation. | pending | pending |  |
| T5.6 | Add tests for helper behavior and feedback failure handling. | pending | pending |  |
| T5.7 | Add slice report and tracker updates. | pending | pending |  |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Add command reference docs. | pending | pending |  |
| T6.2 | Add gate config and report schema docs. | pending | pending |  |
| T6.3 | Add Woodpecker how-to docs. | pending | pending |  |
| T6.4 | Add QA fixture or scripted scenario. | pending | pending |  |
| T6.5 | Run `make verify` and the new gate against this repository. | pending | pending |  |
| T6.6 | Add final cycle report and tracker closure. | pending | pending |  |

## Validation Log

- Slice 1: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 1: `make verify` (pass)

## Notes

- Manual or external changes remain informational by default.
- Forgejo and Woodpecker are first-class targets for this feature.
- The plan must be adjusted before implementation if Forgejo/Woodpecker
  investigation invalidates the Slice 5 feedback mechanism assumptions.
