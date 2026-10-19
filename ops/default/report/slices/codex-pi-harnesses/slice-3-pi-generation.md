# Slice 3: Pi RPC and Structured Generation

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-3-pi-generation`
PR: pending

## Purpose

Allow a prepared native Pi profile to generate validated structured content
through existing operation owners. T3.1 supplies the transport and T3.2 wires
the selected runner. T3.3 resource/tool restrictions remain in progress.

## Delivered Behavior

- T3.1: Pi 1.3.0 RPC uses the shared owned-process supervisor for concurrent
  pipes, bounded input/stdout/stderr, sanitized diagnostic excerpts and cleanup.
- Correlated setup commands select and verify the exact provider/model and
  requested thinking level. Automatic retry and compaction are disabled before
  prompting; an occupied or incompatible session fails before model work.
- Strict LF framing accepts fragmented UTF-8 and CRLF without splitting Unicode
  line/paragraph separators. Malformed, oversized or unterminated records fail.
- Prompt acceptance, streaming deltas, reasoning and `agent_end` cannot become
  final output. The call waits for both its prompt acknowledgment and
  `agent_settled`, including intermediate retry/compaction events.
- Provider failure, abort, early exit, unsupported capability, process exit,
  caller cancellation and deadline remain distinct failures. Even complete
  assistant output is discarded after a failed process exit or shutdown.
- Observed input/output usage is aggregated without certification. Missing,
  negative or overflowing counters stay unknown, including earlier calls and
  unmeasured compaction work.

- T3.2: Specs, acceptance features, Plans, tracking, Quiz questions, contextual
  Quiz, Dump curation, explanations, structured conversation and agentic review
  dispatch through the retained Pi generation selection and original validators.
- Live profile changes preserve admitted model/provider/thinking and executable
  identity. The selected private resource references survive operation recovery;
  model definitions retain a credential-free semantic hash.
- Fast Quiz batches keep buffered questions without another call. Concurrent
  Pi profiles reuse the existing structured-model capacity pool. Each call has
  separate writable session/authentication state that is removed after reaping.
- Successful process output and domain validation remain separate host facts.
  Invalid Spec/acceptance output leaves prior canonical artifacts unchanged.

## Implementation Notes

Each transport call creates its own bounded correlation state and subprocess.
It adds no retry policy, scheduler or capacity pool. The process owner closes
stdin after completion and allows one second for orderly exit before stopping
and reaping the owned group. Cancellation used for internal cleanup is separate
from caller cancellation. Protocol input travels over stdin rather than argv.

The deterministic peer runs as an owned test executable. It exercises actual
pipe backpressure, out-of-order command responses, late acknowledgment, continued
work after a valid earlier answer, and a descendant holding protocol pipes.
It does not establish remote model authentication or a live Pi generation result.

The generation adapter supplies only the explicitly selected static API-key
provider and credential-free model definition. Ambient authentication, home,
project settings and discovered resources are excluded. OAuth generation is
rejected with an actionable refresh-owner requirement. Native version checking
requires Pi 1.3.0; readiness itself never launches a model or executable.

A sibling deterministic RPC peer drives the real application service and CLI
boundaries. Review retains the canonical project's selection while executing
inside its separate workspace. The existing private operation artifact stores
resource references without adding a store, migration or public credential path.

## Contracts Added or Changed

- Shared `process.Executor.Exchange` for bounded bidirectional process calls.
- Sibling `harness/pi` transport pinned to `pi-rpc/1.3.0`, with exact model
  verification and distinguishable protocol/provider/capability failure values.
- Optional private generation launch resources in the existing operation file;
  omitted resources preserve legacy Codex/API operation bytes and checksums.
- Adapter-neutral protocol/abort errors, observed model and optional cache
  counters in host run evidence. Unknown usage never becomes certification.
- Explicit replacement subprocess environments and pre-truncation diagnostic
  redaction for selected file-backed secrets. Result publication stays with core.

## Files of Interest

- `internal/harness/process/exchange.go`
- `internal/harness/pi/transport.go`, `session.go` and `events.go`
- `internal/harness/pi/transport_test.go` and `transport_fixture_test.go`

- `internal/harness/pi/generation.go`, `generation_setup.go`, `resources.go`
- `internal/app/service_harness_generation.go`, `author_generation.go`, `service_work.go`
- `internal/core/harness_generation_operation.go`, `harness_generation_run.go`
- `internal/app/service_pi_generation*_test.go`

## Validation

Passed focused T3.1 implementation checks for source commit
`2a256f9612d20734aa2cdce5562e8aeb96ddfe8c`:

- `go test ./internal/harness/... -count=1 -timeout=1m`
- `go test -race ./internal/harness/pi ./internal/harness/process -run '^(TestPiTransport|TestAgentRunProcess)' -count=1 -timeout=1m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/harness/pi ./internal/harness/process`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Passed focused T3.2 task checks:

- `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestPiGeneration|TestHarnessGenerationBinding|TestHarnessSelection|TestHarnessCodex|TestAgentRunProcess)' -count=1 -timeout=3m`
- `go test -race ./internal/harness/pi ./internal/harness/process ./internal/app -run '^(TestPiGeneration|TestHarnessGenerationBinding|TestAgentRunProcess)' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestPiGenerationCapacity$' -count=1 -timeout=30s`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/harness/... ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Fixtures exposed and now cover an oversized unterminated record at the byte
boundary and cleanup cancellation incorrectly appearing as caller cancellation.
All owned fixtures stopped after their checks. Required slice-level commands
will be recorded through the controller on the final clean PR head after T3.3
and PR-reference follow-up commits.

## Risks and Follow-ups

T3.3 must complete resource/discovery restrictions, explicit read-only file
inspection for callers that require it, and negative capability/usage/resource
fixtures. No canonical PR is ready until that task and the configured checks
are complete. The current T3.2 adapter disables all tools.
