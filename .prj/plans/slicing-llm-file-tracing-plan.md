# LLM File Tracing Slicing Plan

Status: active

## Purpose

Add opt-in LLM request and response tracing for ChrysoCoord model-backed
operations.

The tracing surface must make expensive model calls debuggable without changing
normal runtime behavior. When enabled, ChrysoCoord writes structured rolling
JSONL trace events under the active project userspace:

```text
.chc/user/log/llm/
```

This first implementation uses environment variables as the configuration
surface. A richer per-user configuration system remains future work and is
recorded explicitly as a follow-up.

## Source

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `internal/chc/project_userspace.go`
- `internal/chc/quiz_model.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chcapp/service_quiz.go`
- `internal/chcapp/service_spec.go`
- `internal/chcapp/service_plan.go`
- Existing model-backed quiz, spec, plan, and plan-tracking generation flows.

External implementation reference:

- A sibling project has an LLM trace writer and request/response recorder that
  use opt-in tracing, JSONL records, redaction, and rolling files. This plan
  captures the transferable pattern without naming or documenting that project
  in ChrysoCoord artifacts.

Conversation decisions:

- Tracing must live inside active userspace, not at repository root.
- Initial path should be `.chc/user/log/llm/`.
- The output should include request payload and response payload when payload
  capture is explicitly enabled.
- A trace-level identifier is desirable but may be absent until full flow
  tracing exists.
- Each LLM call should still have a call identifier that lets request and
  response events be correlated.
- Tracing must be opt-in by configuration; environment variables are acceptable
  for the first version.
- Rolling files are required.
- A future per-user configuration system should replace or complement the
  environment-only control surface.

## Current State

Model-backed generation currently flows through `chc.QuizModelAdapter` and the
OpenAI responses adapter.

Important limitations:

- There is no durable request/response trace file.
- Provider request payloads are built inside the OpenAI quiz model adapter and
  are not captured for diagnostics.
- Provider response payloads are parsed into domain responses, but the raw
  provider response is not recorded outside existing generated artifacts.
- Plan/spec/quiz generation failures can be hard to debug without reproducing
  the exact payload.
- Userspace exists as `.chc/user/`, but no LLM log directory exists beneath it.
- Configuration is not yet per-user beyond the active operator environment.

## Target Model

The target model introduces an opt-in trace sink that records one JSON object per
line.

Example event shape:

```json
{"event":"llm_request","timestamp":"2026-05-27T10:00:00.000Z","llm_call_id":"llm-...","request_id":"...","action":"plan_generation","provider":"openai","model":"gpt-5-mini","payload_bytes":1234}
```

When payload capture is enabled, request events include the serialized provider
request payload and response events include the raw provider response body or
normalized output payload.

Environment variables for the first version:

- `CHC_LLM_TRACE=1`: enable LLM file tracing.
- `CHC_LLM_TRACE_PAYLOADS=1`: include request and response payloads.
- `CHC_LLM_TRACE_MAX_BYTES=<bytes>`: override rotation size.
- `CHC_LLM_TRACE_KEEP_FILES=<count>`: override rotated file count.

Default behavior:

- tracing disabled;
- no file or directory created when disabled;
- payload capture disabled unless explicitly enabled;
- redaction protects secret-like fields and environment secret values;
- failed trace writes do not fail the model operation.

Future work:

- per-user configuration under the active userspace;
- flow-level trace IDs across TUI/CLI actions;
- trace inspection commands;
- richer correlation with tickets, plans, specs, and runtime execution records.

## Slice Count Decision

This work should be split into two slices.

Slice 1 creates the reusable trace sink and keeps it behavior-neutral.
Slice 2 wires actual model requests and responses into that sink.

This avoids mixing file-format/rotation concerns with provider invocation
changes and keeps review risk low.

## Slice Map

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

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | llm trace sink | `feat/slice-1-llm-trace-sink` | `feat(slice-1): llm trace sink` | `.prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md` |
| Slice 2 | quiz model trace integration | `feat/slice-2-quiz-model-trace-integration` | `feat(slice-2): quiz model trace integration` | `.prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md` |

## Slice 1: LLM Trace Sink

Add the reusable tracing foundation without wiring it into provider calls.

Expected behavior:

- Resolve trace output beneath the active project userspace.
- Create `.chc/user/log/llm/` only when tracing is enabled and the first event
  is written.
- Write JSONL records.
- Rotate the active file when it exceeds the configured size.
- Keep a bounded number of rotated files.
- Redact secret-like values and known secret environment values.
- Treat trace write failures as non-fatal.
- Add unit tests for disabled tracing, enabled tracing, payload gating,
  redaction, and rotation.

Out of scope:

- Calling the trace sink from the OpenAI adapter.
- Flow-level trace IDs.
- CLI trace inspection.
- Per-user config files.

## Slice 2: Quiz Model Trace Integration

Record actual request and response exchanges around model-backed generation.

Expected behavior:

- Generate an `llm_call_id` for each provider invocation.
- Record `llm_request` before dispatching the provider request.
- Record `llm_response` after a provider response, provider error, or parse
  failure.
- Include request id, action, provider, model, status, latency, payload size,
  and error details.
- Include request/response payloads only when payload capture is enabled.
- Cover quiz next-question, spec generation, plan generation, and plan-tracking
  materialization because they share the quiz model adapter boundary.
- Add tests with a fake HTTP client or fake trace sink.

Out of scope:

- Retrying failed provider calls.
- Persisting trace IDs in generated artifacts.
- TUI browsing for trace files.
- Config file migration.

## Delivery Order

Implement slices sequentially:

1. Slice 1 establishes the trace writer, configuration environment, and rolling
   file behavior.
2. Slice 2 connects the writer to real model calls and verifies request/response
   correlation.

No implementation slice should start before this plan and its tracker are
committed and merged into `dev`.
