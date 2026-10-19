---
id: TKT-20261019103830
title: "Pi Docker workers and private runtime setup"
status: in_progress
kind: task
severity: medium
priority: normal
scope: infra
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:30Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T15:01:47Z
branch: feat/slice-4-pi-worker-runtime
commits: []
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

## Implementation Progress

T4.1 runtime assets and inspection are implemented. Standard, minimal and personal
initialization include the sibling Pi Dockerfile and exact npm dependency lock.
The explicit missing-asset upgrade is non-destructive. Bounded image inspection
rejects unsupported platform/version metadata and returns the immutable ID for
worker admission. Dispatch/readiness integration follows in T4.2 and T4.3.

The selected image built as
`sha256:4e2a6c5c80a972f8668307ad782db247b3869c186267f0a20a4440ce36ea18f8`.
The initial build found a missing `xz` extractor after verifying the Node archive;
adding the required utility produced a successful build without changing pins.

Successful focused checks:

- `docker build --file cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi/Dockerfile --tag chr-agent-pi:1.3.0 cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test ./internal/harness/pi ./internal/core ./internal/app -run '^(TestPiRuntime|TestRuntimeInit|TestInit.*Profile)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/harness/pi ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The real container version check uses no network or credentials. All owned
containers exited and were removed. No model call occurred. T4.2 and T4.3 are
pending; no canonical PR is open yet.
