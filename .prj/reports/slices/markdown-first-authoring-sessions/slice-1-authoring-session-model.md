# Slice 1: Authoring Session Model

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-1-authoring-session-model`
PR: #91

## Purpose

Define the side-effect-free model for Markdown-first authoring sessions before
storage, templates, CLI commands, or promotion behavior are added.

## Delivered Behavior

The repository now has an authoring session model that can represent:

- session identity, project identity, status, mode, target, topic, and times;
- Markdown-direct, static, assisted, and TUI-backed session modes;
- dump, quiz, draft, spec, plan, slice-set, slice, tracker, and ticket targets;
- input and output artifact references;
- provenance links between source and generated or promoted artifacts;
- validation errors for invalid session metadata, artifacts, provenance, and
  time ordering.

## Implementation Notes

`ValidateAuthoringSession` is intentionally side-effect free. It validates the
in-memory contract only and does not read or write files.

This slice includes future-compatible modes for assisted and TUI-backed
authoring so later slices can reuse the same backend without changing the
stored domain vocabulary. It does not implement assisted execution or TUI
behavior.

## Contracts Added or Changed

New internal contracts:

- `AuthoringSession`
- `AuthoringSessionStatus`
- `AuthoringSessionMode`
- `AuthoringTarget`
- `AuthoringArtifactRef`
- `AuthoringArtifactKind`
- `AuthoringArtifactRole`
- `AuthoringProvenanceLink`
- `AuthoringProvenanceRelation`
- `ValidateAuthoringSession`

## Files of Interest

- `internal/chr/authoring_session.go`
- `internal/chr/authoring_session_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chr -run 'TestValidateAuthoringSession'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice defines the contract only. It does not persist sessions, render
Markdown, generate questionnaires, expose CLI commands, or promote artifacts.
