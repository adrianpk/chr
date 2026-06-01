# ChrysoCoord Self-Adoption Baseline

Status: active

ChrysoCoord is being adopted into its own repository as a managed project without deleting the existing `.prj/` delivery history.

## Boundary

- `.chc/user/` is the active ChrysoCoord workspace for this repository.
- `.prj/` remains historical internal delivery context.
- `ops/` remains existing spec, plan, and tracker history.
- No execution reports are fabricated during adoption.
- No quizzes are fabricated during adoption.

## Initial Policy

The initial dev-to-main gate policy is intentionally informational for traceability:

- validation command: `make verify`
- manual or external changes: listed and warned
- missing agent evidence: not blocking during adoption

This avoids treating pre-adoption work as missing ChrysoCoord evidence.

## Baseline Sources

The initial project and use-case metadata was inferred from:

- `README.md`
- `docs/explanation/architecture.md`
- `docs/explanation/document-generation-model.md`
- `docs/explanation/ticket-resolution-lifecycle.md`
- `docs/explanation/agent-sandbox-and-runtime.md`
- `docs/explanation/dev-main-gate-model.md`
- `docs/reference/dev-main-gate.md`
- `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
- `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
- `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`

## Next Adoption Step

Future work can start creating real tickets, specs, plans, and execution evidence under `.chc/user/`. Once new work is consistently executed through ChrysoCoord, the gate policy can be tightened to fail on missing agent evidence.
