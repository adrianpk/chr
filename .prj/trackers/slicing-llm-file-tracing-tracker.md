# LLM File Tracing Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-llm-file-tracing-plan.md`

## Active Slice

Slicing name: LLM File Tracing

Active slice number: Slice 1

Active slice short name: llm trace sink

Expected branch name: `feat/slice-1-llm-trace-sink`

Expected PR title: `feat(slice-1): llm trace sink`

Expected report path:
`.prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md`

Prerequisite PRs or delivered work:

- Board Ticket Operations userspace resolver delivered.
- Quiz, spec, plan, and plan-tracking model-backed generation flows exist.
- OpenAI responses adapter exists as the current model provider boundary.

## Slicing Checklist

- [x] Capture LLM file tracing goals.
- [x] Capture userspace log location.
- [x] Capture opt-in environment configuration.
- [x] Capture rolling file requirement.
- [x] Capture request/response payload gating.
- [x] Capture future per-user configuration need.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [ ] Commit plan and tracker.
- [ ] Merge plan and tracker into `dev`.
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: LLM Trace Sink | Planned | `feat/slice-1-llm-trace-sink` | `feat(slice-1): llm trace sink` | `.prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md` | Adds reusable opt-in writer and tests. |
| Slice 2: Quiz Model Trace Integration | Planned | `feat/slice-2-quiz-model-trace-integration` | `feat(slice-2): quiz model trace integration` | `.prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md` | Wires tracing into real model calls. |

## Slice 1: LLM Trace Sink

Candidate tasks:

- [ ] T1.1 Add trace configuration resolved from environment variables.
- [ ] T1.2 Add active-userspace LLM log path resolution.
- [ ] T1.3 Add JSONL trace writer with non-fatal write behavior.
- [ ] T1.4 Add rolling file rotation with bounded retained files.
- [ ] T1.5 Add redaction for secret-like fields and known secret environment
      values.
- [ ] T1.6 Add unit tests for disabled tracing, enabled tracing, payload gating,
      redaction, and rotation.
- [ ] T1.7 Add slice implementation report.
- [ ] T1.8 Run targeted tests and `go test ./...`.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 2: Quiz Model Trace Integration

Candidate tasks:

- [ ] T2.1 Generate one `llm_call_id` per provider invocation.
- [ ] T2.2 Record `llm_request` events before provider dispatch.
- [ ] T2.3 Record `llm_response` events for success, provider failure, and
      parse failure.
- [ ] T2.4 Include action, request id, provider, model, status, latency,
      payload byte counts, and error details.
- [ ] T2.5 Include request and response payloads only when payload capture is
      enabled.
- [ ] T2.6 Add integration tests covering quiz/spec/plan shared model paths or
      the common OpenAI adapter boundary.
- [ ] T2.7 Add slice implementation report.
- [ ] T2.8 Run targeted tests and `go test ./...`.

Task commits:

- Pending.

Validation:

- Pending.

## Slicing Closure

Pending.
