# Staged Plan Generation Slicing Tracker

Status: delivered

Plan: `.prj/plans/slicing-staged-plan-generation-plan.md`

## Active Slice

Slicing name: Staged Plan Generation

Active slice number: none

Active slice short name: none

Expected branch name: none

Expected PR title: none

Expected report path:
none

Prerequisite PRs or delivered work:

- Plan Generation Safety slicing delivered.
- Model-backed spec and plan generation flow.
- Staged Plan Generation draft spec.

## Slicing Checklist

- [x] Capture staged plan generation draft spec.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with repository quality gates.
- [x] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Outline Generation Contract | Delivered | `feat/slice-1-outline-generation-contract` | `feat(slice-1): outline generation contract` | `.prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md` | Delivered in PR #237. |
| Slice 2: Slice Detail Expansion | Delivered | `feat/slice-2-slice-detail-expansion` | `feat(slice-2): slice detail expansion` | `.prj/reports/slices/staged-plan-generation/slice-2-slice-detail-expansion.md` | Delivered in PR #238. |
| Slice 3: Deterministic Plan Assembly | Delivered | `feat/slice-3-deterministic-plan-assembly` | `feat(slice-3): deterministic plan assembly` | `.prj/reports/slices/staged-plan-generation/slice-3-deterministic-plan-assembly.md` | Delivered in PR #239. |

## Slice 1: Outline Generation Contract

Candidate tasks:

- [x] T1.1 Add a plan outline model action and output contract.
- [x] T1.2 Add outline payload structs and validation.
- [x] T1.3 Add strict JSON schema for outline responses.
- [x] T1.4 Add provider instruction text for outline generation.
- [x] T1.5 Add service request construction that reuses plan generation context.
- [x] T1.6 Add tests for action selection, schema shape, payload validation,
      and request context sections.
- [x] T1.7 Add slice implementation report.
- [x] T1.8 Run targeted tests and repository quality gates.

Task commits:

- `4d80405` docs(lint): require pre-commit quality gates
- `4cc1b1f` feat(chc): add plan outline model contract
- `c32aaea` feat(chc): add plan outline OpenAI contract
- `9c343fb` feat(chcapp): build plan outline requests
- `f1e5dd8` test(chc): cover plan outline generation contract

Validation:

- `gofmt -w internal/chc/quiz_model.go internal/chc/quiz_model_payload.go internal/chc/quiz_openai_model.go internal/chc/plan_outline.go internal/chc/quiz_model_test.go internal/chcapp/service_plan.go internal/chcapp/service_plan_test.go` - pass.
- `go test ./internal/chc ./internal/chcapp` - pass.
- `go test ./...` - pass.
- `git diff --check` - pass.
- `make verify` - fail: `lint-strict` reports pre-existing lint debt.

## Completion

All planned slices are delivered:

- Slice 1 delivered in PR #237.
- Slice 2 delivered in PR #238.
- Slice 3 delivered in PR #239.

Completed slicing validation:

- `go test ./internal/chc ./internal/chcapp`
- `go test ./...`
- `git diff --check`
- `make verify` remains blocked by pre-existing `lint-strict` debt.

Tuning needed: no.

Evidence: all planned staged generation contracts and orchestration slices were
delivered, targeted and repository tests passed, and no post-slice tuning issue
was identified during automated validation.

## Slice 2: Slice Detail Expansion

Candidate tasks:

- [x] T2.1 Add a slice detail model action and output contract.
- [x] T2.2 Add slice detail payload structs and validation.
- [x] T2.3 Add strict JSON schema for slice detail responses.
- [x] T2.4 Add provider instruction text for one-slice expansion.
- [x] T2.5 Add service request construction with full outline and target slice.
- [x] T2.6 Add validation that rejects renamed, unknown, reordered, or
      out-of-scope slice details.
- [x] T2.7 Add tests for valid detail expansion and rejection cases.
- [x] T2.8 Add slice implementation report.
- [x] T2.9 Run targeted tests and repository quality gates.

Task commits:

- `0c62bb4` feat(chc): add plan slice detail contract
- `8166f11` feat(chc): add plan slice detail OpenAI contract
- `e7ada61` feat(chcapp): build plan slice detail requests
- `b505eb0` test(chc): cover plan slice detail contract

Validation:

- `gofmt -w internal/chc/quiz_model.go internal/chc/plan_outline.go internal/chc/plan_slice_detail.go internal/chc/quiz_model_payload.go internal/chc/quiz_openai_model.go internal/chc/quiz_model_test.go internal/chcapp/service_plan.go internal/chcapp/service_plan_test.go` - pass.
- `go test ./internal/chc ./internal/chcapp` - pass.
- `go test ./...` - pass.
- `git diff --check` - pass.
- `make verify` - fail: `lint-strict` reports pre-existing lint debt.

## Slice 3: Deterministic Plan Assembly

Candidate tasks:

- [x] T3.1 Add deterministic Markdown assembly from outline plus slice details.
- [x] T3.2 Connect staged generation orchestration into `GeneratePlan`.
- [x] T3.3 Preserve existing generate/regenerate/archive/failure safety
      semantics.
- [x] T3.4 Record staged request/response provenance in the generated plan.
- [x] T3.5 Ensure no final plan is written when outline or slice detail
      generation fails.
- [x] T3.6 Add integration tests for staged plan generation success and
      failure paths.
- [x] T3.7 Add slice implementation report.
- [x] T3.8 Run targeted tests and repository quality gates.

Task commits:

- `6213b78` feat(chc): assemble staged generated plans
- `9853ae2` feat(chcapp): orchestrate staged plan generation
- `6037cba` test(chcapp): cover staged plan generation

Validation:

- `gofmt -w internal/chc/plan_assembly.go internal/chcapp/service_plan.go internal/chcapp/service_plan_test.go internal/chcapp/service_quiz_test.go` - pass.
- `go test ./internal/chc ./internal/chcapp` - pass.
- `go test ./...` - pass.
- `git diff --check` - pass.
- `make verify` - fail: `lint-strict` reports pre-existing lint debt.
