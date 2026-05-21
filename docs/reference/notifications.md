# Notifications

Notification records are channel-neutral events that describe something the user
may need to see.

The event model does not send Telegram or email messages by itself. Transport,
configuration, audit persistence, and remote command intake are separate
contracts built on top of this event shape.

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
