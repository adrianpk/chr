# Slice 6: Semantic Index Worker Boundary

Status: reviewing
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-6-semantic-index-worker`
PR: pending

## Purpose

Add an explicit semantic digest worker boundary that can generate pending and
stale digests without making TUI startup or workspace refresh depend on hosted
LLM calls.

## Delivered Behavior

- Define per-node semantic digest generation requests.
- Define per-node semantic digest generation results.
- Add a semantic index worker that consumes a structural index, digest cache,
  generator, model id, and prompt hash.
- Generate digests only for pending and stale nodes.
- Preserve ready digest entries without regenerating them.
- Keep generation explicit and disabled by default; startup and refresh paths do
  not invoke the worker.
- Surface pending and stale digest counts through workspace index health.
- Keep tests on the deterministic fake generator only.

## Implementation Notes

The worker is synchronous as a boundary object, but it is not wired into startup
or refresh. Later adapters can call it from a service command, background task,
or daemon path without changing the digest cache contract.

This slice still avoids hosted LLM calls. The generator interface is the
boundary; production generator wiring belongs outside this slice.

## Contracts Added or Changed

- Added `WorkspaceSemanticDigestRequest`.
- Added `WorkspaceSemanticDigestGeneration`.
- Added `WorkspaceSemanticIndexWorker`.
- Added `WorkspaceSemanticIndexWorkerResult`.
- Changed `WorkspaceSemanticDigestGenerator` to accept a request object.
- Added `WorkspaceIndex.HealthWithSemanticDigestStates`.
- Extended `WorkspaceIndexHealth` with semantic pending and stale counts.

## Files of Interest

- `internal/chc/workspace_semantic_digest.go`
- `internal/chc/workspace_semantic_digest_test.go`
- `internal/chc/workspace_index.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-6-semantic-index-worker.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chc` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Worker invocation is intentionally explicit. Slice 7 and Slice 8 can decide
  where query surfaces should read pending/stale digest state.
