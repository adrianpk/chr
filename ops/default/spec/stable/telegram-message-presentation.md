# Telegram Message Presentation

Status: accepted

Date: 2026-10-21

Approved: maintainer, 2026-10-21; specification, model and implementation accepted for Looper execution.

Model: [Presentation and document delivery model](telegram-message-presentation-model.md)

Plan: [Presentation delivery plan](../../plan/telegram-message-presentation.md)

Tracker: [Presentation coverage and delivery tracker](../../tracker/telegram-message-presentation.md)

## Purpose and Inputs

Make chr conversations readable and easy to navigate in Telegram. Users inspect
a short view in the chat and open complete documents in an appropriate external
application. They can understand the current state and next action without
decoding operational metadata.

The [48-family source review](../../report/telegram-message-review.md) supplies
the initial inventory. The existing [conversation specification](telegram-conversational-workflow.md)
continues to own routing, authority and workflow operations. This delivery
extends message presentation and outbound document delivery across receiver
replies and the separate notification sender.

## 1. Flexible Message Composition

A message can contain, in order:

1. Initial text explaining the event, project or work in human terms.
2. Main content: a short Markdown excerpt, code, a list, choices or another
   presentation appropriate to the event.
3. Optional secondary text explaining a consequence, wait, preserved result or
   next step.
4. Action buttons when the recipient has a supported action to take.

Use only the necessary blocks. Separate visible blocks with paragraph breaks.
A brief acknowledgement can be one sentence. A question leads with the actual
question. An error explains its impact and recovery. Progress and final results
use their own concise layouts. The pattern is not a mandatory four-part wrapper.

Use project names and work titles. Keep hashes, internal IDs, source paths,
revisions and raw JSON/YAML in accessible technical details unless a particular
diagnostic or target choice requires them. Show one primary next action.
Labels explain the real effect: preparing a ticket, authorizing development and
starting execution are different operations.

For example, a prepared specification can use these blocks:

| Block | Example |
| --- | --- |
| Initial text | Draft specification ready for the README change in example-project. |
| Main content | Opening lines of the specification, displayed as literal Markdown. |
| Secondary text | Complete specification attached. |
| Action | Prepare ticket |

A short acknowledgement or progress notice may contain only initial text.

Capture-start messages acknowledge the request already retained and explain
that more details are optional. Prepared results identify what was produced
and what input, if any, is required next. Multiple-session choices use human
topics and project names while retaining exact internal target bindings.

## 2. Markdown and Document Views

When showing Markdown source, present it as a literal preformatted block.
Code values use inline or block code as appropriate. Ordinary explanatory text
remains ordinary text. The renderer uses Telegram-supported formatting and
escapes dynamic content; it does not assume that arbitrary document Markdown
is valid Telegram markup.

For a document response, show its title and an opening excerpt taken directly
from the document body. Omit leading YAML front matter and operational
provenance from the excerpt. Do not generate a model summary to replace the
source. Mark a shortened excerpt and state that the complete document is attached.

Attach the complete document as a sensibly named .md file. Preserve its exact
bytes, including source metadata where present. Read/export operations whose
canonical content is another format may attach that original format or a
clearly identified deterministic readable export. Never label an excerpt or
derived export as the complete original.

Attach the primary document relevant to the current operation. Do not resend
all earlier dump, quiz, spec and plan artifacts on every response. Other retained
documents remain available through authorized inspection. A short status
without a document needs no attachment.

For document-backed confirmations, send the short caption, complete file and
existing action keyboard together. If an essential consequence does not fit
the caption, include it in the complete bound document and explicitly identify
the shortened view. Users must have the full scope available before acting.

## 3. Presentation Bounds

These are the accepted chr product limits, below Telegram transport limits.

| Item | Bound |
| --- | --- |
| Opening excerpt | First 6 nonempty body lines, at most 600 UTF-16 units |
| Normal composed text | At most 1,800 UTF-16 units, including secondary text |
| Document caption | At most 900 UTF-16 units, including all visible blocks |
| Document bytes | At most 2 MiB per snapshot |
| Documents per logical response | One primary document |
| Retained snapshot bytes | At most 64 MiB in each owning database |

Count the final visible text, including truncation markers, before sending.
Retain existing hard transport and queue limits. Do not fill the chat with
automatic document chunks to bypass these bounds. Paginated lists may retain
their existing bounded navigation.

If the full document cannot be attached, state why and give a supported next
step. Do not replace it silently with a truncated file. A new confirmation that
requires that document remains unavailable until a complete valid view can be
offered. A domain operation that already succeeded keeps its original receipt;
a presentation failure must not repeat the operation or claim it failed.

## 4. Version, Authority and Delivery

Freeze each presentation and document snapshot with its original request or
notification event before transmission. Reopening, retrying or switching to
plain formatting sends the same document version rather than rereading a
mutable path.

Keep existing peer, operator, project, namespace, source revision, exact scope,
expiration, action-consumption and native receipt checks. A friendly label or
short excerpt is not an authorization guard. A document-backed action binds
the snapshot hash as well as the original prepared operation.

A Telegram acknowledgement proves delivery only. Actual execution, validation
and integration continue to require their own canonical evidence. Lost send
acknowledgements can produce duplicate messages; they must not duplicate a
domain mutation or silently create another retry allowance.

Plain fallback preserves the concise content, full attachment and action
binding. A failed upload, invalid source, revoked access, exhausted retry or
storage bound produces a truthful, bounded outcome. Cleanup cannot remove
bytes needed by an unacknowledged delivery or a live confirmation.

The legacy notification path uses the same presentation rules and document
values through its existing event repository and dispatch owner. It retains its
configured destination and trigger policy. It does not acquire interactive
development authority or a second receiver.

## 5. Coverage and Acceptance

The [tracker](../../tracker/telegram-message-presentation.md) maps all 48 initial
families to implementation and evidence. Every family must have an explicit
outcome, including keep/omit where a block or button is unnecessary. Reconcile
new producers and variants against the actual implementation baseline before
closing coverage.

| Criterion | Observable requirement |
| --- | --- |
| TMP-01 | Messages compose only the blocks needed for their event, in a readable order. |
| TMP-02 | Both receiver responses and notification messages follow the same presentation contract. |
| TMP-03 | Literal Markdown/code is correctly displayed; dynamic delimiters and Unicode cannot corrupt the surrounding message. |
| TMP-04 | Document previews are bounded direct excerpts without front matter or routine audit metadata. |
| TMP-05 | The complete, correctly named attachment matches its frozen source or explicitly identified export. |
| TMP-06 | The user can understand what happened and the next input/action; optional capture input and ambiguous targets are explained. |
| TMP-07 | Buttons describe actual effects and retain all existing target, authority, revision and replay guards. |
| TMP-08 | Reopen, retry, plain fallback and source-file changes preserve the bound document bytes and action identity. |
| TMP-09 | Missing/oversized documents, full storage and failed delivery are explicit; required-document confirmations fail closed. |
| TMP-10 | Existing queued messages and notification records survive additive migration with their original meaning. |
| TMP-11 | All 48 families and their applicable rare/dynamic branches have recorded outcomes and meaningful regression evidence. |
| TMP-12 | Real mobile/desktop inspection and a complete user-operated journey confirm readability and truthful delivery states. |

Visual acceptance includes opening attachments, paragraph spacing, literal
Markdown, long names, emoji, keyboard labels, multiple conversations, fallback,
errors and partial/final results. Record client/build identities and real
observations. HTTP fixtures do not prove client layout.

This specification improves the presentation of existing operations. Workflow
gaps that require a new domain transition or automatic artifact orchestration
must be identified separately rather than hidden by reassuring copy.

## 6. Approval Checkpoint

The maintainer accepted flexible composition, short document views, complete
attachments, the numeric bounds and the model/plan on 2026-10-21. Implementation
and canonical slice PR integration into dev are delegated to Looper for this
delivery set. Actual mobile/desktop observations and the user-operated journey
remain acceptance requirements; they cannot be replaced with fixture evidence.

## References

- [Telegram formatting options](https://core.telegram.org/bots/api#formatting-options)
- [Telegram document delivery](https://core.telegram.org/bots/api#senddocument)
- [Existing review ticket](../../ticket/blocked/20261017162810-review-all-telegram-messages.md)
