# Slice 1: Ops Artifact Boundary

Status: reviewing
Slicing: chr Operational Artifacts to Ops
Plan: `ops/plans/2026-07-14-chr-operational-artifacts-to-ops-plan.md`
Tracker: `ops/trackers/2026-07-14-chr-operational-artifacts-to-ops-tracker.md`
Branch: `fix/slice-1-ops-artifact-boundary`
PR: `#361`

## Purpose

Complete the corrective migration from legacy `.chr/default/...` operational artifact paths to canonical `ops/<namespace>/...` artifact envelopes, while keeping `.chr/` available for CHR-private project configuration and state.

## Delivered Behavior

- Active dump, quiz, spec, plan, tracker, acceptance-feature, gate-report, workspace graph, TUI document, and Web UI document flows now use canonical ops artifact bodies.
- Gate reports write Markdown to `ops/default/report/<id>/artifact.md` and JSON evidence to `ops/default/report/<id>/evidence/gate-report.json`.
- Workspace graph and document discovery read canonical `ops/default/{dump,quiz,spec,plan,tracker,ticket,report}/<id>/artifact.md` bodies and ignore envelope metadata/attachments as documents.
- Tracked operational artifacts and QA reference artifacts were moved into ops envelopes with metadata.
- `.chr/default` is retained only for CHR-private project material such as config, case/project metadata, protocol assets, and runtime Dockerfile state.

## Implementation Notes

- Added deterministic named ops artifact helpers in `internal/core/ops_named_artifact.go` and body-path classification helpers in `internal/core/ops_artifact_paths.go`.
- Retargeted app-level dump creation/resolution from flat `.chr/default/dumps/*.md` files to ops dump envelopes.
- Retargeted generated specs/plans/trackers, quiz artifacts, acceptance feature artifacts, authoring artifacts, and gate reports.
- Updated app/core/TUI tests and QA scripts for canonical ops paths.

## Contracts Added or Changed

- Operational Markdown artifacts are addressed as `ops/<namespace>/<type>/<id>/artifact.md`.
- Gate report JSON is report evidence, not a sibling flat report file.
- `ops/default` is the default no-user namespace.
- `.chr/` remains private chr configuration/state, not a shared operational artifact root.

## Files of Interest

- `internal/core/ops_named_artifact.go`
- `internal/core/ops_artifact_paths.go`
- `internal/app/service_dump.go`
- `internal/core/gate_report.go`
- `internal/core/workspace_graph.go`
- `internal/tui/documents.go`
- `CHANGELOG.md`
- `scripts/qa/gate.sh`
- `scripts/qa/smoke.sh`

## Validation

- `make verify` — passed.
- `GOCACHE=/tmp/chr-ops-boundary-go-cache go test ./...` — passed.
- `python3 /home/adrian/Projects/ops/default/spec/scripts/validate_ops_layout.py --root .` — passed.
- `git diff --check` — passed.
- Active operational `.chr/default` text scan across `internal`, `docs`, `scripts`, `qa`, and `CHANGELOG.md` — no matches.

## Risks and Follow-ups

- Historical legacy snapshots remain under `ops/default/report/.../files/legacy/...` as preserved evidence, not active discovery roots.
