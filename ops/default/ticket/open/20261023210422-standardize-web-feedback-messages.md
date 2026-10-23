---
id: TKT-20261023210422
title: "Standardize shared Web UI feedback messages"
status: open
kind: follow_up
severity: unclassified
priority: unclassified
scope: ui
tags: web-ui, feedback, ux
source: chat
reported_at: 2026-10-23T21:04:22Z
commits: []
---

# Standardize shared Web UI feedback messages

## Problem

Web UI operations present errors, warnings and confirmations inconsistently.
The shared `WriteActionResponse` envelope and `write-result` presentation
already exist, but Quiz, Specs, Plans, tickets, harness selection and workflow
screens also render their own feedback. Severity colors are not applied
consistently across those surfaces.

A guided workflow can end with "This step needs attention. Your progress is
saved." after several minutes of processing. The message does not identify the
failed operation, its cause or the action needed to continue. Retry does not
explain which retained operation it repeats.

## Expected Outcome

Extend the existing shared feedback mechanism instead of creating a separate
Wizard-specific error presentation. All Web UI surfaces should use consistent
messages with three semantic levels:

1. Red: an operation failed, with a concrete cause and useful recovery guidance.
2. Yellow: a warning, missing prerequisite or partial result requiring attention.
3. Green: a confirmed successful outcome of the operation being described.

## Scope

1. Inventory existing feedback producers and result templates, then consolidate
   their presentation around the current operation-result contract.
2. Define consistent severity styling, wording, placement and accessibility.
   Color must accompany meaningful text rather than carry the meaning alone.
3. Keep blocking errors and actionable warnings visible until resolved or
   explicitly dismissed where dismissal is appropriate. Do not silently
   auto-dismiss an impediment to continuing work.
4. Identify the failed or retried operation in human terms, explain what is
   preserved and retain the existing guarded recovery actions.
5. Preserve useful diagnostics without placing raw internal state, opaque
   identifiers or credentials in routine user-facing messages.

## Acceptance Criteria

1. The Wizard, manual authoring screens and other Web UI operation surfaces
   share one severity and presentation convention based on the existing result
   mechanism.
2. Error, warning and success examples are visually consistent in full pages,
   dialogs and HTMX updates, on desktop and narrow viewports.
3. A failure to start development because a required Git remote is missing
   identifies that operation and prerequisite, and states that the prepared
   Spec, Plan and ticket remain saved.
4. Retry messages identify the retained operation without implying that earlier
   completed generation must be repeated or that a retry repaired configuration.
5. Success messages report only the outcome actually confirmed. Admission,
   queued work, validation and an open PR do not imply delivered implementation.
6. Validation errors, stale writes, blocked operations, partial results and
   confirmed writes preserve their existing application semantics and bindings.
7. Accessible labels and live-region behavior convey severity without relying
   solely on red, yellow or green.

## Starting Evidence

1. `internal/webui/views.go`: `WriteActionResponse` and `WriteStatus*`.
2. `internal/webui/assets/templates/shell.html`: shared `write_result.html`
   and Dump result presentation.
3. `internal/webui/assets/templates/quiz.html`, `spec.html`, `plan.html`,
   `ticket_editor.html`, `harness.html` and `workflow.html`: surface-specific
   feedback variants.
4. `internal/webui/assets/templates/guided_use_case.html`: generic recovery
   message and retained-step retry.
5. `internal/webui/assets/static/chr.css`: existing danger/amber palette,
   neutral result styling and isolated status treatments.

## Validation

Exercise the shared result mapping and real rendered feedback across the
inventoried surfaces. Verify recoverable failures, partial results and confirmed
success, including an asynchronous workflow failure. Inspect desktop and narrow
browser views and confirm that stale bindings, draft retention and retry
identity remain protected.

## Execution Boundary

Captured for later work. This ticket does not begin implementation or expand
the current Wizard refinement. Git prerequisites, local integration and remote
publication policy remain separate workflow concerns.
