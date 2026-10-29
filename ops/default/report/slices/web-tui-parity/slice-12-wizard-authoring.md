# Slice 12: Wizard Quiz and preparation

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-12-tui-wizard-authoring`
PR: [#466](https://forge.adrianpk.com/adrianpk/chr/pulls/466)

## Purpose

Resolve G-009 and G-010 through shared Quiz interaction, committed document
presentation and actionable retained authoring recovery in the existing tabs.

## Delivered Behavior

- Wizard Quizzes uses the manual Quiz renderer, textarea, history paging and
  controls. Enter submits, Shift+Enter/Ctrl+J adds a newline, Deep/Fast changes
  through the retained owner, and Skip, Done and Cancel retain their guards.
  Guided Done still requires trusted clarification sufficiency. Guided mode
  exposes no manual Restart. Cancellation or controls that discard an unsent
  answer require the shared confirmation; Esc preserves the complete input.
- Specs and Plans use the existing document detail component over canonical
  committed originals. Generated specification approval, plan publication,
  native ticket materialization and subsequent progression stay application
  owned. Original bytes, reference and commit remain independent of later
  working-copy replacements.
- Each authoring stage explains pending operations, saved input, retained
  specification/plan generation, partial ticket preparation, checkpoint
  acknowledgement, suspension and actual blockers. Eligible controls and
  Continue come from the existing guided owners. A failed acknowledgement
  remains an error, accompanied by the available retained observation.
- Reads preserve failed requests. Retry and restart reconcile the original
  operation and input rather than allocating another plan, ticket set or
  allowance. Older replies cannot replace newer run state. Quiz drafts retain
  their captured question, mode, revision, digest and operation through refresh.
  Alt+Y retries the original failed request; Alt+R confirms discarding the old
  answer before reading the current question. Cancelled reload writes nothing.

## Implementation Notes

T12.1: `6d43124624539535e1c5a1cf3736e21e25488b01`.
T12.2: `fc3e0563baf91c4fcb98bd6253d3473cff7e16ff`.
T12.3: `d21a31eb2820f2ae2030039cef8679fcdda13c1e`.
T12.4: `bfc05ca37ce59466d85010e7b274fde92c0a89d0`.
T12.5 report introduction: `517571b3e4556afa37c502a260435cdb268d95cc`.

Quiz transport mapping reads the neutral native observation and shared answer
history projection. UI requests capture a guided command and use the existing
ChangeGuidedUseCase operation. No Web DTO is transported into the terminal and
no terminal progression controller is introduced. Document tabs use the same
renderer as manual catalogs and canonical committed-original readers.

Real preparation tests inject failures only after actual owner effects or at
native contract generation. The deadline case returns DeadlineExceeded at the
publication acknowledgement boundary. Re-entry is read-only; explicit retry
retains the exact command, one generated plan and one native materialization.
Delayed stale answers are tested against a real concurrent revision change.
Canonical images, revisions, generation counts and complete original bytes are
asserted; provenance alone is not treated as semantic equivalence.

The existing managed production PTY now answers through the shared Quiz editor
with Enter. It waits for that editor's own control marker before typing. Exact
final capture, default approval, checkpoint count, later idea, committed
original and paused restart assertions remain. Its 45-second fixture and
40-second child limits are unchanged; only owned children are stopped.

## Contracts Added or Changed

Transient Quiz requests and drafts can carry a captured guided command. Guided
observations include the native Quiz projection; document projection selects
the canonical committed plan and tracker. Reload is an explicit read-only
request after shared confirmation. No domain command, schema, authority,
publication policy, stage semantics or finite budget changes.

## Files of Interest

- `internal/tui/quiz_actions.go`, `quizzes.go`, `wizard_quiz.go`: shared Quiz IO.
- `internal/tui/wizard_documents.go`: shared document rendering.
- `internal/tui/guided.go`, `wizard_recovery.go`, `operation_feedback.go`:
  captured retries, response ordering and shared confirmation/reload.
- `internal/app/tui_guided_quiz.go`, `tui_guided_authoring.go`,
  `tui_guided_feedback.go`: native observations, committed originals and state.
- `internal/app/tui_wizard_*test.go`, `internal/tui/wizard_recovery_test.go`:
  real retained IO, fault recovery, stale input and terminal evidence.
- User Guide, TUI reference and changelog: delivered controls and recovery.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including captured draft, retry,
  confirmation, reload and regressing/obsolete response checks.
- `go test ./internal/app -run '^Test(TUIWizard.*|GuidedUseCaseTUIJourney|GuidedUseCaseTUITerminal|GuidedUseCaseParityNative)$' -count=1 -timeout=3m` — passed for real persistence, faults, duplicate prevention, production PTY and native parity.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.
- `git diff --check` — passed before task commits.

Required `make verify`, focused race and Web reference checks, and final whitespace
evidence will be recorded by the controller on the clean PR-reference HEAD.
They are pending at report introduction. Integration and closure remain
controller-owned.

## Risks and Follow-ups

None recorded.
