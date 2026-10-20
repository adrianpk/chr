---
id: TKT-20261017162810
title: "Review every Telegram message and conversational UX"
status: blocked
kind: follow_up
severity: unclassified
priority: unclassified
scope: ui
tags: telegram, conversational-ux, message-design
source: chat
reported_at: 2026-10-17T16:28:10Z
commits: []
ready_at: 2026-10-20T20:23:49Z
started_at: 2026-10-20T20:23:49Z
branch: dev
---

# Review every Telegram message and conversational UX

## Problem

The maintainer considers the current Telegram communication unclear and
unpolished. Review every response and message the agent sends through Telegram,
including messages outside the immediate command-response path. Isolated fixes
to project lists or selection confirmations do not establish a coherent
conversation experience.

One observed development-start response combines an opaque conversation ID,
an internal project ID, a parenthesized intent, capture instructions and several
advanced commands in one dense paragraph. It does not clearly explain whether
the agent is waiting for details, preparing work or executing a change, or what
the user must do next.

## Review Scope

1. Build a complete inventory of outbound message producers and message
   families. Include static text, shared renderers, dynamically assembled replies,
   model-produced content, notifications, edited messages, callback responses and
   button labels wherever those surfaces exist. Record the source owner and
   triggering conditions so less common branches are accounted for.
2. Review all command families and complete user journeys: help and usage,
   project lists and selection, conversation creation and focus, collecting
   requirements, clarification questions, artifact preparation, work previews and
   confirmation, execution, progress, validation, review, integration and results.
3. Include empty states, ambiguous input, unavailable or unauthorized targets,
   stale actions, cancellation, capacity waits, provider or transport retries,
   recovery after restart, failures and partial results. Assess proactive updates
   and retained replies as well as immediate responses.
4. Examine wording, terminology, information hierarchy, paragraph breaks,
   headings, lists, emphasis, command presentation, message density, repeated
   context and consistency. Review whole exchanges as well as individual messages.

## Expected Communication

1. Use plain English and human-readable project and work names. Explain any
   necessary technical term. Internal IDs, state names and parenthesized metadata
   must not dominate routine messages. Preserve meaningful exact references and
   the target and authority bindings behind confirmations.
2. State what has actually happened and what is happening now. Distinguish a
   received request from prepared work, execution, passed validation and confirmed
   integration. Make waiting for user input explicit.
3. Highlight the next action needed to continue, the input expected and what
   that action will do. A user must understand how to advance without translating
   internal workflow terminology. Show advanced conversation-management commands
   only when they help the current situation.
4. Separate context, status and next steps with a readable visual hierarchy.
   Avoid compressing identifiers, metadata, explanations and several possible
   commands into one paragraph. Use consistent terminology and concise text.
5. Make errors and waits actionable. Explain what is preserved, whether retry
   is automatic and whether the user must act, without implying progress or
   completion that has not occurred.
6. Keep confirmations clear about the selected project, proposed action, scope
   and consequences. Presentation improvements must preserve authorization,
   revision checks, replay safety and accurate delivery claims.

## Review Deliverables and Acceptance

1. Every inventoried message family has an explicit review outcome: keep with a
   reason, rewrite, restructure or make conditional. Account for dynamic variants
   and rare paths; a handful of selected screenshots is not a complete audit.
2. Provide concrete findings, proposed replacement copy and layout, and a small
   consistent set of communication rules. Identify any workflow gap that wording
   alone cannot resolve and separate it from a presentation issue.
3. Walk through complete conversations, including a development request from
   project selection to its actual result. At every step, the user can tell the
   current state and the next required action.
4. Assess readability in Telegram mobile and desktop layouts, including long
   names, Unicode, message limits, formatting and plain-text fallback. Record
   actual visual evidence separately from renderer or transport test evidence.
5. Produce actionable follow-up recommendations for implementation, with
   coverage traceable to the complete message inventory.

## Execution Boundary

This ticket captures an exhaustive review task and replacement-message
proposals. Keep it open until the review is scheduled; capturing the request
does not begin the audit or change the running Telegram workflow.

## Execution Update: 2026-10-20

The maintainer scheduled this review through the project GTD work queue.
The [source review](../../report/telegram-message-review.md) inventories 48
message families across both outbound transports, records a disposition for
each, proposes replacement layouts and maps bounded implementation follow-ups.
No runtime copy or running receiver was changed.

Status is `blocked` on actual Telegram mobile/desktop visual evidence and
configured-user continuation of the complete live development conversations.
Source and renderer/HTTP fixture evidence do not satisfy those acceptance
items. Replacement proposals are not implementation approval.
