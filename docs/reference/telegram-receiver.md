# Telegram Receiver

The dedicated receiver supports authenticated project reads, retained text
conversations, configured-provider explanations and canonical artifact/native
ticket preparation. `/help` is generated from the available command registries.
Unknown slash commands return help without consulting a model. Bound actions
accept buttons, replies and `/answer` through the same retained decision owner.
Preparation creates no work invocation and executes no saved intake handoff.

## Command help

Send `/help` to read the complete command catalogue without selecting a project
or invoking a model. Commands come from their native definitions and retain
their arguments, aliases and overloads. Populated categories appear in this
order: Execution, Tickets, Slice sets, Use cases, Projects, Authoring,
Conversations, System controls and Other. Each syntax entry occupies its own
line; descriptions and large code blocks are omitted. Entries sort by exact
syntax within a category.

Help uses Telegram's legacy `Markdown` mode, with bold headings and inline
commands where safe. A definite entity-format rejection permits one plain-text
fallback. Network failures, rate limits and uncertain delivery use the normal
retained retry policy. Both renderings are frozen together when queued, so
fallback after restart cannot read a newer command catalogue. Older queued
plain replies keep their previous behavior; other responses remain plain text.

The current catalogue fits one message through the existing 3,200
UTF-16-unit response pipeline. Both forms, including any diagnostic prefix,
must fit that boundary and the sender's 3,900-code-point/UTF-16-unit bounds.
Oversized help returns an explicit diagnostic instead of a truncated catalogue.
There are no help topic pages, navigation buttons or pagination controls.

Direct `/help` remains available during capture, pending authoring questions
and confirmations, including when sent as a reply to the prompt. It does not
change the selected project, focused session, capture, revisions or question
and confirmation bindings. Its response is not a new capture/confirmation
target. Later messages remain subject to normal conversation routing.

Unknown slash commands retain their diagnostic and return the same categorized
help without a model. Existing natural-language help still passes through the
configured intent resolver; it is not guaranteed to be model-free.

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
private credential file. It accepts up to 64 KiB of literal `NAME=value` lines
and `#` comments, with unique environment names. Token replacement preserves
other assignments. Put optional model credentials, backend selection, an
absolute `CHR_CODEX_COMMAND`, or an explicit `PATH` in this same private file
for the user service. Values are not shell commands and do not expand shell
variables. Foreground model settings still use the process environment.

## Optional user service

Install the unit using the executable intended to remain installed:

```sh
chr telegram setup --user-id 100001 --operator-id maintainer --project-id example-project --install-service
```

This writes `chr-telegram.service` under the sibling `systemd/user` directory
of the user config root. `--dev-root` keeps both the unit and receiver files
inside that development environment. The unit invokes this executable by its
absolute path, preserves explicit user environment directories and reads the
optional private `telegram.env`; it does not inherit an interactive shell.
Keep the executable at that path between runs. Differing unit content requires
`--replace`. Installation never enables or starts the service, sends a message
or creates a work database.

For the normal user environment, activate it explicitly after readiness checks:

```sh
systemctl --user daemon-reload
systemctl --user enable --now chr-telegram.service
systemctl --user status chr-telegram.service
journalctl --user -u chr-telegram.service
```

A development-root unit is a fixture outside the normal service search path;
run `chr telegram serve --dev-root <root>` explicitly there. The service restarts
recoverable transport/process failures after five seconds, bounded to five
starts per minute. Invalid configuration, credentials, retained schema, exhausted
polling or ownership conflicts exit with status 78 and prevent automatic restart.
Correct the reported condition before explicitly starting it again. SIGTERM
allows one 30-second grace for owned calls; the service permits 35 seconds
before terminating its main process. It does not signal another adapter's
workers. A failed nonquiescent shutdown retains local exclusion until those
calls finish or the receiver process exits.

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

## Conversations and capture

| Command | Behavior |
| --- | --- |
| `/help`, `/projects`, `/status` | Show commands, enabled allowed projects and receiver state. |
| `/project <project>` | Select an exact authorized ID or unique registered name. |
| `/prepare [--project <id>] <topic>` | Create an independent preparation-only capture. |
| `/develop [--project <id>] <topic>` | Create a capture retaining development intent. |
| `/sessions` | Show bounded session identities, phases, focus, sizes and revisions. |
| `/focus <conversation>` | Focus an exact collecting session. |
| `/reply <conversation> <text>` | Append to an exact current conversation. |
| `/dump [conversation]` | Show bounded capture text. |
| `/done [conversation]` | Finish capture and generate a canonical dump. |
| `/quiz [conversation]` | Ask a real provider-generated clarification question. |
| `/spec [conversation]`, `/plan [conversation]` | Generate a draft spec or plan from canonical context, asking missing information when needed. |
| `/refine <dump\|spec\|plan> <conversation> <instructions>` | Explicitly replace that session's observed output. |
| `/explain [--project <id>\|--conversation <id>] <question>` | Explain canonical context without preparing or starting work. |
| `/prepare-work [conversation]` | Prepare a native backlog ticket from a canonical dump or spec, preserving the session's intent. |
| `/work [--project <id>] [--conversation <id>] [--namespace <name>] <extent> [reference]` | Preview the exact native selection and effective policy, then ask for a bound `start` decision. |
| `/cancel [conversation]` | Cancel an unstarted session and discard its capture. |

Topics contain at most 512 UTF-8 bytes. `--project` accepts one ID; use
`/project <exact registered name>` for a name containing spaces. IDs returned by
`/sessions` are opaque. Other slash commands remain unavailable until their
shared operations are implemented.

Work extents are `ticket`, `slice`, `slice-set`, `use-case` and `loose`.
Only `loose` omits its reference. The preview freezes the target commit,
membership snapshot, selected ticket revision, execution limits and publication
policy. Buttons, replies and `/answer <token> start` use the same retained
decision. Ordinary language proposes the same typed preview and cannot bypass
that decision. A changed source, policy, target or authority rejects admission;
the adapter does not silently refresh the displayed scope.

Without an explicit session, `/work` creates an independent conversation bound
to its source and resolved project/namespace. With an explicit session, it
preserves that session's binding. Admission stores the canonical invocation ID
with the confirmed request, source acknowledgement and outbox result. Another
extent cannot replace that conversation's bound invocation. Replaying the same
native request key observes its original admission before current source guards.
Admission alone establishes neither host validation nor integration. See
[Workflow Selection](workflow-selection.md) for native selection bounds.

While the foreground receiver runs, it reattaches confirmed invocations without
another message. A separate host permits at most sixteen concurrent calls and
requests one shared native checkpoint per activation. Intake remains available
during a long worker call. The common owner controls execution capacity, claims,
recovery and frozen budgets. Live owners, operator stops, occupied capacity and
human review or merger-decision waits preserve their native state. Dead owners
use the common recovery operation.

Each activation checks the original admission, current conversation binding,
numeric peer authorization, mapped operator and registered project root. The
same authority check runs again under native continuation acceptance. Changing
the selected project or `OPS_USER` does not move retained work. Shutdown cancels
and joins the receiver's own calls before closing its local store.

Conversations project canonical progress and terminal outcomes together with a
durable outbox reservation. Never-attempted routine notices coalesce to the
latest revision; new routine notices are delivered at most once per thirty
seconds per conversation. Human waits, terminal results and attempted or uncertain
sends retain their identities. A terminal result can be delivered immediately.
If outbox capacity is exhausted, the observation remains recoverable from the
native invocation. Automatic publication policy advances without another
message; human policy retains its exact candidate and review wait.

Project resolution uses an explicit authorized target, the session/reply's
bound project, peer selection, authorized configured default, then the sole
authorized enabled project. Ambiguous or conflicting targets require explicit
selection. Selection is separate for every bot/chat/topic/sender combination;
changing it preserves existing sessions. Sessions freeze project/root identity,
ops namespace and intent. Namespace uses the existing explicit/`OPS_USER`/default
resolution. Root changes require an explicit new session; existing captures
remain intact.

Plain text routes by valid Telegram reply, explicit `/reply` target, focused
collecting session, then sole collecting session. With multiple sessions and no
focus, use `/sessions` and choose explicitly. Starting another session preserves
older questions and capture. Replies to another user/topic, replaced questions,
changed conversation revisions, finished captures or pruned messages cannot
fall back to another session. Use explicit targeting after a stale prompt.

A capture accepts at most 32 parts, 16 KiB per message and 64 KiB total, including
one newline between parts. Overflow leaves existing capture unchanged. `/done`
uses the configured provider and shared authoring owner to prepare a dump or ask
for missing information. Actual prompts bind the canonical session, turn,
revision and content digest. Reply to the prompt or use `/reply` to answer it.
`/cancel` discards only the selected unstarted capture, including a pending scope
decision. At most eight nonterminal sessions per peer and 128 per bot can exist.
Session lists retain active entries and recent history within a 50-entry view;
rendered output uses the normal Unicode bounds.

Source acknowledgement, capture changes and the response outbox commit together.
Replay cannot append a part twice. A reply association becomes usable only when
a definite send acknowledgement is retained; at most 64 associations remain
active per session. Current allowlists and registry bindings are checked again
before queued input is applied.

## Provider and scope decisions

Ordinary-language proposals use the same typed operations as slash controls.
Leading slash commands bypass intent interpretation; generation and explanation
still use their configured provider. `CHR_CONVERSATION_GENERATION_BACKEND=openai`
selects OpenAI; otherwise generation uses the configured Codex runner. Each
boundary retains up to five attempts with backoff, a 60-second deadline and bounded
input/output. Restart preserves decisions, questions, answers and completed
results. Explanations retain source context independently of intent resolution.

`/prepare-work` uses actual generated dump/spec content and the shared native
record owner. A simple ticket can use its dump directly. It retains the original
instruction, source images, description, native contract and exact scope digest
before application. The result identifies a real backlog ticket and owner
receipt; it creates no invocation.

A literal `/develop` instruction can bind the first exact prepared scope. That
first reservation is retained permanently for the conversation. An ordinary-
language development proposal supplies no authority by itself. Missing authority
or a changed scope produces one expiring preview with a `develop` choice. Use
its button, reply `develop`, or send `/answer <action-token> develop`. All three
bind the same displayed scope and configured operator. Other artifact steps
require no separate development approval. Repeating an unchanged preparation
returns its retained receipt or existing prompt without creating another ticket
or extending the prompt's expiry.

Source edits, replaced prompts, changed operators and rebound project roots
cannot inherit an older decision. A confirmed native receipt reconciles a lost
acknowledgement before later source checks, preserving independent edits.
See [Authoring sessions](authoring-sessions.md) for shared contracts and bounds.

## Readiness and retained state

Status reads local files and database state without migration or network IO.
Doctor checks local project runtime files, model credential/executable presence
and the selected conversation/dump/spec/plan provider prerequisites. Missing,
disabled or unready projects are reported independently. With
`--check-connection`, doctor checks bot identity, webhook absence and both local
configuration/bot exclusions. Remote polling ownership remains unchecked until
`serve` receives an API conflict; doctor does not poll to probe another receiver.
Model connectivity remains unchecked; doctor starts no agent.
A webhook, another receiver, or changed bot identity requires local correction;
the receiver never removes webhooks or takes over polling.

Long polling defaults to 30 seconds, permits 1–50 seconds and has an HTTP
deadline ten seconds longer. Batches request 50 updates, accept at most 100,
and limit the API body to 1 MiB. Input text is at most 16 KiB. Read replies use
plain text and remain within 3,900 code points and UTF-16 units.

Intake commits a complete normalized batch and its next cursor together.
Sparse source IDs remain distinct; changed content under an existing ID blocks
intake. Rejected input retains no message body. Capacity limits are 1,024
pending updates, 1,024 unresolved deliveries, 10,000 receipt-only sources in
addition to pending input, and 64 MiB of logical payload. Capacity failure does
not advance the cursor.

Poll and delivery attempts are reserved before network calls and retained
across restart. A cycle permits five failed attempts with bounded backoff and
Telegram flood-control delays. Permanent errors and exhausted retries remain
visible and blocked. Stop the receiver, correct the reported condition, inspect
`chr telegram status`, then open exactly one local transport cycle:

```sh
chr telegram retry --poll
chr telegram retry --delivery <exact-delivery-id>
```

Choose one option per command. Status lists at most twenty unresolved delivery
IDs, prioritizing blocked/exhausted entries. Retry requires local receiver
exclusion and records the OS user ID, time and increasing cycle. Lifetime
counts, source identities, model attempts, canonical receipts and native budgets
survive. Restart the receiver explicitly afterward; retry does not start it.
Retrying an already open cycle or a delivered response fails.

Status also reports the installed build, last successful intake, cursor epoch,
retained retry timing, active conversation and queue counts, exhausted models
and possible remote retention gaps. It prints metadata rather than message,
model or credential bodies.

After more than 24 hours without successful intake, the next poll opens one
new cursor epoch with no offset and asks for the earliest available remote
update. Telegram may already have lost messages; the local inbox is preserved
and status records the gap. No negative offset or pending-update deletion is
used. Terminal bodies may compact after seven days only when no focused
conversation, unresolved request/action or unsent result depends on them.
Canonical request receipt identities remain retained. Retired source IDs are
kept as at most 1,024 closed intervals, merging only adjacent observed IDs.
Unobserved gaps never become a global high-water mark. When an interval or row
bound cannot admit more work safely, existing authority remains and new input
is refused. An uncertain send can repeat a read response;
it cannot repeat a project mutation. Configuration changes block pending output
rendered under the previous authorization digest.

`notify telegram test` and `notify telegram poll --fixture` retain their
existing notification and saved-handoff behavior.
