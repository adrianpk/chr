# Architecture Boundary Cleanup Slicing Plan

Status: in_progress

## Purpose

Clean up application architecture boundaries discovered during the architecture
review without changing user-facing behavior.

The goal is to reduce hidden coupling between application startup, services,
CLI commands, and TUI adapters while keeping the current CLI and TUI surfaces
stable.

## Source

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.agents/ARCHITECTURE.md`
- `.agents/CONVENTIONS.md`
- `.agents/INVARIANTS.md`
- Architecture review findings from the current session

Review findings:

- `internal/chcapp/app.go` owns a package-level mutable `assetsFS` dependency.
- `internal/chcapp/service.go` stores `context.Context` in the service object.
- `internal/chcapp/metadata_cli.go` contains workflow logic that belongs behind
  service operations.
- `internal/chcapp/tui_adapters.go` mixes adapter functions for several
  unrelated domains in one file.
- `quizScopeFlags` and `quizScopeFromCommand` are used outside quiz-specific
  flows and should use neutral naming.

## Current State

Application startup builds CLI commands and TUI entrypoints through `chcapp`.
The service layer already centralizes many workflows, but several dependencies
and adapters still leak through older boundaries:

- asset embedding is passed partly through package-level state;
- command cancellation is represented as service object state;
- some CLI commands still perform load/patch/upsert logic directly;
- TUI adapter code is broad enough to make domain ownership harder to inspect;
- shared plan/spec scope helpers still use quiz-specific names.

These issues do not currently require product behavior changes. They are
architecture cleanup items intended to make later CLI/TUI, planning, tracking,
and ticket workflows easier to evolve.

## Target Model

The target model is:

- startup dependencies are explicit inputs, not mutable package state;
- service operations that can block or call agent-backed flows receive
  operation-scoped context;
- CLI and TUI call shared service operations instead of duplicating workflow
  logic;
- TUI adapter files are split by domain while remaining thin dispatch layers;
- shared scope helpers use neutral naming.

## Slice Map

Slicing name: Architecture Boundary Cleanup

Active slice number: Slice 1

Active slice short name: dependency context cleanup

Expected branch name: `feat/slice-1-dependency-context-cleanup`

Expected PR title: `feat(slice-1): dependency context cleanup`

Expected report path:
`.prj/reports/slices/architecture-boundary-cleanup/slice-1-dependency-context-cleanup.md`

Prerequisite PRs or delivered work:

- CLI TUI Surface Parity slicing completed.
- Documentation and parity matrix already describe current CLI and TUI
  operation surfaces.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | dependency context cleanup | `feat/slice-1-dependency-context-cleanup` | `feat(slice-1): dependency context cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-1-dependency-context-cleanup.md` |
| Slice 2 | operation context boundary | `feat/slice-2-operation-context-boundary` | `feat(slice-2): operation context boundary` | `.prj/reports/slices/architecture-boundary-cleanup/slice-2-operation-context-boundary.md` |
| Slice 3 | adapter service shape cleanup | `feat/slice-3-adapter-service-shape-cleanup` | `feat(slice-3): adapter service shape cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-3-adapter-service-shape-cleanup.md` |

## Slice 1: Dependency Context Cleanup

Remove startup dependency leakage around embedded assets and project context
resolution.

Expected behavior:

- `internal/chcapp` no longer relies on package-level mutable asset filesystem
  state.
- Project context resolution receives required dependencies explicitly.
- CLI and TUI startup behavior remains unchanged.
- Existing tests and verification continue to pass.

Out of scope:

- Changing command names or TUI behavior.
- Refactoring service operation context signatures.
- Moving metadata command workflow logic.

## Slice 2: Operation Context Boundary

Move cancellable or long-running service operations to operation-scoped context.

Expected behavior:

- Relevant service methods accept `context.Context` as an operation argument.
- Stored service context usage is removed or narrowed to non-operation setup.
- Existing `contextcheck` suppressions in CLI flow code are removed where the
  operation boundary now carries context correctly.
- CLI and TUI behavior remains unchanged.

Out of scope:

- Broad renaming of service methods unrelated to context boundaries.
- Adapter file splitting.

## Slice 3: Adapter Service Shape Cleanup

Move remaining workflow logic behind service operations and reduce adapter
surface noise.

Expected behavior:

- Metadata project and case patch/upsert workflows live behind service methods.
- CLI metadata commands delegate to service operations.
- TUI adapter wrappers are split by domain while staying behavior-preserving.
- Shared scope helpers use neutral names instead of quiz-specific names.
- Documentation or parity matrix entries are updated only if file paths or
  service paths changed.

Out of scope:

- Adding new CLI or TUI actions.
- Changing persisted file formats.
- Reworking ticket, plan, dump, quiz, or spec generation semantics.

## Delivery Order

Implement slices sequentially:

1. Slice 1 removes the hidden startup dependency first.
2. Slice 2 tightens operation context passing after dependencies are explicit.
3. Slice 3 moves residual workflow logic and adapter organization after the
   service boundary is clearer.

No slice should start before the previous slice is reviewed, merged to `dev`,
and validated.
