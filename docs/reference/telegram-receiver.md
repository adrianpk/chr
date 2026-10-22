# Telegram Receiver

The dedicated receiver supports authenticated project reads, retained text
conversations, configured-provider explanations and canonical artifact/native
ticket preparation. `/help` is generated from the available command registries.
Unknown slash commands return help without consulting a model. Bound actions
accept buttons, replies and `/answer` through the same retained decision owner.
Preparation creates no work invocation and executes no saved intake handoff.

## Command help

Send `/help` for the first command view. It keeps common project, conversation
and work commands together. Send `/help all` for the advanced group index, then
choose `/help execution`, `/help tickets`, `/help slices`, `/help cases`,
`/help projects`, `/help authoring`, `/help conversations` or `/help system`.
Commands are generated from their native definitions with complete arguments,
aliases and overloads. Unknown groups point back to `/help all`.

Entries use categorized headings and one exact syntax per line. The existing
help formatter retains legacy `Markdown` and its 3,200 UTF-16-unit hard safety
bound. Oversized help returns a local-reference diagnostic. A definite
entity-format rejection permits one frozen plain-text fallback; network errors,
rate limits and uncertain responses retain the existing transport retry policy.
Previously queued help is not regenerated from a newer command catalogue.

Help remains available during capture, pending questions and confirmations,
including when sent as a reply to their prompt. It does not change project
selection, capture, session revisions or action/question bindings. Unknown slash
commands return a correction without consulting a model. Ordinary-language help
uses the configured intent resolver.

## Message presentation and attachments

Responses compose only the relevant state, content and next-step blocks. Human
project/work names precede diagnostic references. Body text preserves literal
Markdown, code delimiters and Unicode through explicit entities. Short views
retain space for a supported next input or local inspection route.

| Item | Bound |
| --- | --- |
| Opening document excerpt | First 6 nonempty body lines, at most 600 UTF-16 units |
| Newly composed text | 1,800 UTF-16 units including visible secondary text |
| Document caption | 900 UTF-16 units including visible secondary text |
| Complete document | 2 MiB, one primary attachment per logical response |
| Retained document bytes | 64 MiB across both snapshot tables in the owning database |

Opening excerpts omit YAML front matter and routine audit fields. Dump, spec
and plan results attach the complete canonical source. Explanations and record
pages identify their complete Markdown or JSON export. Long bodies are not
split into automatic document chunks. Lists retain their existing page routes.
Legacy queued multipart replies retain their original ordering and retry state.

Asking to show, read, open or download an existing dump, specification or plan
selects its original canonical document in the resolved conversation. The
attachment preserves the complete file, including frontmatter and provenance;
only the caption is shortened. This read does not generate an explanation,
replace an artifact or start work. An ambiguous target needs clarification;
a missing document or changed prepared source cannot fall back to generated
prose. Questions about the document continue to return an identified explanation.

Document-backed confirmations send their caption, complete scope file and
existing keyboard together. Read the attachment before confirming a scope that
is shortened in the caption. Buttons describe actual effects, while original
peer, operator, project/root, namespace, revision, expiry and exact-scope guards
remain authoritative. Required-document failure offers no action. An already
successful operation keeps its original receipt if later document preparation
or delivery fails.

The original request or notification event freezes the view and attachment
before IO. Source edits, reopen and plain fallback retain those same bytes.
Missing images, future presentation versions and partial new mappings are
refused before transmission. Genuine legacy rows keep their old payload and
guards. A Telegram acknowledgement records transport delivery only. Host
validation, work completion and integration require their native evidence.
A lost acknowledgement can repeat the same notice within its retained allowance.

Changed authorization or destination bindings block transmission. Eligible
cleanup purges at most twenty terminal, unreferenced images per pass after at
least seven days. Unresolved delivery or a live confirmation retains its bytes;
hashes and original receipts survive eligible byte cleanup. Purged content
requires a fresh authorized view rather than a silent replacement file.

The configured notification sender uses the same renderer, document values and
uploader under its existing event repository. Workflow notices attach one
complete original event JSON export; manual notification tests remain short and
identify their project and test purpose. The sender retains its configured
triggers and destination and gains no interactive controls or receiver. See
[Notifications](notifications.md#workflow-notifications) for audit outcomes.

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

## Harness defaults

Select an enabled, authorized registered project with `/project <name>`, then
send `/harness` as a new message. Explicit project selection is required even
when a default project is configured. The reply shows worker and generation
defaults, up to eight configured profiles with local readiness for each role,
and up to four retained worker bindings. Local readiness does not verify remote
authentication. Inspection does not allocate a conversation or call a model.

```text
/harness
/harness worker=code generation=writing
/harness generation=-
```

Use your privately configured profile names. Omitted roles stay unchanged;
`-` restores legacy dispatch for that role. The command accepts only these two
named role fields. Executables, runtime images, credentials and model endpoint
setup remain local.

A change produces a ten-minute confirmation bound to the mapped operator,
selected project/root, exact role/profile choices, observed configuration
revision and original request. Reply `confirm` or `cancel`, use its button, or
send `/answer <token> confirm` (or `cancel`). Both changed roles publish together.
Changed project selection, root, actor mapping, access, configuration revision
or unavailable local setup rejects the old confirmation. Inspect `/harness`
and issue a new change after resolving the condition.

Changes become visible to other interfaces and new operations without restarting
the receiver. Running, paused, retried and resumed work keeps its admitted
binding, including nested generation. Retrying the same confirmation recovers
its original result; reusing its token cannot exchange choices. Lost Telegram
send acknowledgements produce a labelled notice retry, preserving the request
identity. Private setup and configuration changes cannot be authorized by a
model proposal.

## Conversations and capture

| Command | Behavior |
| --- | --- |
| `/help`, `/projects`, `/status` | Show commands, enabled allowed projects and receiver state. |
| `/project <project>` | Select an exact authorized ID or unique registered name. |
| `/prepare [--project <id>] <topic>` | Create an independent preparation-only capture. |
| `/develop [--project <id>] <topic>` | Create a capture retaining development intent. |
| `/sessions` | Show human project/topic rows, state, focus and an exact targeting command. |
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

An ordinary-language status question can refer to the most recently delivered
work message. Replying to a progress, completion or conversation-bound status
message selects that message's execution, including after it finishes. The
interpreter receives the observed execution binding separately from preparation
focus. These references permit inspection; they do not reopen capture or grant
control authority. Conflicting peers, projects, namespaces and changed roots
remain rejected.

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

The prepared independent ticket contains an executable contract using the
unchanged description and the project's `runtime.validation_command`. Its
contract deadline uses `work.validation_timeout_seconds`, or 3600 seconds when
that legacy setting is absent. A frozen invocation policy may impose a shorter
deadline. No model supplies the validation command.

`/work ticket <key>` previews that backlog ticket's readiness transition,
contract hash, validation arguments and deadline. Only the bound `start`
confirmation makes it ready, atomically with native admission. Changed records,
policy, validation input or authority reject the decision without changing the
ticket. Replay retains the original transition and invocation. Preparation-only
conversations still require this separate explicit work decision.

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
