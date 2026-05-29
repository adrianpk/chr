# Remote Commands

Remote commands are bounded requests produced from external input channels.

The request model is channel-neutral. Telegram update intake converts Telegram
messages into this model and stores explicit accepted or rejected handoff
records.

## Request Model

A remote command request contains:

- `id`: deterministic request identifier.
- `sender`: channel, external sender ID, and optional display name.
- `project_selector`: optional project selector from the command text.
- `intent`: one supported command intent.
- `ticket_key`: ticket key for ticket commands.
- `comment`: comment text for ticket comment commands.
- `loop`: loop filters and max iteration limit for loop commands.
- `raw_text`: original command text.
- `received_at`: time the channel received the input.
- `created_at`: time the request was created.

The sender channel and external sender ID are required. Text fields must not
contain control characters. Timestamps are required.

## Supported Intents

Supported intent values:

- `status`
- `loop_start`
- `ticket_start`
- `ticket_comment`

Unknown intents are rejected. The model does not expose arbitrary shell
commands or unrestricted CLI execution.

## Text Grammar

The optional project selector is written as the first token:

```text
@project-name status
```

Supported commands:

```text
status
@project-name status
loop start --status ready_for_dev --priority high --max-iterations 1
ticket start CHC-12
ticket comment CHC-12 blocked on CI logs
```

Loop flags may use either separate values or `--name=value` form:

```text
loop start --status=ready_for_dev --tag backend
```

Supported loop flags:

- `--status`
- `--priority`
- `--severity`
- `--project-area`
- `--tag`
- `--type`
- `--use-case`
- `--max-iterations`

Unknown root commands, unknown ticket actions, positional loop arguments, and
unknown loop flags are rejected.

## Safety Boundary

Remote commands are intentionally narrower than the local CLI. They describe
the user's requested intent, but they do not execute it directly.

Accepted requests are mapped to CLI-equivalent handoff text, such as
`chc daemon status`, `chc ticket loop`, `chc ticket start`, or
`chc ticket comment`. Rejected requests store the rejection reason.

The local CLI remains the source of truth for actual project operations.

## Telegram Fixture Intake

Telegram update fixtures use the `getUpdates` response shape:

```json
{
  "ok": true,
  "result": [
    {
      "update_id": 1,
      "message": {
        "message_id": 10,
        "from": {"id": 100, "username": "ops"},
        "chat": {"id": 200},
        "date": 1768000000,
        "text": "@example status"
      }
    }
  ]
}
```

Preview parsing without persistence:

```sh
chc notify telegram poll --fixture updates.json --dry-run
```

Persist accepted and rejected handoff records from a fixture:

```sh
chc notify telegram poll --fixture updates.json
```

This slice does not perform live Telegram polling. The fixture command provides
the bounded intake and persistence boundary used by later transport work.
