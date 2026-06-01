# Ticket Review Integration Slicing Tracker

Status: reviewing

Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Spec Draft: `.prj/specs/drafts/ticket-review-integration-draft.md`
Quiz: `.prj/quizzes/curated/ticket-review-integration-quiz.md`
Mirror: `.chc/user/trackers/ticket-review-integration-tracker.md`

## Slicing Checklist

- [x] Capture ticket review integration draft.
- [x] Capture raw and curated quiz artifacts.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | ticket review metadata contract | delivered | `feat/slice-1-ticket-review-metadata-contract` | `#273` | `.prj/reports/slices/ticket-review-integration/slice-1-ticket-review-metadata-contract.md` |
| Slice 2 | pr reference workflow | delivered | `feat/slice-2-pr-reference-workflow` | `#274` | `.prj/reports/slices/ticket-review-integration/slice-2-pr-reference-workflow.md` |
| Slice 3 | review mode configuration | pending | `feat/slice-3-review-mode-configuration` | pending | `.prj/reports/slices/ticket-review-integration/slice-3-review-mode-configuration.md` |
| Slice 4 | forge provider diagnostics | pending | `feat/slice-4-forge-provider-diagnostics` | pending | `.prj/reports/slices/ticket-review-integration/slice-4-forge-provider-diagnostics.md` |
| Slice 5 | remote review adapter | pending | `feat/slice-5-remote-review-adapter` | pending | `.prj/reports/slices/ticket-review-integration/slice-5-remote-review-adapter.md` |
| Slice 6 | ticket review command | pending | `feat/slice-6-ticket-review-command` | pending | `.prj/reports/slices/ticket-review-integration/slice-6-ticket-review-command.md` |
| Slice 7 | serialized work preflight | pending | `feat/slice-7-serialized-work-preflight` | pending | `.prj/reports/slices/ticket-review-integration/slice-7-serialized-work-preflight.md` |
| Slice 8 | repair handoff and docs | pending | `feat/slice-8-repair-handoff-docs` | pending | `.prj/reports/slices/ticket-review-integration/slice-8-repair-handoff-docs.md` |

## Slice 1: Ticket Review Metadata Contract

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-ticket-review-metadata-contract`
- Expected PR title: `feat(slice-1): ticket review metadata contract`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-1-ticket-review-metadata-contract.md`

Tasks:

- [x] T1.1 Add ticket metadata fields and validation rules. Commit: `1fb54e1`
- [x] T1.2 Update ticket parser and renderer preservation behavior. Commit: `1fb54e1`
- [x] T1.3 Update ticket index/read models and board attention display. Commit: `1fb54e1`
- [x] T1.4 Add tests for metadata parsing, rendering, indexing, and legacy tickets. Commit: `1fb54e1`
- [x] T1.5 Add reference documentation for the ticket review metadata contract. Commit: `ed0404f`
- [x] T1.6 Add slice report and tracker updates. Commit: this commit

## Slice 2: Pull Request Reference Workflow

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-pr-reference-workflow`
- Expected PR title: `feat(slice-2): pr reference workflow`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-2-pr-reference-workflow.md`

Tasks:

- [x] T2.1 Add command/service support for attaching branch and pull request refs. Commit: `f1f4188`
- [x] T2.2 Validate `ready_for_review` requirements for branch and PR metadata. Commit: `f1f4188`
- [x] T2.3 Preserve existing PR refs during execution result application. Commit: `f1f4188`
- [x] T2.4 Add CLI/TUI operation paths for branch and PR metadata updates. Commit: `f1f4188`
- [x] T2.5 Add tests for valid refs, missing refs, and repeated updates. Commit: `f1f4188`
- [x] T2.6 Add slice report and tracker updates. Commit: this commit

## Slice 3: Review Mode Configuration

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-review-mode-configuration`
- Expected PR title: `feat(slice-3): review mode configuration`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-3-review-mode-configuration.md`

Tasks:

- [ ] T3.1 Add review workflow config schema and defaults.
- [ ] T3.2 Add config loader and validation for review modes.
- [ ] T3.3 Add dry-run/status output that explains the active review mode.
- [ ] T3.4 Add tests for default, human, agentic, and trusted modes.
- [ ] T3.5 Document review mode configuration.
- [ ] T3.6 Add slice report and tracker updates.

## Slice 4: Forge Provider Diagnostics

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-forge-provider-diagnostics`
- Expected PR title: `feat(slice-4): forge provider diagnostics`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-4-forge-provider-diagnostics.md`

Tasks:

- [ ] T4.1 Add provider configuration model for Forgejo/Codeberg and GitHub.
- [ ] T4.2 Add provider detection from pull request URLs.
- [ ] T4.3 Add diagnostics for `tea`, Forgejo-compatible tokens, and PR readability.
- [ ] T4.4 Add diagnostics for `gh`, GitHub tokens/auth, and PR readability.
- [ ] T4.5 Add CLI output and tests for pass, missing tooling, missing token, and unsupported provider cases.
- [ ] T4.6 Add slice report and tracker updates.

## Slice 5: Remote Review Adapter

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-remote-review-adapter`
- Expected PR title: `feat(slice-5): remote review adapter`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-5-remote-review-adapter.md`

Tasks:

- [ ] T5.1 Define remote review adapter contracts and result types.
- [ ] T5.2 Implement Forgejo/Codeberg read paths for PR metadata, diff, comments, and checks.
- [ ] T5.3 Implement Forgejo/Codeberg comment writing with inline-first behavior.
- [ ] T5.4 Implement GitHub read and comment paths.
- [ ] T5.5 Add tests with provider fixtures or fakes for read, inline comment, general comment, and unavailable capability cases.
- [ ] T5.6 Add slice report and tracker updates.

## Slice 6: Ticket Review Command

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-ticket-review-command`
- Expected PR title: `feat(slice-6): ticket review command`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-6-ticket-review-command.md`

Tasks:

- [ ] T6.1 Add ticket review command/service boundary.
- [ ] T6.2 Add evidence assembler for ticket, execution, validation, and PR data.
- [ ] T6.3 Add reviewer prompt/policy for inspection-only, high-signal review.
- [ ] T6.4 Add result application for pass, needs changes, insufficient evidence, and human decision cases.
- [ ] T6.5 Add tests for clean pass, blocking findings, non-blocking findings, missing remote access, and dry-run.
- [ ] T6.6 Add slice report and tracker updates.

## Slice 7: Serialized Work Preflight

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-serialized-work-preflight`
- Expected PR title: `feat(slice-7): serialized work preflight`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-7-serialized-work-preflight.md`

Tasks:

- [ ] T7.1 Add workspace-level agent operation lock for review/development paths.
- [ ] T7.2 Add series/dependency metadata to work selection inputs.
- [ ] T7.3 Add preflight checks for unresolved review repairs.
- [ ] T7.4 Add preflight checks for prerequisite PR merged and contained in updated `dev`.
- [ ] T7.5 Add scheduler behavior to select repair work before later series work.
- [ ] T7.6 Add tests for blocked next slice, repair priority, stale dev, and lock conflicts.
- [ ] T7.7 Add slice report and tracker updates.

## Slice 8: Repair Handoff and Docs

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
- Slicing: `Ticket Review Integration`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-repair-handoff-docs`
- Expected PR title: `feat(slice-8): repair handoff docs`
- Expected report: `.prj/reports/slices/ticket-review-integration/slice-8-repair-handoff-docs.md`

Tasks:

- [ ] T8.1 Add repair handoff context for existing branch and pull request.
- [ ] T8.2 Add behavior for remote comment access failure during repair.
- [ ] T8.3 Add tests for repair context, existing branch reuse, and no next-slice selection while repair is pending.
- [ ] T8.4 Add reference docs for ticket review integration.
- [ ] T8.5 Add how-to docs for Forgejo-first setup and review operation.
- [ ] T8.6 Run `make verify` and a ticket review QA scenario.
- [ ] T8.7 Add final cycle report and tracker closure.

## Validation Log

- Slice 1:
  - `go test ./internal/chc -run 'Ticket|Board|Migrat'` - pass
  - `go test ./internal/chcapp -run 'Ticket(List|Show|Board)|Migration'` - pass
  - `go test ./internal/chctui -run 'Ticket|Board'` - pass
  - `go test ./internal/chc ./internal/chcapp ./internal/chctui` - pass
  - `make verify` - pass
- Slice 2:
  - `go test ./internal/chc -run 'Ticket(Review|Status)|ReviewMetadata'` - pass
  - `go test ./internal/chcapp -run 'Ticket(ReviewRef|Status|ApplyResult)'` - pass
  - `go test ./internal/chc ./internal/chcapp ./internal/chctui` - pass
  - `make verify` - pass

## Notes

- The first implementation slice is not executable until this plan and tracker
  are committed to `dev`.
- The reviewer prompt must stay inspection-only and anti-nit by default.
- Remote comments are written only for findings; clean passes stay quiet.
