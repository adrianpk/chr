# Slice 3: Pi RPC and Structured Generation

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-3-pi-generation`
PR: pending

## Purpose

Allow a prepared native Pi profile to generate validated structured content
through existing operation owners. T3.1 supplies the transport; generation
assembly and private resource restrictions remain T3.2 and T3.3.

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

## Contracts Added or Changed

- Shared `process.Executor.Exchange` for bounded bidirectional process calls.
- Sibling `harness/pi` transport pinned to `pi-rpc/1.3.0`, with exact model
  verification and distinguishable protocol/provider/capability failure values.
- No workflow, artifact schema, frozen binding or existing Codex path changed
  during T3.1. Generation result publication remains with its existing owner.

## Files of Interest

- `internal/harness/process/exchange.go`
- `internal/harness/pi/transport.go`, `session.go` and `events.go`
- `internal/harness/pi/transport_test.go` and `transport_fixture_test.go`

## Validation

Passed focused T3.1 implementation checks for source commit
`2a256f9612d20734aa2cdce5562e8aeb96ddfe8c`:

- `go test ./internal/harness/... -count=1 -timeout=1m`
- `go test -race ./internal/harness/pi ./internal/harness/process -run '^(TestPiTransport|TestAgentRunProcess)' -count=1 -timeout=1m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/harness/pi ./internal/harness/process`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Fixtures exposed and now cover an oversized unterminated record at the byte
boundary and cleanup cancellation incorrectly appearing as caller cancellation.
All owned fixtures stopped after their checks. Required slice-level commands
will be recorded through the controller on the final clean PR head after T3.2,
T3.3 and PR-reference follow-up commits.

## Risks and Follow-ups

T3.2 must wire the Pi runner through all generation families while retaining
operation bindings, validators and shared capacity. T3.3 must enforce private
per-call state and resource/tool discovery restrictions. No canonical PR is
ready until those tasks and the configured slice checks are complete.
