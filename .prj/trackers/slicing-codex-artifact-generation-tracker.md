# Codex Artifact Generation Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-codex-artifact-generation-plan.md`

## Active Slice

Slicing name: Codex Artifact Generation

Active slice number: Slice 1

Active slice short name: plan tracking codex backend

Expected branch name: `feat/slice-1-plan-tracking-codex-backend`

Expected PR title: `feat(slice-1): plan tracking codex backend`

Expected report path:
`.prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md`

Prerequisite PRs or delivered work:

- `dev` includes shared Codex agent runner.
- `dev` includes Codex routing for spec and plan generation.
- `feat/codex-dump-generation` exists as a dump curation implementation branch
  and must be reconciled before or during default-policy cleanup.

## Slicing Checklist

- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with repository quality gates.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Plan Tracking Codex Backend | Reviewing | `feat/slice-1-plan-tracking-codex-backend` | `feat(slice-1): plan tracking codex backend` | `.prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md` | Implemented on `feat/slice-1-plan-tracking-codex-backend`; PR pending. |
| Slice 2: Ticket Payload Boundary | Planned | `feat/slice-2-ticket-payload-boundary` | `feat(slice-2): ticket payload boundary` | `.prj/reports/slices/codex-artifact-generation/slice-2-ticket-payload-boundary.md` | Depends on Slice 1. |
| Slice 3: Artifact Backend Defaults | Planned | `feat/slice-3-artifact-backend-defaults` | `feat(slice-3): artifact backend defaults` | `.prj/reports/slices/codex-artifact-generation/slice-3-artifact-backend-defaults.md` | Depends on Slices 1 and 2; reconciles dump curation branch/default policy. |
| Slice 4: Artifact Run Bundle Audit | Planned | `feat/slice-4-artifact-run-bundle-audit` | `feat(slice-4): artifact run bundle audit` | `.prj/reports/slices/codex-artifact-generation/slice-4-artifact-run-bundle-audit.md` | Depends on Slices 1 through 3. |

## Slice 1: Plan Tracking Codex Backend

Candidate tasks:

- [x] T1.1 Add plan tracking schema support to the Codex quiz model adapter.
- [x] T1.2 Add plan tracking routing/configuration in the service.
- [x] T1.3 Add tests that plan tracking can use the Codex route.
- [x] T1.4 Add failure-path tests for parse/validation failure without writes.
- [x] T1.5 Add slice implementation report.
- [x] T1.6 Run targeted tests and repository quality gates.

Task commits:

- Pending commit on `feat/slice-1-plan-tracking-codex-backend`.

Validation:

- `go test ./internal/chc ./internal/chcapp` - pass.
- `make verify` - pass.

## Slice 2: Ticket Payload Boundary

Candidate tasks:

- [ ] T2.1 Audit current plan tracking ticket payload validation.
- [ ] T2.2 Add or strengthen validation around ticket metadata and body fields.
- [ ] T2.3 Add tests proving invalid generated ticket payloads are rejected
      before file creation.
- [ ] T2.4 Add tests proving Go owns ticket paths and keys.
- [ ] T2.5 Add slice implementation report.
- [ ] T2.6 Run targeted tests and repository quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 3: Artifact Backend Defaults

Candidate tasks:

- [ ] T3.1 Decide and implement default routing for dump/spec/plan/tracking.
- [ ] T3.2 Preserve or simplify env var override behavior.
- [ ] T3.3 Reconcile the `feat/codex-dump-generation` implementation with the
      committed slice map.
- [ ] T3.4 Update docs/config reference for artifact generation backends.
- [ ] T3.5 Add routing tests for default and override cases.
- [ ] T3.6 Add slice implementation report.
- [ ] T3.7 Run targeted tests and repository quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Slice 4: Artifact Run Bundle Audit

Candidate tasks:

- [ ] T4.1 Create or update a temporary probe script for the reference
      workspace.
- [ ] T4.2 Run dump/spec/plan/tracking generation with clean trace and bundles.
- [ ] T4.3 Inspect generated artifacts and trace evidence.
- [ ] T4.4 Record results in the slice report.
- [ ] T4.5 Add any missing lightweight docs for manual reproduction.
- [ ] T4.6 Run quality gates.

Task commits:

- None yet.

Validation:

- Not run yet.

## Completion

The slicing is not complete.

Tuning needed: undecided.
