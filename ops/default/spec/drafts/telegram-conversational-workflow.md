# Telegram Conversation and Workflow Control

Status: Draft

Date: 2026-10-16

Approval: Pending

Implementation authorization: Not granted by this document

Model: [Telegram Conversation and Workflow Model](telegram-conversational-workflow-model.md)

## 1. Purpose

An operator can discuss a change with Chryso through its own Telegram bot,
prepare the required artifacts, authorize a bounded work extent, follow its
execution and resolve questions until the change is integrated. The operator
does not need a terminal to advance an otherwise supported development cycle.

Developing Chryso through Chryso is a preferred use of this capability. It is
not mandatory and does not prevent another development workflow when needed.
The bot, credentials, configuration, service and state belong to Chryso.

This specification defines a complete text conversation and execution adapter.
It introduces no alternative ticket scheduler, agent runner or merger. Existing
project policy continues to determine execution capacity, review, validation,
publication, integration method and approval authority.

## 2. Existing Contracts and Delivery Boundary

The current implementation sends Telegram notifications and parses saved
update fixtures into remote command handoffs. Neither starts a live receiver
nor executes those handoffs. The existing foreground daemon continues already
authorized invocations; it does not authorize backlog work.

The application already has a central operation service, authoring and artifact
operations, registered projects and ticket-native work admission and controls.
Shared workflow records own execution, budgets, claims and integration evidence.
Some authoring entrypoints still need shared service coverage or a real model
continuation; their extension is part of this scope when required by the
conversation below. A fixture-only authoring path cannot satisfy live delivery.

Related contracts:

- [Ticket-native workflow](../stable/ticket-native-workflow.md)
- [Native integration queue](../stable/native-integration-queue.md)
- [Authoring sessions](../../../../docs/reference/authoring-sessions.md)
- [Workflow controls](../../../../docs/reference/workflow-controls.md)
- [Daemon coordination](../../../../docs/reference/daemon-coordination.md)
- [Remote command handoffs](../../../../docs/reference/remote-commands.md)

Existing notification and fixture commands keep their present behavior.
Historical accepted handoffs are not retrospectively authorized for execution.

## 3. Required Operator Journey

1. Configure a dedicated bot locally, authorize numeric Telegram identities
   and bind them to existing local operators and registered projects.
2. Start the receiver and receive a useful reply to help or project selection.
3. Describe a change in ordinary language or use an explicit command.
4. Collect context and answer clarification questions in that conversation.
5. Inspect or refine the resulting dump, quiz, spec, plan and ticket artifacts
   as needed. A simple ticket need not manufacture a complete artifact chain.
6. Request preparation only, or authorize development of the displayed scope.
7. Admit exactly that ticket, slice, slice set, use case or loose-ticket extent
   through the common work engine.
8. Receive progress while retaining the ability to ask questions, pause,
   cancel, resume, answer a decision or grant a bounded allowance.
9. Follow review and integration according to the project's existing policy.
10. Receive an outcome that identifies actual integrated work, remaining work,
    blockers and evidence. Worker success alone is not completion.

A single explicit instruction to develop a sufficiently defined change may
authorize its bounded preparation and execution. The request must retain that
intent and its scope. Requests to explain, draft, prepare or inspect never
implicitly authorize implementation. Additional confirmation is needed for an
ambiguous target, changed scope, stale approval or authority the request lacks.
Do not add approval at every internal artifact step.

## 4. Architecture and Ownership

| Concern | Owner and contract |
| --- | --- |
| Process assembly | Command entrypoint assembles configuration, transport, store, service and lifecycle. |
| Telegram IO | HTTP adapter receives updates and sends messages; it owns API formatting and transport errors. |
| Conversation | Adapter maps peer messages, replies, focus and prompts to typed application requests. |
| Intent resolution | Bounded resolver proposes one registered operation with structured arguments; deterministic validation decides whether it is executable. |
| Application operations | Existing central service owns project, authoring, artifact and work operations. Extend that service where an adapter-neutral operation is missing. |
| Durable conversation state | Core models and SQLite repository own inbox, peer/session bindings, requests, confirmations and delivery bookkeeping. |
| Project mutations | Existing project owners retain canonical artifacts and replay evidence under their mutation exclusion. |
| Execution and integration | Existing workflow owner retains worktrees, capacity, attempts, gates, source/base/result guards, review and receipts. |

Adapters do not invoke CLI text through a shell. They call typed service
operations. A declarative operation registry supplies slash syntax, natural
language argument schemas, effects, required state and help. Each descriptor
has one operation ID; both input forms reach the same handler.

Transport and conversation records may reference workflow identities and
observed revisions. They do not copy editable membership, execution leases,
merge queues or budget counters. SQLite conversation state cannot override
canonical project workflow state.

## 5. Installation, Configuration and Service

Proposed commands:

~~~sh
chr telegram setup --user-id 100001 --operator-id maintainer --project-id example-project
chr telegram status
chr telegram doctor --check-connection
chr telegram retry --poll
chr telegram serve
~~~

Setup asks for the bot token through masked terminal input, or reads the named
environment variable. It preserves existing configuration and requires an
explicit replacement option for differing values. It writes private user-local
configuration and can install a user service only with an explicit
--install-service option. Setup does not start work or send a test message.

Default files follow the existing user environment, including --dev-root:

| File | Purpose |
| --- | --- |
| User config directory / telegram.yml | Bot token variable name, authorized identities and project bindings. |
| User config directory / telegram.env | Optional private token and model credential environment file. |
| User data directory / core.db | Existing user database, extended with transport and conversation tables. |
| User state directory / telegram/ | Singleton lock and bounded sanitized runtime diagnostics. |
| User systemd directory / chr-telegram.service | Optional foreground receiver service. |

Example proposed configuration:

~~~yaml
telegram:
  enabled: true
  bot_token_env: CHR_TELEGRAM_BOT_TOKEN
  poll_timeout_seconds: 30
  default_project_id: example-project
  allowed_chat_ids: []
  allowed_users:
    - telegram_user_id: 100001
      operator_id: maintainer
      project_ids:
        - example-project
~~~

Project IDs resolve through the existing registry; this file does not own a
second registry. Each authorized user has an explicit operator mapping and a
non-empty project allowlist. Empty chat allowlists permit private chats only;
groups require an explicitly allowed chat and an individually allowed sender.
Anonymous senders, channel posts and bot senders cannot authorize operations.

Tokens are absent from YAML, argv, project artifacts and logs. Private files
use mode 0600 and private directories 0700. Service credentials come from its
environment file or environment references; an interactive shell is not assumed.
Existing model/provider selection remains project-owned.

Status is local and read-only. Doctor also reports existing project runtime
and configured model readiness without starting an agent. A project readiness
failure blocks that project's work, not another project's status or the bot's
help. Connection doctor and serve can use getMe and
getWebhookInfo without sending a message. A configured webhook or another
receiver is a visible startup conflict. Do not delete a webhook, drop pending
updates or take over another receiver automatically.

The local retry command opens an attributable new bounded transport retry cycle
for polling, or for one delivery selected with --delivery <id>. It retains
lifetime attempt history and never resets domain attempts or token budgets.
Service restart alone cannot clear an exhausted transport retry cycle.

The optional service invokes the installed executable with telegram serve and
restarts only for recoverable transport/process failure. Invalid configuration,
invalid credentials and polling ownership conflicts require correction.
Installation and service control operate under the normal user account.

## 6. Input, Identity and Project Resolution

The live transport admits message and callback_query updates. Other delivered
types are recorded as ignored metadata and cannot mutate project state.
Edited messages, forwarded content, documents, voice and images are not treated
as fresh development instructions in this text delivery.

Authenticate the actual Telegram sender, chat, topic and verified bot identity
before interpretation, model calls, project lookup replies or state changes.
Do not authorize by username, display name, forwarded identity or quoted text.
Recheck current configuration before applying a queued mutation.

The peer key is bot ID, chat ID, topic ID and sender user ID. Project selection:

1. An explicit authorized project ID or exact unique registered name.
2. The project bound to an explicit conversation or valid reply association.
3. The peer's selected project.
4. The configured default, if authorized for this peer.
5. The only project authorized for this peer.

A mismatch between an explicit target and a bound reply asks for clarification.
Missing or ambiguous selection never chooses the first registry entry.
Fuzzy names may produce suggestions; a mutation requires an exact resolved ID.
Disabled or removed projects cannot execute queued work.

A conversation freezes its project identity, namespace and intent. Switching
the peer's active project does not retarget that conversation. Registry root
changes invalidate pending mutation bindings until explicitly resolved.
Telegram identity, operator identity, namespace and Git identity remain distinct.

## 7. Natural Language and Explicit Controls

Ordinary messages are the normal interaction path. A leading slash uses the
deterministic command parser and bypasses semantic intent classification.
Unknown commands return help without consulting a model to invent a command.

The resolver receives bounded current-message context, the selected project,
active conversation references, current prompt and allowed operation schemas.
It returns operation ID, typed arguments, unresolved fields and a concise
explanation. It cannot return executable code, tool names, process identities,
ownership credentials or an arbitrary shell command.

Validate the operation, arguments, authority, current phase and target outside
the model. An unsupported or ambiguous decision asks one focused question.
No model availability means explicit commands remain available; already
authorized work does not become failed merely because intent resolution waits.

| Operation family | Proposed explicit surface | Required behavior |
| --- | --- | --- |
| Help and projects | /help, /projects, /project <project> | Bounded reads and peer-local selection. |
| Explain | /explain <question> | Read project context and answer without admitting work. |
| Prepare or develop | /prepare <topic>, /develop <topic> | Create a conversation with preparation-only or development intent. |
| Drafts and questions | /dump, /quiz, /spec, /plan, /done | Collect, review and generate through shared authoring operations; /done finishes the current capture only. |
| Records | /tickets, /ticket, /slices, /slice-sets, /use-cases | List/show and explicitly requested create/update through shared owners with observed guards. |
| Work admission | /work <extent> <reference> | Show and admit the exact native selection and effective policy. |
| Observe and continue | /status, /continue <execution> | Read exact state or continue retained authorization. |
| Stop controls | /pause, /cancel, /resume | Address an explicit invocation or the unambiguous bound one. |
| Bound decisions | /approve, /grant, /answer, /reconcile | Use the exact displayed action binding and shared control contract. |
| Conversation targeting | /sessions, /focus <conversation>, /reply <conversation> <text> | Preserve independent sessions and provide a deterministic fallback to Telegram reply routing. |

Command argument details are generated from the registry. Legacy fixture text
grammar remains a separate import surface; it cannot bypass live authorization.
Records use their actual lifecycle and ownership rules, including typed request
IDs for retry-safe creation and mutation. An unsupported service operation is
reported accurately until its shared implementation is delivered.

## 8. Conversation Lifecycle and Authoring

The adapter phases are collecting, clarifying, preparing, prepared,
awaiting_confirmation, executing, waiting, completed, canceled and failed.
Executing and waiting are projections of a bound invocation; they cannot
advance its canonical status. Cancellation of an unstarted draft discards that
conversation's pending capture. Cancellation after work admission delegates to
the workflow owner and waits for its actual result.

Plain text follows a valid reply association first, then explicit conversation
targeting, then the focused eligible conversation, then the sole eligible one.
Multiple eligible sessions without focus produce a selection question.
Completed, expired or replaced prompts cannot accept a new mutation.

Persist bot reply associations before treating them as usable routing evidence.
Question answers bind the authoring session, question/turn identity and revision.
One user's answer cannot resolve another user's prompt in a shared chat.
Starting another conversation preserves earlier sessions and pending questions.

Capture supports multiple bounded text parts, explicit completion and cancel.
The same update cannot append a part twice. Generated artifacts use the current
ops namespace, normal provenance and shared writers. An existing artifact is
not replaced without its normal observed revision/hash and replacement authority.
Model context references canonical artifacts; a hidden chat transcript does
not become authoritative project state.

Preparation-only requests stop at their requested artifact or prepared work.
Development requests retain the accepted requirements and bounded scope,
materialize the approved native work and continue that exact extent under its
effective policy. Clarification may refine that scope; expansion beyond the
retained authorization requires a new explicit decision.

## 9. Work, Review and Parallel Execution

Use PreviewSelectedWork, StartBoundWorkExtent, WorkflowStatus and
ControlWorkExtent through the central service. Extend their adapter-neutral
contracts for request identity where required; do not replace them.

Admission retains the real selector, snapshot, selected record revisions,
target, membership policy and effective execution/review/validation policy.
A changed preview rejects rather than silently authorizing a refreshed extent.
Automatic review/integration policy continues without per-PR Telegram approval.
Human review policy waits on an exact candidate binding.

The default project capacity remains one active worker. Explicit project
capacity up to the currently supported eight remains available. Telegram,
daemon, CLI, TUI and Web compete through the same repository exclusion, lane
claims and target publication queue. Several conversations do not imply
several agents or authorize running a single extent twice.

The foreground receiver hosts bounded asynchronous calls to the existing
continuation operation. It can continue eligible retained invocations without
another user message. It uses one checkpoint per activation and preserves
operator pauses, cancellations, approval waits, requirements and budget limits.
Other adapters can observe or control the same invocation.

While an activation is running, incoming status and stop/decision requests
remain responsive. Duplicate activations attach or wait through shared
ownership. A local transport worker registry tracks only in-flight calls.
It is not a durable task queue or execution authority.

Approval, grants and decision answers bind the displayed invocation revision,
candidate and review binding, integration entry, budget revision and decision
revision/hash where required. Actor identity comes from the local authorized
mapping. Stale or expired actions show current state and require a new decision;
never refresh a failed approval and apply it automatically.

## 10. Durable Intake, Replay and Delivery

Before advancing the Telegram polling offset, atomically store the complete
received batch or explicit ignored/rejected metadata and its next cursor.
Do not assume update IDs are consecutive. Inbox capacity or persistence failure
prevents acknowledgement of unretained input.

Uniqueness of bot ID and update ID protects intake. A different payload for an
existing source identity is a conflict. Independent requests have stable opaque
IDs and normalized argument digests. Replaying a request with changed arguments
fails. Replaying a completed request returns its original result.

Use the existing workflow RequestKey for admission. Other mutations require
equivalent replay handling in their shared project owner: reserve target IDs
before writing, journal intended changes under mutation exclusion and retain
the result reference. A crash after a project write but before the SQLite
acknowledgement reconciles that same request; it cannot create a second ticket,
session, grant, approval or execution. Delivery receipts do not prove mutation
receipts.

Transport result delivery uses a durable outbox. A definite successful
send records Telegram message IDs and reply bindings. Ambiguous sends may
duplicate a notification after a bounded retry; each copy carries the same
conversation/request reference and refers to the same single-use action.
No claim of exactly-once Telegram message delivery is made.

Persist retry schedule and attempt count. Respect retry_after for flood
control. Timeouts and transient server failures use bounded backoff; credential,
authorization and malformed-request failures require correction.
Notification failure does not roll back confirmed work or rerun it.

After a polling outage longer than the remote retention window, preserve local
pending records and establish a new cursor epoch from the earliest available
remote update without dropping pending updates. Report possible missing remote
messages. Deduplication still checks retained source identities and digests.

## 11. Recovery, Shutdown and Self-development

Startup validates schema, configuration, credentials, bot identity, webhook
state and the singleton receiver lock before reporting readiness. A bot
identity change requires explicit rebinding; it never inherits another bot's
peers, cursor, approvals or conversations. The same bot cannot have two local
receivers. A remote polling conflict stops receipt without forcing takeover.

Recover unfinished application requests from their shared receipts before
replaying effects. Clear only transport processing claims proven to belong to
a dead local owner. Re-read canonical invocation state and use existing shared
reconciliation for dead workflow ownership. Do not reset budgets, validation
evidence, retry counts, review bindings or operator stop intent.

SIGTERM stops intake, checkpoints conversation/outbox work and shuts down only
owned activations through existing process supervision. It does not cancel
another adapter's worker or revoke retained work authorization. Bounded
shutdown must report a failure if quiescence cannot be established.

Chryso can be registered like any other project. The installed receiver stays
fixed to its running build while an agent changes repository source. Candidate
builds are validated separately; updating the coordinating service occurs
between active executions through an explicit installation/service action.
The user can always inspect or recover retained work through CLI/Web.

## 12. Proposed Bounds and Observable State

These defaults are part of the draft decision set and require approval with
the specification. Configuration may reduce them; documented hard ceilings
cannot silently grow to accommodate a request.

| Resource | Default and hard boundary |
| --- | --- |
| Long poll | 30 seconds; configurable 1-50; HTTP deadline poll + 10 seconds. |
| Update batch / API body | 50 updates, maximum 100; response maximum 1 MiB. |
| Input / staged text | 16 KiB per message; 64 KiB and 32 parts per capture. |
| Live conversations | 8 per peer and 128 per bot. |
| In-flight interpretation | At most 2 model calls per receiver; 60-second deadline. |
| Application activations | At most 16 per receiver; project worker limits remain domain-owned. |
| Intake reserve | At most 1,024 pending updates; control/status input has a reserved 32-request processing allowance. |
| Retained transport data | 64 MiB logical payload ceiling, 10,000 receipt-only source rows and 1,024 exact retired-ID intervals; terminal payloads compact after 7 days. |
| Pending outbound messages | 1,024 per bot; progress notices coalesce by conversation and observed revision. |
| Network retry | At most 5 attempts for one retained delivery/poll failure; backoff 1, 2, 4, 8, 16 seconds unless retry_after is longer. |
| Poll failure exhaustion | Degraded receive state; explicit local telegram retry opens a bounded new cycle without replaying confirmed operations. |
| Confirmation | Expires after 24 hours or immediately when its observed guard changes. |
| Output | Plain text, chunks at most 3,900 Unicode code points and 3,900 UTF-16 units; at most 8 chunks per response, then a summary and a paged artifact reference. |
| Lists / diagnostic tail | Default 20, maximum 50 rows; at most 100 diagnostic events. |
| Progress / shutdown | At most one routine progress notice per conversation per 30 seconds; 30-second shutdown grace. |

Retention never deletes unresolved requests, live confirmations, active
conversation references or unsent terminal results. At capacity, reject new
authoring/admission before accepting an effect and return a useful capacity
response. Status, stop and reconciliation remain prioritized; if persistent
storage itself is unavailable, fail visibly without claiming command acceptance.
Completed-source compaction cannot make an old update executable again.

Status reports installed build, receiver health, bot ID, poll cursor epoch,
last successful intake, retry timing, active conversations, pending delivery
and sanitized failures. Per-conversation output identifies the project,
intent, current question or invocation, validated tasks, confirmed tickets,
integrated units, actual blocker and actionable next step. It separates model,
transport, validation and domain failures.

## 13. Acceptance Criteria

| ID | Required observable result |
| --- | --- |
| TG-01 | A dedicated configured bot receives a live authorized text message and returns the correct project-bound response. |
| TG-02 | Setup preserves existing configuration, masks token entry and keeps secrets out of argv, logs and project artifacts. |
| TG-03 | Unauthorized, anonymous, forwarded-identity and disallowed-chat inputs cannot call a model or mutate state. |
| TG-04 | Explicit commands and equivalent ordinary-language requests reach the same typed operation and authority checks. |
| TG-05 | An unsupported operation, ambiguous project or invalid model payload asks for correction without executing a guessed action. |
| TG-06 | Project switching and concurrent peers do not retarget an existing conversation or leak its context. |
| TG-07 | Multiple sessions for one peer retain independent questions; replies, focus and explicit targeting resolve the right session. |
| TG-08 | Duplicate updates cannot append capture text twice; overflow and cancellation preserve existing artifacts. |
| TG-09 | Preparation can produce and refine actual artifacts using the configured model provider without fake question/summary inputs. |
| TG-10 | A preparation-only instruction produces no work invocation; a bounded development instruction retains its explicit authorization. |
| TG-11 | Ticket, slice, slice-set, use-case and loose extents use shared native previews and admit only the accepted membership. |
| TG-12 | Replaying admission attaches to the same workflow RequestKey; changed input under that key fails. |
| TG-13 | An authorized extent advances while the operator sends no further messages, including the automatic review/integration path. |
| TG-14 | Human review waits on the exact displayed candidate; a stale button or text approval cannot approve a successor. |
| TG-15 | Pause, cancellation, resume, grants and decision answers retain the same meaning and guards as CLI/Web. |
| TG-16 | Serial default and configured parallel capacity share limits with another CLI/Web/daemon caller and prevent duplicate ownership. |
| TG-17 | A long-running worker or unavailable model does not block explicit status and stop controls. |
| TG-18 | A crash after inbox persistence, after project mutation and before response delivery recovers without a duplicated effect. |
| TG-19 | Offset advancement never acknowledges unretained input; duplicate, sparse and restarted streams retain their semantics. |
| TG-20 | Uncertain message delivery can repeat a notice but cannot repeat its mutation or action authorization. |
| TG-21 | Polling, model and delivery retry budgets survive restart; exhausted failure is visible and never becomes a false completion. |
| TG-22 | Receiver restart preserves prepared work, explicit pauses, active validation ownership, waits and consumed budgets. |
| TG-23 | Token replacement by another bot, an existing webhook and another receiver fail with actionable diagnostics and no takeover. |
| TG-24 | Bounded retention preserves unresolved records and prevents retired updates from becoming new commands. |
| TG-25 | Replies correctly handle Unicode, message splitting, group topic isolation and references to replaced prompts. |
| TG-26 | Final output distinguishes confirmed integration from worker exit, validation success or an open PR. |
| TG-27 | A small real change can be requested, clarified, authorized and integrated from Telegram using the normal Chryso engine. |
| TG-28 | Chryso itself can use that journey without replacing the running receiver mid-execution or depending on an external development controller. |
| TG-29 | Existing notification and fixture-import commands keep their documented behavior and do not execute old handoffs. |
| TG-30 | CLI/TUI/Web can observe and control the Telegram-created invocation using the same canonical references. |
| TG-31 | The selected transport scope and proposed bounds are documented by installed help and connection diagnostics. |
| TG-32 | The first authorized live demonstration records exact build, request, invocation and integration evidence separately from fixture checks. |

## 14. Validation and Decisions for Approval

Use deterministic HTTP fixtures and real temporary SQLite/project stores for
routing, authorization, replay, crash windows, guarded actions and recovery.
Use targeted race checks for shared ownership and receiver shutdown. Keep
interactive validation focused; full/race-wide campaigns belong to the existing
nightly. Live bot or forge actions require the designated explicit authorization
and report their evidence separately.

The proposed delivery is text-first, supports private chats and explicitly
authorized group topics, uses long polling and the existing registered-project
store, and includes the complete preparation-to-integration journey. Approve
these boundaries and the numeric defaults before planning implementation.
Rich media/transcription, webhook hosting and remote shell access are outside
this text adapter.

No plan, tracker, ticket batch or service activation is created by this draft.
Approval of the document and authorization to implement remain separate states.

Transport facts are checked against the Telegram Bot API: offset acknowledgement
and long polling, exclusive webhook delivery, a remote update retention window
of at most 24 hours, message text limits, callback payload limits and retry_after.
Reference: https://core.telegram.org/bots/api
