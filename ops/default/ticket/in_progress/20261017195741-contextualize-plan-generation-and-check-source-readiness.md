---
id: TKT-20261017195741
title: "Generate plans from use-case context with sufficient source material"
status: in_progress
kind: task
severity: medium
priority: normal
scope: domain
tags: plans, generation, context
source: chat
reported_at: 2026-10-17T19:57:41Z
commits: []
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Generate plans from use-case context with sufficient source material

## Context and Evidence

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
The supplied Plans HTML places separate Generate Plan and Regenerate Plan
forms below the catalog. Both ask the user to type a Case slug. These are
contextless operational forms in a document-listing screen.

The maintainer wants generation associated with a use case and based on
meaningful discovery material. Additional dump content or quiz answers can
improve the spec and inform a new plan. Generation and regeneration can be
useful operations; their placement and prerequisites must reflect that workflow.

Capture only. Do not implement, invoke generation or update services here.
Tracking creation, tracking behavior and ticket materialization are outside
this ticket and will be discussed separately.

## Use-Case Context and Actions

1. Place generation and regeneration in the relevant use-case context rather
   than as generic forms below the Plans catalog. The action must make the
   selected use case and the document being generated clear.
2. Use the current selection or a selector of existing use-case names. Do not
   require the user to discover or type an internal slug. Internal identifiers
   may still be used to submit the selected scope.
3. Generate creates a plan for the selected use case; Regenerate updates an
   existing plan for that same use case. Keep these actions contextual instead
   of exposing both as unrelated global commands in a listing.

## Source Prerequisites and Sufficiency

1. Do not generate a plan for a use case with neither meaningful dump content
   nor quiz answers. At least one of those sources must supply usable context.
   A placeholder file or an unanswered quiz is not sufficient evidence.
2. Source presence and source sufficiency are separate checks. A nonempty dump
   alone does not automatically establish that enough is known to plan the
   work. Evaluate the available information against the use case's scope.
3. Both sources are not mandatory. A sufficiently complete dump can make a
   quiz unnecessary. A short dump can also suffice for a narrowly scoped use
   case; document length alone does not establish readiness.
4. When information is insufficient, explain what is missing in readable terms
   and direct the user to add context or answer relevant questions. Do not
   generate a plan by filling important gaps with unsupported assumptions.
5. Enforce the prerequisite through the shared generation owner as well as the
   Web action. Apply it to generation and regeneration before producing or
   replacing a plan; a disabled browser button alone is not enforcement.

## Regeneration After Additional Context

1. Support the maintainer's primary regeneration scenario: a dump is expanded
   or further quiz questions are answered, making more information available
   for the same use case.
2. Keep spec and plan generation distinct. Additional discovery material can
   improve the spec, which in turn informs planning. A Regenerate Plan label
   must not conceal a different document operation or silently imply that
   refreshing the plan also refreshed the spec.
3. When refining this ticket, verify how the current spec, dump and quiz are
   included in shared plan generation. Define the contextual entrypoint and
   the sufficiency decision before moving the ticket to ready. Do not invent
   an automatic cascade or arbitrary minimum word count during implementation.

Manual spec creation remains governed by the
[Specs authoring ticket](20261017194552-present-spec-cards-and-user-selected-creation-state.md).
This ticket's plan-generation prerequisite does not impose dump or quiz
requirements on writing a spec manually. Catalog appearance and read-only plan
viewing are covered by the separate
[Plans catalog ticket](20261017195458-present-plans-and-tracking-as-read-only-cards.md).

## Acceptance

1. Generate and Regenerate operate on the selected named use case without a
   typed slug. The Plans listing has no standalone slug-based generation forms.
2. No useful dump and no quiz answers prevents generation. An insufficient
   dump also prevents it and produces an understandable next step.
3. A comprehensive dump without a quiz, a sufficient short dump for a small
   use case, and sufficient quiz answers without a dump are valid scenarios.
   Do not require both source types or apply a fixed document-length threshold.
4. Added dump content or quiz answers are available to the relevant subsequent
   generation operation. Verify which document changes and which context it
   uses, distinguishing spec improvement from plan regeneration.
5. A failed readiness check does not generate or replace the selected plan.
   Verify the shared operation and the browser entrypoint when implementation
   is authorized. Do not create or modify tracking artifacts for this ticket.

## Consolidated Design

Selected for coordinated Web UI specification work on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. The current
request authorizes classification, this in-progress status, specification and
commit; application implementation remains a separate step.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/drafts/web-ui-workflow-consolidation-draft.md).
