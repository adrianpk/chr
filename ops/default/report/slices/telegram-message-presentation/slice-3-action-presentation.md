# Slice 3: Clear Guarded Confirmations and Controls

Status: drafting
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-3-telegram-action-presentation`
PR: pending

## Purpose

Present clear scope and consequences for Telegram confirmations and controls,
while preserving the original operation, document, destination and action guards.

## Delivered Behavior

None. T3.1-T3.3 have not started because the activation prerequisite is blocked.

## Implementation Notes

PR #425 is verified merged into `dev` at
`a5c6a7e068a8e5ec547d13d73d39afe1f0477335`. Its report and tracker entries
are closed. This branch records that closure and Slice 3 activation only.

The approved plan's section 2 requires canonical integration of the earlier
preparation/admission and executor-handoff corrections before Slice 3. Their
canonical branch `test/slice-8-telegram-delivery` still points to
`5bde31118db47c4fe52bff416555049a66d22911`, outside `dev` ancestry.
The canonical forge returned no open PRs; the retained earlier tracker still
records Slice 8 as blocked. The current delivery set cannot copy those
corrections or take over that controller.

## Contracts Added or Changed

None. Implementation awaits the recorded dependency.

## Files of Interest

- `ops/default/plan/telegram-message-presentation.md`: section 2 activation prerequisite.
- `ops/default/tracker/telegram-message-presentation.md`: Slice 2 closure and Slice 3 blocker.
- `ops/default/report/slices/telegram-message-presentation/slice-2-document-delivery.md`: verified integration and exact-head evidence.

## Validation

`git merge-base --is-ancestor 5bde31118db47c4fe52bff416555049a66d22911 dev`
returned exit 1 on 2026-10-21. Canonical refs agree with the local identities.
This establishes an unmet dependency, not a failed runtime check.

Local Markdown links and `git diff --check` passed for these documentation
updates. No Slice 3 runtime or race checks ran, and no PR is ready.

## Risks and Follow-ups

Verify the earlier corrections' canonical integration before starting T3.1.
Preserve its controller, live QA invocations and service ownership. The Slice 2
validation receipt belongs only to its original merged candidate.
