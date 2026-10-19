---
id: TKT-20261019103830
title: "Pi Docker workers and private runtime setup"
status: reviewing
kind: task
severity: medium
priority: normal
scope: infra
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:30Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T15:01:47Z
reviewing_at: 2026-10-19T16:38:20Z
branch: feat/slice-4-pi-worker-runtime
commits: [77bea3e585b26aed356594524e9162dbd2cb914c, 4e6a54bb85c4784564877bc1dcac01080f81d858, 29ee737a40ee280057376cd16bc35a69e2c075ae, c9dd7d4f950721de3bc1d24b44fb2c6b96c66f86]
---

# Pi Docker workers and private runtime setup

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 4 of 8 — Pi Docker workers and private runtime setup.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-4-pi-worker-runtime`.

Expected PR title: `feat(slice-4): execute Pi workers in their selected Docker runtime`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-4-pi-worker-runtime.md`.

Activated after canonical PR #414 was verified merged into dev.
Tasks proceed in recorded order; isolated container fixtures use deterministic
owned endpoints without paid provider calls.

Prerequisite: Slice 3 integrated into dev; implementation authorization for this set.

## Required Outcome

Pi worker RPC runs inside the chosen container with pinned runtime inputs and isolated authentication/session state; existing Codex runtimes remain usable.

Owners: Bundled standard/minimal/personal runtime assets and explicit upgrade path; core runtime assembly, worker bridge and app readiness.

Acceptance: HP-01, HP-02, HP-07, HP-08, HP-12, HP-13, HP-14, HP-16, HP-17.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T4.1 | Add sibling Pi runtime assets using the resolved pins below and a committed npm dependency lock; update all bundled profile initialization and explicit upgrade/readiness paths while preserving existing Codex Dockerfile/image references. |
| T4.2 | Execute the selected worker bridge and Pi RPC inside the selected container through owned pipes; mount only explicit profile credentials/model settings, allocate private writable call state and implement tested OAuth refresh exclusion without importing host home or arbitrary resources. |
| T4.3 | Retain invocation/claim/retry/result-validation authority and shared capacity; enforce finite-budget and merger-broker capability restrictions, verify container-reachable custom endpoints and provide actionable setup failures with no automatic fallback. |

## Validation

Build and inspect one selected Pi image, exercise a real container with a deterministic local protocol/model endpoint, and run two isolated calls against the shared capacity boundary. Verify only the intended mounts/flags and credential owner. An authorized real-model disposable task supplies HP-08 during final acceptance.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^(TestPiWorker|TestPiRuntime|TestHarnessBinding)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/core ./internal/app -run '^(TestPiWorkerIsolation|TestPiCredentialRefresh)' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.

## Implementation Evidence

T4.1: sibling runtime assets, non-destructive explicit upgrades and bounded local
image inspection are implemented. The pinned image built as
`sha256:4e2a6c5c80a972f8668307ad782db247b3869c186267f0a20a4440ce36ea18f8`.
The build exposed a missing archive extractor, corrected without changing pins.

T4.2: frozen bridge/image dispatch, explicit mounts and private call state are
implemented. Worker tools use the pinned SDK at the checkout; Pi runs from
private state. Existing reservation files survive the host process until the
runtime stops. Connection loss cancels RPC, and adopted detached tool processes
are reaped before capacity or refreshed credentials are released.

The refresh owner uses cross-process exclusion and atomic publication. A stress
run exposed an inspector retaining a replaced inode. Canonical identity checks
and reopening preserve private-file rules. Two concurrent network-disabled
containers use Pi 1.3.0's actual `AuthStorage.modify` API on private copies and
retain both synthetic OAuth rotations. No real account was used.

T4.3: role preparation distinguishes native Pi generation from the static Linux
CHR worker bridge without launching it. Missing or incompatible preparation gives
setup guidance without fallback. Finite budgets reject uncertified workers before
launch; merger-broker input is rejected. Custom container-network endpoints work
without host networking; container loopback endpoints fail preparation.

Two real Pi workers with different credentials reached an endpoint barrier
concurrently. Docker inspection checked explicit mounts and process restrictions.
Each retained private state, committed a separate output and passed host Git
validation. A new Codex claim and a native operation could not bypass occupied
Pi slots. Capacity returned after reaping. Poisoned project Pi resources were
excluded while explicit project instructions reached model context.

Successful focused commands:

- `go test ./internal/harness/pi ./internal/core ./internal/app -run '^(TestPiWorker|TestHarnessBindingBudget|TestHarnessSelection)' -count=1 -timeout=3m`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test -race ./internal/app -run '^(TestPiWorkerIsolationContainer|TestPiWorkerContainer)$' -count=1 -timeout=3m`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test ./internal/harness/pi -run '^TestPiCredentialRefreshStorage$' -count=1 -timeout=3m`
- `go test -race ./internal/harness/pi -run '^TestPiCredentialRefresh' -count=3 -timeout=1m`

Source task commits are retained in frontmatter. The slice report carries image,
protocol and development failure/correction evidence. The controller records
final ordered checks against the clean PR HEAD. Owned fixtures stopped after
their checks. Real-model acceptance remains the approved final opt-in item.

Canonical PR: pending.
