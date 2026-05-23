# Workspace Index And Semantic Cache Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`

## Slicing Checklist

- [x] Capture workspace index and semantic cache direction.
- [x] Capture `Draft` artifact removal direction.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Workspace Index Contract | Delivered | `feat/slice-1-workspace-index-contract` | `feat(slice-1): workspace index contract` | `.prj/reports/slices/workspace-index-semantic-cache/slice-1-workspace-index-contract.md` | PR #164. |
| Slice 2: Structural Workspace Index Model | Planned | `feat/slice-2-structural-workspace-index` | `feat(slice-2): structural workspace index` | `.prj/reports/slices/workspace-index-semantic-cache/slice-2-structural-workspace-index.md` | Adds index model shape. |
| Slice 3: Index Load And Refresh Semantics | Planned | `feat/slice-3-index-refresh-semantics` | `feat(slice-3): index refresh semantics` | `.prj/reports/slices/workspace-index-semantic-cache/slice-3-index-refresh-semantics.md` | Handles direct file changes. |
| Slice 4: Discardable Structural Cache | Planned | `feat/slice-4-structural-index-cache` | `feat(slice-4): structural index cache` | `.prj/reports/slices/workspace-index-semantic-cache/slice-4-structural-index-cache.md` | Adds local structural cache. |
| Slice 5: Semantic Digest Cache Contract | Planned | `feat/slice-5-semantic-digest-cache` | `feat(slice-5): semantic digest cache` | `.prj/reports/slices/workspace-index-semantic-cache/slice-5-semantic-digest-cache.md` | Adds digest schema and invalidation. |
| Slice 6: Semantic Index Worker Boundary | Planned | `feat/slice-6-semantic-index-worker` | `feat(slice-6): semantic index worker` | `.prj/reports/slices/workspace-index-semantic-cache/slice-6-semantic-index-worker.md` | Adds optional LLM boundary. |
| Slice 7: Service Queries And Suggestions | Planned | `feat/slice-7-index-service-queries` | `feat(slice-7): index service queries` | `.prj/reports/slices/workspace-index-semantic-cache/slice-7-index-service-queries.md` | Exposes adapter-facing queries. |
| Slice 8: TUI Index Consumers | Planned | `feat/slice-8-tui-index-consumers` | `feat(slice-8): tui index consumers` | `.prj/reports/slices/workspace-index-semantic-cache/slice-8-tui-index-consumers.md` | Uses queries in TUI selectors/status. |
| Slice 9: Draft Artifact Removal | Planned | `feat/slice-9-remove-draft-artifact` | `feat(slice-9): remove draft artifact` | `.prj/reports/slices/workspace-index-semantic-cache/slice-9-remove-draft-artifact.md` | Removes first-class Draft semantics. |

## Slice 1: Workspace Index Contract

Candidate tasks:

- [x] Update architecture/product documentation with the workspace index model.
- [x] Define structural vs semantic index responsibilities.
- [x] Define cache discardability and invalidation rules.
- [x] Define refresh behavior for direct filesystem edits.
- [x] Define optional/background semantic digest generation behavior.
- [x] Document that `Draft` is not a first-class document kind.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `18a4351` documents workspace index architecture.
- T1.2: `18a4351` documents semantic cache and refresh contracts.
- T1.3: `18a4351` documents draft artifact removal direction.

## Slice 2: Structural Workspace Index Model

Candidate tasks:

- [ ] Add workspace index model types.
- [ ] Map existing workspace graph nodes into index nodes.
- [ ] Add explicit edge types.
- [ ] Add lookup maps for kind, path, case slug, ticket key, and case slug.
- [ ] Preserve existing graph API where needed during transition.
- [ ] Add focused model and builder tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1: pending.
- T2.2: pending.
- T2.3: pending.

## Slice 3: Index Load And Refresh Semantics

Candidate tasks:

- [ ] Load or build the structural index at app startup.
- [ ] Compute modified time and source hash for indexed files.
- [ ] Detect changed, added, and removed files.
- [ ] Rebuild or target-update the structural index after refresh.
- [ ] Preserve direct-file-edit workflows.
- [ ] Expose index health and diagnostics through service results.
- [ ] Add tests for changed files and stale cache inputs.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending.
- T3.2: pending.
- T3.3: pending.

## Slice 4: Discardable Structural Cache

Candidate tasks:

- [ ] Define cache manifest schema.
- [ ] Write `structural-index.json` under `.chc/tmp/cache/workspace-index/`.
- [ ] Load cache when schema and source hashes remain valid.
- [ ] Fall back to filesystem rebuild when cache is missing or stale.
- [ ] Add maintenance or debug visibility for cache state where practical.
- [ ] Add tests for cache hit, cache miss, stale cache, and cache deletion.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending.
- T4.2: pending.
- T4.3: pending.

## Slice 5: Semantic Digest Cache Contract

Candidate tasks:

- [ ] Add semantic digest types.
- [ ] Add cache entry schema with version, model id, prompt hash, source hash,
      path, and digest.
- [ ] Add deterministic fake digest generator for tests.
- [ ] Mark missing or stale semantic digests as pending.
- [ ] Keep digest generation optional and non-blocking.
- [ ] Add tests for digest cache invalidation paths.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending.
- T5.2: pending.
- T5.3: pending.

## Slice 6: Semantic Index Worker Boundary

Candidate tasks:

- [ ] Define semantic digest generator interface.
- [ ] Define request and response contracts for per-node digest generation.
- [ ] Wire a disabled-by-default background worker or service method.
- [ ] Ensure tests use fake generators only.
- [ ] Surface pending/stale digest counts in index health.
- [ ] Prevent UI startup from blocking on semantic indexing.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending.
- T6.2: pending.
- T6.3: pending.

## Slice 7: Service Queries And Suggestions

Candidate tasks:

- [ ] Add service query methods for listing and suggesting cases.
- [ ] Add service query methods for resolving paths and related documents.
- [ ] Add search over structural metadata and semantic digests when available.
- [ ] Ensure adapters do not parse workspace files directly for these queries.
- [ ] Add tests for case suggestions and related documents.
- [ ] Add tests for unresolved references and stale semantic status.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending.
- T7.2: pending.
- T7.3: pending.

## Slice 8: TUI Index Consumers

Candidate tasks:

- [ ] Replace direct graph projection call sites with service/index query
      results where appropriate.
- [ ] Add a use-case selector or suggestion surface for artifact creation flows
      that need case association.
- [ ] Show index health and stale semantic digest status in an appropriate view.
- [ ] Keep TUI interaction non-blocking when semantic digests are pending.
- [ ] Add TUI tests for suggestions and index health display.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending.
- T8.2: pending.
- T8.3: pending.

## Slice 9: Draft Artifact Removal

Candidate tasks:

- [ ] Remove draft-specific TUI command palette actions or mark them unavailable
      according to current command behavior.
- [ ] Migrate active docs from draft workflow language to dump/quiz/spec
      language.
- [ ] Update authoring references so quiz/dump material can produce specs with
      `status: draft`.
- [ ] Remove draft document family from active index classification if present.
- [ ] Preserve historical notes only when marked historical or superseded.
- [ ] Add tests that active document classification and TUI surfaces do not
      expose `Draft` as a first-class document kind.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T9.1: pending.
- T9.2: pending.
- T9.3: pending.

## Open Decisions

- Whether the structural index should fully replace `WorkspaceGraph` or wrap it
  for one compatibility phase.
- Whether structural cache should include all lookup maps or only enough source
  facts to rebuild maps quickly.
- Exact semantic digest schema fields after the first fake-generator pass.
- Whether semantic digest generation should be a service command, background TUI
  worker, daemon job, or all three over the same service boundary.
- Whether draft cleanup should happen before or after semantic indexing if
  active draft references cause classification churn.

## Validation Log

- 2026-05-23: `make verify` passed for plan and tracker creation.
- 2026-05-23: `make verify` passed for Slice 1.
