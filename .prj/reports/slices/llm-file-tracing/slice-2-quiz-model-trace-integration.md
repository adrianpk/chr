# Slice 2: LLM Call Trace Integration

Status: delivered
Slicing: LLM File Tracing
Plan: `.prj/plans/slicing-llm-file-tracing-plan.md`
Tracker: `.prj/trackers/slicing-llm-file-tracing-tracker.md`
Branch: `feat/slice-2-quiz-model-trace-integration`
PR: #230

## Purpose

Wire the reusable LLM trace sink into chr-owned LLM calls so quiz,
spec, plan, plan-tracking, and local authoring quiz generation calls can be
inspected from the active project userspace when tracing is enabled.

## Delivered Behavior

- Generates one `llm_call_id` per OpenAI provider invocation.
- Records `llm_request` before dispatching the provider request.
- Records `llm_response` after provider success, provider failure, and domain
  parse failure.
- Includes request id, action, provider, model, status, HTTP status, latency,
  payload byte counts, response id, and error details.
- Includes request, response, and output payloads only when
  `CHR_LLM_TRACE_PAYLOADS=1`.
- Routes the default application OpenAI adapter with the active project
  directory so trace files resolve under `.chr/default/log/llm/`.
- Records local Codex authoring quiz generation calls outside the ticket
  execution container.

## Implementation Notes

- The OpenAI adapter now owns trace correlation because it has access to the
  serialized provider request and raw provider response.
- `RecordLLMTraceEvent` keeps trace writes best-effort; model calls continue
  even if trace configuration or file writes fail.
- Provider errors and parse errors share the same `llm_response` event shape
  with distinct `status` values.
- The service constructor passes `projectContext.ProjectDir` into the default
  OpenAI adapter.
- Ticket execution worker invocations are intentionally outside this trace
  surface because they run as the execution container workflow.

## Contracts Added or Changed

- Added `OpenAIQuizModelAdapter.ProjectDir`.
- Added `llm_call_id` correlation across request and response trace events.
- Added trace response statuses:
  - `ready`
  - `error`
  - `parse_error`
- Added `output_payload` to the payload-gated trace fields.
- Added trace recording for `CodexAuthoringQuizQuestionProvider`.

## Files of Interest

- `internal/chr/quiz_openai_model.go`
- `internal/chr/quiz_model_test.go`
- `internal/chr/authoring_contextual_quiz.go`
- `internal/chr/authoring_contextual_quiz_test.go`
- `internal/chr/llm_trace.go`
- `internal/chrapp/service.go`
- `.prj/trackers/slicing-llm-file-tracing-tracker.md`
- `.prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md`

## Validation

- `go test ./internal/chr -run 'TestOpenAIQuizModel(AdapterRecords|RequestBody|Timeout)|TestWriteLLMTraceEvent'` - pass.
- `go test ./internal/chr -run 'Test(CodexAuthoringQuizQuestionProviderRecords|OpenAIQuizModelAdapterRecords|WriteLLMTraceEvent)'` - pass.
- `go test ./internal/chrapp` - pass.
- `go test ./...` - pass.

## Risks and Follow-ups

- Flow-level trace IDs remain future work.
- Per-user configuration files remain future work.
- Trace inspection commands remain future work.
