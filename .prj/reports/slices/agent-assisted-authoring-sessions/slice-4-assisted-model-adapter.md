# Slice 4: Model Invocation Adapter

Status: delivered
Slicing: Agent-Assisted Authoring Sessions
Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
Branch: `feat/slice-4-assisted-model-adapter`
PR: #102

## Purpose

Define the first host-side adapter boundary for model-backed assisted authoring.

## Delivered Behavior

The core package now supports:

- assisted model request and response envelopes;
- response status validation for ready and blocked model outcomes;
- provider isolation behind `AssistedModelAdapter`;
- deterministic fake adapter for tests and later loop development;
- invocation wrapper with request, adapter, and response validation;
- classified assisted model failures;
- deterministic Markdown rendering for request and response artifacts;
- writing request and response artifacts under a session-scoped model invocation directory.

## Implementation Notes

This slice intentionally does not bind to a real model provider. Provider
details remain behind the adapter interface so later slices can add OpenAI or
Codex-backed invocation without changing the assisted authoring loop contract.

Invocation artifacts are planned under:

```text
.chc/authoring/sessions/<session-id>/model/<request-id>/
```

with:

```text
request.md
response.md
```

## Contracts Added or Changed

New internal contracts:

- `AssistedModelRequest`
- `AssistedModelResponse`
- `AssistedModelResponseStatus`
- `AssistedModelAdapter`
- `AssistedModelFailure`
- `AssistedModelFailureKind`
- `FakeAssistedModelAdapter`
- `InvokeAssistedModel`
- `PlanAssistedModelInvocationArtifacts`
- `RecordAssistedModelInvocation`
- `RenderAssistedModelRequest`
- `RenderAssistedModelResponse`

## Files of Interest

- `internal/chc/assisted_model.go`
- `internal/chc/assisted_model_test.go`
- `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestInvokeAssistedModel|TestRecordAssistedModelInvocation|TestRenderAssistedModelArtifacts'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice provides the boundary and recording contract only. It does not run
the adaptive quiz loop, curate answers, synthesize artifacts, expose CLI
commands, or call a real provider.
