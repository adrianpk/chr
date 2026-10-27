---
id: TKT-20261026084827
title: "Quality slice 7: Telegram transport, receiver and retained delivery"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-7
source: review
reported_at: 2026-10-26T08:48:27Z
commits: []
started_at: 2026-10-27T09:28:39Z
branch: refactor/slice-7-quality-telegram
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 7: Telegram transport, receiver and retained delivery

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R6, R7, R9, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T7.1:** Move the core Telegram receiver, transport and retained operation/outbox behavior into internal/core/telegram. Introduce a nested outbox owner only if storage and transport separation actually need it; preserve one transactional owner for each existing operation.
2. **T7.2:** Replace SendDocument, SendAction, presentation senders and notice insert argument lists with named request/row values. Convert all assigned result shapes while preserving missing, stale, duplicate and acknowledgement outcomes.
3. **T7.3:** Apply the row mappings from Slice 3 to notice, document, action and response queues. Preserve exact original-document attachment bytes, captions, fallback behavior, finite quotas and restart reconciliation.
4. **T7.4:** Resolve duplicate dispatch pair D06 through a narrow helper with separate handler tables. Verify that removing telegramGuidedResultExport and telegramHelpReply does not remove their live replacements or useful behavioral tests.
5. **T7.5:** Consolidate error type names, typed receiver/operation/outbox states, unused parameters and justified close handling. Update app adapters directly; do not re-export the moved package through core.
6. **T7.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Telegram code has an owned package boundary, cohesive request values and compatible retained delivery; duplicate replies or attachments are not introduced by retries or restarts.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 6 into dev.
Planned branch: `refactor/slice-7-quality-telegram`.
Expected PR title: `refactor(slice-7): isolate Telegram ownership and request values`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-7-telegram.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing receiver/action/outbox and document multipart fixtures with real local HTTP/SQLite; stale and lost-ack recovery; focused race where state ownership changes; make verify.

All commands above are planned. None ran as implementation validation during this planning task.
