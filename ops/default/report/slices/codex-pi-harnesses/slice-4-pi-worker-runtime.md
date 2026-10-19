# Slice 4: Pi Docker Workers and Private Runtime Setup

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-4-pi-worker-runtime`
PR: pending

## Purpose

Execute selected Pi workers inside their prepared Docker runtime while retaining
host-owned invocation, capacity, retry and result validation.

## Delivered Behavior

- All three bundled profiles install sibling Pi runtime assets. Explicit
  `chr runtime init` adds missing files without replacing Codex runtime files.
- The Pi image uses the approved Go manifest, verified platform-specific Node
  22.23.3 archive and Pi 1.3.0 with a committed transitive npm lock. Installation
  disables dependency lifecycle scripts. Asset installation performs no probes,
  image builds, authentication or role changes.
- Bounded local image inspection validates platform/version labels and returns
  the immutable image ID. Unsupported or Codex-only images fail preparation.
- A separate worker transport profile admits only pinned builtin tools and
  correlates their names and IDs. Generation retains its restricted tools and
  stricter output ceiling.
- The worker credential owner serializes OAuth actions across processes and
  atomically publishes valid refreshed credentials after process reaping.
  Each call has a private copy; failure, cancellation and conflicting operator
  replacement preserve the owner boundary.

## Implementation Notes

T4.1 adds `runtime/agent/pi/` to standard, minimal and personal profiles. Its
Dockerfile uses that directory as the build context. Existing runtime paths and
images remain valid. The standard image built as
`sha256:4e2a6c5c80a972f8668307ad782db247b3869c186267f0a20a4440ce36ea18f8`.
The first build verified the Node checksum but exposed a missing `xz` extractor;
adding the required archive utility made the unchanged pinned input build pass.

The T4.2 worker contract component adds selected-provider credentials, private
copies and a dedicated OAuth directory with cross-process exclusion. Its action
must reap Pi before returning. Valid refresh writes survive a later model failure;
invalid updates or conflicting operator replacement cannot overwrite the source.

Container bridge dispatch, private runtime state, actual pinned Pi refresh
integration and T4.3 admission/capacity/endpoint checks remain in this slice.
The new owner and transport profiles are not yet connected to worker dispatch.

## Contracts Added or Changed

- Sibling Pi image preparation and local immutable image inspection.
- Missing-asset upgrade preserves project-owned runtime files.
- Explicit worker tool correlation and separate worker/generation output bounds.
- Private selected-provider worker resources and atomic OAuth credential owner.

## Files of Interest

- `cmd/chr/assets/workspace/profiles/*/runtime/agent/pi/`
- `internal/harness/pi/runtime.go`, `worker_resources.go`
- `internal/harness/pi/transport.go`, `inspection_events.go`
- `internal/app/pi_runtime_test.go`
- `docs/reference/project-workspace.md`

## Validation

Passed focused T4.1 checks:

- `docker build --file cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi/Dockerfile --tag chr-agent-pi:1.3.0 cmd/chr/assets/workspace/profiles/standard/runtime/agent/pi`
- `CHR_TEST_PI_RUNTIME_IMAGE=chr-agent-pi:1.3.0 go test ./internal/harness/pi ./internal/core ./internal/app -run '^(TestPiRuntime|TestRuntimeInit|TestInit.*Profile)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/harness/pi ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The opt-in image check inspects the real image and verifies both binary versions
inside a container with networking disabled. Owned containers exited and were
removed. This establishes image preparation; it does not establish worker model
execution or authentication. Final exact-head checks remain pending.

Passed focused T4.2 component checks:

- `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestPiTransport|TestPiGeneration|TestPiCredentialRefresh|TestPiRuntime|TestHarnessBinding)' -count=1 -timeout=3m`
- `go test -race ./internal/harness/pi -run '^(TestPiCredentialRefresh|TestPiTransport|TestPiGeneration)' -count=1 -timeout=2m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/harness/pi`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Eight concurrent callers and one independent subprocess preserve every simulated
refresh update. Failure/cancellation, unknown usage, tool authority and output
bounds also pass. The provider's actual refresh behavior still needs the planned
pinned-container integration; these fixtures do not establish OAuth authentication.
Owned subprocesses ended.

## Risks and Follow-ups

T4.2 and T4.3 remain within this slice. Real paid model acceptance remains the
approved opt-in final acceptance item.
