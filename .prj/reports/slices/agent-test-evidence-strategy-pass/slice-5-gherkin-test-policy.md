# Slice 5: Gherkin Test Policy

Status: delivered
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-5-gherkin-test-policy`
PR: `#324`

## Purpose

Make the Test/Evidence Strategy pass explicit about when Gherkin applies, when
it does not apply, and which stronger test evidence defaults should influence a
worker before implementation starts.

This slice keeps the strategy pass inline in the existing worker context. It
does not generate Gherkin automatically and it does not add a second Codex
container.

## Delivered Behavior

- Added project-level Gherkin detection based on existing
  `.chr/user/acceptance/*.feature` artifacts.
- Added ticket-level Gherkin applicability for observable behavior tickets.
- Added default evidence guidance for change kind, risk, test timing, tests,
  and Gherkin.
- Added the evidence decision matrix to worker context.
- Worker context now shows active Gherkin state, required feature path, and
  suggested strategy before implementation.
- Review evidence readiness now rejects `Gherkin: not_applicable` when active
  Gherkin applies to the ticket.
- App result application and direct status gating pass the real command/service
  context through ticket evidence readiness.
- Acceptance/Gherkin reference docs now describe the evidence strategy
  interaction.

## Policy Notes

- Gherkin is active only when `gherkin_policy` is `active_when_applicable` and
  the project has existing `.feature` artifacts.
- Gherkin applies to user-visible CLI, TUI, API, domain, acceptance workflow,
  and user-visible bugfix behavior.
- Docs-only, Godoc-only, metadata-only, internal refactor, and low-level
  parser or validator tickets can still record Gherkin as `not_applicable`
  when no observable behavior scenario is being specified.
- The decision matrix is guidance for the worker, not a universal TDD mandate.

## Files of Interest

- `internal/core/evidence_policy.go`
- `internal/core/evidence_policy_test.go`
- `internal/core/codex_worker_context.go`
- `internal/core/codex_worker_context_test.go`
- `internal/core/evidence_readiness.go`
- `internal/core/evidence_readiness_test.go`
- `internal/app/ticket_evidence_readiness.go`
- `internal/app/cli_test.go`
- `docs/reference/acceptance-feature-generation.md`
- `docs/reference/evidence-strategy-config.md`

## Validation

- `go test ./internal/core -run 'Evidence.*Policy|EvidenceReadiness|CodexWorkerContext'` (pass)
- `go test ./internal/app -run 'TestTicketApplyResult'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The worker is guided to update `.feature` artifacts but this slice does not
  generate or execute Gherkin.
- Evidence status is still mostly visible through runtime context and ticket
  state; Board/Dash visibility remains in the next slice.
