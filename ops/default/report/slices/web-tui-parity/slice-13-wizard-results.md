# Slice 13: Wizard execution, results and recovery

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-13-tui-wizard-results`
PR: [#467](https://forge.adrianpk.com/adrianpk/chr/pulls/467)

## Purpose

Resolve G-011 through G-014 with shared Board execution, evidence-backed results,
confirmed cancellation and read-only observation that preserves terminal input.

## Delivered Behavior

- Execution and results use the existing case-filtered Board, including wide
  columns, narrow layout, ticket/detail focus and filters. Progress separates
  validated tasks, confirmed delivery, review waits and remaining work. Actual
  branch/PR links and eligible approval, reconciliation, allowance and decision
  controls use the existing guided/native owners and exact confirmations.
- Results show the verified summary, repository, branch, commit, PR, delivery
  receipt/reference/hash, task and closing validation commands/exit codes and
  closing evidence. Worker success and an open PR remain unconfirmed. `v`
  toggles the existing JSON export and Board; committed originals and later
  ideas remain separate secondary actions.
- Cancellation confirms with “Saved documents and completed changes will be
  kept.” and explains that unsent text will be discarded. Esc preserves both
  run and draft. Successful cancellation retains saved artifacts and completed
  changes. Pause is offered only during eligible active execution; the existing
  resume, cancellation, scope and consumed-attempt guards remain authoritative.
- Passive reads preserve complete unsent input, selected case, tab and focus.
  Typing and explicit guarded submission remain available while a read is in
  flight. Older guided/native replies cannot replace newer observations.
  Refresh never submits an answer, advances work or redirects an earlier
  document; a visible notice explains that unsent input remains available.

## Implementation Notes

T13.1: `ed20cd540f319ecc74809656233f01b53fe771e3`.
T13.2: `8972f55df1ff14316f8c2282953e3abbebfe772a`.
T13.3: `66d738e145954fec713ace69f28beb17907b3b74`.
T13.4: `24112fcb7cc1f4688bc1846ce2a60925214e6730`.
T13.5 report introduction: `034ffe4db3a18d410b16337d15457e77616784e8`.

Transport projection reads canonical tickets and the neutral verified native
result inspection. It does not derive delivery from counters or import Web DTOs.
Board and confirmation handlers remain shared with manual/native operation.
The TUI projects the existing Web pause policy without changing domain commands.

At most one passive read overlaps an explicit request. The explicit request
takes presentation ownership; sequence and revision checks ignore older replies.
Confirmation contexts suppress new passive reads, and drafts retain their
captured question, revision, digest and operation. No queued progression engine
or new retry budget is introduced.

The production cancellation PTY pauses actual waiting-review execution, declines
one cancellation without writes, then confirms cancellation. It verifies exact
committed document bytes/reference/commit, four checkpoints, the completed
worker commit and unchanged native policy/attempt count. The earlier journey's
paused authoring restart precondition is now seeded through the existing owner,
because authoring no longer offers Pause. Its original exact capture, approval,
checkpoint, later-idea, original-image and restart assertions remain intact.

The production polling PTY types a draft while an explicit peer records an idea.
Seeing that idea proves a real periodic read occurred; typing and cancelled
discard retain the complete input. A second peer idea appears while the earlier
Dumps tab remains selected. Canonical images/revisions show zero passive writes
or submitted answers. These fixtures keep their 45-second fixture and 40-second
child limits and stop only owned children. UI regressions also cover newer stages,
case selection, native revision ordering and submission during a passive read.

## Contracts Added or Changed

Guided observations can carry canonical tickets and structured verified result
data. A transient passive-read flag distinguishes inspection from explicit
requests. No domain schema, command authority, integration semantics, persistence,
finite resource limits or dependency changes.

## Files of Interest

- `internal/tui/wizard_board.go`, `ticket_render.go`: shared Board and native controls.
- `internal/tui/guided_results.go`, `internal/app/tui_guided_result.go`:
  structured verified delivery and validation presentation.
- `internal/tui/wizard_cancel.go`, `internal/app/tui_guided_controls.go`:
  retained cancellation confirmation and visible Pause policy.
- `internal/tui/wizard_polling.go`, `wizard_recovery.go`, `guided.go`:
  input/navigation preservation and ordered read presentation.
- `internal/app/tui_wizard_*test.go`, `internal/tui/wizard_*test.go`:
  real canonical IO, production PTY and shared component regressions.
- User Guide, TUI reference and changelog: delivered controls and results.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including Board, result toggle,
  captured input, passive overlap, newer stages and native revision ordering.
- `go test ./internal/app -run '^Test(TUIWizard.*|GuidedUseCaseTUI.*|GuidedUseCaseParityNative|GuidedUseCaseSharedQuizWeb)$' -count=1 -timeout=3m` — passed for canonical persistence, delivery receipts, production cancellation/polling PTYs and existing guided/native/Web Quiz regressions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed with zero issues.
- `git diff --check` — passed before task commits.

Final controller validation `BV6G4MNWIBE7TQXVHIDFMORIVQ` passed on clean
HEAD `91df4d01d2838180430f9259eeab1569ed6618e5` in this order:

- `make verify`
- `go test -race ./internal/tui -count=1`
- `go test -race ./internal/app -run '^Test(TUIWizard.*|GuidedUseCaseTUI.*|GuidedUseCaseParityNative|GuidedUseCaseSharedQuizWeb)$' -count=1 -timeout=3m`
- `git diff --check`

PR #467 merged into dev at that same commit. Acta merge event:
`01M64DPQMB4K5FCCCT1BTDBER6`. No checks were repeated for documentation closure.

## Risks and Follow-ups

None recorded.
