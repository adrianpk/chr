# Slice 2: Forge publication profiles

Status: reviewing
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-2-native-forge-profiles`
PR: [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376)

## Purpose

Admit an explicitly authorized owned-source publication profile and observe the original canonical PR after an exact atomic Git transaction through the existing Forgejo and GitHub adapters.

## Delivered Behavior

`work.integration.publication_profile` is disabled by default, independently of review mode and merge strategy. Initialization and legacy upgrades preserve that disabled policy and existing review choices. Retained thirteen-field upgrade snapshots remain readable without changing their bytes or implicitly enabling the new profile.

The explicit `git_atomic_dual_ref` profile requires current grants for both exact full refs, source rewrite authority, host-bound candidate review, credential identity, method certification, effective repository protections and a real receive-pack advertisement. Unsupported combinations identify the condition and required operator decision before sending updates. The capability freezes S, B and P, ownership revision, current policy, publisher identity, negotiated transaction features and adapter observation contract; a recheck invalidates changed authorization.

Observation retains the original PR and S separately from published P, fetches the canonical target and verifies exact method/history/tree proof. Metadata lag remains observing. Close, retarget, incompatible results and expired observation produce explicit reconciliation. Later source and target descendants preserve the original published proof while recording later source movement separately. The observer never issues receipts or finalizes tasks.

## Implementation Notes

Read-only transport inspection ends receive-pack with an empty command list. Actual publication repeats the expected-old checks and atomic two-ref transaction. Repository policy and identity API checks are read-only; uncertified native protections, approvals and effective branch rules block the baseline profile. Existing source-only forge merge API paths remain blocked.

GitHub appends one empty transport flush after its two ref statuses. The parser accepts precisely that form and rejects missing, duplicate, malformed and additional status data. Capability fingerprints bind negotiated transaction features rather than session nonces. A speculative open-PR `merge_commit_sha` is excluded from source-binding comparison; merged proof still requires exact P and actual target history.

Forgejo 16.0.3 requires repository-level automatic Git-merge detection. That option was enabled only on the designated disposable repository under the approved fixture-configuration authority. A disabled-option probe applied Git but retained an open PR; a fresh enabled-option probe observed the same original PR merged at exact P. One GitHub probe briefly appeared closed/unmerged before converging to merged at P; observations preserve that explicit reconciliation condition until authoritative metadata converges. Applied diagnostic transactions were observed independently and never republished.

Task implementation references:

- [T2.1: profile admission and policy-preserving upgrade](https://forge.adrianpk.com/adrianpk/chr/commit/131dd1733a087585add492b68ba5fa3177ef3ec7).
- [T2.2: original-PR observation and exact target history](https://forge.adrianpk.com/adrianpk/chr/commit/17a308198809683a0bd8adff965379bbd6dd90a6).
- [T2.3: GitHub wire compatibility and shared live certification](https://forge.adrianpk.com/adrianpk/chr/commit/a7e90f2ee18ba5d09e46677c101b38f981916b1b).

## Contracts Added or Changed

- Disabled/atomic profile configuration without changing review policy.
- Trusted current host-policy and read-only receive-pack inspection ports.
- Current capability bindings for canonical PR/full refs, S/B/P, ownership, roles and policy.
- Existing adapter observations combining same-PR facts with actual target history and reconciliation outcomes.
- Strict recognition of GitHub's additional empty report terminator.

## Files of Interest

- `internal/core/workflow_native_profile.go` and `review_workflow_native_capability.go`: admission and current capability binding.
- `internal/core/workflow_git_capability.go`: real read-only receive-pack inspection.
- `internal/core/review_workflow_native_observation.go`: same-PR metadata and actual target proof.
- `internal/core/config_native_profile_test.go`: disabled initialization, defaults and policy-preserving upgrade compatibility.
- `internal/core/workflow_native_profile_live_test.go`: real-forge rebase certification with retained host-owned evidence.

## Validation

Development checks passed:

- `go test ./internal/core ./internal/app -run 'TestInitWorkflow|Test.*Config|Test.*Upgrade|TestNative|TestWorkflowGit' -count=1`
- `CHR_NATIVE_INTEGRATION_QA=1 go test ./internal/core -run '^TestWorkflowNativeRebase$' -count=1 -v`
- `make verify`

The first pre-commit `make verify` failed an initialization assertion that still expected thirteen effective configuration fields. It now expects fourteen and explicitly checks the profile remains disabled; the repeated complete gate passed. Strict lint uses the repository-pinned golangci-lint v2.14.0. Execution uses Go 1.27.1, an isolated Git configuration retaining host settings with a neutral fixture identity, a thirty-minute Go test timeout and the supported absent default-agent command fixture. Positive provider/executor fixtures remain explicit. These execution settings do not alter repository assertions or policy.

Final candidate outcomes belong to Looper's ledger for the exact clean PR head. The ordered command list is:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `CHR_NATIVE_INTEGRATION_QA=1 go test ./internal/core -run '^TestWorkflowNativeRebase$' -count=1 -v`
- `git diff --check`

The [fixture report](../../native-integration-fixtures.md) binds both designated private repositories, server/API profile, credential roles and branch policy. Development certification passed through the actual publisher and existing adapters; full S/B/P and ordered-history proof is retained below and in host-owned JSON evidence. The original canonical PRs were observed merged at exact P without a native merge API or provider merge queue.

- forgejo: [original PR](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/5), S `82067e2c536a2dbd7625de05a3bb1e01c6bc5c70`, B `8bdbe30c8a44418b7208a1b6e1f17698d5ad4937`, P `6647bd965f819403d9e8cf88a73d4b618dde7cf8`, ordered history hash `97a6d3e7758f1dac487cec73b07226dc1f00ac11ade403185dee43288e3cbf1f`; retained host file `forgejo-native-rebase-dm28t4n85rtu.json`.
- github: [original PR](https://github.com/adrianpk/chr-integration-qa/pull/9), S `a5456ae7af376bcc874cee9858375f1a11f1a790`, B `9a10f0740792b042d3ed50a2b47415d1427ad551`, P `cbc3d631d8866119f5a69dbfd40b3c5fe4106f0a`, ordered history hash `2627117e8c14e89217b24edd24b1f28372f4b3e06d71801abd7e7a457faae435`; retained host file `github-native-rebase-dm28t77ztoqf.json`.

| Acceptance boundary | Evidence in this slice |
| --- | --- |
| AC-032, AC-044..048, AC-080..081 | Disabled defaults/upgrade compatibility, exact ownership, fork/native-policy denials and current authorization invalidation. Signed/protected equivalence remains uncertified by the baseline. |
| AC-061..063, AC-099 | Pending/closed/retargeted/result-mismatch observations; retained S/P binding and actual target-history verification. |
| AC-083, AC-086, AC-090 | Real default rebase on both designated repositories with original PR, exact commits, current role/policy, negotiated capability and actual method/history/tree proof. |
| AC-082, AC-084..085, AC-087, AC-089 | Adapter contracts and conservative denials are present; the remaining real method/denial matrix is assigned to the approved Slice 12. |

## Risks and Follow-ups

Forge metadata follows Git application asynchronously. Missing or inconsistent merged proof retains reconciliation; target rollback or repeated publication is unsafe. The complete method and denial matrix belongs to Slice 12. Durable queue, intent and receipt integration belongs to the recorded successor slices.
