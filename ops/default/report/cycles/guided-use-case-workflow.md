# Guided Use-Case Workflow Cycle Report

Acceptance source commit: `f5c51e8a5ea9754f996bf4967f9581da8ff1fbb7`
Status: implementation integrated; controller full gate pending
Delivery set: guided-use-case-workflow
Plan: [Approved slice map](../../plan/guided-use-case-workflow.md)
Tracker: [Task and criterion evidence](../../tracker/guided-use-case-workflow.md)
Specification: [Approved behavior](../../spec/stable/guided-use-case-workflow.md)

## Delivered Outcome

One retained guided run now spans Web, Telegram and TUI: name, explicit nonempty
capture, sufficient clarification, exact generated-specification approval,
committed work preparation, native admission/execution and verified results.
Adapters call shared commands and queries; original input checkpoints, native
membership, actor, policy and finite attempts survive interface changes and
recovery. Later ideas remain separate pending review.

All six slices are integrated through canonical PRs #430-#435. The final slice
merged at `4b622cc433e828b8451350d439f5b488e26b5911` after all six configured ordered
checks passed. Forgejo, the controller checkpoint and Acta confirm that exact
integration. The tracker and Slice 6 report retain its evidence. The full gate
on the documentation-closure candidate remains pending with the controller.

## Observed Acceptance

| Surface | Observed scope | Boundary |
| --- | --- | --- |
| Web | Complete HTTP journey, real temporary Git checkpoints/publication, lost-response recovery; owned Chromium capture, answers, controls, native approval and verified results at desktop/narrow widths. | Deterministic content generation and local publication provider. See Slice 4 report. |
| Telegram | Real receiver/action/outbox and multipart transport fixtures; duplicate/stale/foreign input, lost acknowledgements and SQLite reopen; shared Web-run continuation. | No user-operated Telegram client observation. See Slice 5 report. |
| TUI | Actual model keys and Linux PTY child running production Bubble Tea and managed-host assembly; capture, clarification, approval, pause, ideas, original reads, graceful shutdown and reopening the same paused run. | Deterministic generation and local Git. See Slice 6 report. |
| Shared parity | One prepared Web run is paused in TUI, resumed in Telegram, and given a separate TUI idea. Native review is preserved; only original receipt reconciliation establishes completion across all three adapters. | Native work/publication fixtures; no external live-provider certification. |

The tracker reconciles GUC-01 through GUC-16 with their owning tasks and concrete
regressions. Manual authoring/approval/plan operations retain their existing
semantics. Stale revision, project, namespace, reply and button bindings cannot
overwrite newer input or change another run.

## Concrete Checkpoint and Native Evidence

The successful verbose T6.3 task fixture used the retained run `guided-quiz` and
produced the following actual temporary-repository inputs. These object IDs
identify fixture observations, not commits in the chr source repository.
Re-running a fixture can generate different IDs without changing its contract.

| Stage | Full input commit | Representative repository-relative input |
| --- | --- | --- |
| Dump | `7c06978f0e0900f01d20bfc07c4558984021888a` | `ops/default/dump/00000-guided-clarification/artifact.md` |
| Quiz | `6375e8b21107ee5ccc37332bd08c74820c712a72` | `ops/default/quiz/00000-guided-clarification-quiz/artifact.md` |
| Approved specification | `33ce9d98e199007f99a8de03d9d559a62b5c21ae` | `ops/default/spec/00000-spec-e3a401f28fb96b55b14fe491e4aba5d2/artifact.md` |
| Preparation | `9b7481c746e7260c339cd4885cf37a0b51f1d4ca` | `ops/default/plan/00000-guided-clarification-plan/artifact.md` and `ops/default/tracker/00000-guided-clarification-tracking/artifact.md` |

The native candidate `00003-workflow-invocation-attempt-2` retained submission
`chr-submission-df10bff6498889b38988ca09cd51d4a96a1b62d04e747fd30c8ce2bb75b996f1`. Its actual fixture Git target was `dev` at
`8fad9414f8c074c01dcb8bf6e33c04e7e2eeb608`. Publication applied before the injected
response loss; recovery observed the original submission rather than issuing
another one. Native review and one consumed worker attempt remained effective.

- Verified delivery receipt hash: `4412fe42399ebbe15631469709d8d8d64014fa702d5e0382c17523c7d6d17f4f`.
- Native closing receipt: `completed`, hash `5a534e959bc9e7a59a987ccc29c04dd28426887e6250ac64debcef42516ce9fc`.
- Host validation receipt: `task_validated`, hash `50f5107c1db1eb8513a47e08950b179d0c67ce1d127de49e6f0a1f9a8d19d03d`.

The parity test requires nonempty delivery, validation and closing receipts before
reporting completion. Original plan/tracker reads remain at their committed
bytes after later working-copy replacement. Missing, partial and invalid result
fixtures remain unconfirmed. Native counters, worker exits and open PRs cannot
substitute for these proofs.

## Validation and Completion Boundary

T6.3 task checks and source candidate are recorded in the Slice 6 report. All six
configured Slice 6 checks passed on `4b622cc433e828b8451350d439f5b488e26b5911`:
development gate, focused terminal/parity normal and race checks, bounded
cross-slice guided normal/race regressions, then whitespace validation. Evidence
belongs to that exact clean HEAD and command list.

The worker verified the local Go, Python, Git and make prerequisites and required
repository scripts for the configured full gate, `make verify`. Short tests own
their temporary fixtures; no retained external service is required. The worker
has not run the full gate during preparation. After the documentation closure is
pushed, the controller owns the unchanged full gate on its exact clean dev
candidate and the final approved documentation closure. Scheduled exhaustive
certification and live external-forge/provider acceptance remain separate.
