# Slice 16: Final enforcement, documentation and integrated closure

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-16-quality-closure`
PR: pending

## Purpose

Enforce the final production signature and duplication contracts, reconcile every
approved baseline finding and publish the final capability map and glossary.
Prepare exact-candidate evidence without preempting controller-owned integration
or the unchanged final delivery-set gate.

## Delivered Behavior

All 771 baseline signature rows now resolve to actual current declarations and
owning contracts: 121 A, 230 B and 420 C. There are 570 conforming named-contract
dispositions, 200 retained cohesive five-input cases and one exact external
Bubble Tea Model.Update exception. The two-declaration capability recheck maps
both its real adapter and stage policy owner. Current production has 9,411
signature declarations, no function above five inputs and no other invalid
result shape. This is source reconciliation, not runtime or integrated acceptance.

## Implementation Notes

Predecessor PR #452 is verified merged into dev at
`a41e62be3b2a14daf04b470e8d9344d350480f34`. Its report, tracker and ticket are
closed as delivered. Slice 16 uses its approved branch and dedicated worktree.
The configured report path had no existing introduction in Git history; this
commit introduces the canonical report at that exact path.

## Contracts Added or Changed

The production checker now has a permanent five-parameter maximum and the
existing semantic result policy. The phase configuration, debt JSON file,
allowance branch and ratchet tests are removed. Exact external-interface
exceptions remain mandatory, signature-bound and stale/duplicate/unnecessary
entries fail. Source/build-variant/alias/dependency coverage remains unchanged.

Pinned revive argument-limit and function-result-limit enforce five inputs and
two results; the AST check retains interface/callback/literal and second-result
coverage. dupl is enabled at threshold 150. The new production rules exclude
only test fixtures; existing defect-oriented lint scope remains unchanged.
Supported wire, persistence, authority, hashing and process contracts are unchanged.

## Files of Interest

- [Final baseline signatures](../../../plan/code-quality-remediation-slice-16-signatures.csv): all 771 original identities, current signatures and code-backed contract provenance.

- `scripts/goquality/`: complete production source and dependency policy.
- `.golangci.yml`: pinned repository lint settings.
- `ops/default/plan/code-quality-remediation-*.csv`: baseline and source evidence.

## Validation

Activation is metadata only. The predecessor's exact-head evidence is retained
in its report and controller receipt. The final 771-row ledger verifies current
identities/signatures against the production inventory and the owning source-move
and semantic disposition ledgers. All 230 B rows have individual existing
cohesion decisions. Ten remaining qualified-method/relocation gaps are resolved
from the actual Slice 9/12 move entries and the shared database extraction commit.
The secondary-file/type/duplicate/dead-code/minor reconciliation remains active.
No complete Slice 16 runtime or integrated acceptance is claimed.


Permanent-policy task verification:

- `go test ./scripts/goquality -run '^TestBranchIndependent$' -count=1` initially failed as expected: the phase-six policy accepted a new six-input function on a dev self-comparison. The existing fixture now tests this exact final boundary.
- `go test ./scripts/goquality -count=1` passed after permanent enforcement, including exact exception, alias/build-variant, nested contract and branch-independent coverage.
- `go run ./scripts/goquality` passed: 9,411 declarations, maximum five, one external exception and no migration manifest.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...` passed with zero issues under the final new rules.
- `git diff --check` passed.

These are task checks. Required clean-head controller validation and the later
integrated delivery-set gate remain distinct.


Minor-item source correction completes the reported eight error-type names:
AssistedModelError, execution.Error and providerError replace the three remaining
names, with constructor/caller updates and no compatibility aliases. Error text,
cause unwrapping, failure-kind strings and transport sentinels are unchanged.
The other five names were corrected by the existing Telegram owner.

The reported 23 bare errcheck directives comprise 11 production sites and 12
test-fixture sites at the approved baseline. Eight production sites already have
resource-specific handling/reasons. The remaining three production fallback
pipe/read-handle closes and all 12 fixture sites now state their exact cleanup
purpose. Checked permission, process, output and durable-write errors remain
observable. No generic close helper replaces their distinct failure contracts.

- Identical non-comment token streams were verified for all 11 cleanup-comment files.
- Exact error-identifier normalization matches all tokens in the 11 naming files; raw literals and assertions are unchanged.
- `go test -p=1 ./internal/core/authoring ./internal/core/execution ./internal/harness/pi -run '^Test(InvokeAssistedModel.*|RecordAssistedModelInvocation.*|RenderAssistedModelArtifacts.*|ExecutionFailure.*|ExecutionRunner.*|PiTransport(Failures|Worker|WorkerOutput))$' -count=1 -timeout=3m -json` passed all 19 selected roots without failures/skips; owned fixture cleanup passed.
- `go run ./scripts/goquality` passed with 9,411 declarations, maximum five and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...` passed with zero issues.

Current declaration/hash provenance now identifies source commit
`436f6742f7f224ac8c25bfde7cc61bb6ab328332`. The reconciliation verifies all 4,445 workflow backbone
hashes, 217 current compatibility images and all 771 final signature rows.
Thirty current rows across ten existing ledgers are updated for the exact error
names, constructor consumers or cleanup comments. Frozen source/baseline/entry
fields, original introduction commits and CSV line endings are unchanged. The
method map remains 908 assessed source methods and 777 guarded Store methods.
Runtime checks are not repeated for this metadata reconciliation.

## Risks and Follow-ups

All Slice 16 tasks remain open. Final integrated gate evidence and delivery-set
closure require the controller's exact integrated candidate and later phase.
