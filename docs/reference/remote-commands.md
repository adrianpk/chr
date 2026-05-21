# Remote Commands

Remote commands are bounded requests produced from external input channels.

The model is channel-neutral. Telegram update intake, persistence, and command
handoff are separate concerns. This contract only defines the request shape,
validation rules, and the small text grammar accepted by the parser.

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
loop start --status ready --priority high --max-iterations 1
ticket start CHC-12
ticket comment CHC-12 blocked on CI logs
```

Loop flags may use either separate values or `--name=value` form:

```text
loop start --status=ready --tag backend
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

Later channel-specific intake can persist accepted or rejected requests and map
accepted requests to existing local command behavior. The local CLI remains the
source of truth for actual project operations.
