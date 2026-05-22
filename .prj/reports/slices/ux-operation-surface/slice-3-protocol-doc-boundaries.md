# Slice 3: Protocol Document Boundary Cleanup

Status: delivered
Slicing: UX Operation Surface
Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
Branch: `feat/slice-3-protocol-doc-boundaries`
PR: #136

## Purpose

Separate generated project protocol guidance from ChrysoCoord operational
workspace contracts.

## Delivered Behavior

- Generated `DOCUMENTING.md` now focuses on durable project documentation.
- Generated `DOCUMENTING.md` no longer defines `.chc/` operational artifact
  placement.
- Generated `ARCHITECTURE.md` treats `.chc/` as tool-managed project state
  instead of application package layout.
- Generated `INVARIANTS.md` avoids naming `.chc/` layout as an editable
  protocol rule.
- `docs/reference/project-workspace.md` now states that operational artifact
  placement is a product contract owned by `chc`.

## Implementation Notes

- The existing context composition candidate list did not require code changes:
  `.chc/manifest.md`, `.chc/config.yml`, runtime Dockerfile, and `.chc/agents`
  files remain valid context inputs.
- The behavioral change is in generated protocol content, not in context source
  selection.
- CLI init tests now assert that generated protocol files do not reintroduce
  the old self-referential workspace rules.

## Contracts Added or Changed

- Project documentation guidance lives in generated `DOCUMENTING.md`.
- ChrysoCoord artifact layout is documented in product reference docs and
  enforced by code paths, not by generated documentation convention text.
- Generated protocol files may mention tool-managed workspace state, but they
  must not define `.chc/` layout as a project convention.

## Files of Interest

- `cmd/chc/assets/workspace/profiles/standard/agents/DOCUMENTING.md`
- `cmd/chc/assets/workspace/profiles/standard/agents/ARCHITECTURE.md`
- `cmd/chc/assets/workspace/profiles/standard/agents/INVARIANTS.md`
- `internal/chcapp/cli_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chcapp -run TestInitStandardProfileInstallsProtocolAssets` - pass
- `make verify` - pass

## Risks and Follow-ups

- Later slices may still refine how much protocol context is included in ticket
  execution prompts.
- Historical reports and plans may still describe old generated protocol
  content as historical context.
