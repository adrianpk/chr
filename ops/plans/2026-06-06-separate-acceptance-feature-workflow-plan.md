# Separate Acceptance Feature Workflow Plan

Status: planned

Active Slice: Slice 3 - acceptance feature cli

Spec: `ops/specs/2026-06-06-separate-acceptance-feature-workflow.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`

## Purpose

Replace the combined spec-plus-Gherkin generation path with a separated
workflow:

```text
dump -> quiz -> spec -> acceptance feature round -> .feature
```

This keeps requirements discovery and specification generation general, then
generates dedicated `.feature` artifacts from richer context when acceptance
testing is useful.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice must update its implementation report under
  `.prj/reports/slices/separate-acceptance-feature-workflow/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before continuing.
- `godog` execution and step-definition generation are out of scope.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | spec feature decoupling | `feat/slice-1-spec-feature-decoupling` | `feat(slice-1): spec feature decoupling` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-1-spec-feature-decoupling.md` |
| Slice 2 | acceptance feature service | `feat/slice-2-acceptance-feature-service` | `feat(slice-2): acceptance feature service` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-2-acceptance-feature-service.md` |
| Slice 3 | acceptance feature cli | `feat/slice-3-acceptance-feature-cli` | `feat(slice-3): acceptance feature cli` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-3-acceptance-feature-cli.md` |
| Slice 4 | feature round defaults | `feat/slice-4-feature-round-defaults` | `feat(slice-4): feature round defaults` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-4-feature-round-defaults.md` |
| Slice 5 | acceptance feature tui action | `feat/slice-5-acceptance-feature-tui-action` | `feat(slice-5): acceptance feature tui action` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-5-acceptance-feature-tui-action.md` |
| Slice 6 | acceptance workflow docs and fixture | `feat/slice-6-acceptance-workflow-docs-and-fixture` | `feat(slice-6): acceptance workflow docs and fixture` | `.prj/reports/slices/separate-acceptance-feature-workflow/slice-6-acceptance-workflow-docs-and-fixture.md` |

## Slice 1: Spec Feature Decoupling

Goal: make spec generation produce only the spec and stop returning or writing
acceptance feature output.

Behavior:

- `spec generate` writes only `.chc/user/specs/<case>-spec.md`.
- The spec model output no longer treats feature output as part of normal spec
  generation.
- Spec body content remains governed by normal spec quality rules, not
  Gherkin-specific rejection.
- Existing static feature validation code remains available for later slices.

Tasks:

- T1.1 Remove feature writing from `GenerateSpec`.
- T1.2 Update spec generation instructions and schema expectations.
- T1.3 Add tests that spec output remains neutral about Gherkin examples while
  rejecting structured feature payloads.
- T1.4 Update affected docs or references that claim spec generation writes
  feature files.
- T1.5 Add slice report and tracker updates.

## Slice 2: Acceptance Feature Service

Goal: add a dedicated service operation for feature generation.

Behavior:

- A service operation generates acceptance features independently from specs.
- The operation loads project, use-case, dump, quiz, and spec context.
- The operation returns `generated`, `needs_more_info`, or `not_applicable`.
- Generated features are written to `.chc/user/acceptance/<case>.feature` only
  after static checks pass.
- The spec is not rewritten as part of feature generation.

Tasks:

- T2.1 Add feature generation request, result, and model contract.
- T2.2 Build feature-generation context from spec, quiz, dump, project, and
  use-case artifacts.
- T2.3 Write checked features to `.chc/user/acceptance/<case>.feature`.
- T2.4 Add service tests for generated, rejected, needs-more-info, and
  not-applicable flows.
- T2.5 Add slice report and tracker updates.

## Slice 3: Acceptance Feature CLI

Goal: expose feature generation without the TUI.

Behavior:

- A CLI command can generate the feature for a selected use case.
- The command prints feature status and path when generated.
- The command returns clear output for `needs_more_info` and `not_applicable`.

Preferred command namespace:

```sh
chc acceptance feature generate --use-case <case>
```

Tasks:

- T3.1 Add `acceptance feature generate` command wiring.
- T3.2 Add command output for generated, needs-more-info, and not-applicable.
- T3.3 Add CLI tests.
- T3.4 Add command reference updates if needed for this slice.
- T3.5 Add slice report and tracker updates.

## Slice 4: Feature Round Defaults

Goal: add a feature-specific clarification round that uses defaults inferred
from the spec and quiz.

Behavior:

- Feature generation can report missing scenario details as default-oriented
  prompts.
- Feature-round questions ask the user to confirm or adjust inferred defaults.
- Feature-round notes are persisted separately from the general quiz.
- The general quiz no longer needs strong early Gherkin bias.

Tasks:

- T4.1 Add feature readiness/default prompt model.
- T4.2 Persist feature-round notes under a stable acceptance artifact path.
- T4.3 Use feature-round notes as input for subsequent feature generation.
- T4.4 Add tests for default confirmation and missing-info persistence.
- T4.5 Add slice report and tracker updates.

## Slice 5: Acceptance Feature TUI Action

Goal: expose feature generation from the TUI.

Behavior:

- The Specs view exposes a `Generate feature` action for the selected use case.
- The TUI shows generated feature status and path when available.
- If the feature operation needs more information, the TUI surfaces the
  default-oriented feature prompt.
- The action reuses the same service boundary as the CLI.

Tasks:

- T5.1 Add acceptance feature read model fields for Specs or a focused
  Acceptance action surface.
- T5.2 Add TUI action executor support for feature generation.
- T5.3 Render feature status/path and needs-more-info prompts.
- T5.4 Add TUI tests for command/action availability and result rendering.
- T5.5 Add slice report and tracker updates.

## Slice 6: Acceptance Workflow Docs And Fixture

Goal: document and validate the final separated workflow.

Behavior:

- User guide describes `dump -> quiz -> spec -> acceptance feature`.
- Reference docs describe acceptance artifact paths and CLI commands.
- Explanation docs clarify why specs and `.feature` artifacts are separate.
- Dialer `call-outcome-capture` remains a resettable testbed.

Tasks:

- T6.1 Update the User Guide with the separated workflow.
- T6.2 Update command and artifact reference docs.
- T6.3 Update explanation docs with the design rationale.
- T6.4 Validate the dialer fixture against the new flow.
- T6.5 Run `make verify` and add final cycle report/tracker closure.

## Open Implementation Notes

- Prefer strict generation contracts over post-hoc Markdown parsing.
- Keep `.feature` files as first-class acceptance artifacts.
- Do not execute `godog`; only preserve a layout that makes later execution
  straightforward.
- Do not add content-based Gherkin rejection to spec generation; the separated
  workflow boundary is the structured feature payload and artifact writing.
