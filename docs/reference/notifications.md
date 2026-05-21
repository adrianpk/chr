# Notifications

Notification records are channel-neutral events that describe something the user
may need to see.

The event model does not send Telegram or email messages by itself. Transport,
configuration, audit persistence, and remote command intake are separate
contracts built on top of this event shape.

Remote command request parsing is described in
[Remote commands](remote-commands.md).

## Event Fields

Required fields:

- `id`: notification event identifier.
- `project_id`: managed project identifier.
- `type`: reason the notification exists.
- `severity`: user attention level.
- `status`: notification lifecycle status.
- `source`: system component that produced the event.
- `summary`: short human-readable message.
- `created_at`: creation time.
- `updated_at`: last update time.

Optional fields:

- `detail`: longer explanation or handoff text.
- `ticket_key`: related ticket key.
- `execution_id`: related ticket execution ID.
- `loop_run_id`: related loop run ID.
- `daemon_pid`: daemon process ID when relevant.
- `failure_reason`: dispatch or processing failure reason.
- `delivered_at`: delivery time for delivered events.

## Event Types

Supported notification event types:

- `ticket_blocked`
- `validation_failed`
- `loop_finished`
- `loop_blocked`
- `loop_failed`
- `daemon_attention`
- `remote_command`
- `manual_test`

## Severity

Supported severity values:

- `info`
- `warning`
- `error`
- `critical`

Severity is channel-neutral. Channel renderers may later decide how to display
or route different severities.

## Status

Supported status values:

- `pending`
- `skipped`
- `delivered`
- `failed`

Failed notifications require a failure reason. Delivered notifications may
record `delivered_at`, and that timestamp must not be earlier than `created_at`.

## Source

Supported source values:

- `ticket_execution`
- `ticket_loop`
- `daemon`
- `remote_command`
- `manual`

The source identifies where the event came from. It is not the delivery channel.

## Reference Rules

Some event types require specific references:

- `ticket_blocked` requires `ticket_key`.
- `validation_failed` requires `execution_id`.
- `loop_finished`, `loop_blocked`, and `loop_failed` require `loop_run_id`.

Channel-specific identifiers, such as Telegram chat IDs, do not belong in the
core event model. They belong to later channel configuration, dispatch, or audit
records.

## Audit Records

Notification audit records persist notification events in user-level state.

Audit records store the channel separately from the core event. This keeps the
event model channel-neutral while allowing later dispatch code to record where a
notification was sent or attempted.

The user-level SQLite table is `notification_events`.

Stored fields include:

- event identity, project, type, severity, status, and source;
- optional channel name;
- ticket, execution, loop, and daemon references;
- summary and detail text;
- failure reason;
- creation, delivery, and update timestamps.

The repository supports:

- creating a notification audit record;
- updating a notification audit record;
- getting a record by ID;
- listing records for one project ordered by creation time.

Audit persistence does not send messages. Dispatch behavior belongs to the
channel-specific slices that follow this contract.

## Workflow Notifications

Ticket and loop workflows create notification audit records for user-visible
attention points.

Current workflow-triggered events:

- `ticket_blocked`: recorded when an applied execution result leaves a ticket
  blocked.
- `loop_finished`: recorded when a foreground ticket loop finishes after
  processing at least one ticket.
- `loop_blocked`: recorded when a foreground ticket loop stops after a blocked
  ticket result.
- `loop_failed`: recorded when a foreground ticket loop stops because ticket
  execution failed.

Workflow notifications are configuration-gated for delivery:

- when Telegram notifications are disabled, events are recorded as `skipped`;
- when Telegram is enabled and valid, events are recorded as `pending`, sent,
  then updated to `delivered` or `failed`;
- dispatch failures are recorded in the notification audit event and do not
  replace the ticket or loop result.

Ticket-blocked events include ticket and execution references. Loop events
include the loop run reference. All workflow events include the managed project
identifier.

## Notify Commands

The `notify` command group exposes setup and inspection surfaces.

Report notification configuration readiness:

```sh
chc notify status
```

Preview a Telegram setup-check payload:

```sh
chc notify telegram test --dry-run
```

`notify telegram test --dry-run` renders the planned payload from project
configuration and does not contact Telegram.

Non-dry-run Telegram dispatch belongs to the Telegram dispatch boundary.

Send a Telegram setup-check message:

```sh
chc notify telegram test
```

This command reads the bot token from the environment variable named by
`telegram_bot_token_env`, sends the rendered test message through Telegram, and
records a notification audit event with `channel` set to `telegram`.

If dispatch fails after the audit event is created, the event is updated to
`failed` with the failure reason.
