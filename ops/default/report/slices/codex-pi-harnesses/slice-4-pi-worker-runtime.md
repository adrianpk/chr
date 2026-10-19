# Slice 4: Pi Docker Workers and Private Runtime Setup

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-4-pi-worker-runtime`
PR: [#415](https://forge.adrianpk.com/adrianpk/chr/pulls/415)

## Purpose

Execute selected Pi workers inside their prepared Docker runtime while retaining
host-owned invocation, capacity, retry and result validation.

## Delivered Behavior

- Standard, minimal and personal profiles include sibling Pi runtime assets.
  Explicit `chr runtime init` adds missing files without replacing Codex assets.
  The locked image contains Node 22.23.3 and Pi 1.3.0; dependency lifecycle
  scripts are disabled and unsupported platforms fail preparation.
- Worker admission checks an explicitly prepared static Linux CHR bridge and
  local Docker metadata, then freezes the executable, immutable image identity
  and selected model semantics. Native generation uses a separate Pi profile.
  Missing or incompatible preparation fails with setup guidance.
- The host bridge runs Pi RPC and its seven explicit builtin tools inside the
  selected container. Each call has private writable state and selected profile
  resources. Project instructions are supplied explicitly; discovered Pi
  settings/extensions cannot replace the frozen selection.
- OAuth has a dedicated private credential owner with cross-process exclusion.
  Each call uses a private copy. Valid refreshes are atomically published after
  process reaping, including a refresh followed by model failure. Invalid updates
  or conflicting operator replacement cannot overwrite the owner.
- Containers retain the existing admitted capacity files until their processes
  stop. Concurrent Pi, Codex and native operations use the same repository
  capacity boundary. Host disconnection cancels RPC; adopted detached tools are
  reaped before capacity or refreshed credentials are released.
- Frozen work survives live profile replacement. Claims, attempts, validation,
  artifacts and completion remain host-owned. Pi remains uncertified for finite
  whole-call budgets and merger-broker authority. Container-reachable custom
  endpoints work without enabling host networking.

## Implementation Notes

T4.1 installs the sibling runtime and bounded local inspection. The tested image
is `sha256:4e2a6c5c80a972f8668307ad782db247b3869c186267f0a20a4440ce36ea18f8`.
The first image build exposed a missing `xz` extractor; the required utility was
added while retaining all approved pins and the committed dependency lock.

T4.2 delegates transport to the existing Pi RPC adapter through owned pipes.
A host-owned extension reuses the pinned SDK builtin tools with execution
context bound to the checkout while Pi runs from private state. Only the
checkout, Git metadata, existing reservation files, bridge and selected resources
are mounted. The container has a read-only root, private tmpfs state, dropped
capabilities and no host home, Docker socket or blanket environment forwarding.

Slot conversion and delegation happen under the existing capacity coordinator.
The independently owned container bridge holds those same files and adopts and
reaps detached tool children. No second scheduler, workflow engine, credential
store or capacity pool was introduced.

Credential inspection takes a short shared publication lock; atomic refresh
publication takes its exclusive counterpart. Inspectors verify canonical inode
identity after locking and reopen a source replaced by rename. This avoids
observing publication temporaries without waiting through the whole model call.

T4.3 inspects executable build metadata without running a configured program,
keeps finite-budget and merger restrictions, and rejects container loopback
endpoints. Explicit host-gateway mapping still depends on host listener/firewall
policy. The working owned fixture uses an explicit container-network endpoint.

## Contracts Added or Changed

- Pinned sibling Pi image preparation and immutable local runtime inspection.
- Distinct worker tool correlation and worker/generation output bounds.
- Selected-provider credentials, private copies and atomic OAuth publication.
- Existing capacity reservation delegated to an independently owned runtime.
- Bounded connection framing, cancellation, explicit instructions and reaping.
- Separate native generation and static Linux worker bridge preparation.

## Files of Interest

- `cmd/chr/assets/workspace/profiles/*/runtime/agent/pi/`
- `internal/harness/pi/runtime.go`, `worker_resources.go`, `worker_container.go`
- `internal/harness/pi/worker_bridge.go`, `worker_connection.go`, `worker.mjs`
- `internal/app/pi_worker.go`, `pi_worker_preparation.go`, `service_harness_worker.go`
- `internal/core/worker_runtime.go`, `worker_runtime_reservation.go`
- `internal/harness/process/subreaper_linux.go`
- `internal/app/pi_worker_container_test.go`, `pi_worker_isolation_test.go`
- `internal/harness/pi/worker_storage_test.go`, `docs/reference/project-workspace.md`

## Validation

Successful focused task checks include:

- `docker build --file cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi/Dockerfile --tag chr-agent-pi:1.3.0 cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test ./internal/harness/pi ./internal/core ./internal/app -run '^(TestPiRuntime|TestRuntimeInit|TestInit.*Profile)' -count=1 -timeout=3m`
- `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestPiTransport|TestPiGeneration|TestPiRuntime|TestPiWorker|TestPiCredentialRefresh|TestHarnessBinding|TestHarnessSelection)' -count=1 -timeout=3m`
- `go test -race ./internal/harness/pi -run '^TestPiCredentialRefresh' -count=3 -timeout=1m`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test ./internal/harness/pi -run '^TestPiCredentialRefreshStorage$' -count=1 -timeout=3m`
- `go test ./internal/harness/pi ./internal/core ./internal/app -run '^(TestPiWorker|TestHarnessBindingBudget|TestHarnessSelection)' -count=1 -timeout=3m`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test -race ./internal/app -run '^(TestPiWorkerIsolationContainer|TestPiWorkerContainer)$' -count=1 -timeout=3m`

The actual image completed a disposable task after service recovery and profile
replacement. Two further workers reached an endpoint barrier concurrently with
different credentials, separate private state and independent Git outputs. Both
committed results passed host validation. Docker inspection checked the intended
mounts and process restrictions. Occupied Pi slots rejected a new Codex claim and
a native operation; the pool became available after the workers stopped.
Poisoned project Pi settings/extensions were excluded while explicit project
instructions reached the model context.

Two concurrent network-disabled containers also used Pi 1.3.0's actual
`AuthStorage.modify` API. Both synthetic OAuth rotations survived publication,
and the second owner observed the first rotation. Eight goroutines plus an
independent subprocess cover owner exclusion, failure and cancellation. This
establishes pinned storage behavior with synthetic provider IO, not real-account
authentication. All owned fixtures stopped after their checks.

Development checks caught and corrected the SDK tool-context mismatch and the
credential inspection/rename race. The original host validation and private-file
assertions were retained. Initial concurrent fixtures also retained slice-owned
contract fields and a wrong auth mount assertion; their setup was corrected
without changing production ownership or mount policy.

The controller records final validation against the clean PR HEAD and its exact
ordered command list. The canonical PR carries that candidate's successful
commands; task checks above are component evidence.

## Risks and Follow-ups

Real paid-model acceptance remains the approved opt-in final acceptance item.
OAuth provider network refresh is not certified by synthetic storage fixtures.
Host-gateway reachability remains subject to explicit host networking policy.
