# Telegram Conversation and Workflow Model

Status: Approved

Date: 2026-10-16

Approved: maintainer, 2026-10-16; model accepted with the behavior specification for planning.

Specification: [Telegram Conversation and Workflow Control](telegram-conversational-workflow.md)

## 1. Authority and Storage

Extend the existing user-level core.db with additive migrations. User-local
SQLite owns Telegram transport and conversation bookkeeping. The current
project registry remains the only registry; project Markdown and workflow
records remain authoritative for artifacts, selection, execution and delivery.

Bot tokens and provider credentials are environment-owned and never stored in
these tables. Numeric Telegram identities stay in private user state. Project
receipts contain opaque request IDs and configured operator identities rather
than chat IDs, sender IDs, bot tokens or conversation transcripts.

Every table below has a schema version where an encoded payload needs one.
All IDs and external integers use validated representations: Telegram numeric
IDs use signed 64-bit storage, timestamps are UTC and digests are lowercase
SHA-256 hex. Opaque local IDs have at most 128 bytes. JSON uses a strict typed
decoder with bounded size and known schema; arbitrary model fields are rejected.

## 2. Bot Receiver

Business entity: TelegramReceiver. Stored table: telegram_receivers.

| Field | Stored representation and constraint |
| --- | --- |
| bot_id | Integer primary key; identity returned by authenticated getMe. |
| schema_version | Positive integer, initially 1. |
| config_digest | Digest of effective non-secret authorization/configuration. |
| cursor_epoch | Positive integer; increases when a stale remote cursor is deliberately replaced under the recovery contract. |
| next_update_id | Nullable non-negative integer; null requests the earliest available unconfirmed update. |
| retired_update_ranges | At most 1,024 exact closed intervals of retired source IDs; gaps are never merged. |
| health | starting, ready, degraded, stopped or blocked. |
| last_poll_at | Nullable UTC timestamp of successful intake. |
| retry_attempt | Integer 0-5 for the current retained poll failure cycle. |
| retry_at | Nullable UTC retry deadline. |
| retry_cycle, total_retry_attempts | Monotonic local cycle identity and lifetime count; explicit local retry starts a new cycle. |
| last_error_code | Bounded sanitized error classification, not a raw HTTP URL. |
| updated_at | UTC timestamp. |

The process holds an OS-backed lock scoped to this bot under the user state
directory. A PID or heartbeat is diagnostic information, not exclusive
ownership proof. Acquire the lock before polling. A bot identity change cannot
reuse the previous row as if it were the same receiver.

Exhaustion survives automatic restart. An explicit local retry records its
actor/time, advances retry_cycle and resets only that cycle's 0-5 counter.
The next cursor changes in the same transaction that retains a complete batch.
Sparse IDs do not imply missing rows. After an outage longer than 24 hours,
start a new cursor epoch with null offset, preserve the inbox and report the
remote retention gap. Never use a negative offset or drop_pending_updates to
perform this recovery. Retained source IDs still deduplicate across epochs.

## 3. Peer and Conversation

Business entity: TelegramPeer. Stored table: telegram_peers.

| Field | Stored representation and constraint |
| --- | --- |
| bot_id, chat_id, topic_id, user_id | Composite primary key; topic_id is 0 for a non-topic chat. |
| selected_project_id | Nullable existing registry project ID. |
| focused_conversation_id | Nullable conversation ID owned by this exact peer. |
| revision | Positive integer used for compare-and-swap updates. |
| updated_at | UTC timestamp. |

The current config maps user_id to operator identity and allowed projects.
That mapping is rechecked, not copied into a permanent privilege grant.
Selecting a project or focus changes only this peer row.

Business entity: TelegramConversation. Stored table: telegram_conversations.

| Field | Stored representation and constraint |
| --- | --- |
| id | Opaque primary key. |
| bot_id, chat_id, topic_id, user_id | Peer foreign key. |
| project_id | Existing registry identity, immutable after binding. |
| project_binding_digest | Observed registry/root identity; a root change invalidates pending writes. |
| namespace | Resolved ops namespace, frozen for this conversation. |
| intent | explain, prepare or develop; changing to a wider intent requires explicit authorization. |
| phase | collecting, clarifying, preparing, prepared, awaiting_confirmation, executing, waiting, completed, canceled or failed. |
| topic | Human description, at most 512 UTF-8 bytes. |
| authoring_session_ref | Nullable canonical authoring session reference. |
| invocation_id | Nullable canonical invocation reference; observation of another run does not replace it. |
| question_binding | Nullable typed question/turn ID, revision and content digest. |
| capture | Nullable bounded text and ordered source-part bindings. |
| authorization_ref | Nullable request/decision that explicitly authorizes the retained development scope. |
| last_observed_revision | Nullable revision of the bound canonical invocation. |
| revision | Positive integer; mutation requires the observed value. |
| created_at, updated_at, finished_at | UTC timestamps; finished_at only for terminal phases. |

Capture is at most 64 KiB and 32 parts. Each part records its inbox source ID,
text digest and byte range; an inbox replay cannot append a second copy.
Canonical artifact creation uses the shared writer, not the conversation row.
Terminal conversation payload can compact only after its requests and outbox
are terminal. Preserve a concise receipt and canonical references.

Indexes support peer/phase/updated_at, project/phase and invocation lookups.
Enforce at most eight non-terminal conversations per peer and 128 per bot in
the allocation transaction. Multiple sessions for the same project are valid.

## 4. Durable Update Inbox

Business entity: TelegramUpdateReceipt. Stored table: telegram_updates.

| Field | Stored representation and constraint |
| --- | --- |
| bot_id, update_id | Composite primary key, stable across cursor epochs. |
| cursor_epoch | Epoch under which this update was observed. |
| payload_digest | Digest of the normalized received update. |
| update_kind | message, callback_query, ignored or rejected. |
| chat_id, topic_id, user_id, message_id | Nullable routing metadata; absent for unsupported update kinds. |
| received_at | Local UTC receipt timestamp. |
| payload | Nullable bounded normalized text/callback data; raw API responses are not retained. |
| state | pending, processing, processed, ignored or rejected. |
| request_id | Nullable resulting application request. |
| reason_code | Nullable bounded rejection/ignore reason. |
| interpretation_attempt, interpretation_retry_at | Retained model-resolution attempt count 0-5 and nullable UTC retry deadline; reserved before each call. |
| interpretation_error_code | Nullable bounded sanitized model failure classification. |
| processing_owner | Nullable private local processing claim. |
| updated_at | UTC timestamp. |

Insert a fetched batch and update the receiver cursor atomically. Duplicate
source with the same digest reuses its receipt. Different content under the
same identity blocks interpretation. Disallowed senders retain only minimal
receipt metadata, never their message body or model context.

Processing claims serialize this peer's conversation transition. Claims expire
only after proving their owner is gone; a wall-clock timeout cannot authorize
a second concurrently applying operation.

Indexes: bot/state/received_at for bounded drain; peer/received_at for ordered
conversation processing. Pending row capacity is 1,024. Intake, interpretation
and application capacity are checked independently.

Interpretation retries retain their count and deadline across restart. Once a
typed request exists, replay observes that request instead of asking the model
to choose another operation. Exhaustion rejects resolution with a visible
provider failure and preserves the conversation; replay of the same source
cannot reset it. A new explicit user instruction has a new source identity and
still passes current scope and authority checks. Artifact-generation retries
remain owned by the shared authoring operation and its retained request, not
by an independent transport attempt budget.

## 5. Application Request and Domain Replay

Business entity: RemoteOperationRequest. Stored table: telegram_requests.

| Field | Stored representation and constraint |
| --- | --- |
| id | Stable opaque primary key; derived/reserved once from source plus operation ordinal. |
| bot_id, update_id | Inbox foreign key; one input may emit a bounded ordered set of operations. |
| operation_ordinal | Non-negative integer, maximum 7; unique with bot_id/update_id. |
| conversation_id | Nullable local conversation reference. |
| operator_id | Configured actor, never a model-provided identity. |
| project_id, namespace | Frozen target identity. |
| operation_id | One declarative registry operation. |
| arguments | Strict typed JSON, maximum 64 KiB. |
| argument_digest | Digest including operation ID, normalized arguments and target. |
| expected_conversation_revision | Nullable observed conversation revision. |
| domain_guard | Typed snapshot/record/question/review/integration/budget bindings. |
| state | prepared, awaiting_confirmation, ready, applying, confirmed, rejected, blocked or reconcile_required. |
| domain_receipt_ref | Nullable shared owner receipt or invocation RequestKey/result reference. |
| result | Nullable bounded original result facts and canonical artifact references. |
| created_at, updated_at | UTC timestamps. |

The normal interpretation emits one operation. A retained preparation sequence
may derive up to eight ordered operations from the same input. Each step is
reserved before dispatch and requires its own satisfied predecessor receipt.
Expansion beyond that bound continues as a new durable preparation step under
the same explicitly retained conversation scope.

Confirmed and rejected states are terminal for their exact payload. A new user
request uses a new ID. Retrying an ID with changed arguments is always a
conflict, including after a failed attempt.

Project operation replay is enforced by the shared domain owner:

1. Use the existing WorkflowStartRequest.RequestKey for work admission.
2. For other mutations, add a caller-neutral request binding to the existing
   owner. Reserve canonical artifact IDs before any creation or append.
3. Journal expected before images and intended after images under that owner's
   existing mutation lock; commit its result receipt with the change.
4. Recovery observes or completes that same journal before returning a result.
   A divergent image blocks; it does not infer success or start a new write.
5. SQLite records the confirmed result after shared-owner confirmation. A crash
   in between repeats observation of the same receipt, not the effect.

The domain binding contains schema, opaque request ID, operation/argument
digest, operator, target references, observed revisions, state and result
references. It follows the existing owner's project-local record/journal layout
and bounds. It contains no Telegram routing metadata and is reusable by other
adapters. Do not hold a SQLite transaction over a model call, agent run or
forge operation.

Indexes: bot/state/updated_at; conversation/created_at; project/operation_id.
Admission request keys, grant IDs and decision-answer record IDs remain stable
across transport restarts. Existing workflow receipts remain canonical.

## 6. Confirmation and Reply Binding

Business entity: TelegramActionBinding. Stored table: telegram_actions.

| Field | Stored representation and constraint |
| --- | --- |
| token | Random opaque 128-bit identifier; primary key, encoded within callback byte limits. |
| request_id | Exact request foreign key. |
| bot_id, chat_id, topic_id, user_id | Exact authorized peer. |
| conversation_revision | Revision at which the prompt was created. |
| guard_digest | Digest of the complete retained target/candidate/decision guard. |
| allowed_choices | Bounded typed choices defined by the operation. |
| state | pending, applying, consumed, expired or superseded. |
| expires_at | UTC timestamp, at most 24 hours from creation. |
| consumed_by_request | Nullable unique answer request binding. |
| created_at, updated_at | UTC timestamps. |

Callbacks contain only the opaque token and a bounded choice identifier.
Text answers resolve the same binding through their valid reply or explicit
target. They cannot submit a replacement actor, candidate or guard.
Authorization, expiry and domain revisions are checked again at application.

Claiming a pending action and creating its answer request is one transaction.
An applying action recovers through that same request's domain receipt. Only
a confirmed result consumes it. Stale or denied results supersede it and
require a new observed prompt; they never automatically approve a fresh image.

Business entity: TelegramReplyAssociation. Stored table: telegram_replies.

Fields: bot_id, chat_id, topic_id, message_id, user_id, conversation_id,
request_id, purpose, question/action binding, created_at and invalidated_at.
Primary key is bot_id/chat_id/topic_id/message_id/user_id.

A reply to a different peer, replaced question, terminal capture or superseded
approval fails without choosing a different session. Retain at most 64 active
message associations per conversation. Pruning invalidates association-based
answers; explicit current conversation targeting remains available.

## 7. Durable Outbox

Business entity: TelegramDelivery. Stored table: telegram_outbox.

| Field | Stored representation and constraint |
| --- | --- |
| id | Stable opaque primary key. |
| bot_id, chat_id, topic_id, user_id | Destination peer. |
| conversation_id, request_id | Nullable source references. |
| event_key | Stable domain revision/event or request-result identity. |
| kind | acknowledgement, question, progress, result or diagnostic. |
| part_index, part_count | Ordered chunk coordinates, 1-8. |
| text | Bounded rendered plain text under the specification's character limits. |
| action_token | Nullable existing action binding. |
| state | pending, sending, delivered, uncertain or blocked. |
| attempts, retry_at | Persisted 0-5 count and nullable UTC deadline. |
| retry_cycle, total_attempts | Local retry cycle and lifetime attempt count, preserved across restart. |
| telegram_message_id | Nullable definite successful API result. |
| last_error_code | Bounded sanitized classification. |
| created_at, updated_at | UTC timestamps. |

Unique key: bot_id/destination peer/event_key/part_index. Retry reuses the
same delivery, event key and action token. On definite success, write delivery
and reply association together. An uncertain send retains its uncertainty;
retry may repeat a notice but cannot generate another application request.

Progress for an unobserved intermediate revision may coalesce into the latest
state. Do not coalesce away questions, operator decisions or terminal results.
Delivered routine progress payloads can compact when no active reply binding
references them; keep the current observation and bounded reply associations.
Capacity is 1,024 pending deliveries. Retention and send failures cannot turn
an unconfirmed operation into a completed one.

Indexes: bot/state/retry_at for drain; conversation/created_at for inspection.
No independent durable execution subscription is needed: conversations bind
canonical invocations, and event_key deduplicates observed notices.

## 8. State Transitions and Restart

| Boundary | Durable transition |
| --- | --- |
| API batch received | Inbox rows plus next cursor in one transaction. |
| Input interpreted | Typed request and peer/conversation transition guarded by revision. |
| Confirmation required | Pending action plus question outbox, without applying the mutation. |
| User answers | Claim action and reserve one answer request atomically. |
| Domain application begins | Stable request binding; shared owner reserves/journals its effect. |
| Domain result confirmed | Original result references retained; outbox reserved without replaying effects. |
| Telegram send confirmed | Delivery state and reply associations retained atomically. |
| Service restart | Recover applying requests/actions, observe canonical execution, resume due delivery within retained budgets. |
| Work terminal | Project owner confirms outcome; conversation projects it and reserves final notification. |

Transport tables have no authority to mark a workflow integrated, clear an
operator stop, change membership or grant additional attempts. Work state is
always reread through the existing service. Configuration revocation blocks
queued mutations and delivery to revoked peers.

## 9. Retention, Queries and Migration

Required bounded queries:

- Receiver readiness, last intake and next retry for one bot.
- Authorized peer's focused and non-terminal conversations.
- Pending updates ordered within a peer.
- Applying requests requiring domain-receipt reconciliation.
- Due outbox deliveries and unresolved terminal notifications.
- Current action and reply association for one exact peer/message/token.
- Bound invocations needing observation or an eligible continuation call.

Keep at most 10,000 receipt-only source rows and 64 MiB of logical transport
payload. Compact terminal input/capture/result bodies after seven days when
references permit. Preserve unresolved requests, active actions, current
conversation references and undelivered terminal output regardless of age.
Capacity pressure rejects new work; it does not evict an unresolved effect.

Retiring source rows atomically adds their exact update IDs to the receiver's
retired_update_ranges. Merge adjacent retired IDs only; never bridge an
unobserved numeric gap or use the previous epoch's maximum as a universal
rejection watermark. An update covered by a retired range is rejected as
retired in every epoch, never recreated. New sparse IDs remain admissible.
If interval capacity prevents safe retirement, retain the source rows and
apply capacity backpressure rather than losing duplicate protection.
Explicit recovery of an unresolved receipt takes precedence over compaction.

Migrations are additive and transactional. Existing project, remote-command
handoff, notification and daemon tables keep their meaning. The receiver
refuses an unknown future schema. Test failed/interrupted migration and reopen
with real SQLite, including preservation of existing user registry/state.

Use database constraints for primary/unique keys, required fields and state
enums; use domain validation and guarded transactions for cross-record
ownership, sizes, quotas and valid transitions. Do not rely on a single
in-memory mutex for cross-process protection.
