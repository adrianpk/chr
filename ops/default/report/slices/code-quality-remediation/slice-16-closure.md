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

## Risks and Follow-ups

All Slice 16 tasks remain open. Final integrated gate evidence and delivery-set
closure require the controller's exact integrated candidate and later phase.
