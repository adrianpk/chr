# Slice 2: Guided Quiz and Approved Specification

Status: delivered
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-2-guided-quiz-spec`
PR: `#431`
Integrated commit: `2246651001dd6002f783fd86d0dcc13f5bef02e6`

## Purpose

Continue a newly captured guided use case through clarification and a committed,
automatically approved specification using the shared application service and
canonical authoring owners.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Task and acceptance evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Finishing a Web dump automatically asks the current provider-generated
  clarification question. Answers bind run, project, namespace, revision,
  question and committed dump digest. Duplicate replies reuse their receipt;
  foreign, empty and stale replies cannot reserve effects.
- Sufficient answered clarification seals the quiz, raw answers, metadata and
  runtime evidence. Empty generation, skipped/incomplete answers and budget
  exhaustion cannot stand in for sufficient clarification.
- The service generates a specification from committed dump and quiz inputs,
  saves its draft and approves that exact revision under retained development
  intent. The approved specification and metadata receive a Git checkpoint
  with the original dump and quiz dependency bindings.
- Web inspection exposes complete original documents after working copies
  change. Retained-operation retry preserves the original request and revision.
  The current flow reports unavailable work preparation after approval.

## Implementation Notes

`ApplyGuidedUseCaseStep` composes the existing quiz and specification owners
under the shared authoring exclusion and journal. Pending quiz requests and
validated specification generation are retained before downstream artifact
writes. Existing owner receipts reconcile lost acknowledgements without
allocating replacement identities. Stage commits reconcile their retained Git
refs before consulting later working copies.

Quiz storage supports an explicit namespace while absent optional fields retain
legacy JSON receipt inputs. The provider receives committed document sections
and answered history. Automatic Web progression lives in the shared service;
rendering, reconnecting and artifact reads do not dispatch generation.

## Contracts Added or Changed

- Guided quiz state retains its canonical session/revision, current question and
  input digest; provider-confirmed clarification is required for sealing.
- Guided specification approval binds actor, run, `guided-default/v1`, generated
  owner revision, body hash and committed input digest. The canonical mutation
  owner rejects different content, scope, actor, policy or generated revision.
- Optional guided fields preserve old manual session and mutation serialization.
  Standalone specification approval remains an explicit independent operation.
- The specification checkpoint carries the dump and quiz source repository,
  full commit, path and content-hash chain. These artifact commits do not select
  a native code execution base.

## Files of Interest

- `internal/core/guided_use_case_progress.go` and `guided_spec.go`: retained
  progression, sufficient quiz sealing, exact approval guards and stage seals.
- `internal/app/service_guided_quiz.go`, `service_guided_spec.go` and
  `service_guided_progress.go`: committed-input generation and shared progression.
- `internal/app/web_guided_use_case.go` and
  `internal/webui/assets/templates/guided_use_case.html`: current input, retries
  and complete original inspection.
- `docs/reference/authoring-sessions.md`: clarification, approval and recovery
  contract.

## Validation

Successful task checks:

- `go test ./internal/core ./internal/app -run '^TestGuidedUseCaseQuiz' -count=1 -timeout=3m` — passed.
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCaseQuiz' -count=1 -timeout=3m` — passed.
- `go test ./internal/core ./internal/app -run '^TestGuidedUseCaseSpec' -count=1 -timeout=3m` — passed.
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCase(Quiz|Spec)' -count=1 -timeout=3m` — passed.
- `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Quiz|Spec|Capture)' -count=1 -timeout=3m` — passed.
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Quiz|Spec|Capture)' -count=1 -timeout=3m` — passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — passed, zero issues.

Focused fixtures use deterministic local generation and real temporary Git and
filesystem stores. The older capture fixture initially failed its obsolete
unavailable-clarification assertion after reaching the configured generation
adapter; it now uses the same deterministic provider boundary and passes.

The [canonical PR](https://forge.adrianpk.com/adrianpk/chr/pulls/431) lists the successful required commands for its final clean
HEAD; Looper retains the exact candidate and ordered validation evidence.
The controller verified all required checks on the exact integrated commit
`2246651001dd6002f783fd86d0dcc13f5bef02e6`:

- `make verify` — passed.
- `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Quiz|Spec)' -count=1 -timeout=3m` — passed.
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Quiz|Spec)' -count=1 -timeout=3m` — passed.
- `git diff --check` — passed. Browser interaction belongs to
the later Web acceptance checkpoint and is not established by these HTTP tests.

## Risks and Follow-ups

Work preparation and execution remain the next recorded slice. Complete Web
host recovery and browser acceptance remain assigned to Slice 4. No additional
follow-up work is introduced by this report.
