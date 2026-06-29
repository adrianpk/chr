# Slice 5: QA Runbook And Smoke Validation

Status: reviewing

Slicing: QA Agent Work Scenario

Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`

Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

Branch: `feat/slice-5-qa-runbook-smoke-validation`

PR: pending

## Scope

This slice documents the manual QA scenario and adds smoke validation for the
reset workspace.

## Delivered

- Added a runbook for resetting the blog QA scenario, opening CHR, running the
  seeded bug ticket, generating comments artifacts, and running generated work.
- Documented the generic and blog-specific QA Make targets.
- Added `qa-smoke` and `qa-smoke-blog` targets.
- Added `scripts/qa/smoke.sh` to verify reset behavior, parent repository
  ignore behavior, independent Git initialization, and the expected failing
  unknown-post test.

## Validation

- `make qa-smoke-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test expects `404`
- `make verify`
- `git diff --check`
