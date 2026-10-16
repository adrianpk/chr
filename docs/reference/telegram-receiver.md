# Telegram Receiver

The dedicated receiver currently supports authenticated text reads:
`/help`, `/projects` and `/status`. Unknown commands and ordinary text return
the available command list. Callbacks return an unavailable-action diagnostic.
These inputs do not admit work or execute saved remote-command handoffs.

## Private setup

```sh
chr telegram setup --user-id 100001 --operator-id maintainer --project-id example-project
chr telegram status
chr telegram doctor --check-connection
chr telegram serve
```

The first command reads the token with terminal echo disabled. For automated
setup, use `--bot-token-env EXAMPLE_BOT_TOKEN` with an existing environment
reference. Never pass the token as an argument. Setup does not contact Telegram,
send a message or start the receiver. It requires `--replace` to change differing
configuration or private credentials. Replacement cannot run while the local
receiver holds its configuration lock.

All commands accept `--dev-root`. Files use the existing user environment:
`telegram.yml` and optional `telegram.env` in the user config directory,
`core.db` in the user data directory, and locks and the verified bot binding in
the `telegram/` user state directory. New private directories use mode 0700;
configuration, credentials and the database use mode 0600. Tokens remain absent
from YAML and command output. Environment values take precedence over the
private credential file, which contains one literal `NAME=token` assignment.

## Configuration

The loader accepts the following strict YAML subset. Numeric chat/topic lists
use inline arrays. User entries start with `telegram_user_id`, followed by
`operator_id` and a block or quoted inline `project_ids` list. Unknown keys,
duplicate fields, aliases and tags are rejected.

```yaml
telegram:
  enabled: true
  bot_token_env: EXAMPLE_BOT_TOKEN
  poll_timeout_seconds: 30
  default_project_id: "example-project"
  allowed_chat_ids: [-100001]
  allowed_topic_ids: [7]
  allowed_users:
    - telegram_user_id: 100001
      operator_id: maintainer
      project_ids:
        - example-project
```

Each user requires an explicit operator and one to 50 registered project IDs.
`--project-id`, `--chat-id` and `--topic-id` are repeatable setup options.
Private chats require the actual sender's chat ID. Groups require an explicitly
allowed chat and user; topics additionally require their configured numeric ID.
An empty topic list permits non-topic messages only. Bots, anonymous senders,
channel posts, forwarded messages and unsupported content cannot enter dispatch.
Project listing returns only enabled, explicitly allowed registry entries.

## Readiness and retained state

Status reads local files and database state without migration or network IO.
Doctor checks local project runtime files, model credential/executable presence
and, with `--check-connection`, bot identity, webhook absence and receiver
exclusion. Model connectivity remains unchecked; doctor starts no agent.
A webhook, another receiver, or changed bot identity requires local correction;
the receiver never removes webhooks or takes over polling.

Long polling defaults to 30 seconds, permits 1–50 seconds and has an HTTP
deadline ten seconds longer. Batches request 50 updates, accept at most 100,
and limit the API body to 1 MiB. Input text is at most 16 KiB. Read replies use
plain text and remain within 3,900 code points and UTF-16 units.

Intake commits a complete normalized batch and its next cursor together.
Sparse source IDs remain distinct; changed content under an existing ID blocks
intake. Rejected input retains no message body. Capacity limits are 1,024
pending updates, 1,024 unresolved deliveries, 10,000 retained updates and 64 MiB
of logical payload. Capacity failure does not advance the cursor.

Poll and delivery attempts are reserved before network calls and retained
across restart. A cycle permits five failed attempts with bounded backoff and
Telegram flood-control delays. Permanent errors and exhausted retries remain
visible and blocked. Local retry-cycle controls and long-outage retirement are
delivered by the recovery slice. An uncertain send can repeat a read response;
it cannot repeat a project mutation. Configuration changes block pending output
rendered under the previous authorization digest.

`notify telegram test` and `notify telegram poll --fixture` retain their
existing notification and saved-handoff behavior.
