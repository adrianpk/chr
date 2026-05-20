# Slice 8: Execution Readiness and Handoff

Status: delivered
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-8-context-readiness-handoff`
PR: #85

## Purpose

Block low-quality context handoffs before Codex starts and expose reviewable
context evidence for humans and later agents.

## Delivered Behavior

The repository now has a readiness and handoff model that can:

- define required context source kinds;
- report missing ticket, protocol, project, document, validation, or runtime
  fields;
- compute context completeness as passed checks over total checks;
- carry composition warnings into the final handoff;
- link main and sibling context artifacts with host and container paths;
- render a handoff Markdown artifact;
- document mounted-workspace constraints and future PAT/worktree push behavior.

## Implementation Notes

The default required source kinds are agent identity, protocol, project, and
document context. Callers may override the required kinds when a narrower or
specialized execution mode is needed.

This slice does not introduce clone/worktree execution or PAT-backed push. It
records those as explicit remaining runtime-workspace concerns so they are not
lost after the context hardening set closes.

## Contracts Added or Changed

New internal contracts:

- `ContextReadinessHandoffOptions`
- `ContextReadinessHandoff`
- `ContextCompleteness`
- `ContextReadinessIssue`
- `ContextArtifactLink`
- `EvaluateContextReadinessHandoff`
- `RenderContextReadinessHandoff`

## Files of Interest

- `internal/chc/context_readiness_handoff.go`
- `internal/chc/context_readiness_handoff_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestEvaluateContextReadinessHandoff'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice provides the readiness/handoff contract. Wiring it into the live
ticket execution command path remains an integration step after this context
hardening slicing is delivered.
