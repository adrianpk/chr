# Notification and Remote Commands Slicing Tracker

Status: in progress

Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm daemon, loop, ticket result, and context prerequisites are delivered.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Notification Event Model | Delivered | `feat/slice-1-notification-event-model` | `feat(slice-1): notification event model` | `.prj/reports/slices/notification-and-remote-commands/slice-1-notification-event-model.md` | Repository PR #108 merged. |
| Slice 2: Notification Configuration | Delivered | `feat/slice-2-notification-config` | `feat(slice-2): notification config` | `.prj/reports/slices/notification-and-remote-commands/slice-2-notification-config.md` | Repository PR #109 merged. |
| Slice 3: Notification Audit Repository | Delivered | `feat/slice-3-notification-audit-repository` | `feat(slice-3): notification audit repository` | `.prj/reports/slices/notification-and-remote-commands/slice-3-notification-audit-repository.md` | Repository PR #110 merged. |
| Slice 4: Notify Command Skeleton | Delivered | `feat/slice-4-notify-command-skeleton` | `feat(slice-4): notify command skeleton` | `.prj/reports/slices/notification-and-remote-commands/slice-4-notify-command-skeleton.md` | Repository PR #111 merged. |
| Slice 5: Telegram Dispatch Boundary | Delivered | `feat/slice-5-telegram-dispatch-boundary` | `feat(slice-5): telegram dispatch boundary` | `.prj/reports/slices/notification-and-remote-commands/slice-5-telegram-dispatch-boundary.md` | Repository PR #112 merged. |
| Slice 6: Loop and Ticket Notifications | Delivered | `feat/slice-6-loop-ticket-notifications` | `feat(slice-6): loop ticket notifications` | `.prj/reports/slices/notification-and-remote-commands/slice-6-loop-ticket-notifications.md` | Repository PR #113 merged. |
| Slice 7: Remote Command Intake Model | In Review | `feat/slice-7-remote-command-intake-model` | `feat(slice-7): remote command intake model` | `.prj/reports/slices/notification-and-remote-commands/slice-7-remote-command-intake-model.md` | PR pending review. |
| Slice 8: Telegram Remote Command Handoff | Planned | `feat/slice-8-telegram-remote-command-handoff` | `feat(slice-8): telegram remote command handoff` | `.prj/reports/slices/notification-and-remote-commands/slice-8-telegram-remote-command-handoff.md` | Pending implementation. |

## Slice 1: Notification Event Model

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-notification-event-model`
- Expected PR title: `feat(slice-1): notification event model`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-1-notification-event-model.md`

Candidate tasks:

- [x] Add notification event type, severity, status, and source enums.
- [x] Add notification event identity and project/ticket/loop/execution references.
- [x] Validate required fields and lifecycle timestamps.
- [x] Add payload summary and detail fields without channel-specific formatting.
- [x] Add focused model tests.
- [x] Document the event contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `be8fa65`
- T1.2: `be8fa65`
- T1.3: `be8fa65`
- T1.4: `be8fa65`
- T1.5: `8cc19fe`
- T1.6: `bee1d8c`
- T1.7: `docs(slice-1): record notification event report`
- T1.8: `make verify`

## Slice 2: Notification Configuration

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-notification-config`
- Expected PR title: `feat(slice-2): notification config`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-2-notification-config.md`

Candidate tasks:

- [x] Extend workspace config with notification settings.
- [x] Support Telegram channel fields with environment-variable based token lookup.
- [x] Support disabled channels by default.
- [x] Validate channel configuration without contacting external services.
- [x] Add config tests.
- [x] Document configuration fields.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `75ba182`
- T2.2: `75ba182`
- T2.3: `75ba182`
- T2.4: `75ba182`
- T2.5: `75ba182`
- T2.6: `ff93114`
- T2.7: `docs(slice-2): record notification config report`
- T2.8: `make verify`

## Slice 3: Notification Audit Repository

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-notification-audit-repository`
- Expected PR title: `feat(slice-3): notification audit repository`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-3-notification-audit-repository.md`

Candidate tasks:

- [x] Add SQLite migration for notification events.
- [x] Add repository create, update, get, and list operations.
- [x] Store channel, status, timestamps, references, summary, and failure reason.
- [x] Keep payload storage bounded and text-based.
- [x] Add repository tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `059e597`
- T3.2: `059e597`
- T3.3: `059e597`
- T3.4: `059e597`
- T3.5: `059e597`
- T3.6: `docs(slice-3): record notification audit report`
- T3.7: `make verify`

## Slice 4: Notify Command Skeleton

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-notify-command-skeleton`
- Expected PR title: `feat(slice-4): notify command skeleton`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-4-notify-command-skeleton.md`

Candidate tasks:

- [x] Add `chc notify` command group.
- [x] Add `chc notify status`.
- [x] Add `chc notify telegram test --dry-run`.
- [x] Render the planned Telegram test payload from config.
- [x] Avoid network calls in dry-run mode.
- [x] Add focused CLI tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `8ad7bb7`
- T4.2: `8ad7bb7`
- T4.3: `8ad7bb7`
- T4.4: `8ad7bb7`
- T4.5: `8ad7bb7`
- T4.6: `8ad7bb7`
- T4.7: `docs(slice-4): record notify command report`
- T4.8: `make verify`

## Slice 5: Telegram Dispatch Boundary

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-telegram-dispatch-boundary`
- Expected PR title: `feat(slice-5): telegram dispatch boundary`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-5-telegram-dispatch-boundary.md`

Candidate tasks:

- [x] Add Telegram message renderer.
- [x] Add transport interface for sending messages.
- [x] Add HTTP sender implementation with injectable client.
- [x] Wire non-dry-run `notify telegram test`.
- [x] Record audit success or failure.
- [x] Add transport and CLI tests without requiring real credentials.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `132d07f`
- T5.2: `132d07f`
- T5.3: `132d07f`
- T5.4: `132d07f`
- T5.5: `132d07f`
- T5.6: `132d07f`
- T5.7: `docs(slice-5): record telegram dispatch report`
- T5.8: `make verify`

## Slice 6: Loop and Ticket Notifications

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-loop-ticket-notifications`
- Expected PR title: `feat(slice-6): loop ticket notifications`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-6-loop-ticket-notifications.md`

Candidate tasks:

- [x] Emit events when a ticket execution is blocked.
- [x] Emit events when a loop stops on blocker or failure.
- [x] Emit events when a loop finishes with processed tickets.
- [x] Include ticket, loop, execution, and project references.
- [x] Keep dispatch optional and configuration-gated.
- [x] Add integration tests with fake notifier.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: this PR
- T6.2: this PR
- T6.3: this PR
- T6.4: this PR
- T6.5: this PR
- T6.6: this PR
- T6.7: this PR
- T6.8: `make verify`

## Slice 7: Remote Command Intake Model

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-remote-command-intake-model`
- Expected PR title: `feat(slice-7): remote command intake model`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-7-remote-command-intake-model.md`

Candidate tasks:

- [x] Add remote command request model.
- [x] Support status, loop start, ticket start, and ticket comment intents.
- [x] Validate sender, project selector, arguments, and timestamps.
- [x] Add parser for a small text command grammar.
- [x] Reject unknown or unsafe commands clearly.
- [x] Add focused parser and validation tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T7.1: this PR
- T7.2: this PR
- T7.3: this PR
- T7.4: this PR
- T7.5: this PR
- T7.6: this PR
- T7.7: this PR
- T7.8: `make verify`

## Slice 8: Telegram Remote Command Handoff

Pre-flight:

- Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
- Slicing: `Notification and Remote Commands`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-telegram-remote-command-handoff`
- Expected PR title: `feat(slice-8): telegram remote command handoff`
- Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-8-telegram-remote-command-handoff.md`

Candidate tasks:

- [ ] Parse Telegram updates into remote command requests.
- [ ] Persist accepted and rejected remote command records.
- [ ] Map supported commands to existing CLI-equivalent handlers or explicit handoff records.
- [ ] Add `notify telegram poll --dry-run` or equivalent bounded intake preview.
- [ ] Document the remote command safety boundary.
- [ ] Add tests with fixture Telegram updates.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending
- T8.2: pending
- T8.3: pending
- T8.4: pending
- T8.5: pending
- T8.6: pending
- T8.7: pending
- T8.8: pending
