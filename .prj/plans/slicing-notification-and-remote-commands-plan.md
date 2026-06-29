# Notification and Remote Commands Slicing Plan

Status: completed

## Purpose

Plan the `Notification and Remote Commands` slicing from the feature slicing set
map.

This slicing covers the use case:

> `chr` notifies the user and accepts remote input when asynchronous execution
> needs attention.

## Source

Primary sources:

- `.prj/drafts/chr-feature-slicing-set-map.md`
- `.prj/drafts/chr-command-surfaces-and-priority.md`

Related documents:

- `.prj/plans/slicing-daemon-coordination-plan.md`
- `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- `.prj/plans/slicing-ticket-result-application-plan.md`
- `docs/reference/daemon-coordination.md`
- `docs/reference/ticket-loop.md`
- `docs/explanation/ticket-resolution-lifecycle.md`

Prerequisite delivered work:

- `Daemon Coordination` slicing.
- `Filtered Ticket Loop` slicing.
- `Ticket Result Application` slicing.
- `Ticket Execution Context Hardening` slicing.

## Current State

Delivered prerequisites:

- foreground ticket loops can stop on blockers and failures;
- daemon coordination can run foreground loop iterations over registered
  projects;
- tickets can record comments, blockers, validation evidence, and changed files;
- loop and execution records persist durable status;
- context and evidence artifacts are materialized for agent execution.

Delivered:

- notification event model;
- notification audit records;
- channel configuration for Telegram;
- notification dispatch from loop and ticket outcomes;
- remote command intake;
- remote command handoff into existing CLI-equivalent operations.

## Slicing Scope

This slicing includes:

- a core notification event model;
- channel-neutral notification payloads;
- user-level notification audit persistence;
- project-local notification configuration shape;
- setup-check commands for notification channels;
- Telegram as the first planned remote command channel;
- event emission for blocked tickets and completed loops;
- bounded remote command parsing for status, loop start, ticket start, and
  ticket comment handoff;
- documentation for current and future notification behavior.

## Out of Scope

This slicing does not include:

- multi-user coordination;
- multiple agents mutating a repository at the same time;
- hosted services;
- background OS service installation;
- full email delivery unless a later slice explicitly narrows the scope;
- forge or pull-request automation;
- secret management beyond reading configured environment variable names;
- unbounded conversational control over the project.

## Design Direction

Manual `notify` commands are setup checks. They should prove configuration and
transport wiring, but they are not the primary workflow.

The primary workflow is system-triggered:

- a ticket becomes blocked;
- validation repeatedly fails;
- a loop finishes;
- a loop stops on failure or blocker;
- a daemon state needs attention;
- a remote reply should become a ticket comment or explicit command request.

Remote commands should reuse existing local command semantics. They should not
define a separate product surface that behaves differently from the CLI.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Notification Event Model

Status: delivered.

Expected branch: `feat/slice-1-notification-event-model`
Expected PR title: `feat(slice-1): notification event model`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-1-notification-event-model.md`

Goal:

Define the channel-neutral notification event contract before adding transport
or CLI behavior.

Candidate scope:

- add notification event type, severity, status, and source enums;
- add notification event identity and project/ticket/loop/execution references;
- validate required fields and lifecycle timestamps;
- add payload summary and detail fields without channel-specific formatting;
- add focused model tests;
- document the event contract.

Expected outcome:

- later slices can persist, render, and dispatch notifications without inventing
  event shape in the channel layer.

### Slice 2: Notification Configuration

Status: delivered.

Expected branch: `feat/slice-2-notification-config`
Expected PR title: `feat(slice-2): notification config`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-2-notification-config.md`

Goal:

Add project-local notification configuration and channel enablement rules.

Candidate scope:

- extend workspace config with notification settings;
- support Telegram channel fields with environment-variable based token lookup;
- support disabled channels by default;
- validate channel configuration without contacting external services;
- add config tests;
- document configuration fields.

Expected outcome:

- notification channels can be configured explicitly while keeping default
  projects silent.

### Slice 3: Notification Audit Repository

Status: delivered.

Expected branch: `feat/slice-3-notification-audit-repository`
Expected PR title: `feat(slice-3): notification audit repository`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-3-notification-audit-repository.md`

Goal:

Persist notification events and dispatch outcomes in user-level state.

Candidate scope:

- add SQLite migration for notification events;
- add repository create, update, get, and list operations;
- store channel, status, timestamps, references, summary, and failure reason;
- keep payload storage bounded and text-based;
- add repository tests.

Expected outcome:

- users and daemon flows have durable notification history for audit and
  troubleshooting.

### Slice 4: Notify Command Skeleton

Status: delivered.

Expected branch: `feat/slice-4-notify-command-skeleton`
Expected PR title: `feat(slice-4): notify command skeleton`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-4-notify-command-skeleton.md`

Goal:

Expose setup-check notification commands without treating manual notification as
the primary workflow.

Candidate scope:

- add `chr notify` command group;
- add `chr notify status`;
- add `chr notify telegram test --dry-run`;
- render the planned Telegram test payload from config;
- avoid network calls in dry-run mode;
- add focused CLI tests.

Expected outcome:

- users can inspect notification readiness and preview Telegram setup checks.

### Slice 5: Telegram Dispatch Boundary

Status: delivered.

Expected branch: `feat/slice-5-telegram-dispatch-boundary`
Expected PR title: `feat(slice-5): telegram dispatch boundary`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-5-telegram-dispatch-boundary.md`

Goal:

Add the first concrete channel boundary for sending notification events.

Candidate scope:

- add Telegram message renderer;
- add transport interface for sending messages;
- add HTTP sender implementation with injectable client;
- wire non-dry-run `notify telegram test`;
- record audit success or failure;
- add transport and CLI tests without requiring real credentials.

Expected outcome:

- Telegram can be used as the first external notification transport while tests
  remain deterministic.

### Slice 6: Loop and Ticket Notifications

Status: delivered.

Expected branch: `feat/slice-6-loop-ticket-notifications`
Expected PR title: `feat(slice-6): loop ticket notifications`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-6-loop-ticket-notifications.md`

Goal:

Emit notification events from the workflows that actually need user attention.

Candidate scope:

- emit events when a ticket execution is blocked;
- emit events when a loop stops on blocker or failure;
- emit events when a loop finishes with processed tickets;
- include ticket, loop, execution, and project references;
- keep dispatch optional and configuration-gated;
- add integration tests with fake notifier.

Expected outcome:

- notification is connected to loop and ticket outcomes rather than only manual
  test commands.

### Slice 7: Remote Command Intake Model

Status: delivered.

Expected branch: `feat/slice-7-remote-command-intake-model`
Expected PR title: `feat(slice-7): remote command intake model`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-7-remote-command-intake-model.md`

Goal:

Define bounded remote command requests independently from Telegram update
transport details.

Candidate scope:

- add remote command request model;
- support status, loop start, ticket start, and ticket comment intents;
- validate sender, project selector, arguments, and timestamps;
- add parser for a small text command grammar;
- reject unknown or unsafe commands clearly;
- add focused parser and validation tests.

Expected outcome:

- remote input has a durable, bounded command contract before any channel polls
  or applies commands.

### Slice 8: Telegram Remote Command Handoff

Status: delivered.

Expected branch: `feat/slice-8-telegram-remote-command-handoff`
Expected PR title: `feat(slice-8): telegram remote command handoff`
Expected report: `.prj/reports/slices/notification-and-remote-commands/slice-8-telegram-remote-command-handoff.md`

Goal:

Connect Telegram update parsing to the bounded remote command model without
creating a new execution engine.

Candidate scope:

- parse Telegram updates into remote command requests;
- persist accepted and rejected remote command records;
- map supported commands to existing CLI-equivalent handlers or explicit
  handoff records;
- add `notify telegram poll --dry-run` or equivalent bounded intake preview;
- document the remote command safety boundary;
- add tests with fixture Telegram updates.

Expected outcome:

- Telegram can serve as a remote command intake surface while local CLI
  semantics remain the source of truth.

## Completion Criteria

The slicing is complete when:

- notification events have a stable core contract;
- notification configuration is explicit and disabled by default;
- notification audit records are durable;
- Telegram setup checks can be previewed and executed with test doubles;
- loop and ticket outcomes can emit notification events;
- remote commands have a bounded model and parser;
- Telegram update intake can produce remote command requests;
- documentation describes current behavior and future email scope.

## Validation

Required validation for every implementation slice:

```sh
make verify
```

Focused tests should be added for each new model, repository, renderer,
transport, command parser, and CLI surface.
