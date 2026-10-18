---
id: TKT-20261017194552
title: "Present individual spec cards and allow choosing the initial state"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: specs, catalog, creation-state
source: chat
reported_at: 2026-10-17T19:45:52Z
commits: []
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Present individual spec cards and allow choosing the initial state

## Context and Evidence Boundary

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
The maintainer supplied screenshots to explain the desired card presentation
and creation entrypoint for Specs. The requirements below describe that target
independently. The actual CHR Specs screen has not yet been visually reviewed
in this walkthrough; do not invent current-screen defects from the examples.

Capture only. No application change, creation of a product spec, lifecycle
mutation or service update is authorized by this development ticket.

## Spec Catalog Presentation

1. Treat each card as one individual spec, identified by that spec's title.
   A sample title that happens to be Specs is not a category or collection.
2. Put the title above a bounded inset area that looks like the beginning of
   a small document. Show the actual first lines of the spec body, with
   readable line structure and bounded overflow.
3. Do not substitute generated summaries, artifact paths or technical metadata
   for the opening document text. Long documents must not expand cards into
   full-page document views.
4. Keep a coherent card grid, with consistent dimensions, spacing, borders
   and title/preview hierarchy. This describes the visual composition without
   requiring another product's palette or decorative details.
5. The page is Specs. Do not introduce separate Drafts, Specs and Plan Trackers
   category cards or require the user to enter a different document category
   merely to create a spec in a particular state.
6. Clicking a spec card opens that exact spec for editing, with its stored
   content prefilled and a Save action. This applies to both manually authored
   and generated specs; the catalog is the entrypoint to their editor.

## New Action and Initial State

1. Use one New action at the top right of the Specs page, aligned with the
   section heading. The section already supplies the document type.
2. New opens the spec-creation modal. One compact Draft switch or themed
   checkbox lets the user choose draft or approved in that same form.
3. Honor the chosen state on creation. A user who selects approved can create
   an approved spec directly; do not force a draft-first creation followed by
   a separate approval or promotion step.
4. A draft is a state of a spec rather than a separate first-class document
   type, creation route or mandatory precursor artifact.
5. Persist the selected state and any required native lifecycle metadata
   through the shared spec owner. The behavior must be real, not just a label
   displayed by the Web adapter while storage silently uses draft.
6. Map user-facing state choices to the existing canonical status vocabulary
   when refining the implementation. Do not invent a parallel Web-only state
   model or silently turn an approved selection into a draft.

## Creation Modal

The creation modal is one simple vertical form. Use this conceptual structure:

```text
New Spec                                                   [x]

Title
[ Single-line title                                         ]

Use Case
[ Project-wide or an existing use case                     v ]

Body
+-----------------------------------------------------------+
|                                                           |
| One large editable spec body                              |
|                                                           |
|                                                           |
+-----------------------------------------------------------+
Draft [ on / off ]                                  [ Create ]
```

1. Title is a full-width single-line input.
2. Use Case is a selector containing Project-wide and existing use cases,
   identified by their names. Project-wide means no particular case
   association; do not require typing an internal case slug.
3. At most one compact Draft control supplements those fields. Use a switch
   or checkbox styled with the application theme, rather than a separate
   full-width State selector or lifecycle panel. On means draft; off permits
   direct approved creation through the native spec lifecycle.
4. Body is one large multiline editor below the small metadata fields and
   occupies most of the form. Do not add a simultaneous read-only copy of
   the same text or require a separate append form.
5. Create is a compact action at the bottom right, below the body editor.
   Creating an approved spec uses this same action, without an extra draft
   creation or promotion screen.
6. Keep the layout as one coherent form with a normal modal close control,
   consistent field spacing and enough room to write the document. Do not
   add document-category navigation or additional workflow panels inside it.

These requirements describe the conceptual arrangement; they do not require
an external palette, typography package or exact decorative reproduction.
The creation target and editing behavior are specified below. CHR's existing
spec editor has not yet been visually reviewed in this walkthrough.

## Manual and Generated Specs

1. A user can write a spec manually for the project or an existing use case.
   Manual creation does not require a dump, a completed quiz or an LLM request.
2. CHR also generates specs from the available context: project context plus
   the relevant dump, if present, and the quiz for the selected scope, if
   present. Missing optional dump or quiz content must not become an
   artificial prerequisite. Keep unrelated use-case material out of the
   selected use case's context.
3. Manual and generated specs use the same canonical spec model and appear
   in the same catalog. Generation is another authoring path, not a separate
   document type or a requirement of the manual New form.
4. Users can open and edit the text of either kind of spec, including one
   generated by CHR. They can correct errors, supply omitted details and
   expand the document. Save changes to that spec and its existing scope;
   do not require recreating it as a manually authored copy.
5. Editing provides one editable body and a Save action. Do not duplicate the
   same body in a read-only panel or require a separate append interaction.

This capture does not define a new automatic-generation trigger or require
generation controls in the minimal manual creation modal. Direct approved
creation is an explicit user choice; it does not imply automatic approval of
generated specs.

## Shared Creation Boundary

Inspect the shared spec creation, generation, editing and persistence paths
when implementation is explicitly authorized. The shared service already
exposes GenerateSpec in internal/app/service_spec.go; verify its actual context
assembly against the requirements above. If manual creation, editing or direct
approved-state persistence needs domain/service work, include that work rather
than implementing only cosmetic Web controls. Do not impose model-generated
authoring or an artificial draft/promotion ceremony that the requested flow
does not require.

## Acceptance

1. Each catalog card is an individual spec with a title and its actual opening
   text in a bounded inset preview. Verify short and long document bodies.
2. One New action appears in the Specs heading region. There is no New Draft
   action or mandatory separate Draft collection in this flow.
3. The creation modal offers one compact Draft switch or themed checkbox.
   Create one draft and one approved spec through that same entrypoint and
   verify their retained native states after reload.
4. Direct approved creation does not create an extra precursor draft or force
   a later interactive approval transition.
5. The catalog includes the created specs as individual cards, independent of
   their initial state, with opening text drawn from their stored document body.
6. The creation modal has Title, Use Case, one large Body editor, the compact
   Draft control and Create at the bottom right. Verify both Project-wide and
   an existing case destination without requiring a typed case slug.
7. Manual creation works without a dump, a quiz or a model call. Verify that
   generation receives project context and the relevant optional dump and quiz
   when available, without including another use case's material.
8. Click a generated spec's catalog card and verify that it opens that spec's
   editor with the stored content prefilled. Correct and extend its text, save
   and reload, then verify the changes on the same spec and scope. Its catalog
   preview reflects its current opening text. Repeat through a manual spec's
   card to verify the same editing behavior and correct document selection.
9. Capture visual evidence of the catalog, the New entrypoint and the selected
   modal state when implementation is authorized. Validate the shared state
   and scope persistence as well as the visible UI.

## Consolidated Design

Selected for coordinated Web UI specification work on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. The current
request authorizes classification, this in-progress status, specification and
commit; application implementation remains a separate step.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/drafts/web-ui-workflow-consolidation-draft.md).
