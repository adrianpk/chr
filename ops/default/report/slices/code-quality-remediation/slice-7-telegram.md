# Slice 7: Telegram transport, receiver and retained delivery

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-7-quality-telegram`
PR: pending

## Purpose

Give Telegram transport and retained delivery one bounded package owner while
preserving native authority, source acknowledgement and durable replay behavior.

## Delivered Behavior

The receiver, HTTP transport, conversations, operations, confirmations and outbox
live in core/telegram. App/domain consumers import the owner directly. The store
keeps each original transaction across source, request, action, notice and frozen
document admission. A nested outbox would split those transactions without an
independent consumer, so the same owner retains them.

Named send/completion/projection/intake/reply records replace long inputs. Named
parser, routing, lookup, provider reservation and host results retain recognition,
start, target need, pending, missing and found as distinct facts. All 156 assigned
occurrences conform: 37 A inputs, 52 legal B cohesion watch items and 67 C results.
The disposition ledger records 105 fixed and 51 retained-contract occurrences.
Temporary debt decreases from 410 to 307 with the same ceiling/external exception.

Typed SQL rows retain exact peer/owner fields, defaults, NULLs, hashes, bytes and
finite quotas. Real SQLite column-name assertions distinguish values after insert
and update; a positive topic/user swap fails. Existing multipart, caption/entity
fallback, lost-ack, receipt identity and restart tests remain. D06 shares only
intake mechanics; explanation and work keep their separate execution handlers.
The two test-only forwarding symbols are removed with live coverage retained.

## Implementation Notes

A separate move commit preserves 453 complete production declaration texts across
71 files before name/contract changes. Fifty-one private owner tests move with
them; mixed domain/migration/native integration tests stay at root. Integration
fixtures own SQL inspection connections and feed the real public polling codec.
Frozen wire fixtures, legacy SQLite bytes and schema assets remain unchanged.
The SQL ledger records 174 unchanged expressions at the row/input checkpoint.

Shared authoring identity/intent, structured-provider error identities, namespace,
project directory and workspace spelling have pure owners with bounded imports.
Complete guided validation stays in app composition before a narrow association;
its invalid-domain regression proves rejected attachment cannot change peer focus.
Legacy update decoding stays with transport; generic command records stay in app.
The finite ledger accounts for 714 production core files.

Three stale non-short assertions reproduced on the unchanged predecessor. Their
current capacity/readiness wording and stronger original-request/receipt/reply
identity checks replace stale presentation expectations without runtime changes.

## Contracts Added or Changed

Direct ReceiverStore/ReceiverAPI contracts use named operation inputs and results.
Context and transaction handles remain explicit. Six transport state types keep
exact stored strings. Error types keep their identity and text with consistent
names. SQL cursor cleanup joins errors; one documented HTTP body-close suppression
preserves an observed acknowledgement rather than granting another send attempt.
The unused serve inputs were already resolved by Slice 4's typed command factory.

Production dependency checks reject root/coordinator/adapter imports, including
descendant bypasses; message keeps its stricter pure-value bounds. Native execution,
approval, receipt authority, provider attempts and model/OS exclusion stay at their
existing owners. Boolean outcomes with different contracts remain distinct.

## Files of Interest

- `internal/core/telegram/`: transport, store, named requests/results/states and rows.
- `internal/app/telegram_command_route.go`, guarded Telegram adapters: shared intake
  and direct operation composition.
- `internal/core/authoring/state/`, `model/`, `workspace/`, `ops/namespace/`:
  shared immutable contracts without coordinator imports.
- `internal/core/telegram/storage_row_test.go`, root frozen-record fixtures:
  distinguishable SQL identities and original representations.
- [All 156 dispositions](../../../plan/code-quality-remediation-slice-7-dispositions.csv),
  [mechanical moves](../../../plan/code-quality-remediation-slice-7-moves.csv),
  [SQL expressions](../../../plan/code-quality-remediation-slice-7-storage.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-7-signatures.csv),
  [cleanup contracts](../../../plan/code-quality-remediation-slice-7-cleanup.csv).
- [Current owners](../../../plan/code-quality-remediation-core-owners.csv),
  [dependency/glossary decisions](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./internal/core/telegram ./internal/core ./internal/app -run '^Test(Telegram|RenderTelegram|GuidedUseCaseTelegram|QualityRecordWire|QualityStorage)' -count=1 -timeout=3m`
  — passed for the named operation inputs and original non-short integrations.
- `go test ./internal/core/telegram ./internal/core ./internal/app -run '^Test(TelegramStorageRows|TelegramReceiver(PollBudget|Outbox|HTTP.*)|TelegramOperationRequestRecovery|TelegramOperationActionLostAcknowledgement|TelegramRecovery.*|QualityStorage.*|QualityRecordWire|GuidedUseCaseTelegramScope)$' -count=1 -timeout=3m`
  — passed after typed states, error names and cleanup changes.
- `go test ./internal/core/telegram -run '^TestTelegramStorageRows$' -count=1 -timeout=3m`
  — passed; a temporary positive topic/user column swap failed its named-column assertion.
- `make check-go-style` — passed with 307 exact remaining debts and one exception.
- Pinned strict core/app/owner/tooling lint passed at task checkpoints.

Final PR-head validation uses make verify, the non-short Telegram/core/app batch,
focused retained-state concurrency/recovery race tests, then git diff --check.
The controller validation ledger and PR Validation own the exact clean HEAD,
ordered commands, outcomes and logs. Integration remains controller-owned.

## Risks and Follow-ups

None recorded.
