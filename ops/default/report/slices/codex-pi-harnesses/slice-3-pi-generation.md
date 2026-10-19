# Slice 3: Pi RPC and Structured Generation

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-3-pi-generation`
PR: [#414](https://forge.adrianpk.com/adrianpk/chr/pulls/414)

## Purpose

Generate structured content through a prepared native Pi profile while keeping
operation selection, domain validation, publication, capacity and deadlines
under their existing owners.

## Delivered Behavior

- Pi serves Specs, acceptance features, Plans, tracking, fast/deep Quiz,
  contextual Quiz, Dump curation, explanations, conversation and agentic review.
  Codex and direct API dispatch retain their established paths.
- Provider/model/thinking, executable identity and private resource references
  remain frozen across switches, retries and service recovery. Credential
  rotation at the retained reference preserves the model-definition identity.
- Correlated Pi 1.3.0 RPC waits for prompt acceptance, a completed assistant
  answer and `agent_settled`. Intermediate answers, retries, compaction,
  reasoning and tool results cannot become the final output.
- Ordinary structured calls disable tools. Discovery/review callers receive
  only bounded read-only inspection of their project and exact host-provided
  evidence files. Automatic local resources cannot change the call's authority.
- Per-call home, credentials, settings and session state are private and removed
  after process reaping. Unsupported or unavailable preparation fails without
  fallback. Invalid output preserves prior canonical artifacts.
- Optional usage/cache counters remain absent when unknown. Observations never
  certify finite-budget or whole-call token enforcement. Concurrent selections
  reuse the existing structured-model capacity pool.

## Implementation Notes

T3.1 introduced bounded full-duplex transport through `process.Executor`.
Strict LF framing accepts fragmented UTF-8 and CRLF while preserving Unicode
line/paragraph separators. Records, input, diagnostics and complete streams have
explicit ceilings. Setup verifies exact model/thinking and idle private state;
retry and compaction are disabled before prompting. Internal cleanup cancellation
is distinct from caller cancellation. Closing stdin allows one second for
orderly exit before stopping and reaping the owned group.

T3.2 wires the runner through existing shared services and caller parsers.
The existing private operation artifact retains launch references and a hash
of only the selected credential-free model semantics. It adds no store or
migration. Absent resources preserve existing Codex/API operation bytes and
checksums. Agentic review keeps the canonical project's selection inside its
separate execution workspace.

T3.3 (`c1f0495e670804d2925f3f018f3c35fc434b36b2`) loads only explicitly selected static API-key resources into a private
per-call agent directory with a replacement environment. OAuth generation
requires a refresh owner and is rejected. Credential commands, interpolation,
embedded endpoint credentials, authentication headers and discovered resources
are rejected or disabled. Input/artifact preparation rejects redirected,
non-regular or non-private files without replacing their targets.

The explicit inspection profile loads only the embedded host extension. It
opens files read-only, checks opened descriptors, confines access to the project
and at most 16 exact evidence references, and excludes selected credential/model
files. Scans/results are bounded to 4096 entries and 64 KiB; RPC permits at most
64 correlated inspection calls. Intermediate tool output remains unapproved.
Contextual Quiz applies its existing domain validator before recording output
validation, including an empty-question rejection.

## Contracts Added or Changed

- Sibling Pi generation runner and `pi-rpc/1.3.0` settled RPC transport.
- Private launch-resource references in the existing generation operation file.
- Explicit subprocess environments and selected-secret diagnostic redaction
  before excerpt truncation.
- Host-requested read-only inspection and exact evidence-file capabilities.
- Adapter-neutral protocol/abort errors, observed model and optional cache usage
  in host run metadata. Process completion and domain validation remain distinct.

## Files of Interest

- `internal/harness/process/exchange.go`, `diagnostics.go`
- `internal/harness/pi/transport.go`, `session.go`, `events.go`
- `internal/harness/pi/generation.go`, `resources.go`, `generation_files.go`, `readonly.mjs`
- `internal/core/harness_generation_operation.go`, `harness_generation_run.go`
- `internal/app/service_harness_generation.go`, `author_generation.go`, `service_work.go`
- `internal/harness/generation_contract_test.go`, `internal/harness/pi/generation_pinned_test.go`
- `internal/app/service_pi_generation*_test.go`

## Validation

Focused task checks passed for real subprocess transport, real application/CLI
callers, retained bindings, shared capacity, cancellation, filesystem failures,
usage gaps, invalid output and redaction. Exact commands and source task commits
are retained in the slice ticket. Owned fixtures stopped after their checks.

The prepared Pi 1.3.0 binary also passed an owned loopback model fixture for
allowed file reads, exact outside evidence, path/symlink/hard-link escape,
selected credential denial, oversized/non-text/FIFO input, list/find and invalid
tool arguments. It exposed only `chr_inspect` and excluded poisoned project
context/extensions. This establishes local native protocol/resource behavior;
it does not establish paid provider authentication.

The initial controller run on `612d16edb59d3842f414f785d04f9cf23f52857e`
passed `make verify` and the required normal selection, then reproduced a
three-second race-mode backpressure timeout. A CPU profile located credential
regexp scans over the diagnostic lookahead. An ASCII prefilter skips scans only
when no marker can match; Unicode folding and secret redaction are unchanged.
The unchanged completion/backpressure and redaction checks passed three race
runs. Record volume, byte limits, instrumentation and deadlines were retained.

Before PR readiness, record the following ordered checks through the controller
on the final clean PR head; the exact result belongs to that head:

- `make verify`
- `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestPiTransport|TestPiGeneration|TestHarnessGenerationBinding)' -count=1 -timeout=3m`
- `go test -race ./internal/harness/pi ./internal/app -run '^(TestPiTransport|TestPiGeneration)' -count=1 -timeout=3m`
- `CHR_TEST_PI_EXECUTABLE="$(mise where pi@1.3.0)/pi/pi" go test ./internal/harness/pi -run '^TestPiGenerationPinnedInspection$' -count=1 -timeout=1m`
- `git diff --check`

## Risks and Follow-ups

Native generation requires an explicit private static API-key reference. OAuth
is unsupported without an owned refresh protocol. Worker/runtime behavior
remains the next recorded slice; real-model acceptance remains the approved
opt-in final acceptance item.
