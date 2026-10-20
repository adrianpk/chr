# Telegram Message Review

Status: source review complete; live conversation and visual acceptance pending

Ticket: [TKT-20261017162810](../ticket/blocked/20261017162810-review-all-telegram-messages.md)

Source baseline: `5817e477190f` on `dev`, inspected on 2026-10-20. Pending
Slice 8 admission behavior is described separately in its delivery report.
This report proposes presentation changes; it changes no runtime copy,
authorization, active receiver, controller or workflow policy.

## Findings

1. Routine responses lead with internal identifiers and implementation state.
   Capture start, session lists, work status and control confirmations need a
   human project/work name, actual state and next action before audit details.
2. Capture, artifact preparation, ticket creation, admission, validation and
   integration are distinct events, but the vocabulary alternates between
   "session", "authoring", "native work", "record mutation" and "invocation".
   Use plain event descriptions while retaining exact references in details.
3. Prepared-artifact messages list several commands without choosing the next
   useful step. A development request can appear to be progressing while it is
   actually waiting for `/quiz`, `/spec`, `/prepare-work` or explicit admission.
4. Record reads and mutation receipts expose JSON as the primary interface.
   User-facing summaries need titles, fields that changed and an inspection
   reference. Preserve JSON as canonical data rather than conversational copy.
5. Provider retries often produce no intermediate response. Wording cannot
   reveal whether a request is queued, retrying, exhausted or lost unless the
   lifecycle owner exposes and durably projects that state.
6. Error messages correctly avoid claiming success, but often defer to "local
   diagnostics" or ask for several exact IDs without identifying the command
   the user can send now. A stopped receiver also cannot deliver its own outage
   notice through that receiver; local health remains necessary.
7. Exact confirmation scope and revision bindings are valuable. Presentation
   should explain the change and its consequences while preserving those
   bindings, expiration, operator mapping and replay protection.
8. Unicode bounds and retained plain-text fallback exist in the conversational
   receiver. The separate legacy notification sender has a different transport
   path and needs its own long/dynamic-content coverage.

## Outbound Boundary and Inventory

Receiver replies reach `telegramReceiver.sendDeliveryBody` in
`internal/app/telegram_drain.go`, then `TelegramReceiverAPI.SendAction` or
`SendMarkdown`. The durable outbox owns retry, replay and format fallback.
Separate notifications reach `TelegramSender.SendMessage` from
`notification_emit.go` and `notify.go`. Both transports call `sendMessage`.
The inspected production source has no `editMessage*` or
`answerCallbackQuery` call. Callback updates resolve retained actions and
produce ordinary messages; edited-message and callback-popup copy therefore
have no current producer. Silent authorization rejection is intentional.

The table groups dynamic branches by their producer and meaning. Filenames
are relative to `internal/app/` unless prefixed with `core/` (relative to
`internal/`). Every row has a disposition and concrete implementation target;
examples below are proposals, not screenshots or existing behavior.

| # | Trigger and variants | Source owner | Decision and replacement direction |
| --- | --- | --- | --- |
| 01 | `/help`, natural-language help; categorized commands | `telegram_help.go`, `telegram_help_metadata.go` | Keep categorized layout; make advanced commands conditional on relevant work. Preserve complete discoverability. |
| 02 | Unknown command, unsupported callback | `telegram_operations.go` | Rewrite: "I could not use that command. Send /help to see available commands." Keep callback wording specific to an unavailable action. |
| 03 | `/projects`; populated and empty registry | `telegram_operations.go` | Keep human names; restructure list heading and selection example. Empty: "No projects are available to you. Project access must be configured locally." |
| 04 | `/project`; selected and current selection | `core/telegram_conversation.go` | Keep "Selected project: {name}". Add a next step only when selection blocks a conversation. |
| 05 | Ambiguous/unauthorized project, conflicting reply, unavailable or changed root | `core/telegram_conversation.go`, `core/telegram_conversation_target.go` | Rewrite with `/projects`, a matching current session or a new session as appropriate. Do not reveal inaccessible project details. Root changes preserve the old session and require explicit rebinding. |
| 06 | `/status`; receiver health, retries and error | `telegram_operations.go` | Restructure: receiver health first, actionable retry/error state next. Bot ID and cursor epoch belong in diagnostic details. |
| 07 | `/prepare` or `/develop`; capture begins | `core/telegram_conversation.go` | Rewrite using example A. State that details are being collected; show `/done` and the selected project. Show session ID only as a secondary reference. |
| 08 | Start input too long/missing, conversation allocation full | `core/telegram_conversation.go` | Rewrite limits in plain terms. Capacity: "Finish or cancel an existing conversation before starting another. Send /sessions to choose one." |
| 09 | `/sessions`; empty, active, focused and retained sessions | `core/telegram_conversation_target.go` | Restructure into project, topic, state and one targeting command per session. Keep exact ID where selecting among sessions requires it. Empty state keeps `/prepare` and `/develop`. |
| 10 | `/focus`; success, missing target, non-collecting target | `core/telegram_conversation_target.go` | Rewrite: "Your next details will go to {topic} in {project}." Refuse non-collecting focus with its actual next step; do not imply that prepared sessions accept new capture. |
| 11 | Explicit/replied target mismatch, stale prompt, unavailable peer/session | `core/telegram_conversation_target.go`, `core/telegram_reply_resolution.go` | Rewrite: explain which prompt is obsolete, preserve current work and direct to `/sessions` or an exact current target. Keep peer scoping. |
| 12 | Text appended to capture | `core/telegram_conversation_capture.go` | Rewrite: "Details saved. Send more details, or /done to finish." Show byte/part usage only near limits or on explicit inspection. |
| 13 | `/show`; capture and state | `core/telegram_conversation_capture.go` | Restructure: project/topic, human state, captured text, reference. Move byte counts and revisions to details. |
| 14 | Invalid capture text; 16 KiB part, 64 KiB total, 32-part limit; already finished | `core/telegram_conversation_capture.go` | Keep exact limits in the relevant error. Explain that saved details are unchanged, and offer `/done`, `/cancel` or a new session according to actual state. |
| 15 | `/done`; saved capture and canonical authoring result | `core/telegram_conversation_capture.go`, `telegram_authoring_apply.go` | Rewrite using example B. State what was saved and whether any generation actually ran; choose a useful next command. |
| 16 | `/cancel`; unstarted conversation capture discarded | `core/telegram_conversation_capture.go` | Keep explicit discard semantics: "Conversation canceled. Its collected details were discarded." Distinguish this from execution cancellation. |
| 17 | Conversation/authoring command usage; unknown conversation operation | `telegram_conversation.go`, `telegram_authoring_dispatch.go`, `telegram_authoring_command.go` | Keep exact syntax, but show only the relevant command and one valid example. Do not present model argument schemas as usage. |
| 18 | Natural-language clarification, ambiguous target, malformed model decision | `telegram_intent_dispatch.go`, `telegram_intent_guard.go` | Rewrite one concrete question; use example C. Preserve the rule that interpretation cannot invent development authority or resolve conflicting targets silently. |
| 19 | Intent interpretation unavailable; five-attempt exhaustion | `telegram_intent_dispatch.go` | Rewrite: request not understood, prior conversations retained, retry budget exhausted when true; suggest `/help` or a new corrected instruction. |
| 20 | `/quiz`, generated missing-information question, reply/answer routing | `telegram_authoring_apply.go`, `core/telegram_conversation_question.go` | Restructure question first, with one reply instruction. Show `/reply {session} <answer>` only when direct reply is unavailable or ambiguous. |
| 21 | Generated dump/spec/plan and explicit refinement | `telegram_authoring_apply.go`, `telegram_authoring_dispatch.go` | Rewrite with artifact kind/name, draft state where applicable, what was preserved/replaced and a next step justified by the current intent. Never label a new draft approved. |
| 22 | Provider wait, retry and five-attempt exhaustion; authoring binding changed | `telegram_authoring_apply.go` | Rewrite exhaustion/rebinding diagnostics. Conditional durable retry notices need lifecycle support; do not claim an automatic retry when no retry remains. |
| 23 | Owner success after conversation advanced | `telegram_authoring_apply.go`, `telegram_prepared_work_result.go` | Restructure: saved artifact/ticket first, then "This conversation has moved on; its current selection was preserved." Do not overwrite the current binding. |
| 24 | `/explain`; model text, canonical sources, unsupported service | `telegram_explanation_dispatch.go` | Keep model answer separate from a short deterministic state wrapper. Group source references below. Explanation must not imply a spec, ticket or execution was created. |
| 25 | `/prepare-work`; description/scope preview and development authorization | `telegram_prepared_work_result.go`, `telegram_prepared_work_apply.go` | Restructure using example D. Name the project, ticket scope, original instruction and consequence. Retain exact description/contract binding and excerpts with a complete source reference. |
| 26 | Prepared backlog ticket success, replayed receipt | `telegram_prepared_work_result.go`, `telegram_prepared_work_dispatch.go` | Rewrite: ticket created in backlog, execution not started, next command for that exact ticket. Replayed success refers to the original result; do not print raw JSON as the main reply. |
| 27 | Prepared-work unavailable, unsupported intent, missing original instruction | `telegram_prepared_work_dispatch.go`, `core/telegram_conversation_instruction.go` | Rewrite with preserved context and the missing explicit instruction. Do not replace a missing development instruction with model inference. |
| 28 | Ticket/slice-set/use-case reads and pagination | `telegram_record_dispatch.go`, `telegram_record_command.go` | Restructure JSON into readable record titles, states, essential fields, page position and next-page syntax. Exact record references remain available. |
| 29 | Record create/update preview and confirmation | `telegram_action_record.go`, `telegram_record_dispatch.go`, `telegram_record_apply.go` | Restructure proposed changes and explicit consequences. Success: "Updated {kind} {title}: {changes}." Include canonical reference; retain request/revision guards privately. |
| 30 | Record invalid JSON, unsupported lifecycle, changed revision/source or owner unavailable | `telegram_record_dispatch.go` | Rewrite with precise correction: reread the record for changed guards; correct input for malformed requests; identify local setup only when remote retry cannot help. Do not imply mutation occurred. |
| 31 | `/work`; exact selection preview and confirmation | `telegram_work_action.go`, `telegram_work_dispatch.go` | Restructure using example E. Display human scope, count, target, validation and execution/publication policy. Preserve complete sealed bounds and exact confirmation authority. |
| 32 | Work admitted, waiting capacity or current owner, binding changed | `telegram_work_action.go`, `telegram_work_activation.go`, `telegram_work_failure.go` | Rewrite admitted state separately from execution and integration. Waiting/current-owner responses should link the existing work rather than encourage duplicate admission. |
| 33 | Execution progress, validation, review, partial results and final result | `telegram_work_progress.go`, `telegram_work_evidence.go` | Restructure using example F. Counts and next responsible party first. Completed integration requires the confirmed receipt; no receipt means no delivery claim. |
| 34 | Proactive work updates; reopened receiver or worker recovery | `telegram_work_progress.go`, `telegram_work_host.go`, `core/telegram_work_projection.go`, `core/telegram_processing_recovery.go` | Keep durable monotonic projection and replay behavior. Add a concise recovery explanation only when it changes what the user should do. Never replay an action as new work. |
| 35 | `/control status`, pause/cancel/resume/continue/reconcile/approve/grant/decision previews | `telegram_control_dispatch.go`, `telegram_control_prepare.go`, `telegram_control_request.go` | Restructure with exact action, work name, consequence, scope and limits. Preserve question/option, grant amount/reason and candidate binding. "Pause requested" is distinct from quiescent workers. |
| 36 | Control confirmed, receipt recovery, unavailable/stale control | `telegram_control_action.go`, `telegram_control_diagnostic.go` | Rewrite using the actual action and resulting state. Recovered receipts confirm the prior action, not a second application. Stale previews require a new current preview. |
| 37 | Integration questions/options, truncated issue text, more than three questions | `telegram_control_status.go` | Restructure one question and bounded choices at a time. Show the exact answer command/reference. Preserve complete canonical questions and candidate-bound answer semantics. |
| 38 | `/harness`; role defaults, readiness, available profiles, retained bindings | `telegram_harness_dispatch.go` | Restructure current worker/generation selections first. Keep local readiness distinct from remote authentication; hide unrelated running bindings unless they explain retained behavior. |
| 39 | Harness change preview, confirm, cancel, stale revision, unsupported setup | `telegram_harness_prepare.go`, `telegram_harness_action.go`, `core/telegram_harness_action.go` | Keep the new-operations-only consequence; rewrite labels around worker and generation roles. Preserve clearing versus leaving unchanged and exact revision checks. |
| 40 | Action already confirmed; explicit `/answer`, reply or callback; expired/stale action | `telegram_action_dispatch.go`, `telegram_control_diagnostic.go`, `core/telegram_action_answer.go` | Rewrite: "This action was already confirmed" plus its original result, or "This confirmation has expired. Request a new preview." Keep replay safety and matching peer/prompt. |
| 41 | Inline keyboard labels for start/develop/control/confirm/cancel | `core/telegram_receiver_action_http.go` | Rewrite visible labels as "Start work", "Prepare ticket", "Pause work", "Confirm change", "Cancel", as appropriate to the actual effect. Keep opaque callback tokens and accepted choice keys unchanged. |
| 42 | Pending-intake, model-request and conversation capacity | `core/telegram_intake_reserve.go`, `telegram_drain.go`, `core/telegram_request_allowance.go` | Rewrite rejection separately from waiting: "The queue is full. This request was not accepted. Retry after pending requests finish." Preserve reserved control capacity and existing work. |
| 43 | Uncertain send acknowledgement, transport retry, Markdown fallback | `telegram_drain.go`, `core/telegram_receiver_outbox.go`, `core/telegram_help_outbox.go` | Rewrite duplicate warning: "Delivery confirmation was lost; this notice may be repeated." Keep bounded retries, retained fallback and operator-only exhausted-delivery recovery. |
| 44 | Empty output, long output, multipart reference headings, malformed Unicode | `core/telegram_response_chunks.go` | Keep Unicode accounting and retained reference; restructure numbered parts to avoid repeating long opaque context. Summary must identify omitted content and a real inspection route. |
| 45 | General read/service unavailable | `telegram_drain.go`, `telegram_conversation.go`, preparation dispatchers | Rewrite with the unavailable operation, preserved state and a usable next step. Avoid promising progress while a local operator fix is required. |
| 46 | Unauthorized inbound/delivery, removed access, invalid config, poll outage | `telegram_drain.go`, `telegram_serve.go`, `telegram_failure.go` | Keep silent authorization rejection and local-only service diagnostics. An outage notification needs an independent channel; do not promise self-reporting through a stopped receiver. |
| 47 | Ticket-blocked and loop finished/blocked/failed notifications, including dynamic worker summary | `notification_emit.go`, `core/telegram_notification.go` | Restructure with project/work name, actual counts, blocker and next owner. "Loop finished" must not imply integration. Treat worker summary as reported content, not host certification. |
| 48 | Explicit notification test | `notify.go`, `core/telegram_notification.go` | Rewrite: "Telegram notification test for {project name}." Keep the test purpose; show local directory only when needed for diagnostics. |

## Proposed Message Layouts

Placeholders below require validated source data. They do not authorize a new
lookup, target change, hidden default or command. Commands must use existing
syntax and exact retained references. Human names complement exact bindings;
they never replace them in the action or receipt.

### A. Development capture begins

```text
Development request for telegram-journey-qa

I am collecting the details for: {topic}
Execution has not started.

Send the requirements here. Send /done when the details are complete.

Conversation reference: {conversation}
```

For preparation-only intent, use "Preparation request" and do not imply
development authority. Focus/list commands appear only if several sessions
make the destination ambiguous.

### B. Capture is saved

```text
Details saved for {topic} in {project name}.

Saved: {dump reference}
Specification generation and execution have not started.

Next: send /quiz {conversation} to clarify the requirements.
```

Choose this next action only when clarification is appropriate. If all input
is already sufficient, offer `/spec {conversation}` instead. A successful
`/spec` response should identify the generated draft, not reuse this capture
message or suggest that its generation is still pending.

### C. Clarification question

```text
One detail is needed for {topic}:

{question}

Reply to this message with your answer.
```

Do not add an invented question. Display the actual retained question and its
choices; use the exact current conversation fallback when direct reply is
ambiguous or unavailable.

### D. Preparing a ticket

```text
Prepare a ticket in {project name}

{short description}

Original request: {instruction excerpt}
Full description: {canonical source}

Confirming creates this ticket in the backlog. It does not start execution.
```

When development authority is missing, explicitly explain what a confirmation
would authorize. Never label the button "Start work" for ticket preparation.
Keep scope/description/contract hashes and operator/revision bindings in the
retained action and accessible audit details.

### E. Admission confirmation

```text
Start work in {project name}?

Scope: {exact work title and ticket count}
Target: {repository and target branch}
Validation: {configured command}
Integration: {actual publication and review policy}
Limits: {actual attempts and execution/validation deadlines}

Confirming starts this exact selection under the displayed policy.
```

Pending Slice 8 additionally needs to explain a backlog-to-ready transition
when the sealed admission includes it. Omitted members or truncated scope
must remain inspectable before confirmation. Keep the original native work
extent, readiness image and policy bindings intact.

### F. Progress and results

```text
{work title} — {human execution state}

Validated: {validated}/{total} tasks
Integrated: {confirmed units}; remaining: {remaining}
{current blocker or next step, when present}

{PR URL, when one exists}
{confirmed integration receipt, only when one exists}
Reference: {exact work reference}
```

Use state-specific next actions: waiting for user input names the exact
question or approval; waiting for capacity says whether resumption is
automatic; failed validation points to retained evidence; partial results
identify integrated and remaining scope. Without a confirmed receipt, say
"Integration is not confirmed". Never infer delivery from worker exit,
passed validation, an open PR or a model's success statement.

## Communication Rules for Implementation

- Lead with the project/work name and the event that actually occurred.
- Use separate paragraphs for state, relevant context and the next action.
- Offer one primary next action; show alternate targeting commands conditionally.
- Retain exact scope, target, authority, revision, expiration and evidence.
  Move routine implementation metadata into accessible details.
- Use "saved", "draft prepared", "ticket created", "execution started",
  "validation passed" and "integration confirmed" for their distinct events.
- For errors, say what was preserved and whether retry is automatic, explicit
  or exhausted. A new request must not reset the old request's budget.
- Treat model questions, explanations, blocker descriptions and worker
  summaries as dynamic content. Add deterministic state and next-step text
  without promoting those statements into host execution evidence.

## Journey Review and Remaining Evidence

The ordinary development path is project selection → request capture →
clarification → draft artifacts → prepared ticket → scope confirmation →
admission → execution → validation/review → confirmed integration. An
explanation-only request exits with an explanation. Preparation-only requests
stop before development authority and execution. Every pause needs to say
which input or owner can continue it.

Current retained live evidence covers a genuine development request and a
saved canonical dump, with zero model attempts at that checkpoint. The
isolated receiver was observed `ready` on 2026-10-20, with one conversation,
zero pending updates/deliveries and no exhausted model requests. Its executable
digest still matches the private descriptor. No new incoming message was
fabricated and no real end-to-end result is claimed by this audit.

| Evidence | State and limitation |
| --- | --- |
| Production source and outbound send boundaries | Inspected; inventory and proposals above. No runtime presentation changes. |
| Existing deterministic journey and recovery fixtures | Identified in Slice 8 report and Telegram tests. Historical passes apply to their recorded candidates; no new pass is inferred. |
| Mobile Telegram layout | Pending actual client evidence. Source review cannot establish line wrapping, keyboard usability or readability. |
| Desktop Telegram layout | Pending actual client evidence; the same limitation applies. |
| Long/duplicate names, accents, CJK and emoji | Existing Unicode bounds inspected; actual client examples still required. Distinguish byte capture limits from UTF-16 transport bounds. |
| Markdown help and plain fallback | Retained paths inspected. Live prior help acknowledgement does not prove all formatting, long text or fallback cases. |
| Complete live model-to-integration journey | Pending configured-user continuation and canonical result/receipt. |
| Source self-development with fixed receiver | Pending separately bound source-QA request and result; do not rebuild the receiver during that proof. |

For visual acceptance, inspect a short capture exchange, a long question,
prepared-artifact/ticket output, work confirmation and blocked/partial/final
status on both clients. Include long names and emoji, multipart output,
plain fallback and visible buttons. Retain genuine client evidence separately
from HTTP/rendering fixtures and record the actual source candidate.

## Recommended Implementation Follow-ups

| Priority | Bounded change | Inventory coverage | Required proof |
| --- | --- | --- | --- |
| First | Shared human project/work context and readable capture/session/artifact replies | 03–23 | Multiple-session routing, stale replies, complete preparation exchange and both client layouts. |
| First | Ticket/work previews, explicit next steps and human confirmation labels | 25–27, 31–32, 35–41 | Exact scope/target/revision guards, missing authority, changed candidate, replay and actual admission. |
| First | Readable progress, review decisions and truthful result summaries | 33–37, 47 | Partial results, no receipt, confirmed integration, restart and stale-control recovery. |
| Next | Record lists, changes and errors without raw JSON as primary copy | 28–30 | Pagination, long records, strict mutation guards and unsupported lifecycle errors. |
| Next | State-aware retry/capacity/error notices | 05–06, 18–19, 22–24, 42–46 | Owner-backed wait/retry state, exhaustion, revoked access and unchanged retry budgets. |
| Next | Notifications, help and long-content presentation consistency | 01–02, 38–39, 43–44, 47–48 | Both send paths, Unicode limits, long dynamic summaries, Markdown fallback and real mobile/desktop evidence. |

These are review recommendations, not newly approved tickets or implementation.
The review ticket remains blocked on actual visual acceptance and the complete
live conversations. Source review alone does not satisfy its full acceptance.
