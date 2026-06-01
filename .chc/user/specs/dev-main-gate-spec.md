---
title: "Dev To Main Gate"
project_id: "chrysocoord"
case_slug: "dev-main-gate"
status: "accepted"
source_request_id: "self-adoption-baseline"
source_response_id: "existing-ops-spec"
created_at: "2026-06-01T00:00:00Z"
---

# Dev To Main Gate

ChrysoCoord needs a dev-to-main promotion gate that checks objective evidence before accumulated work is promoted.

The gate must be tooling-first and agent-aware. It should not behave like a generic PR reviewer, and it should not require every project change to have a ChrysoCoord ticket. The board is primarily for agent-managed work. Human or external changes may exist outside ChrysoCoord and must not produce warnings or failures by default.

## Expected Behavior

- Collect Git range evidence for a base/head pair.
- Run the configured validation command.
- Write Markdown and JSON reports under `.chc/user/reports/gates/`.
- Classify ChrysoCoord-managed work separately from manual or external changes.
- Fail when validation fails.
- Fail when ChrysoCoord-managed work lacks required evidence under strict policy.
- Keep manual or external changes informational by default.
- Support `.chc/user/config/gates.yml`.
- Provide a Woodpecker setup helper.

## Adoption Policy

Existing projects should start with missing agent evidence as non-blocking and manual/external changes visible. After new work is consistently produced through ChrysoCoord, projects can enable stricter traceability.

## Provenance

- source: ops/specs/2026-06-01-dev-main-agent-evidence-gate.md
- source: docs/reference/dev-main-gate.md
- source: docs/explanation/dev-main-gate-model.md
