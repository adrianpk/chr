# Slice 5: Semantic Digest Cache Contract

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-5-semantic-digest-cache`
PR: #168

## Purpose

Define the semantic digest cache contract used by later indexing workers and
query surfaces without introducing live LLM calls or blocking startup behavior.

## Delivered Behavior

- Add semantic digest types keyed by workspace index node id.
- Store digest cache entries with model id, prompt hash, source hash, path,
  node kind, digest text, and generation timestamp.
- Read and write `semantic-digests.json` under
  `.chr/tmp/cache/workspace-index/`.
- Treat missing, invalid, and schema-mismatched cache files as empty
  discardable caches.
- Mark digest states as `ready`, `pending`, or `stale`.
- Mark missing digest entries as pending.
- Mark digest entries stale when source hash, model id, or prompt hash changes.
- Add a deterministic fake digest generator for tests and future worker tests.

## Implementation Notes

This slice only defines cache contracts and deterministic test support. It does
not schedule background digest generation and does not call a hosted LLM.

The cache remains local and discardable. Later slices can use the same read/write
contract from a worker boundary or service command.

## Contracts Added or Changed

- Added `WorkspaceSemanticDigestStatus`.
- Added `WorkspaceSemanticDigest`.
- Added `WorkspaceSemanticDigestState`.
- Added `WorkspaceSemanticDigestCache`.
- Added `WorkspaceSemanticDigestGenerator`.
- Added `DeterministicWorkspaceSemanticDigestGenerator`.
- Added semantic digest cache read/write helpers.
- Added digest state evaluation for workspace indexes.

## Files of Interest

- `internal/chr/workspace_semantic_digest.go`
- `internal/chr/workspace_semantic_digest_test.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-5-semantic-digest-cache.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chr` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Digest content is not generated automatically in this slice. Worker
  scheduling, non-blocking generation, and stale digest counts belong to Slice 6.
