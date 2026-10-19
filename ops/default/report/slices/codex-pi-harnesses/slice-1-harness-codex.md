# Slice 1: Neutral contracts and Codex extraction

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `refactor/slice-1-harness-codex`
PR: pending

## Purpose

Run existing Codex generation and workers through injected neutral contracts,
with workflow authority and semantic validation retained by the host.

## Delivered Behavior

- Artifact generation, contextual Quiz, structured conversation, Dump curation
  and agentic review use an injected runner. App assembles Codex while keeping
  the existing action-specific direct OpenAI defaults and overrides.
- Native workflow and Docker/proving callers use the Codex worker command
  adapter. Frozen input, runtime selection, claims, budgets, validation and
  delivery receipts retain their existing host owners.
- Codex generation retains read-only, ephemeral execution and legacy command,
  model and timeout handling. No other harness installation is required.
- Output overflow, cancellation, nonzero exit and missing output fail the call.
  A reused request cannot accept a stale output file. Diagnostics are sanitized
  before retention in result values, errors or log artifacts.

## Implementation Notes

Generation uses private bounded stdout/stderr spools and owned process groups.
The stream ceilings are 67,108,864 and 2,097,152 bytes respectively; diagnostics
retain at most 262,144 sanitized bytes. Generation output is limited to 65,536
bytes or a stricter caller limit. Worker context is limited to 4,194,304 bytes;
existing stricter native input/result and diagnostic limits remain in force.
Overflow cancels and reaps the owned process rather than accepting partial JSON.

The adapter receives explicit run directories. Core applies its existing run
layout and records request traces before process launch through a synchronous
start observer. Artifact schemas, prompts, parse errors and semantic checks stay
with their existing owners. Unknown usage and an unobserved CLI model remain
unknown. Concrete runner defaults and permanent core compatibility aliases were
removed. Codex credential discovery moved to the adapter; app passes resolved
read-only mounts into generic runtime planning.

## Contracts Added or Changed

`internal/harness` owns agent request/result/runner, command executor, worker,
capability, observed usage and runtime-mount contracts. The root does not import
its child packages. `process` owns bounded generation subprocess supervision and
shell rendering; `codex` imports only neutral contracts and process support.
Core imports neutral contracts/process support without importing Codex or Pi.

Generic domain owners are now `HarnessQuizModelAdapter`,
`HarnessAuthoringQuizQuestionProvider`, `HarnessAgenticReviewRunner` and the
worker context materializer. Missing injection fails instead of constructing a
hidden concrete runner. The shared structured-model capacity and deadline remain
unchanged.

## Files of Interest

- `internal/harness/contracts.go`
- `internal/harness/process/executor.go`
- `internal/harness/codex/generation.go`, `worker.go` and `runtime.go`
- `internal/core/agent_run.go`, `quiz_harness_model.go` and `worker_context.go`
- `internal/core/harness_invocation.go` and `runtime_plan.go`
- `internal/app/service.go`, `harness.go` and `service_dump_curation_harness.go`

## Validation

Successful task checks:

- `go test ./internal/harness/... -run '^TestAgentRunProcess' -count=1 -timeout=30s`
- `go test ./internal/harness/... -run '^(TestHarnessCodex|TestAgentRunProcess)' -count=1 -timeout=30s`
- `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestHarness|TestStructured|TestNewHarnessInvocation|Test.*Proving|Test.*Runtime|Test.*WorkerContext|TestNewQuizModelAdapter|TestNewDumpCurationProvider)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestHarnessCodexConversation$' -count=1 -timeout=30s`

The subprocess fixtures exercise stdin, generation and worker success, output
fallback, stale output rejection, bounded streams/results, deadline cancellation,
child cleanup, credential redaction, quoted worker input and read-only credential
mounts. Domain fixtures cover schema/JSON rejection, contextual Quiz trace and
parse failures, artifact validation, runtime paths and unchanged OpenAI dispatch.
The app conversation fixture exercises default service assembly through a real
owned executable. Import and caller inventories confirmed the dependency path.

Integration requires controller-retained evidence for the exact clean PR HEAD,
including `make verify`, both configured focused commands, the additional
caller/runtime and process-race checks, and `git diff --check`. The canonical PR
lists the exact successful ordered commands. This development evidence does not
replace scheduled exhaustive or live-provider certification.

## Risks and Follow-ups

None recorded.
