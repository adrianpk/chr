# Slice 8: Harness Acceptance and Operator Guidance

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `test/slice-8-harness-acceptance`
PR: pending

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

T8.2 guidance/evidence mapping and T8.3 authorized live proof remain in progress.

## Contracts Added or Changed

- `TestHarnessJourney` covers active work, role combinations, all four adapters,
  persisted reads, worker resume and legacy direct API dispatch.

## Files of Interest

- `internal/app/harness_journey_test.go`
- `internal/app/harness_journey_fixture_test.go`
- `internal/app/harness_journey_adapters_test.go`

## Validation

Passed focused task checks:

- `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m -v`
- `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/app`
- `git diff --check`

The four named adapter cases executed. Owned HTTP servers, SQLite handles,
claims and generation subprocesses finished their cleanup. The actual generation
caller validates its typed output before retaining successful parser evidence.
Routine fixtures use synthetic credentials and perform no provider/bot calls.

## Risks and Follow-ups

T8.3 requires explicit operator authorization for one real Pi structured call
and one bounded disposable Docker worker task. Native Pi 1.3.0 is available,
but this worker cannot access the Docker socket and passwordless sudo is not
available. Real-provider proof has not run. The original feature ticket remains
open until required evidence and controller-owned integration/closure complete.
