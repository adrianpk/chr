---
title: "Dev To Main Gate Context"
status: "curated"
type: "dump"
case_slug: "dev-main-gate"
---

# Dev To Main Gate Context

The dev-to-main gate collects objective promotion evidence before aligning `dev` into `main`.

It records Git evidence, validation evidence, ChrysoCoord execution evidence, manual or external changes, missing evidence, warnings, and failures. Reports are written as JSON and Markdown under `.chc/user/reports/gates/`.

The default adoption posture for existing projects is permissive:

- run validation;
- list manual or external changes;
- warn when configured;
- do not fail on missing agent evidence until new work is consistently produced through ChrysoCoord.

Woodpecker is the first Forgejo feedback path. A generated Woodpecker workflow runs the gate and exposes pass/fail state through CI status.

Primary sources:

- `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
- `docs/reference/dev-main-gate.md`
- `docs/explanation/dev-main-gate-model.md`
- `docs/how-to/use-dev-main-gate-with-woodpecker.md`
