# Slice 2: Live Role Selection and Retained Bindings

Status: delivered
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-2-harness-selection`
PR: [#413](https://forge.adrianpk.com/adrianpk/chr/pulls/413)

## Purpose

Apply saved role selections to new worker and generation operations while retaining
the identity of admitted work through live changes, retries and host recovery.

## Delivered Behavior

- Private named profiles expose local readiness and effective worker/generation
  selections. Unsupported explicit preparation fails before dispatch. Clearing a
  role restores its established action defaults and direct API behavior.
- Shared inspect/change operations and CLI status/select/clear commands use an
  authorized actor, complete configuration revision, cross-process exclusion,
  atomic two-role publication and bounded request replay. Unrelated configuration
  and unknown fields survive changes and interrupted publication.
- Selected workers use sealed `chr-invocation/v2` and `chr-task-input/v2` identities.
  Pending tickets, subsequent attempts, manual handoffs and independent-process
  recovery retain model, profile, executable, protocol and capability evidence.
  Existing v1 checkpoints keep their bytes, request hashes and command semantics.
- Generation operations retain one selection across nested calls and recovery.
  This covers quiz questions and contextual quizzes, Dump curation, Spec and
  acceptance generation, plan stages/tracking, conversation intent, explanation
  and agentic review. Independent requests read current settings.
- Host run metadata separates process/network outcome, parser validity, requested
  model identity and observed usage. Structured-model selections share the existing
  capacity channel; original owners retain attempt, publication and budget guards.

## Implementation Notes

The configuration owner serializes selection changes with first admissions. Worker
checkpoints privately retain their command; public bindings expose hashes. Claims
and dispatch reject command overrides and changed executable bytes before launch.

Generation bindings live within existing private run artifacts. Host conversations
retain their artifact owner across selected-project changes and recheck current
project access before retries. Review uses the retained project's configuration and
its host-owned execution worktree. Quiz state retains its pending generation ID so
a failed continuation can resume without discarding the confirmed answer.

Run bundles include bounded `chr-harness-run/v1` metadata and complete normalized
output. Schema/parser failure cannot imply artifact approval. Unknown counters and
observed models remain absent. Requested CLI-default provenance does not claim an
exact underlying model. Private profiles and credential values are excluded.

## Contracts Added or Changed

- Workspace role selections, complete-image revision/audit and bounded replay
  receipts; strict private profile parsing and local readiness.
- Versioned worker invocation/task bindings and explicit v1 compatibility.
- Sealed generation operations, private retained host routing, namespaced call IDs
  and host metadata; 16 KiB records and 4,096 retained operation identities.
- Neutral frozen-selection transport input and selected Codex executable/model
  assembly. No second task store, retry authority or capacity pool was added.

## Files of Interest

- `internal/core/harness_selection_change.go` and `harness_selection_journal.go`
- `internal/core/workflow_harness_admission.go` and `harness_binding.go`
- `internal/core/harness_generation_operation.go` and `harness_generation_run.go`
- `internal/app/service_harness_generation.go` and `service_conversation_generation.go`
- `internal/app/service_quiz_generation.go` and `harness_config.go`

## Validation

Passed local focused implementation checks:

- `go test ./internal/core ./internal/app ./internal/harness/codex -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding|TestHarnessQuizModel|TestAgenticReview|TestAuthoringQuiz|TestStructured|TestHarnessCodex|TestService.*(Quiz|Spec|Plan|Dump|Acceptance)|TestTelegram.*(Intent|Authoring))' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/harness/codex`
- `git diff --check`

Coverage includes independent processes, concurrent admissions/switches, stale and
unauthorized updates, crash/replay, unchanged v1 records, executable/schema fences,
staged generation, authoring/quiz/intent recovery and shared model capacity. The
v1-only compatibility consumer is a fixed reader fixture. All owned fixtures ended.

The initial final-head `make verify` at `e24cdc6a47e8e1a6ca7b5d8424b6e8b7df7b384a`
failed strict lint because the shared request validator exceeded the complexity
limit. Extracting its output-name check preserves validation and error order.
The focused `go test ./internal/harness/... -count=1 -timeout=3m` passed afterward.

The next final-head attempt at `e86c9e97903a8ef270ba156e34ff8da8e03a3195`
exposed legacy Web/review fake-provider fixtures that did not inject the generation
binder and a 256 MiB ceiling below supported CLI footprints. Domain fixtures now
inject that boundary explicitly. Production binding fixtures still use prepared
profiles and owned drivers. The streamed ceiling is 512 MiB, with a positive
300 MiB fixture and the existing oversized/non-regular/hash rejection checks.
The targeted caller regression selection passed after these corrections.

The controller recorded all five checks passing on the clean PR head
`698cda11aab88b7b5f0994fa5d7faa8c9aa361fe` (evidence
`XRCUOWBCLESSXUH6I7GROBHLCY`):

- `make verify` (pass)
- `go test ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m` (pass)
- `go test -race ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m` (pass)
- `go test ./internal/harness/codex ./internal/core ./internal/app -run '^(TestHarnessCodex|TestHarnessQuizModel|TestAgenticReview|TestAuthoringQuiz|TestStructured|TestService.*(Quiz|Spec|Plan|Dump|Acceptance)|TestTelegram.*(Intent|Authoring))' -count=1 -timeout=3m` (pass)
- `git diff --check` (pass)

PR #413 was verified merged into dev at that exact commit on 2026-10-19.
Slice 2 of 8 completed; Slice 3 is the next recorded unit.

## Risks and Follow-ups

Pi dispatch and interface controls follow their recorded successor slices.
CLI-default leaves the exact provider-selected model unknown; explicit model
selection is required when that identity must be reproducible.
