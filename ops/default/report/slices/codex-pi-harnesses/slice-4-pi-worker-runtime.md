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

## Implementation Notes

T4.1 adds `runtime/agent/pi/` to standard, minimal and personal profiles. Its
Dockerfile uses that directory as the build context. Existing runtime paths and
images remain valid. The standard image built as
`sha256:4e2a6c5c80a972f8668307ad782db247b3869c186267f0a20a4440ce36ea18f8`.
The first build verified the Node checksum but exposed a missing `xz` extractor;
adding the required archive utility made the unchanged pinned input build pass.

Worker dispatch, private call state and credential refresh ownership are the
remaining ordered tasks. Runtime inspection is ready for their admission path.

## Contracts Added or Changed

- Sibling Pi image preparation and local immutable image inspection.
- Missing-asset upgrade preserves project-owned runtime files.

## Files of Interest

- `cmd/chr/assets/workspace/profiles/*/runtime/agent/pi/`
- `internal/harness/pi/runtime.go`
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

## Risks and Follow-ups

T4.2 and T4.3 remain within this slice. Real paid model acceptance remains the
approved opt-in final acceptance item.
