# Slice 4: Typed commands and central application operations

Status: delivered
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-4-quality-commands`
PR: `#441`

## Purpose

Bind command dependencies at assembly time and share ticket operations through
the existing application service while preserving command and execution contracts.

## Delivered Behavior

All 134 command constructors are typed commandFactory methods. Process IO and
embedded assets are captured once; actions use those typed dependencies without
reading root Metadata. The application entrypoint groups its process streams.
Seventy-four forwarding action helpers and ignored dependency/output arguments
are removed. The command factory regression uses conflicting legacy metadata
and distinct embedded assets to prove the constructor binding controls execution.

Ticket selection, start preflight, workspace preparation, execution, result
application, legacy loop, review and merge stages now share the existing central
service. CLI handlers parse input and render operation results. One ticketLoop
value owns retained run state, repository, events and partial results; contexts
remain explicit. Manual and automatic review/merge retain their distinct policy,
approval, lock and refresh contracts. TUI, Web and Telegram work adapters use the
same service stages through their existing operations.

The assigned signature inventory has 69 row dispositions: 47 fixed occurrences
and 22 retained cohesive five-input contracts. Named results preserve partial
execution and integration state, initial eligibility, selected queues, release
callbacks, stop decisions, evidence diagnostics, provider response IDs and Web
handled decisions. Typed daemon ownership reuses daemonWorkflowRunner, gate
report sources group creation time/artifact paths, and intent-generation binding
keeps its context local to the operation. Global signature debt falls from 517
to 472, with the ceiling still seven and one external-interface exception.

## Implementation Notes

A separate mechanical task split ticket.go, author.go, maintenance.go and
service_work.go into 29 responsibility-owned files. All 292 complete declaration
texts were identical across that move; the largest resulting file was 617 lines.
Subsequent semantic ticket extractions have their own commits and identity ledger.
No new package, service hierarchy or migration facade is introduced.

The two identical command service/project resolvers become projectService; the
service already owns the resolved project. The unreachable dumpDirectory helper
is removed while live dump listing/content checks remain. Other named dead
symbols and Telegram-private error names retain their explicit later slice owners.
Native admission, legacy guards, CLI flags/output, finite iteration limits,
worktree metadata and configured merge strategies retain their existing meaning.
The original source review and detection files retain their SHA-256 digests.

## Contracts Added or Changed

Internal command construction, operation inputs and result signatures are typed.
No Metadata locator or tuple compatibility forwarder remains for the converted
operations. New reusable result values contain no context.Context. Cleanup and
partial failures retain their existing caller branches. Wire/durable formats,
publication authority and provider contracts keep their original owners.

## Files of Interest

- `internal/app/command_factory.go`, `app.go` and `command_factory_test.go`:
  typed assembly, process IO and constructor-binding regression.
- `internal/app/service_ticket*.go`: central preflight, execution, loop,
  application, review and merge operations.
- `internal/app/service_work*.go`, `daemon.go`, `gate.go`, `web_actions.go`:
  named execution outcomes and adapter-owned inputs/results.
- [All constructors](../../../plan/code-quality-remediation-command-constructors.csv).
- [Mechanical moves](../../../plan/code-quality-remediation-slice-4-moves.csv).
- [Operation ownership](../../../plan/code-quality-remediation-slice-4-operations.csv).
- [All 69 dispositions](../../../plan/code-quality-remediation-slice-4-dispositions.csv).
- [Final signature transitions](../../../plan/code-quality-remediation-slice-4-signatures.csv).
- [Compatibility model](../../../spec/code-quality-remediation-model.md).

## Validation

- `go test ./internal/app -run '^Test(CommandFactoryBinding|RootCommand|Ticket|AgentWork|Work|Gate|Daemon|Notify|PiGenerationConversation|PiGenerationDump|WebServeWriteAction|ServiceDump)' -count=1 -timeout=3m`
  — passed in 92.5 seconds; real local Git, SQLite, HTTP/provider and command
  boundaries, including partial failures, review/merge and foreground daemon runs.
- `go test ./internal/app -run '^Test(PiGenerationConversation|ConversationIntent)' -count=1 -timeout=3m`
  — passed after separating bound-project validation from local context derivation.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app/...`
  — passed with zero issues after formatting and signature changes.
- `go run ./scripts/goquality` — passed; 472 exact debt entries, one exception.
- Constructor identity reconciliation — all 134 current declarations/signatures
  match the committed constructor ledger. Every assigned current signature is
  within five parameters and two results with error/bool second when applicable.

The first final-candidate `make verify` attempt stopped at the signature-format
check on moved multiline headers. All 28 added headers were normalized without
changing parameter/result types or bodies; the controller retains that failure
log and regenerates evidence for the corrected commit.

Final clean-head validation runs through the controller after the PR-reference
commit: make verify, the explicit non-short focused command above, then git diff
--check. The controller retains exact HEAD, ordered commands, outcomes and timings;
that record and the PR Validation section own final candidate results.

## Risks and Follow-ups

None recorded.

Integration: PR #441 was verified merged into dev at `6b287a79bbe8b545a4d36633d24ceccea01469d0`. The controller retained successful exact-candidate checks in this order:

- `make verify`
- `go test ./internal/app -run '^Test(CommandFactoryBinding|RootCommand|Ticket|AgentWork|Work|Gate|Daemon|Notify|PiGenerationConversation|PiGenerationDump|WebServeWriteAction|ServiceDump)' -count=1 -timeout=3m`
- `git diff --check`
