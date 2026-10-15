# Slice 12: Live provider certification

Status: blocked
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `test/slice-12-native-provider-certification`
PR: pending

## Purpose

Certify exact merge, rebase and squash publication through the existing adapters
on the two designated private repositories, including contention and explicit
unsupported-profile boundaries.

## Delivered Behavior

T12.1 is committed after passing pre-commit validation. The Forgejo live harness creates separate owned source and
isolated target refs for every method and scenario. Positive cases require exact
prepared result publication, convergence of the same original PR to merged,
independent remote history/tree verification and compatibility with the existing
native receipt observation and transaction validators. The probe does not issue
a runtime delivery receipt.

Source and target contention use an independent authenticated writer after the
publisher reads the actual receive-pack advertisement. Each case requires an
atomic rejection, preservation of the other ref and an unmerged original PR.

The real Forgejo probe exposed an incorrect merge-policy field mapping. The
adapter now reads Forgejo's `allow_merge_commits` and GitHub's
`allow_merge_commit` according to the existing observation contract. An explicit
denial or missing provider-specific field remains unsupported.

T12.1 implementation: [ae806985740281a4a77ff7e06c1d448f4a6a1a86](https://forge.adrianpk.com/adrianpk/chr/commit/ae806985740281a4a77ff7e06c1d448f4a6a1a86). Acta `01M4YGSZ9CG7RDMTYE49P9WKY5`.

T12.2 is committed after passing all nine GitHub cases under the race detector.
A subsequent repeatability run produced a GitHub rebase counterexample. Its
canonical merge evidence contradicts exact P, so GitHub rebase certification is
revoked and T12.3 is blocked. The six-path delivery contract is not satisfied.

T12.2 implementation: [a9b780729b8d7d1fab45e0cde03ba02240e8a625](https://forge.adrianpk.com/adrianpk/chr/commit/a9b780729b8d7d1fab45e0cde03ba02240e8a625). Acta `01M4YHRTEX8AH2R7Y2CM0P3CPT`.

## Implementation Notes

All live mutations are limited to the [designated certification fixtures](../../native-integration-fixtures.md):
[Forgejo](https://forge.adrianpk.com/adrianpk/chr-integration-qa) and
[GitHub](https://github.com/adrianpk/chr-integration-qa). Source refs use
`refs/heads/chr/work/`; isolated targets use `refs/heads/chr/qa/`. Existing host
credentials and the repository owner push/admin role supply access. Baseline
repository policy permits all three methods and retains source branches.

The test-only method probe uses the current descriptor's explicit method
authorization and the real adapter's repository/branch policy checks. It leaves
the host's previously certified method list unchanged. Certification is an
outcome of the real publication and observation checks.

Each case retains host-owned JSON under the state directory's
`chr/qa/native-integration-slice-12/` subtree. Evidence includes the canonical PR,
method, public server/API profile, credential role, sealed policy/transport
bindings, Git objects and observations, time bounds and pass/failure result.
Private credential references and transport locations remain host-owned.
Fixture refs and failed PRs are retained for diagnosis.

The first live runner lacked the current SSH-agent environment. Its fixture
pushes failed before creating PRs. The owned runner now inherits the existing
agent socket. Setup failure evidence is retained before the initial push.

## Contracts Added or Changed

- Provider-specific merge-policy mapping preserves explicit denial and missing
  policy instead of admitting another provider's field.
- Live certification is opt-in and binds the exact two designated repository
  URLs and owned ref prefixes.
- Source and target races occur after advertisement and require both-ref atomic
  rejection without a native merge API or another publication attempt.
- Positive evidence must satisfy the existing native receipt proof validators;
  fixture observation does not finalize a task delivery.

## Files of Interest

- `internal/core/review_workflow_native_policy.go` and its regression test:
  provider-specific merge flags, explicit denial and absent-field checks.
- `internal/core/workflow_native_certification_live_test.go`: three Forgejo
  methods and actual source/target contention.
- `internal/core/workflow_native_certification_policy_test.go` and
  `workflow_native_certification_evidence_test.go`: current authorized profile
  probes, remote proof compatibility and retained bounded evidence.
- `internal/core/workflow_native_profile_live_*test.go`: shared method fixtures
  and original real rebase certification.

## Acceptance Evidence

| Acceptance IDs | Current evidence boundary |
| --- | --- |
| AC-082 | Corrected real merge publication converged on original fixture PR #18 with exact P and receipt-compatible proof. |
| AC-083, AC-084 | Corrected real rebase and squash publications converged on original fixture PRs #21 and #24 with exact Git and receipt-compatible proof. |
| AC-037, AC-038, AC-088 | Both providers and all three methods passed actual source/target races after advertisement over SSH and HTTPS, rejecting both publication updates while retaining the independent writer. |
| AC-036, AC-052, AC-063 | Successful cases independently verified exact remote history/tree and original S versus published P through native receipt compatibility checks; the counterexample produced no compatible receipt. |
| AC-061, AC-099 | Real closed/unmerged and externally closed/retargeted PR observations retained reconciliation and did not roll back P or produce receipt-compatible proof. |
| AC-097 | Actual SSH/HTTPS publication and contention passed; actual HTTPS lost-report cases reconciled one POST without resubmission. Real local SSH/HTTPS unknown-result guards passed. |
| AC-090 | Live evidence binds Forgejo 16.0.3+gitea-1.22.0, REST api/v1 through the existing adapter, the authorized role/policy, selected method, PR and exact Git objects. |
| AC-085, AC-087 | GitHub merge and squash passed through SSH and HTTPS with exact P and the original PR observed merged. |
| AC-086 | BLOCKED: subsequent GitHub rebase PR #27 reports original S as its canonical merge result while both refs contain P. Earlier successful cases do not establish repeatable certification. |
| AC-046, AC-047, AC-089 | Both actual private repositories rejected anonymous HTTPS access before commands and rejected native-review policies on exact owned targets. Those created protections remain enforced. |

T12.1 certified positive cases:

| Method | Original canonical PR | Original S | Expected B | Exact P |
| --- | --- | --- | --- | --- |
| Merge | [Fixture PR #18](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/18) | `0a8209a5005e4851c8adc40d993509d0beed84a5` | `1ddcfe452d994329db6e33f91c3770221a6a68e7` | `c0fe057a94c830bacec21cbc3f3c388abe76ad02` |
| Rebase | [Fixture PR #21](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/21) | `475abc9fa896fed112c5d3ebb7e4b28bccb6a3ce` | `e2d3adf630969805172b0f82d7d491dc52c1a62c` | `4d1c2d7b86ac879bd18991f1b8afbfe2db4f8f62` |
| Squash | [Fixture PR #24](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/24) | `54e5810877d602247e9e66321f900f848f1e579f` | `c04fb538747a766d9ca901c4e4995d97178d94f3` | `33933f284b53a9febb2ee1ccb3fe96b3e82bfcb7` |

All six rejected race cases retained an open original PR:
- merge source race: [Fixture PR #19](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/19).
- merge target race: [Fixture PR #20](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/20).
- rebase source race: [Fixture PR #22](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/22).
- rebase target race: [Fixture PR #23](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/23).
- squash source race: [Fixture PR #25](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/25).
- squash target race: [Fixture PR #26](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/26).

Initial merge-policy failures remain retained as fixture PRs
[9](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/9),
[10](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/10) and
[11](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/11).

T12.2 certified positive cases:

| Method | Original canonical PR | Original S | Expected B | Exact P |
| --- | --- | --- | --- | --- |
| Merge | [Fixture PR #13](https://github.com/adrianpk/chr-integration-qa/pull/13) | `c1180e527a7675ab833900731f424128c3022a69` | `e2f4365811cb0c22d1482629d8919089e5597430` | `386d9dc7879509706362b7cfa2d0a4b63238e263` |
| Rebase | [Fixture PR #16](https://github.com/adrianpk/chr-integration-qa/pull/16) | `225a213f645ab669e7810f54800ae884387e0aab` | `cd99d2f0ac4473999cfe9c63f84be7a9fa5c8502` | `c7df537f350354d1ae31f52e6ec8294355747fcc` |
| Squash | [Fixture PR #19](https://github.com/adrianpk/chr-integration-qa/pull/19) | `98892871a01e36c89c6ccbe48f75dc5ec83ab47e` | `2df1c5cd2d533e9c770a47e3ecd7496d526706b4` | `29143aede86e5e0ab3b78ca456ed9d0712adfbaf` |

All six GitHub source/target races passed with atomic rejection and an open
original PR:
- merge source race: [Fixture PR #14](https://github.com/adrianpk/chr-integration-qa/pull/14).
- merge target race: [Fixture PR #15](https://github.com/adrianpk/chr-integration-qa/pull/15).
- rebase source race: [Fixture PR #17](https://github.com/adrianpk/chr-integration-qa/pull/17).
- rebase target race: [Fixture PR #18](https://github.com/adrianpk/chr-integration-qa/pull/18).
- squash source race: [Fixture PR #20](https://github.com/adrianpk/chr-integration-qa/pull/20).
- squash target race: [Fixture PR #21](https://github.com/adrianpk/chr-integration-qa/pull/21).

## Validation

Development validation:

- `go test ./internal/core -run '^TestNativeForgejoMethods$' -count=1` — compiled
  and skipped live writes without opt-in; this does not certify a remote path.
- `go test ./internal/core -run '^(TestWorkflowGitAtomic|TestWorkflowGitAdvertisement|TestNativeReceiptProof|TestNativeCapability|TestNativeCapabilityDenials)$' -count=1` — passed in 7.621 s.
- `make lint-strict` — passed with zero issues after correcting callback context
  propagation and test formatting. A later regression test required one spacing
  correction; strict lint is included in the pending full pre-commit check.
- `CHR_NATIVE_INTEGRATION_QA=1 go test ./internal/core -run '^TestNativeForgejoMethods$' -count=1 -v` — first runner failed before remote PR creation due to its missing SSH-agent environment. The authenticated runner completed in 40.687 s on unchanged source: all rebase and squash cases passed; merge admission failed on the actual provider field. Acta `01M4YDZ3E3PV2H6BYB3CMWD698`.
- `go test ./internal/core -run '^TestNativeMergePolicy$' -count=1` — reproduced
  false denial and wrong-provider-field acceptance before the mapping fix.
- `go test ./internal/core -run '^(TestNativeMergePolicy|TestWorkflowGitAtomic|TestWorkflowGitAdvertisement|TestNativeReceiptProof|TestNativeCapability|TestNativeCapabilityDenials)$' -count=1` — passed after the mapping fix in 9.717 s.

The owned T12.1 validation completed on unchanged seven-file source
`2acd17be65196ad0f760c11bb43ff4a380fe502b8be0150260e6299cdc474f53`:

- `make verify` — passed with zero lint issues; core 2747.252 s, app 197.636 s,
  TUI 0.065 s and Web 0.084 s. All remaining Web packages passed.
- `CHR_NATIVE_INTEGRATION_QA=1 go test ./internal/core -run '^TestNativeForgejoMethods$' -count=1 -v` — passed all nine live cases in 51.915 s.
- `go test -race ./internal/core -run '^(TestNativeMergePolicy|TestWorkflowGitAtomic|TestWorkflowGitAdvertisement|TestNativeReceiptProof|TestNativeCapability|TestNativeCapabilityDenials)$' -count=1` — passed in 10.642 s.

Acta `01M4YGRAHG25NY28CDGH3JQKKN`. The owned validation unit exited successfully;
its command, source, process and result evidence is retained on the host.
The Forgejo descriptor now records the three successfully certified methods and
their canonical PR/source/base/result evidence.

T12.2 uses the same exact method, contention and receipt-proof checks through the
existing GitHub adapter. Strict lint passed with zero issues. Its first live
race run failed in 580.665 s because the local `gh` wrapper's tool-manager step
exceeded each bounded provider deadline; canonical recovery confirmed no PR was
created. The passing run selected the same installed GitHub CLI directly in the
fixture PATH. No wrapper or global configuration was edited. Failed evidence is
retained as Acta `01M4YHEHZC13K2TXE6MMFXZAVD`.

- `CHR_NATIVE_INTEGRATION_QA=1 go test -race ./internal/core -run '^TestNativeGitHubMethods$' -count=1 -v` — passed all nine live cases; package 221.147 s and command 222.095 s on unchanged eight-file source `d4dc984fbd33e29ff8f0d6612c6bfe8bd07073b7b60191d5879affae715692e5`. Acta `01M4YHQ5TQDSP5NXAZ10HHQVTZ`. The owned unit exited successfully.

T12.3 remains blocked. `make qa-native-integration` explicitly opts into the designated
live profiles and includes both forge/method matrices, smart HTTPS repeats,
anonymous permission denial, retained native-review policies on exact owned
targets, original-PR close/retarget races, and lost HTTPS report reconciliation.
The entrypoint also reuses actual local SSH/HTTPS atomic, missing-capability and
lost-report guards. Fork rejection is a local declared cross-repository boundary;
no live fork or hosted no-atomic profile is claimed.

Strict lint passed with zero issues after context propagation and bounded helper
corrections. Local protocol and capability guards passed in 4.744 s; live cases
were skipped in that opt-out command and are certified separately. Acta
`01M4YJ32QVMTX11ZKDXW9SC47Q`.

- `make qa-native-integration` — failed in 636.930 s on unchanged 14-file source `435d082c526a6065bec62aee4a54ab9c4f3df988a682c6c68913feb1104c715f`: 45 of 46 real cases passed; GitHub SSH rebase publication PR #27 failed exact canonical result proof.
- `go test -race ./internal/core -run '^(TestNativeMergePolicy|TestWorkflowGit(Atomic|Advertisement|HTTPS|SSH|SSHUnknown)|TestNativeCapabilityDenials|TestNativeReceiptProof)$' -count=1` — the live-matrix runner skipped this command after failure. A subsequent independent focused run passed in 12.113 s package time and 34.826 s command time on the same unchanged 14-file source. Acta `01M4YKF8DNFZTDZ6YFY5VJ7D0G`.

The subsequent canonical check still found S in GitHub PR #27's REST merge
result, with P in its head and base. The focused race result supplies local
development evidence; it does not resolve live certification or provide final
clean-HEAD controller validation.

The required clean-HEAD controller checks will run after task and PR-reference
commits in the exact configured order. No delivery PR is ready and no reserved
full delivery gate has run.

## Certification Blocker

[GitHub fixture PR #27](https://github.com/adrianpk/chr-integration-qa/pull/27)
retains this actual counterexample:

| Identity | Exact value |
| --- | --- |
| Original S | `6a2188ec06a78566c85f952d1d2a637e273b077f` |
| Expected B | `04498b30d4798d5c646d7eb30b506cfbe4758867` |
| Prepared and actually published P | `1bf65f7c73daa9259d21101e63784d95be38f331` |
| REST merge result | `6a2188ec06a78566c85f952d1d2a637e273b077f` |
| GraphQL mergeCommit and MergedEvent commit | `6a2188ec06a78566c85f952d1d2a637e273b077f` |

The publisher reported applied. Authenticated Git reads and canonical API reads
confirm both source and target equal P. The original PR first appeared closed
and unmerged, then merged with the result S. Later REST, GraphQL and immutable
merge-event reads agree on S. Waiting for head convergence does not supply the
required exact result proof. The observer correctly returns
`canonical_pr_result_mismatch`; accepting S as P would weaken the approved
contract.

Acta failure/proof activity: `01M4YJX5T41SA4VH4Q32D96T3Z`. Canonical merged-state
activity records the incompatible result explicitly:
`01M4YJX3S3X1TZ9QEZJ3D6FJGJ`. Full case, process and source evidence remains on the
host. GitHub rebase was removed from the descriptor's certified method list;
its earlier passing evidence and this counterexample are both retained.

The blocked T12.3 implementation remains uncommitted in the dedicated slice
worktree. The QA entrypoint and all exact-P assertions remain intact. A provider
profile that supplies the required canonical result, or an explicitly revised
approved compatibility contract, is required before this slice can be ready.

## Risks and Follow-ups

T12.3 and Slice 12 remain blocked by the concrete GitHub rebase result-proof
counterexample. There is no ready canonical delivery PR or exact-HEAD final
validation. The remaining source changes are retained for continuation.
