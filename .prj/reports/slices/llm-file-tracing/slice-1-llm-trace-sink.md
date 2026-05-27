# Slice 1: LLM Trace Sink

Status: delivered
Slicing: LLM File Tracing
Plan: `.prj/plans/slicing-llm-file-tracing-plan.md`
Tracker: `.prj/trackers/slicing-llm-file-tracing-tracker.md`
Branch: `feat/slice-1-llm-trace-sink`
PR: #229

## Purpose

Add the reusable local trace sink for model-backed operations before wiring it
into provider calls.

This slice keeps tracing behavior opt-in, userspace-local, and non-fatal so
later integration can record expensive LLM exchanges without changing normal
generation behavior.

## Delivered Behavior

- Added environment-controlled LLM tracing through `CHC_LLM_TRACE`.
- Added explicit payload capture gating through `CHC_LLM_TRACE_PAYLOADS`.
- Added rotation controls through `CHC_LLM_TRACE_MAX_BYTES` and
  `CHC_LLM_TRACE_KEEP_FILES`.
- Resolved trace output under `.chc/user/log/llm/llm.ndjson`.
- Added JSONL event writing with automatic timestamp injection.
- Added best-effort event recording that ignores tracing failures.
- Added secret redaction for secret-like fields, bearer tokens, and known secret
  environment values.
- Added bounded file rotation for retained trace files.

## Implementation Notes

- `LoadLLMTraceConfig` resolves environment configuration and active userspace
  paths without creating directories when tracing is disabled.
- `WriteLLMTraceEvent` is the low-level checked writer used by tests.
- `RecordLLMTraceEvent` is the best-effort helper intended for runtime
  integration in the next slice.
- Payload fields are omitted by default and only written when payload capture is
  explicitly enabled.
- The OpenAI adapter is intentionally untouched in this slice.

## Contracts Added or Changed

- Added `LLMTraceConfig`.
- Added `LoadLLMTraceConfig`.
- Added `WriteLLMTraceEvent`.
- Added `RecordLLMTraceEvent`.
- Added environment variables:
  - `CHC_LLM_TRACE`
  - `CHC_LLM_TRACE_PAYLOADS`
  - `CHC_LLM_TRACE_MAX_BYTES`
  - `CHC_LLM_TRACE_KEEP_FILES`

## Files of Interest

- `internal/chc/llm_trace.go`
- `internal/chc/llm_trace_test.go`
- `.prj/trackers/slicing-llm-file-tracing-tracker.md`
- `.prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md`

## Validation

- `go test ./internal/chc -run 'Test(LoadLLMTraceConfig|WriteLLMTraceEvent|RecordLLMTraceEvent)'` - pass.
- `go test ./...` - pass.

## Risks and Follow-ups

- Slice 2 must wire request and response recording into the shared model
  provider boundary.
- Flow-level trace IDs remain future work.
- Per-user configuration files remain future work; this slice only adds the
  environment variable control surface.
