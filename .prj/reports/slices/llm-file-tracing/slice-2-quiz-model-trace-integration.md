# Slice 2: Quiz Model Trace Integration

Status: reviewing
Slicing: LLM File Tracing
Plan: `.prj/plans/slicing-llm-file-tracing-plan.md`
Tracker: `.prj/trackers/slicing-llm-file-tracing-tracker.md`
Branch: `feat/slice-2-quiz-model-trace-integration`
PR: pending

## Purpose

Wire the reusable LLM trace sink into the shared OpenAI quiz model adapter so
quiz, spec, plan, and plan-tracking generation calls can be inspected from the
active project userspace when tracing is enabled.

## Delivered Behavior

- Generates one `llm_call_id` per OpenAI provider invocation.
- Records `llm_request` before dispatching the provider request.
- Records `llm_response` after provider success, provider failure, and domain
  parse failure.
- Includes request id, action, provider, model, status, HTTP status, latency,
  payload byte counts, response id, and error details.
- Includes request, response, and output payloads only when
  `CHC_LLM_TRACE_PAYLOADS=1`.
- Routes the default application OpenAI adapter with the active project
  directory so trace files resolve under `.chc/user/log/llm/`.

## Implementation Notes

- The OpenAI adapter now owns trace correlation because it has access to the
  serialized provider request and raw provider response.
- `RecordLLMTraceEvent` keeps trace writes best-effort; model calls continue
  even if trace configuration or file writes fail.
- Provider errors and parse errors share the same `llm_response` event shape
  with distinct `status` values.
- The service constructor passes `projectContext.ProjectDir` into the default
  OpenAI adapter.

## Contracts Added or Changed

- Added `OpenAIQuizModelAdapter.ProjectDir`.
- Added `llm_call_id` correlation across request and response trace events.
- Added trace response statuses:
  - `ready`
  - `error`
  - `parse_error`
- Added `output_payload` to the payload-gated trace fields.

## Files of Interest

- `internal/chc/quiz_openai_model.go`
- `internal/chc/quiz_model_test.go`
- `internal/chc/llm_trace.go`
- `internal/chcapp/service.go`
- `.prj/trackers/slicing-llm-file-tracing-tracker.md`
- `.prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md`

## Validation

- `go test ./internal/chc -run 'TestOpenAIQuizModel(AdapterRecords|RequestBody|Timeout)|TestWriteLLMTraceEvent'` - pass.
- `go test ./internal/chcapp` - pass.
- `go test ./...` - pass.

## Risks and Follow-ups

- Flow-level trace IDs remain future work.
- Per-user configuration files remain future work.
- Trace inspection commands remain future work.
