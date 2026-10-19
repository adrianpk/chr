# Slice 8: Harness Acceptance and Operator Guidance

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `test/slice-8-harness-acceptance`
PR: [#419](https://forge.adrianpk.com/adrianpk/chr/pulls/419)

## Purpose

Establish bounded cross-interface evidence, usable profile/runtime guidance and
explicitly authorized live Pi proof for the approved harness contract.

## Delivered Behavior

- One deterministic journey switches all four worker/generation combinations
  through Web, TUI, Telegram and CLI using the same registered project.
- A real generation subprocess remains blocked while defaults change; its
  completion, nested calls and resumed artifact keep the original executable,
  model and binding. Fresh generation calls use the newly selected transport.
- A claimed worker remains unchanged during switches, then pauses, resumes and
  executes its original Codex driver with host-validated Git output.
- Fresh worker admissions bind one complete selection revision. Pi admissions
  retain their selected immutable runtime identity.
- Codex-only preparation does not inspect Docker or require Pi profiles. Clearing
  generation restores the existing direct API adapter through owned HTTP IO.
- All interfaces and an independent process observe persisted defaults; unrelated
  configuration bytes remain intact.

## Implementation Notes

T8.1 reuses existing production service, receiver, profile and protocol fixtures.
The root has four named transport cases; it does not duplicate the adapter
contract suites. Native Codex/Pi protocol peers and a narrowly constrained
Docker metadata executable make routine checks deterministic. The selected
static Linux worker bridge is built once. Actual Docker execution and remote
model proof are separate from this fixture's admission evidence.

The CLI selects the registered project through its existing active-project
store. An explicit directory has its own directory-derived project identity;
it cannot replace the selected registry identity in this journey.

T8.2 adds the setup/recovery guide and corrects the generation reference to
state explicit-profile precedence, legacy direct API behavior, native Pi's
static-key limitation and observed-versus-certified usage. Registered-project
selection is documented consistently across adapters.

T8.3 provides `TestHarnessLivePi`, an opt-in real-provider fixture. Its closed
input requires a private root, explicit native Pi/bridge/image, provider/model
and private auth source. It copies only the selected provider into private
fixture credentials. Both roles and Docker readiness are checked before any
model call. One structured call uses the existing 60-second deadline; one
worker task uses 90 seconds, one attempt and one checkpoint. Host validation
requires the committed `result.txt` to contain exactly `Delivered` and no other
file change. No push or external repository mutation is requested.

The fixture retains candidate, provider/model, pinned Pi version, immutable
runtime, worker commit/output hash, observed generation usage and caller
validation in private `evidence.json`. The missing-access guard runs an owned
child with synthetic configuration and verifies failure before a model launches.
The real test skips without opt-in; that skip is not live evidence.

T8.3 authorized live execution remains pending. The deterministic fixtures establish
shared application/transport behavior and do not stand in for a real model call.

Task commits: T8.1 `c1374d643d4db2c3bcfa4cdef9a600b9c039fdc2`,
T8.2 `297042e63d220d553f830928626e68ac9d55c908`, and T8.3 preparation
`d131373dc63526920bb4c88728a9f477d0e56f5e`.

## Contracts Added or Changed

- `TestHarnessJourney` covers active work, role combinations, all four adapters,
  persisted reads, worker resume and legacy direct API dispatch.

## Files of Interest

- `internal/app/harness_journey_test.go`
- `internal/app/harness_journey_fixture_test.go`
- `internal/app/harness_journey_adapters_test.go`
- `docs/how-to/select-harnesses.md`
- `docs/reference/artifact-generation-backends.md`
- `docs/how-to/user-guide/index.md`, `docs/index.md`
- `internal/app/harness_live_pi_test.go`, `harness_live_pi_guard_test.go`

## Validation

Passed focused task checks:

- `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m -v`
- `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`
- `go test ./internal/app -run '^(TestHarnessJourney|TestHarnessLivePi|TestHarnessLivePiGuard)$' -count=1 -timeout=3m -v`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/app`
- `git diff --check`

Documentation checks verified 129 local Markdown link targets.
`go list -f '{{join .Imports "\n"}}' ./internal/core` showed only neutral
harness/process dependencies, with no concrete Codex/Pi adapter import.

The four named adapter cases executed. Owned HTTP servers, SQLite handles,
claims and generation subprocesses finished their cleanup. The actual generation
caller validates its typed output before retaining successful parser evidence.
Routine fixtures use synthetic credentials and perform no provider/bot calls.
The combined check runs the journey and missing-access guard; the real-provider
root is explicitly skipped until its separate authorization and access exist.

Final validation will be recorded through the controller on the clean final
PR head after live-proof authorization and access are available. The required
prefix is:

- `make verify`
- `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`

The focused missing-access guard and one opted-in live proof follow that prefix,
then `git diff --check`. These final checks have not run for this candidate.
After controller integration, the controller owns the unchanged development gate
and final documentation closure. This PR does not close the original ticket.

## Acceptance Evidence

These rows point to the owning merged-slice evidence and the bounded current
journey. Earlier successful commands were not rerun as an exhaustive matrix.
Final clean-candidate checks and post-merge development-gate evidence remain
separate controller records.

| Criterion | Owning evidence and current result |
| --- | --- |
| HP-01 | [Slice 1](slice-1-harness-codex.md) Codex contract regressions; current journey prepares and binds Codex without Pi profiles or Docker inspection. |
| HP-02 | Current journey admits and generates all four independent role combinations through real shared operations. |
| HP-03 | [Slice 3](slice-3-pi-generation.md) covers all generation families, contextual Quiz, conversation, explanations, review and tracking; current journey dispatches both native generation protocols. |
| HP-04 | [Slice 2](slice-2-harness-selection.md) legacy action dispatch and Slice 3 API-to-Pi Plan transitions; current journey clears a profile and calls the original direct API path. |
| HP-05 | Slice 1 common process/contract suite and Slice 3 Pi transport cancellation, output and capability coverage. |
| HP-06 | Slice 1 neutral core ownership; current production core import graph has no concrete Codex/Pi adapter dependency. |
| HP-07 | [Slice 4](slice-4-pi-worker-runtime.md) real container task/host validation; current resumed worker produces host-validated Git output without an integration claim. |
| HP-08 | Slice 4 actual pinned Docker image and disposable task passed against an owned model endpoint. T8.3 additionally requires the authorized real-provider Docker task; pending. |
| HP-09 | Slice 3 pinned Pi structured output and invalid-artifact preservation passed with owned provider IO. T8.3 additionally requires the authorized real-provider structured call; pending. |
| HP-10 | Slice 3 shared JSONL framing, bounds, interleaving, provider/exit/abort cases and Slice 4 worker tool correlation. |
| HP-11 | Slice 3 native generation inspection/isolation and Slice 4 poisoned-context container assertions keep discovered tools/extensions/settings out. |
| HP-12 | Slice 2 readiness, Slice 3 explicit resource/profile rejection and Slice 4 bridge/image/network preparation checks reject missing setup without fallback. |
| HP-13 | Slice 3 shared capacity/private state; Slice 4 concurrent real containers use distinct credentials/state and the same admitted capacity pool. |
| HP-14 | Slice 2 retained invocation/generation artifacts and current journey's active completion, nested generation, pause/resume and original worker execution. |
| HP-15 | Slice 2 `TestHarnessBindingOldReader` rejects the mandatory new invocation format before execution. |
| HP-16 | Slice 3 unknown/observed usage and Slice 4 worker budget/merger refusal; guidance states that observations do not certify whole-call enforcement. |
| HP-17 | Slice 4 actual container-network endpoint and loopback/host-gateway boundary tests; remote model access remains part of pending T8.3 proof. |
| HP-18 | Setup/recovery guide, corrected generation reference, User Guide links and the existing Unreleased Codex/Pi capability cover both roles and interfaces. |
| HP-19 | [Slice 5](slice-5-web-harness-selection.md), [Slice 6](slice-6-tui-harness-selection.md), [Slice 7](slice-7-telegram-harness-selection.md) adapter guards; current journey writes/reads all four adapters and an independent process. |
| HP-20 | Current journey switches while the original generation subprocess and worker claim remain active; fresh admissions change while nested/resumed work keeps its binding/policy. |
| HP-21 | Slice 2 atomic image/replay and unrelated bytes, Slices 5-7 stale/unauthorized/lost acknowledgements, and current journey's complete-revision admissions/unrelated-byte preservation. |

## Risks and Follow-ups

T8.3 requires explicit operator authorization for one real Pi structured call
and one bounded disposable Docker worker task. Native Pi 1.3.0 is available,
but this worker cannot access the Docker socket and passwordless sudo is not
available. The proposed fixed proof is `zai/glm-4.7` with thinking off, using the existing
selected static credential. This is a proposal, not authorization. Real-provider
proof has not run. The original feature ticket remains
open until required evidence and controller-owned integration/closure complete.
