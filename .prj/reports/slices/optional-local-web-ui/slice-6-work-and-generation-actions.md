# Slice 6: Work And Generation Actions

Status: reviewing

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-6-work-and-generation-actions`

Expected PR Title: `feat(slice-6): work and generation actions`

PR: `#348`

## Action Inventory

Approved first Web UI generation actions:

- `spec.generate`
  - Service: `GenerateSpec`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: generated spec artifact.
- `plan.generate`
  - Service: `GeneratePlan`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: generated plan artifact when no canonical plan exists.
- `plan.regenerate`
  - Service: `RegeneratePlan`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: archives/replaces generated plan.
- `plan.tracking.create`
  - Service: `MaterializePlanTracking`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: tracking artifact and generated tickets.
- `quiz.start`
  - Service: `StartQuiz`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: quiz runtime/raw/curated artifacts.
- `quiz.resume`
  - Service: `ResumeQuiz`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: none by itself.

Approved first Web UI work actions:

- `work.preview`
  - Service: `PreviewAgentWorkLoop`.
  - Scope: ready-work queue.
  - Confirmation: required.
  - Writes: none.
- `work.daemon.stop`
  - Service: `RequestDaemonStop`.
  - Scope: current daemon runtime.
  - Confirmation: required.
  - Writes: daemon stop request state.

Deferred Web UI actions:

- `work.start.selected`, `work.start.ready`, and multi-ticket work loops stay
  out of this slice because they can run agent development and apply results.
- Quiz answer, skip, complete, cancel, and mode-switch actions stay out of this
  slice because they require a fuller active-question interaction surface.
- Arbitrary shell commands are not exposed.

## Summary

- Added approved Web UI actions for spec generation, plan generation,
  plan regeneration, plan tracking creation, quiz start/resume, work preview,
  and daemon stop requests.
- Added operation-state fields to Web UI action results.
- Added command surfaces to Specs, Plans, Quizzes, and Dash.
- Kept agent work-starting actions out of scope and explicitly blocked
  unsupported work-start requests.

## Changes

- `internal/app/web_actions.go`
  - Adds grouped dispatch for document, generation, and work actions.
  - Calls existing application services only: `GenerateSpec`, `GeneratePlan`,
    `RegeneratePlan`, `MaterializePlanTracking`, `StartQuiz`, `ResumeQuiz`,
    `PreviewAgentWorkLoop`, and `RequestDaemonStop`.
  - Returns operation kind, status, id, and detail in the stable write
    envelope.
- `internal/app/web_views.go`
  - Adds an internal service factory hook so tests can run Web UI actions with
    service-level fakes while production keeps using `newService`.
- `internal/webui/views.go`
  - Adds operation-state fields to write responses.
- `internal/webui/assets/templates/shell.html`
  - Adds command forms for Specs, Plans, Quizzes, work preview, and daemon stop.
- `internal/webui/assets/static/chc.css`
  - Adds command-strip and select styling.
- `internal/webui/webui_test.go`
  - Covers command form rendering on Specs, Plans, Quizzes, and Dash.
- `internal/app/cli_test.go`
  - Covers service-backed spec generation, work preview without execution, and
    unsupported action rejection.

## Validation

- `go test ./internal/webui ./internal/app -run 'WebServe|Handler|WriteAction|Generation|WorkAction'` (pass)
- `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Temporary-project smoke with `go run ./cmd/chc web serve --dir /tmp/tmp.IJQmMJO7zb --port 7336` plus:
  - `curl -fsS http://127.0.0.1:7336/specs` found `spec.generate` (pass)
  - `curl -fsS http://127.0.0.1:7336/plans` found `plan.tracking.create` (pass)
  - `curl -fsS http://127.0.0.1:7336/quizzes` found `quiz.start` (pass)
  - `curl -fsS http://127.0.0.1:7336/dash` found `work.preview` (pass)
  - `curl -X POST http://127.0.0.1:7336/actions` for `work.preview`
    returned `303` (pass)
  - `curl -X POST http://127.0.0.1:7336/actions` for unsupported
    `work.start.ready` returned `400` with an unsupported-action envelope
    (pass)
- `make verify` (pass)

## Notes

- The action inventory is committed before endpoint expansion as required by
  the slice tracker.
- Work-starting actions remain deferred; this slice exposes work preview and
  daemon stop only.
