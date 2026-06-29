# Slice 3: Authoring Artifact Layout

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-3-authoring-artifact-layout`
PR: #93

## Purpose

Define canonical project-local paths and inventory behavior for generated or
promoted authoring artifacts.

## Delivered Behavior

The repository now has an authoring artifact layout model that can:

- resolve canonical `.chr/` paths for dumps, quizzes, drafts, specs, plans,
  slice sets, slices, trackers, and tickets;
- derive stable artifact IDs from kind, session ID, and topic or name;
- return absolute filesystem paths and relative project paths;
- detect existing artifact collisions;
- allow explicit replacement planning without mutating files;
- render inventory entries that record session ID, artifact ID, kind, role,
  path, existence, and collision status.

## Implementation Notes

`PlanAuthoringArtifactLayout` validates inputs and inspects the target path for
collision state. It does not create directories or write files.

The path layout is intentionally independent from static quiz generation and
promotion. Later slices can reuse this contract when they need predictable
artifact destinations.

## Contracts Added or Changed

New internal contracts:

- `AuthoringArtifactLayoutOptions`
- `AuthoringArtifactLayout`
- `AuthoringArtifactCollision`
- `AuthoringArtifactInventoryEntry`
- `PlanAuthoringArtifactLayout`
- `AuthoringArtifactInventory`

## Files of Interest

- `internal/chr/authoring_artifact_layout.go`
- `internal/chr/authoring_artifact_layout_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chr -run 'TestPlanAuthoringArtifactLayout|TestAuthoringArtifactInventory'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice plans artifact paths only. It does not generate quiz content, create
artifact files, expose CLI commands, or promote one artifact kind into another.
