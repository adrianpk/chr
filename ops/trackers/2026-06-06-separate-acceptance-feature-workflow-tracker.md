# Separate Acceptance Feature Workflow Tracker

Status: in_progress

Spec: `ops/specs/2026-06-06-separate-acceptance-feature-workflow.md`
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | spec feature decoupling | delivered | `feat/slice-1-spec-feature-decoupling` | `#294` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-1-spec-feature-decoupling.md` |
| Slice 2 | acceptance feature service | delivered | `feat/slice-2-acceptance-feature-service` | `#295` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-2-acceptance-feature-service.md` |
| Slice 3 | acceptance feature cli | reviewing | `feat/slice-3-acceptance-feature-cli` | pending | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-3-acceptance-feature-cli.md` |
| Slice 4 | feature round defaults | pending | `feat/slice-4-feature-round-defaults` | pending | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-4-feature-round-defaults.md` |
| Slice 5 | acceptance feature tui action | pending | `feat/slice-5-acceptance-feature-tui-action` | pending | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-5-acceptance-feature-tui-action.md` |
| Slice 6 | acceptance workflow docs and fixture | pending | `feat/slice-6-acceptance-workflow-docs-and-fixture` | pending | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-6-acceptance-workflow-docs-and-fixture.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Remove feature writing from `GenerateSpec`. | done | `95b8871` | Spec generation now writes only the Markdown spec. |
| T1.2 | Update spec generation instructions and schema expectations. | done | `c9914e0` | Spec output schema no longer includes `feature`; stale feature fields are rejected. |
| T1.3 | Add tests that spec output remains neutral about Gherkin examples while rejecting structured feature payloads. | done | `e210b73`, corrected by `9626bcc` | Spec body content is not rejected for Gherkin examples. |
| T1.4 | Update affected docs or references that claim spec generation writes feature files. | done | `ff0fc93` | Docs now describe spec generation as spec-only. |
| T1.5 | Add slice report and tracker updates. | done | this commit | Added Slice 1 report, tracker status, and validation log. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add feature generation request, result, and model contract. | done | `648644c`, corrected by `f8573a3`, `5ed3d64` | Added feature action, statuses, schema, decoder, and provider routing. |
| T2.2 | Build feature-generation context from spec, quiz, dump, project, and use-case artifacts. | done | `50436c1`, corrected by `ef35761`, `5ed3d64` | Added service request construction and context sections including specs. |
| T2.3 | Write checked features to `.chc/user/acceptance/<case>.feature`. | done | `adec408`, corrected by `5ed3d64` | Generated features write only after static checks pass. |
| T2.4 | Add service tests for generated, rejected, needs-more-info, and not-applicable flows. | done | `336d472`, corrected by `6c298de` | Added dedicated service tests for all feature generation outcomes. |
| T2.5 | Add slice report and tracker updates. | done | this commit | Added Slice 2 report, tracker status, and validation log. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add `acceptance feature generate` command wiring. | done | `c98209d` | Added top-level `acceptance feature generate` command. |
| T3.2 | Add command output for generated, needs-more-info, and not-applicable. | done | `c98209d` | Command prints status, feature path, and reason output. |
| T3.3 | Add CLI tests. | done | `fe91585` | Added command registration and output tests. |
| T3.4 | Add command reference updates if needed for this slice. | done | `561cb1b` | Documented acceptance feature command and removed old spec-generation feature wording. |
| T3.5 | Add slice report and tracker updates. | done | this commit | Added Slice 3 report, tracker status, and validation log. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add feature readiness/default prompt model. | pending | pending |  |
| T4.2 | Persist feature-round notes under a stable acceptance artifact path. | pending | pending |  |
| T4.3 | Use feature-round notes as input for subsequent feature generation. | pending | pending |  |
| T4.4 | Add tests for default confirmation and missing-info persistence. | pending | pending |  |
| T4.5 | Add slice report and tracker updates. | pending | pending |  |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add acceptance feature read model fields for Specs or a focused Acceptance action surface. | pending | pending |  |
| T5.2 | Add TUI action executor support for feature generation. | pending | pending |  |
| T5.3 | Render feature status/path and needs-more-info prompts. | pending | pending |  |
| T5.4 | Add TUI tests for command/action availability and result rendering. | pending | pending |  |
| T5.5 | Add slice report and tracker updates. | pending | pending |  |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Update the User Guide with the separated workflow. | pending | pending |  |
| T6.2 | Update command and artifact reference docs. | pending | pending |  |
| T6.3 | Update explanation docs with the design rationale. | pending | pending |  |
| T6.4 | Validate the dialer fixture against the new flow. | pending | pending |  |
| T6.5 | Run `make verify` and add final cycle report/tracker closure. | pending | pending |  |

## Validation Log

- Slice 1:
  - `go test ./internal/chc ./internal/chcapp -run 'Spec|Feature|Gherkin'` (pass)
  - `go test ./internal/chc ./internal/chcapp` (pass)
  - `make verify` (pass)
- Slice 2:
  - `go test ./internal/chc ./internal/chcapp` (pass)
  - `make verify` (pass)
- Slice 3:
  - `go test ./internal/chcapp -run 'RootCommand|AcceptanceFeatureGenerate'` (pass)
  - `go test ./internal/chc ./internal/chcapp` (pass)
  - `make verify` (pass)

## Notes

- `godog` execution is intentionally deferred.
- Specs remain canonical requirements artifacts.
- Feature files are separate acceptance artifacts.
