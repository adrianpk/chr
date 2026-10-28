# Go Quality Gates

This repository keeps interactive development checks separate from exhaustive
scheduled verification. Use `.agents/playbooks/testing-strategy.md` for test
selection and `.agents/playbooks/nightly-verification.md` for CI operation.

## Development Gate

`make verify` runs strict lint, short Go tests, CI script regressions and the
application build. `make test`
uses `go test -short -timeout=3m ./...`. Short mode skips real Git workflow
fixtures and live-forge certification at their shared setup boundaries. Cheap
logic, parser, policy and filesystem tests remain enabled.

Before a task commit, format changed Go files and run focused regressions for
changed behavior. Before a slice merge or delivery-set closure, run `make verify`
and the smallest relevant regression selection. For changed IO or workflow
behavior, explicitly select the affected integration tests without `-short`.
For changed concurrency behavior, run that focused selection with `-race`.

Do not add `go test ./...`, `go test -race ./...`, coverage campaigns or full QA
suites to each task, slice or controller final gate. Future controller
configurations use `make verify` as their local final gate. Preserve frozen
configurations and evidence for already active runs.

Documentation-only changes require diff and whitespace validation.

## Nightly Gate

`make nightly` runs lint, build, one uncached complete local suite, one complete
race suite, one browser workflow check and the opted-in live-forge suite once. The normal and race passes
exclude browser and live certification; separate passes run those named tests.
Acceptance and fault manifests are checked before execution. Their tests are
already part of the complete suite and are not repeated as a third local pass.

The Forgejo schedule runs once daily against the captured `dev` commit. Record
that exact hash and retain per-test timings and failure evidence. A successful
short gate does not establish exhaustive or live verification. A new commit
cannot reuse an older commit's nightly result.

Run long gates locally only when the maintainer explicitly requests that run.
Do not dispatch, emulate or repeat a nightly as part of ordinary delivery.
Nightly failures require a focused correction followed by the next scheduled
run, without imposing the whole suite on interactive development.

## Commands

| Command | Purpose |
| --- | --- |
| `make test` | Short tests, with a three-minute timeout per package |
| `make verify` | Strict lint, short Go tests, CI script regressions and build |
| `go test ./internal/core/workflow -run '^TestWorkflowRoundTrip$' -count=1` | Explicit focused integration regression |
| `go test -race ./internal/core/workflow -run '^TestWorkflowRoundTrip$' -count=1` | Focused race mode when relevant |
| `make nightly` | Complete scheduled local gate; live enabled only by fixture opt-in |
| `make test-full` | Uncached full local tests, for explicitly requested diagnostics |
| `make test-race` | Uncached full race tests, for explicitly requested diagnostics |
| `make qa-ticket-native-workflow` | Explicit acceptance/fault diagnosis |
| `make qa-native-integration` | Explicit provider and transport diagnosis |

Keep defect-oriented lint checks enabled. Do not remove assertions or reduce
nightly coverage to make development checks faster. Report the exact command,
result and evidence boundary when closing work.
