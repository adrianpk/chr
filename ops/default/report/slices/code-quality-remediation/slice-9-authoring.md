# Slice 9: Authoring, document generation and guided preparation

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-9-quality-authoring`
PR: `#446`

## Purpose

Give authoring sessions, generated documents and guided preparation one owned
implementation while preserving their distinct approval, provenance and retry
contracts.

## Delivered Behavior

Core/authoring owns the complete original file-backed session aggregate and
manual/guided dump, quiz, spec, plan and case artifact behavior. The mechanical
move relocates 95 source and 53 test files. Shared exclusion, guarded paths,
atomic journal replacement, recovery, source observations, Git checkpoints and
retry receipts remain one implementation. Native consumers and application/TUI
adapters import the real owner directly; root aliases and forwarding APIs are
absent. Acceptance artifacts retain their cohesive child owner.

Authoring uses three read-only native workflow observations for stop quiescence,
original admission lookup and verified results. Expected absence and partial
errors remain explicit. Native admission/preview, materialization projection,
result proofs, original plan resume, credentials and signing remain with their
actual native owners. Native admission/control read a session only under the same
active authoring claim. QuizContextProvider supplies the original workspace/index
observations at the original point before model invocation; request construction
then uses the observed sections without IO. The actual harness adapter stays
outside authoring decisions. Runtime/execution context retains its RuntimePlan,
ticket parser/index and complete execution/result validation at existing owners.

D02 shares bounded concrete catalog traversal; D04 shares payload/context and
model request preparation; D05 shares latest scoped selection; D07 shares
retained-step entry; D09 shares metadata/directory/body writes; D10 shares dump
CLI revision/setup/output. Their governing-contract comparisons retain different
image accounting, schemas, source observations, approval, collision diagnostics,
publication, authority and receipt rules. Spec, acceptance and plan generation
share only their actual common steps.

All 89 assigned signature occurrences have exact current dispositions: 60 tuple
fixes, three parameter fixes (two already delivered by Slice 4) and 26 cohesive
five-parameter contracts. Exact production debt decreases from 301 to 237 with
the same ceiling of seven and one external-interface exception. Required shared
harness/parser counterparts move with their consumers; successor tasks are not
marked delivered. Package-qualified names include SessionRepository, WithMutation
and ReadOwnedSession. The repository remains the original single aggregate.

Shared ops owners hold envelopes/layout, scalar/legacy metadata codecs, bounded
IO/locks and original command buffers. Project owns lexical paths, adjacent
replacement and the existing bounded Git transport; workspace owns default-user
paths. Native process flags, cancellation, output bounds, optional per-write
sync, permissions and cleanup stay unchanged. Workflow/state owns retained
control/evidence/admission and complete native set/plan-binding values; harness
identity has its real shared value owner. StartInput separates durable request
fields from host-only callbacks and commands. Guided/session/request and native
invocation wire bytes, enum spelling, nil/empty values, field order, hashes and
receipt identity remain unchanged. Legacy required headers, optional Markdown
slicing and context/tracker escaping keep their different rules.

## Implementation Notes

[Canonical PR #446](https://forge.adrianpk.com/adrianpk/chr/pulls/446) targets dev.

The move ledger retains 1,243 declaration preimages with current names/files and
stages, including 923 declarations from the complete ownership move. Mechanical
relocation and package naming have separate commits. The plan service separates
API, generation, model preparation, publication, tracking and views. Catalog
observation remains with its artifact owner because its concrete readers,
pending-journal checks and recovery form that responsibility. No separate
coordinator or generic generation framework is introduced.

Native integration tests stay with the native adapters. Authoring journal fault
and case metadata tests move with their aggregate, regardless of their former
Telegram-oriented filenames. Shared test setup is retained on each side of the
package boundary. Existing assertions and frozen compatibility fixtures remain.
A native scope test's case table moved into a helper to meet its existing length
bound; every case and authority assertion remains.

## Contracts Added or Changed

Named results preserve found/changed/cached, partial stage output, provenance and
expected absence separately. All callers retain their original error and replay
branches. Contextual request construction remains a cohesive five-input contract
with already observed context; the explicit context provider owns cancellation
and native observation. API naming changes follow actual ownership.

The original session/reference and guided validators move with their values.
Native admission keeps its narrower intent predicate. Scope, actor/progression,
finite retained-state bounds, guard order and diagnostics stay exact. Catalog
traversal retains namespace/numeric-prefix selection, 4096 entries, aggregate
8 MiB, reference order and final pending-journal inspection; listing never adopts
records. State-to-harness and actual parent dependencies have enforced import
bounds and reverse-cycle fixtures.

## Files of Interest

- `internal/core/authoring/`: session repository, owner/journals, artifact and
  guided operations, context and narrow workflow observation contracts.
- `internal/core/authoring/acceptance/`, `authoring/generation/` and
  `authoring/state/`: distinct artifact checks, payloads and private retained images.
- `internal/core/guided_native_*.go`, `guided_workflow_observer.go` and
  `authoring_quiz_{context,harness}.go`: actual native adapters and proofs.
- `internal/core/{ops,project,workspace}/` and `workflow/state/`: actual shared
  codecs, paths, transport, buffers and retained values.
- `internal/app/service_document_model.go`, `service_document_view.go`, plan
  responsibility files, `service_guided_work_step.go` and `dump_cli_update.go`.
- [Move evidence](../../../plan/code-quality-remediation-slice-9-moves.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-9-signatures.csv),
  [current dispositions](../../../plan/code-quality-remediation-slice-9-dispositions.csv),
  [duplicate comparisons](../../../plan/code-quality-remediation-slice-9-duplicates.csv),
  [current owners](../../../plan/code-quality-remediation-core-owners.csv) and
  [dependency ledger](../../../plan/code-quality-remediation-boundaries.md).

## Validation

Focused task checks passed for real authoring publication/source revision,
guided admission/control/recovery/results, session/metadata/context, native
credential/process behavior and frozen wire records. Pinned strict lint passed
across authoring/core/app/TUI and the checker. Exact dispositions, retained
preimages, import-cycle fixtures, style and package README checks passed.

The controller validation ledger and PR Validation bind final successful outcomes
to the exact clean PR head and this ordered command list:

- `make verify`
- `go test ./internal/core/authoring ./internal/core ./internal/app -run '^Test(QualityRecordWire|AuthoringSession|RenderAndParseAuthoringSession|Assisted|SynthesizeAssisted|LoadAssisted|PlanPublication|PlanOrder|SpecMutation|ProjectDump|GuidedUseCase(Preparation|Admission|Control|Recovery|Result)|GuidedAdmission|TelegramAuthoring|TelegramOperationCase|PlanContextualQuiz|HarnessAuthoringQuiz|QuizToDraft|WriteQuizToDraft)' -count=1 -timeout=3m`
- `git diff --check`

## Risks and Follow-ups

Integration into dev is pending.
