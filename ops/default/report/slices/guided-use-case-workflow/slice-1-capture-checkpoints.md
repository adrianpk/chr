# Slice 1: Guided capture and Git checkpoints

Status: reviewing
Delivery set: guided-use-case-workflow
Plan: [Approved delivery plan](../../../plan/guided-use-case-workflow.md)
Tracker: [Tasks and acceptance](../../../tracker/guided-use-case-workflow.md)
Branch: `feat/slice-1-guided-capture-checkpoints`
PR: `#430`

## Purpose

Let a user start a new use case through Web, retain its name and multiple capture inputs, explicitly finish a meaningful dump, and inspect the exact committed version after reload.

## Delivered Behavior

1. Starting retains the selected project/namespace, initiating actor and automatic progression intent in a canonical authoring session. Duplicate starts replay the same run; conflicting scopes and implicit manual-session adoption are rejected.
2. Naming uses the canonical case mutation owner and retains its original result. Revision-bound capture commands reject stale writes, serialize concurrent input and reconcile interrupted journal replacements.
3. Finishing saves canonical dump/raw components and seals their content, metadata and case name in a Git checkpoint. A private index and a dedicated retained ref preserve the user's branch, staged changes, unstaged edits and untracked files.
4. Closed input reads use the recorded commit/path/hash bindings. Later working-copy edits cannot replace those bytes; missing or corrupt objects fail instead of falling back to current files.
5. Web exposes start/name/capture, explicit completion, readiness explanations, resumable navigation and complete committed dump inspection. After sealing, the run waits at the unavailable quiz stage and does not claim execution.

## Implementation Notes

The optional `guided_state` JSON field extends the existing session frontmatter with schema version 1. Existing sessions keep their encoding and manual semantics. The shared authoring exclusion, revision guards and mutation journal own the state and transcript; generic session/transcript writes cannot replace guided state.

The case and dump owners retain stable operation bindings before effects. Stage checkpoint intent records artifact hashes and the original parent commit before Git writes. Commits use the existing managed Git process boundary, an isolated temporary index, and `refs/chr/guided/<run-hash>/<operation-digest>`. Updating that ref retains the commit independently of the caller's branch. Recovery checks the original ref and committed hashes before sealing the stage, including when Git succeeded but its response was lost.

The repository identity is derived from the resolved Git common directory, so worktrees sharing the repository share the binding. Artifact checkpoints are input history; they do not select a native code target or grant admission/publication authority.

Adapters call the central application service. Inspection reads retained state, local prerequisites and committed artifacts without recovering a journal, starting a host or advancing a stage. No provider or external forge is called to finish a raw capture.

## Contracts Added or Changed

1. `AuthoringSession.Guided` carries bounded scope/intent, progression, capture, operation and checkpoint state. A run uses the session identity and positive session revision.
2. Guided commands bind the operation ID, exact arguments and observed revision. Retried name/finish effects retain their original request identities and expected revisions.
3. Stage checkpoints retain repository identity, full commit, dedicated ref, original parent and artifact path/content hashes, with source-reference fields for later stages.
4. `/guided` is read-only; `/guided/actions` accepts bounded, single-valued forms. Existing Web access and confirmation checks still apply.

## Files of Interest

1. `internal/core/guided_use_case.go` and `guided_use_case_owner.go`: retained schema, commands, inspection and canonical owner integration.
2. `internal/core/guided_use_case_checkpoint.go` and `guided_use_case_git.go`: dump sealing, private-index Git writes, retention, recovery and committed input reads.
3. `internal/app/service_guided_use_case.go` and `web_guided_use_case.go`: shared operations, readiness and Web projection.
4. `internal/webui/guided_use_case.go` and `assets/templates/guided_use_case.html`: HTTP bounds and current-stage forms.
5. Adjacent `guided_use_case*_test.go` files: concurrency, journal/Git faults, HTTP journey, immutable reload and rendering regressions.

## Validation

Task regressions passed:

1. `go test ./internal/core -run '^Test(GuidedUseCaseCapture|AuthoringSession)' -count=1 -timeout=3m`
2. `go test ./internal/core -run '^TestGuidedUseCaseCheckpoint' -count=1 -timeout=3m`
3. `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Capture|Checkpoint)' -count=1 -timeout=3m`
4. `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` completed with zero remaining issues.

Final PR-head acceptance uses the controller's retained exact-clean-HEAD evidence for this ordered command list:

1. `make verify`
2. `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Capture|Checkpoint)' -count=1 -timeout=3m`
3. `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Capture|Checkpoint)' -count=1 -timeout=3m`
4. `git diff --check`

The Git and HTTP tests use real temporary repositories and filesystem owners. Real Git fixtures are selected without short mode; rendering/HTTP evidence does not establish browser or live-client acceptance.

## Risks and Follow-ups

1. Quiz, default specification approval, work preparation/execution, full recovery/controls and other adapters remain assigned to the later approved slices. The current Web journey exposes that stop explicitly.
2. Checkpoint commits require an initialized repository and configured Git author identity. Missing local setup is visible; capture can be retained before setup is repaired.
3. Input retention refs remain reachable for the run's evidence lifetime. Deleting those refs or Git objects outside the owner requires recovery; working copies are not substitute evidence.
