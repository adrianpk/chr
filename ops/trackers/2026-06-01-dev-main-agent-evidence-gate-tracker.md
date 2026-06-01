# Dev to Main Agent Evidence Gate Tracker

Status: active

Spec: `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | gate command and git evidence | delivered | `feat/slice-1-gate-command-git-evidence` | #264 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-1-gate-command-git-evidence.md` |
| Slice 2 | validation and report output | delivered | `feat/slice-2-validation-report-output` | #265 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-2-validation-report-output.md` |
| Slice 3 | agent evidence classification | delivered | `feat/slice-3-agent-evidence-classification` | #266 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-3-agent-evidence-classification.md` |
| Slice 4 | gate policy configuration | delivered | `feat/slice-4-gate-policy-configuration` | #267 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-4-gate-policy-configuration.md` |
| Slice 5 | forgejo woodpecker integration | delivered | `feat/slice-5-forgejo-woodpecker-integration` | #268 | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-integration.md` |
| Slice 6 | docs qa and end to end validation | reviewing | `feat/slice-6-docs-qa-e2e-validation` | pending | `.prj/reports/slices/dev-main-agent-evidence-gate/slice-6-docs-qa-e2e-validation.md` |

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
| T2.1 | Add validation command runner with default `go test ./...`. | done | `cd63af7` | Added default gate validation execution. |
| T2.2 | Add gate JSON report model and writer. | done | `590d75e` | Added JSON report model, timestamped paths, and writer coverage. |
| T2.3 | Add gate Markdown report renderer. | done | `b5ce589` | Added Markdown renderer and writer coverage. |
| T2.4 | Add exit code handling. | done | `6af64c3` | Connected validation and report writing to `chc gate dev-main`. |
| T2.5 | Add tests for pass, validation fail, report write behavior, and stale or invalid ref failures. | done | `51a7165` | Added CLI coverage for pass, validation fail, and invalid ref report flows. |
| T2.6 | Add slice report and tracker updates. | done | this commit | Added the Slice 2 report and validation log. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Discover gate-relevant reports under `.chc/user/reports`. | done | `9fccea6` | Reused project report discovery and resolved execution result paths from report metadata. |
| T3.2 | Classify agent-managed evidence from available run reports. | done | `ba4342f` | Added report fields and command integration for agent-managed work. |
| T3.3 | Classify unmatched changed files as manual or external info. | done | `ba4342f` | Unmatched Git changes are reported as manual or external context. |
| T3.4 | Add failure checks for missing ChrysoCoord-owned run evidence. | done | `e2a9299` | Missing execution evidence now fails the gate. |
| T3.5 | Add tests for agent-managed, manual/external, and missing-evidence cases. | done | `9bf60d3` | Added CLI assertions for manual context and missing evidence failures. |
| T3.6 | Add slice report and tracker updates. | done | this commit | Added the Slice 3 report and validation log. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add gate config parser. | done | `1f0521c` | Added `.chc/user/config/gates.yml` parser and defaults. |
| T4.2 | Support validation command override. | done | `28e1a4d` | Gate validation now reads configured command. |
| T4.3 | Support traceability mode and manual/external warning settings. | done | `16a739e` | Added policy behavior for missing evidence, manual warnings, and manual listing. |
| T4.4 | Support high-risk path configuration as warning-only when enabled. | done | `5066087` | Added high-risk manual change matching and report output. |
| T4.5 | Add tests for default policy and configured policy. | done | `5066087` | Covered defaults, configured policy, CLI behavior, and high-risk matching. |
| T4.6 | Add slice report and tracker updates. | done | this commit | Added the Slice 4 report and validation log. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Investigate Forgejo/Woodpecker status, comments, and artifact link mechanisms. | done | `a0261c4` | Selected Woodpecker CI status as the first feedback path. |
| T5.2 | Add `chc gate woodpecker init`. | done | `2112a09` | Added the Woodpecker setup command under `gate`. |
| T5.3 | Generate Woodpecker snippet and minimal config. | done | `2112a09` | Added deterministic snippet generation and missing-config creation. |
| T5.4 | Add non-destructive existing-config behavior. | done | `2112a09` | Existing configs are not modified without `--write`; unmarked configs fail safely. |
| T5.5 | Add Forgejo/Woodpecker feedback integration selected by investigation. | done | `2112a09` | Command output exposes `woodpecker-status` as the feedback mechanism. |
| T5.6 | Add tests for helper behavior and feedback failure handling. | done | `2112a09` | Added helper and CLI coverage for dry-run, create, existing config, and marked updates. |
| T5.7 | Add slice report and tracker updates. | done | this commit | Added the Slice 5 report and validation log. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Add command reference docs. | done | `ce681ad` | Added `gate dev-main` and `gate woodpecker init` command coverage. |
| T6.2 | Add gate config and report schema docs. | done | `ce681ad` | Added `docs/reference/dev-main-gate.md`. |
| T6.3 | Add Woodpecker how-to docs. | done | `a548b02` | Added the task-oriented Woodpecker setup guide. |
| T6.4 | Add QA fixture or scripted scenario. | done | `0c02534` | Added `make qa-gate` and `scripts/qa/gate.sh`. |
| T6.5 | Run `make verify` and the new gate against this repository. | done | this commit | Validation recorded in the Slice 6 report and log below. |
| T6.6 | Add final cycle report and tracker closure. | done | this commit | Added the Slice 6 report and moved the slice to review. |

## Validation Log

- Slice 1: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 1: `make verify` (pass)
- Slice 2: `go test ./internal/chcapp -run GateDevMain` (pass)
- Slice 2: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 2: `make verify` (pass)
- Slice 3: `go test ./internal/chc -run 'GateEvidence|ReportLister'` (pass)
- Slice 3: `go test ./internal/chc ./internal/chcapp -run 'GateReport|GateEvidence|GateDevMain'` (pass)
- Slice 3: `go test ./internal/chcapp -run 'GateDevMain'` (pass)
- Slice 3: `make verify` (pass)
- Slice 4: `go test ./internal/chc -run GateConfig` (pass)
- Slice 4: `go test ./internal/chc ./internal/chcapp -run 'GateValidation|GateDevMain'` (pass)
- Slice 4: `go test ./internal/chc ./internal/chcapp -run 'GateEvidence|GateReport|GateDevMain'` (pass)
- Slice 4: `make verify` (pass)
- Slice 5: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 5: `make verify` (pass)
- Slice 6: `make verify` (pass)
- Slice 6: `make qa-gate` (pass)
- Slice 6: `GOCACHE=/tmp/chc-gate-self.aRZ3ao/.go-build-cache go run /home/adrian/Projects/chrysocoord/cmd/chc gate dev-main --dir /tmp/chc-gate-self.aRZ3ao --base origin/dev --head HEAD` (pass)

## Notes

- Manual or external changes remain informational by default.
- Forgejo and Woodpecker are first-class targets for this feature.
- The plan must be adjusted before implementation if Forgejo/Woodpecker
  investigation invalidates the Slice 5 feedback mechanism assumptions.
