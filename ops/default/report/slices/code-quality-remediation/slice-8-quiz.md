# Slice 8: Quiz domain and model adapter

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-8-quality-quiz`
PR: pending

## Purpose

Separate clarification state and shared model contracts from provider transport
while preserving finite questions, answer meaning, generation identity and guarded
authoring publication.

## Delivered Behavior

Core/quiz owns runtime values, validation, raw-capture curation, entry/answer
transitions, active-question routing and tag fallback. Direct app consumers retain
the original Deep/Fast counters, finite bounds, question identity and branch return
behavior. Provider IO and retained publication are not moved into that domain.

Core/quiz/model owns shared Request, Response and Adapter.Invoke contracts,
context/output payloads, strict decoding, instructions, schemas and output mapping
used by both the API and harness paths. Shared generated artifact values and pure
validators live in core/authoring/generation. File writers, approval, locks,
journals, capacity and workflow execution stay with their existing owners.

The application service is split by model context, readiness, artifacts and
runtime IO. The unused newQuizModelRequest production wrapper is removed. Its
four readiness tests load persisted entries and use the live entry-aware request
path. Existing manual, Deep/Fast and guided service checks pass.

The OpenAI transport extraction and remaining signature conversions are pending.
This report is an implementation checkpoint, not PR readiness.

## Implementation Notes

The move ledger records 363 declaration extraction checkpoints, including 71
runtime/curation declarations, 95 shared-value prerequisites, 160 model
contracts/helpers and 37 application/transition declarations. Original declaration
hashes identify the code before package qualification, required visibility and
subsequent naming changes. Current names and files remain linked to each entry.

Separate commits establish the model owner and then name its contracts. The large
payload file is split into decoding, output contracts, context, questions, specs,
plans and tracking. The justified schema-literal exception remains. Private
conductor tests live beside the model owner; parser, schema and instruction tests
use the actual shared entrypoints.

Shared helpers retain one owner: workflow reference/argv values and size/count
bounds, strict authoring-state JSON decoding, model JSON-object extraction and
Unicode slug normalization. Import fixtures reject coordinator and provider
imports from the domain/model owners.

## Contracts Added or Changed

Runtime and staged generation records retain their original fields, tags and
state spellings. Model adapters implement Invoke(context.Context, Request), with
all local implementations and consumers updated. The raw-turn parser returns a
named metadata/body result. Question routing returns named parent/trunk identity.
Normal empty bodies, root questions, active parents and retained trails keep their
original branches. Temporary signature debt decreases from 307 to 305.

## Files of Interest

- `internal/core/quiz/` owns runtime validation, curation and pure transitions.
- `internal/core/quiz/model/` owns shared generation payloads and response mapping.
- `internal/core/authoring/generation/` owns staged values and pure validation.
- `internal/app/service_quiz*.go` separates shared service responsibilities.
- [Move evidence](../../../plan/code-quality-remediation-slice-8-moves.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-8-signatures.csv),
  [current owners](../../../plan/code-quality-remediation-core-owners.csv) and
  [dependency/glossary decisions](../../../plan/code-quality-remediation-boundaries.md)
  retain the implementation mapping.

## Validation

- `go test ./internal/core/authoring/generation ./internal/core ./internal/app -run '^Test(Quiz|ValidateQuiz|OpenAIQuiz|GeneratedPlan|PlanOrder|AcceptanceReadiness)' -count=1 -timeout=3m`
  — passed after shared-value extraction.
- `go test ./internal/core/quiz/model ./internal/core ./internal/app -run '^Test(InvokeQuiz|ValidateQuiz|Quiz|OpenAIQuiz|Service.*Quiz|GeneratedPlan|PlanOrder|AcceptanceReadiness|HarnessGeneration)' -count=1 -timeout=3m`
  — passed after model naming and the routing result conversion.
- `go test ./internal/core/quiz ./internal/core/quiz/model ./internal/app -run '^Test(Quiz|Service.*Quiz|NewQuizModelRequest|GuidedUseCase.*Quiz)' -count=1 -timeout=3m`
  — passed after domain transitions, service split and wrapper deletion.
- `go test ./scripts/goquality -count=1` — passed with owner import fixtures.
- `make check-go-style` — passed with 305 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core/... ./internal/app/... ./internal/tui/...`
  — passed, zero issues after the service split.
- `git diff --check` — passed for these task checkpoints.

Final clean-head make verify and focused slice validation have not run. Their
ordered evidence will be recorded through the controller after PR references.

## Risks and Follow-ups

T8.1-T8.3 and T8.5 remain in progress; T8.4 is implemented; T8.6 is pending.
Extract the OpenAI IO owner with narrow ports to existing generation admission,
output validation and trace recording. Preserve the distinct quiz and structured
API body limits and existing shared capacity. Finish the retained action/domain
value boundary and assigned response, bound-call, trace-input and receipt-result
conversions. Complete dispositions, final validation and canonical PR references.
