# Slice 8: Quiz domain and model adapter

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-8-quality-quiz`
PR: pending

## Purpose

Separate clarification and shared model contracts from OpenAI transport while
preserving finite questions, generation admission and guarded authoring receipts.

## Delivered Behavior

Core/quiz owns runtime values, curation, pure transitions and retained action
values. Core/quiz/model owns requests, context/output payloads, strict decoding,
instructions, schemas and response mapping used by API and harness adapters.
The shared generated artifact values and pure validators have one acyclic owner
in core/authoring/generation. Consumers import those owners directly.

Core/quiz/openai owns credentials, API envelopes, HTTP execution and response
text. Its explicit Host uses the original generation admission, request identity,
output validation and redacted trace owner. Structured generation keeps its
caller-owned capacity and distinct body limit. The model decoder receives actual
provider identity; the shared domain has no OpenAI dependency.

Named transport/trace/receipt/metadata/routing outcomes retain raw and bound output,
HTTP accounting, failure, normal absence and parent/trunk identity. All eight
assigned occurrences conform: one A, two B watch items and five C entries. Six
are fixed and two cohesive five-input operations are retained. Temporary debt
decreases from 307 to 301; the ceiling and external exception are unchanged.

Deep/Fast/manual/guided behavior, question identity, answer meaning and finite
limits remain. The service is split by context, readiness, artifacts and runtime
IO. The unused request wrapper is removed; its four readiness tests use the live
persisted-entry path. Private provider request tests follow their owner.

## Implementation Notes

The move ledger links 404 extraction checkpoints to original declaration hashes
and current names/files. Package moves and naming changes have separate task
commits. The large provider and payload files are split by actual responsibility;
the justified shared schema-literal exception remains.

Observation contains only domain snapshot values. Original filesystem images
stay private in the authoring journal, preserving one guarded publication of
runtime, raw capture, curation, metadata and acknowledgement. Receipt schema,
checksum, digest, operation identity, revision restoration and missing/conflict
behavior remain. No migration or retained-record rewrite is introduced.

Shared reference/argv/bound values, strict authoring JSON decoding, generation
identity validation, model JSON extraction and Unicode slug normalization each
have one owner. Production import checks reject coordinator/provider dependencies
from the domain and reject coordinator imports from the OpenAI owner.

## Contracts Added or Changed

Adapters implement Invoke(context.Context, Request). OpenAI requires an explicit
Host. ResponseFromOutput receives provider provenance. OutputSchema describes the
neutral named schema; OpenAI alone wraps strict Responses wire fields, preserving
the original encoded field order and values. No context is stored.

Runtime/action/generated values retain their original fields, tags and state
spellings. The original 90-second transition context, API deadlines, question and
branch limits, retained Fast capacity and schema compatibility remain. Real HTTP
checks cover configured authorization and absence of credentials from provider
and opted-in trace payloads.

## Files of Interest

- `internal/core/quiz/`, `internal/core/quiz/model/`, `internal/core/quiz/openai/`.
- `internal/core/authoring/generation/` and `internal/core/quiz_generation_host.go`.
- `internal/core/quiz_action_commit.go`, `quiz_action_receipt.go` and `quiz_observation.go`
  retain the original authoring journal and private image ownership.
- `internal/app/service_quiz*.go` separates shared application operations.
- [All eight dispositions](../../../plan/code-quality-remediation-slice-8-dispositions.csv),
  [move evidence](../../../plan/code-quality-remediation-slice-8-moves.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-8-signatures.csv),
  [cleanup decisions](../../../plan/code-quality-remediation-slice-8-cleanup.csv),
  [current owners](../../../plan/code-quality-remediation-core-owners.csv) and
  [dependency/glossary decisions](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./internal/core/quiz/... ./internal/core ./internal/app -run '^Test(HostRequired|Quiz|OpenAIQuiz|Structured|HarnessGeneration|Service.*Quiz|Guided.*Quiz|PlanOrderModel)' -count=1 -timeout=3m`
  — passed after named outputs, retained action separation and provider mapping.
- `go test ./internal/core -run '^TestOpenAIQuizModelAdapterRecordsTracePayloadsWhenEnabled$' -count=1 -timeout=3m`
  — passed after the explicit authorization/payload secrecy assertions.
- `go test ./scripts/goquality -count=1` — passed with owner import fixtures.
- `make check-go-style` — passed with 301 exact debts and one external exception.
- Pinned strict lint passed at the owning task checkpoints.
- `git diff --check` — passed for task commits.

Final clean-head validation uses make verify, focused non-short model/provider/
service/retained-state checks, a focused capacity/deadline/recovery race selection,
then git diff --check. The controller ledger and PR Validation own the exact
ordered commands, HEAD, outcomes and logs. Integration remains controller-owned.

## Risks and Follow-ups

None recorded.
