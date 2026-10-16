# Slice 6: Guarded Controls and Capacity

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-6-telegram-controls`

PR: pending

## Purpose

Expose exact-target native controls, bound decisions and shared capacity through
responsive Telegram conversations and the common workflow owner.

## Delivered Behavior

Implementation pending. Slice 5 is delivered after verified canonical PR #394
integration; T6.1 is active under the approved Slice 6 identity.

## Implementation Notes

Use the shared `ControlWorkExtent` operation and domain request receipts. The
transport retains only exact displayed guards and conversation projections.

## Contracts Added or Changed

None yet.

## Files of Interest

- `internal/app/service_workflow_control.go`: existing common control owner.
- `internal/app/telegram_*`: existing typed operations and bound action mapping.

## Validation

Activation documentation passed `git diff --check`.

Required slice validation will be recorded through the controller for the final
clean PR HEAD: `make verify`, focused `TestTelegramControl` normal and race
commands, then `git diff --check`. Live and scheduled exhaustive checks remain
separate.

## Risks and Follow-ups

T6.1 through T6.4 remain pending implementation and focused validation.
