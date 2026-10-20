---
id: TKT-20261019232506
title: "Keep generated spec prose consistent with saved approval"
status: solved
kind: bug
severity: unclassified
priority: unclassified
scope: domain
tags: spec-lifecycle, plan-generation, web-ui
source: chat
reported_at: 2026-10-19T23:25:06Z
commits: [4b3604e46fee57b651b0ef846d0a35b62c7ec0a7]
ready_at: 2026-10-20T20:06:00Z
started_at: 2026-10-20T20:06:00Z
branch: dev
resolution: fixed
closed_at: 2026-10-20T20:11:39Z
---

# Keep generated spec prose consistent with saved approval

## Problem

An approved generated spec can be rejected during plan readiness because its
body still calls itself a draft awaiting user review. The Web Draft checkbox
correctly saves `status: approved` when unchecked; the generated body remains
unchanged. Readiness treats the stale sentence as a conflicting approval gate.

## Reproduction and Confirmed Cause

1. Generate a use-case spec with prose stating that it is a draft for review,
   rather than an accepted implementation contract.
2. Open its editor, uncheck Draft and save. The persisted status is approved.
3. Select Generate Plan. Readiness can reject the spec for an approval conflict
   between its metadata and the unchanged body.
4. The stale sentence was present in the retained original model output from
   Pi generation. It was not appended by the spec writer or approval action.
   The generation prompt explicitly requests draft status and user review.

Relevant owners:

1. `internal/core/quiz_openai_model.go`: `quizModelSpecGenerationInstructions`.
2. `internal/core/spec_generation.go`: `RenderGeneratedSpec`.
3. `internal/app/web_spec.go`: Draft selection and save dispatch.
4. `internal/core/spec_mutation_content.go`: body and lifecycle persistence.
5. `internal/app/service_plan_readiness.go`: semantic planning assessment.

## Expected Outcome

1. Keep lifecycle state authoritative in the saved spec metadata. Approval must
   remain effective when older generated prose still describes the spec as a
   draft; body text must not create a second approval state.
2. Prevent newly generated prose from embedding an approval disclaimer that
   becomes stale when the user approves the document.
3. Preserve user-authored content, source provenance, revision checks and saved
   approval when correcting this behavior. No manual body edit or repeated
   model generation should be required solely to make approval effective.
4. Preserve context checks for genuinely unresolved behavioral requirements.
   An approval correction must not turn every insufficient-context result into
   a successful planning result or silently approve an unapproved document.
5. Apply the behavior across generation backends and shared authoring adapters;
   the observed Pi case is a reproduction, not a backend-specific exception.

## Validation

Reproduce draft generation, approval through the Web editor, and plan-readiness
assessment with a stale draft disclaimer. Verify that approval is recognized,
that unapproved state is not promoted by body text, and that genuine missing
requirements still produce a useful rejection. Cover the shared operation path
with focused regressions.

## Current Boundary

The contradictory wording in the manual test spec was corrected with explicit
permission. The permanent correction is now implemented in the selected-source
projection and shared generation/readiness instructions. Focused regressions
pass for preserved stale prose, saved approval, unapproved documents and genuine
missing behavior. The development gate passed and the correction is committed on dev.

## Delivery Evidence

Implementation: `4b3604e46fee57b651b0ef846d0a35b62c7ec0a7`.

Passed: `go test ./internal/core ./internal/app -run '^Test(PlanSpecApproval|PlanSources|PlanReadiness|PlanGenerationDiscovery|SpecLifecycleAliases|WebSpecAuthoring)$' -count=1 -timeout=3m` and `make verify`.

The model boundary was substituted for deterministic regressions; no live
provider request or running-service change was performed.
