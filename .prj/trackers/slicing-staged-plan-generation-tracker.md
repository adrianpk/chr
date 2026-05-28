# Staged Plan Generation Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-staged-plan-generation-plan.md`

## Active Slice

Slicing name: Staged Plan Generation

Active slice number: Slice 2

Active slice short name: slice detail expansion

Expected branch name: `feat/slice-2-slice-detail-expansion`

Expected PR title: `feat(slice-2): slice detail expansion`

Expected report path:
`.prj/reports/slices/staged-plan-generation/slice-2-slice-detail-expansion.md`

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
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with repository quality gates.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Outline Generation Contract | Delivered | `feat/slice-1-outline-generation-contract` | `feat(slice-1): outline generation contract` | `.prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md` | Delivered in PR #237. |
| Slice 2: Slice Detail Expansion | Planned | `feat/slice-2-slice-detail-expansion` | `feat(slice-2): slice detail expansion` | `.prj/reports/slices/staged-plan-generation/slice-2-slice-detail-expansion.md` | Adds one-slice expansion action/schema/request construction and validation. |
| Slice 3: Deterministic Plan Assembly | Planned | `feat/slice-3-deterministic-plan-assembly` | `feat(slice-3): deterministic plan assembly` | `.prj/reports/slices/staged-plan-generation/slice-3-deterministic-plan-assembly.md` | Connects staged generation into final plan assembly and write behavior. |

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

## Slice 2: Slice Detail Expansion

Candidate tasks:

- [ ] T2.1 Add a slice detail model action and output contract.
- [ ] T2.2 Add slice detail payload structs and validation.
- [ ] T2.3 Add strict JSON schema for slice detail responses.
- [ ] T2.4 Add provider instruction text for one-slice expansion.
- [ ] T2.5 Add service request construction with full outline and target slice.
- [ ] T2.6 Add validation that rejects renamed, unknown, reordered, or
      out-of-scope slice details.
- [ ] T2.7 Add tests for valid detail expansion and rejection cases.
- [ ] T2.8 Add slice implementation report.
- [ ] T2.9 Run targeted tests and repository quality gates.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 3: Deterministic Plan Assembly

Candidate tasks:

- [ ] T3.1 Add deterministic Markdown assembly from outline plus slice details.
- [ ] T3.2 Connect staged generation orchestration into `GeneratePlan`.
- [ ] T3.3 Preserve existing generate/regenerate/archive/failure safety
      semantics.
- [ ] T3.4 Record staged request/response provenance in the generated plan.
- [ ] T3.5 Ensure no final plan is written when outline or slice detail
      generation fails.
- [ ] T3.6 Add integration tests for staged plan generation success and
      failure paths.
- [ ] T3.7 Add slice implementation report.
- [ ] T3.8 Run targeted tests and repository quality gates.

Task commits:

- Pending.

Validation:

- Pending.
