# Slice 8: Live Journey and Closure

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `test/slice-8-telegram-delivery`

PR: pending

## Purpose

Certify the complete Telegram preparation and native development journey,
including bounded live QA, source self-development and exact-candidate closure.

## Delivered Behavior

Slice activation only. No live journey or Slice 8 acceptance result is claimed.
The dedicated private bot passes read-only credential, identity and webhook
checks. Its current project mapping points to the registered source checkout;
live validation must use the designated disposable QA target instead.

## Implementation Notes

PR #396 integration is verified at `c14f1aea640657e27c97bb64e40b46f30f698451`.
Its three tasks are closed with their controller validation receipt. Slice 8
uses its own worktree and leaves the primary checkout on dev. The existing
installed CLI predates the Telegram commands; live validation requires a fixed
receiver executable built from the integrated source.

## Contracts Added or Changed

None yet. Deterministic fixture results cannot replace real bot/model reception,
QA integration, source self-development or the scheduled final-candidate gate.

## Files of Interest

- `docs/how-to/user-guide/index.md`: existing product User Guide.
- `docs/reference/telegram-receiver.md`: receiver operation and bounds.
- `internal/app/telegram_work_automatic_test.go`: bounded native integration fixture.

## Validation

Activation documentation passed `git diff --check`. The integrated Slice 7
binary's read-only `telegram status` and `telegram doctor --check-connection`
checks passed. They establish local bot readiness, not live execution or model
connectivity. Slice 8 focused checks and exact-HEAD controller validation remain
pending.

## Risks and Follow-ups

Prepare an isolated binding to an existing designated QA repository and a fixed
installed receiver. The authorized Telegram user must supply the live request
and any bound clarification/admission messages. Record actual source/base/result
and canonical integration receipt separately from fixtures. Scheduled exhaustive
certification and delivery-set closure remain pending the final integrated
candidate.
