# Notification and Remote Commands Slicing Tracker

Status: planned

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
| Slice 1: Notification Event Model | Reviewing | `feat/slice-1-notification-event-model` | `feat(slice-1): notification event model` | `.prj/reports/slices/notification-and-remote-commands/slice-1-notification-event-model.md` | Repository PR #108. |
| Slice 2: Notification Configuration | Planned | `feat/slice-2-notification-config` | `feat(slice-2): notification config` | `.prj/reports/slices/notification-and-remote-commands/slice-2-notification-config.md` | Pending implementation. |
| Slice 3: Notification Audit Repository | Planned | `feat/slice-3-notification-audit-repository` | `feat(slice-3): notification audit repository` | `.prj/reports/slices/notification-and-remote-commands/slice-3-notification-audit-repository.md` | Pending implementation. |
| Slice 4: Notify Command Skeleton | Planned | `feat/slice-4-notify-command-skeleton` | `feat(slice-4): notify command skeleton` | `.prj/reports/slices/notification-and-remote-commands/slice-4-notify-command-skeleton.md` | Pending implementation. |
| Slice 5: Telegram Dispatch Boundary | Planned | `feat/slice-5-telegram-dispatch-boundary` | `feat(slice-5): telegram dispatch boundary` | `.prj/reports/slices/notification-and-remote-commands/slice-5-telegram-dispatch-boundary.md` | Pending implementation. |
| Slice 6: Loop and Ticket Notifications | Planned | `feat/slice-6-loop-ticket-notifications` | `feat(slice-6): loop ticket notifications` | `.prj/reports/slices/notification-and-remote-commands/slice-6-loop-ticket-notifications.md` | Pending implementation. |
| Slice 7: Remote Command Intake Model | Planned | `feat/slice-7-remote-command-intake-model` | `feat(slice-7): remote command intake model` | `.prj/reports/slices/notification-and-remote-commands/slice-7-remote-command-intake-model.md` | Pending implementation. |
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

- [ ] Extend workspace config with notification settings.
- [ ] Support Telegram channel fields with environment-variable based token lookup.
- [ ] Support disabled channels by default.
- [ ] Validate channel configuration without contacting external services.
- [ ] Add config tests.
- [ ] Document configuration fields.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1: pending
- T2.2: pending
- T2.3: pending
- T2.4: pending
- T2.5: pending
- T2.6: pending
- T2.7: pending
- T2.8: pending

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

- [ ] Add SQLite migration for notification events.
- [ ] Add repository create, update, get, and list operations.
- [ ] Store channel, status, timestamps, references, summary, and failure reason.
- [ ] Keep payload storage bounded and text-based.
- [ ] Add repository tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending
- T3.7: pending

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

- [ ] Add `chc notify` command group.
- [ ] Add `chc notify status`.
- [ ] Add `chc notify telegram test --dry-run`.
- [ ] Render the planned Telegram test payload from config.
- [ ] Avoid network calls in dry-run mode.
- [ ] Add focused CLI tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending
- T4.2: pending
- T4.3: pending
- T4.4: pending
- T4.5: pending
- T4.6: pending
- T4.7: pending
- T4.8: pending

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

- [ ] Add Telegram message renderer.
- [ ] Add transport interface for sending messages.
- [ ] Add HTTP sender implementation with injectable client.
- [ ] Wire non-dry-run `notify telegram test`.
- [ ] Record audit success or failure.
- [ ] Add transport and CLI tests without requiring real credentials.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending
- T5.2: pending
- T5.3: pending
- T5.4: pending
- T5.5: pending
- T5.6: pending
- T5.7: pending
- T5.8: pending

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

- [ ] Emit events when a ticket execution is blocked.
- [ ] Emit events when a loop stops on blocker or failure.
- [ ] Emit events when a loop finishes with processed tickets.
- [ ] Include ticket, loop, execution, and project references.
- [ ] Keep dispatch optional and configuration-gated.
- [ ] Add integration tests with fake notifier.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending
- T6.7: pending
- T6.8: pending

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

- [ ] Add remote command request model.
- [ ] Support status, loop start, ticket start, and ticket comment intents.
- [ ] Validate sender, project selector, arguments, and timestamps.
- [ ] Add parser for a small text command grammar.
- [ ] Reject unknown or unsafe commands clearly.
- [ ] Add focused parser and validation tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending
- T7.7: pending
- T7.8: pending

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
