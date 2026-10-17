---
id: TKT-20261017094606
title: "Render categorized Telegram help from native command metadata"
status: open
kind: task
severity: unclassified
priority: normal
scope: api
tags: telegram, help, command-registry
source: chat
reported_at: 2026-10-17T09:46:06Z
commits: []
---

# Render categorized Telegram help from native command metadata

## Problem

The Telegram receiver assembles `/help` by concatenating command-family help
strings in `internal/app/telegram_operations.go`. The result mixes syntax and
descriptions without category headings. Some families derive text from typed
registries; others keep usage in separate strings. The receiver bounds this
combined text before retaining the reply, so a growing list can lose commands
through truncation. The HTTP sender currently sends plain text only.

Provide a complete, readable command reference inside Telegram while retaining
CHR's existing command semantics, authorization and durable delivery model.

## Requested Work

1. Project the authoritative CHR command definitions into read-only help
   metadata with a stable identity, category and canonical syntax. Derive text
   from the existing help/usage fields without rewriting argument names,
   optional parameters or whitespace. Add missing usage metadata beside the
   owning parser/descriptor where necessary; do not introduce a separate
   hand-maintained command catalogue or redesign the entire dispatcher.
   Include usage-only definitions and preserve existing alias and overload
   semantics. Preserve descriptor multiplicity rather than silently deduplicating.
2. Render short headings and one canonical syntax entry per line, with a blank
   line between sections. Use this section order: Execution, Tickets, Slice
   sets, Use cases, Projects, Authoring, Conversations, System controls, Other.
   Emit only populated sections. Assign categories from CHR's command owners;
   unknown owners use Other. Sort exact syntax text within each section.
   Keep descriptions out of the rendered list. Do not use a large code fence.
   Rendering must be deterministic, leave input metadata unchanged and return
   empty unformatted output for empty input.
3. Use Telegram's legacy `Markdown` mode for help only: safe headings are
   bold and safe syntax entries use inline code. Escape text that cannot use
   those entities safely, including literal backticks. Do not change command
   syntax to satisfy the markup. Ordinary responses retain their existing
   plain-text behavior. Reuse the renderer wherever the receiver already
   returns the full help list, preserving any existing diagnostic prefix.
4. Keep direct `/help` model-free and available to an authorized peer without
   a selected project. Preserve selected project, focused conversation,
   revisions, capture, pending questions, confirmation bindings and native work
   state. Help must neither admit work nor become a capture/confirmation reply
   target. Preserve current parsing and routing behavior, including existing
   natural-language interpretation and unknown-command diagnostics. Do not
   add `/start`, command suffix handling or help-topic semantics as part of
   this presentation change.
5. Carry the help format explicitly through the existing durable outbox and
   sender so pending replies retain their meaning across restart. Retain the
   formatted body and plain fallback from the same immutable metadata snapshot
   atomically with the logical response. After restart, fallback uses that
   retained snapshot rather than the current command registry. Include any
   existing diagnostic prefix in both forms. Keep source acknowledgement and
   response retention atomic. Replayed input must not
   enqueue another reply. Keep authorization checks at intake and delivery,
   peer isolation, ordered delivery, capacity bounds and retained retry budgets.
   Retained replies created before this change must remain deliverable as
   plain text.
6. Permit one plain-text fallback after a definite Telegram entity/format
   rejection, identified from the API response rather than a generic HTTP
   status. Render that plain body from the same canonical metadata; do not
   strip delimiters or unescape command punctuation heuristically. The fallback
   must contain every syntax entry. Never immediately
   resend merely because of a timeout, rate limit, transport/server error or
   uncertain delivery. Record actual attempts using the existing delivery
   owner; restart must not reset the attempt budget or repeat a known-successful
   formatted send. The format-specific fallback must not affect other replies.
7. Demonstrate that the complete current CHR help fits one message through the
   actual retention and delivery pipeline. Account for the existing 3,200
   UTF-16-unit chunk threshold and the sender's 3,900-code-point/UTF-16-unit
   limits for both formatted and plain bodies, including diagnostic prefixes,
   headings and delimiters. Do not pass full help through
   a truncation helper or split inside markup. If future metadata exceeds this
   bounded contract, return an explicit bounded diagnostic instead of a partial
   command list. Add a synthetic overflow regression. Arbitrary-size formatted
   pagination is outside this ticket.
8. Update the Telegram reference, relevant user guidance and the Unreleased
   changelog when implementation is delivered. Document the actual entry
   points, formatting, fallback and size boundary without promising topic
   pages or client presentation that was not verified.

## Native Integration Points

- `internal/app/telegram_operations.go`: read registry and full-help assembly.
- `internal/app/telegram_conversation.go`, `telegram_authoring_command.go`,
  `telegram_explanation_command.go`, `telegram_prepared_work_command.go`,
  `telegram_work_command.go`, `telegram_control_command.go`,
  `telegram_record_command.go` and `telegram_action_dispatch.go` in the same
  package: existing command-family definitions and syntax owners.
- `internal/app/telegram_drain.go`: authorization, dispatch and durable send.
- `internal/core/telegram_receiver_outbox.go`: response retention, replay and
  attempt accounting; any format metadata must preserve existing records.
- `internal/core/telegram_receiver_action_http.go`: outgoing payload and
  action-button behavior, which ordinary responses must retain.
- `internal/core/telegram_response_chunks.go` and
  `internal/core/telegram_receiver_update.go`: existing output bounds.
- `docs/reference/telegram-receiver.md`: current external contract.

## Acceptance and Validation

1. Compare the rendered syntax multiset with the actual CHR definitions,
   including usage-only entries, aliases, overloaded commands and scoped
   arguments. Inject and modify test registry entries to prove that the
   renderer does not own another command list. Cover ordering, empty input,
   unknown categories, duplicate entries and input nonmutation.
2. Check markup using punctuation, underscores, brackets, backticks and Unicode.
   Prove current output survives the real outbox path as one complete valid
   message. Exercise synthetic overflow without truncation or invalid entities.
3. Use local HTTP fixtures and real retained receiver state to assert the
   outgoing `parse_mode`, exact command coverage, ordinary plain-text replies,
   definite format rejection and its single plain fallback. Assert that
   timeout/uncertainty, HTTP 429 and server errors retain the normal retry
   behavior instead of causing an immediate second send.
4. Cover restart with a pending formatted reply, old plain-text outbox rows,
   input replay, attempt accounting, unauthorized peers and authorization
   revoked before delivery. Change the registry after enqueue and verify that
   fallback after restart still uses the retained original syntax and prefix.
   These are focused receiver regressions, not a new generic reliability
   framework.
5. Assert `/help` during a collecting/clarifying conversation and a pending
   confirmation leaves the selected project, capture, revisions and bindings
   unchanged, invokes no model and admits no work. Test existing reply-routing
   guards, including `/help` sent as a reply to a question/confirmation, so
   displaying help cannot consume an answer or a confirmation. A later message
   replying to help may still follow ordinary focused-conversation routing;
   help itself must not create a new capture or confirmation binding.
6. During implementation, run the new focused metadata/renderer, receiver and
   HTTP regressions plus the repository development gate `make verify`.
   Record exact commands and results. Keep exhaustive suites in the scheduled
   Forgejo nightly. A real Telegram client visual check is separate evidence;
   local HTTP fixtures do not prove mobile/desktop appearance.

## Scope Boundary

One focused implementation ticket. This capture authorizes no implementation,
service activation, bot messages, token changes or controller restart. It does
not add help search, topic pages, navigation buttons, menu registration,
localization or model-generated help. No runtime, package or configuration
dependency on another application is required.

## Discovery Evidence

The requested Iris exchange with the KAL project is conversation
`e67cd52b8628018af2a8e5b0527a1362`. The first verified answer is message
`3e1b0088349b1363f063538dfa68d907`; the follow-up answer
`f8a353d5f210ddb08d6bb0c0a3bae11a` resolves registry ownership, fallback safety,
output bounds and active-conversation isolation. Final scope review
`3ef10deb031b6a27f338191516f75606` confirmed the bounded contract and identified
the retained-snapshot clarification now included above. Discovery is complete;
implementation remains unstarted.

The delivered reference is KAL commit
`eda72ae69993f1483f4d37e7476dea68fde54e8e`, following PRs
[207](https://forge.adrianpk.com/adrianpk/kal/pulls/207),
[208](https://forge.adrianpk.com/adrianpk/kal/pulls/208),
[209](https://forge.adrianpk.com/adrianpk/kal/pulls/209) and
[210](https://forge.adrianpk.com/adrianpk/kal/pulls/210).
Reviewed source includes `internal/telegram/help/{metadata,render,escape}.go`,
`internal/app/telegram_help_metadata.go`, the help HTTP/acceptance tests and
`docs/reference/telegram-help.md` in that repository. These are provenance;
the requirements above define the self-contained CHR change.

The reference's generic fallback retries on any send error and its common
splitter uses byte boundaries. Neither limitation is a required CHR behavior.
The narrower fallback and output bounds above preserve CHR's native delivery
guarantees. No runtime tests, live bot requests or client visual checks were
run during ticket capture.
