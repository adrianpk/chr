# Slice 7: Recovery and User Operation

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-7-telegram-recovery`

PR: pending

## Purpose

Recover retained Telegram work and operate the receiver across restart,
shutdown, retry exhaustion and retention pressure.

## Delivered Behavior

Implementation in progress. Slice 6 is verified integrated through PR #395.

## Implementation Notes

T7.1 is active. T7.2 and T7.3 remain pending in the approved order.

## Contracts Added or Changed

Pending implementation evidence.

## Files of Interest

Pending implementation evidence.

## Validation

Activation documentation passed `git diff --check`. Final slice evidence will
use the configured ordered commands at the exact clean PR HEAD.

## Risks and Follow-ups

Startup/shutdown, retention/retry and operational tooling remain in scope for
this slice. Live journey certification belongs to the next recorded slice.
