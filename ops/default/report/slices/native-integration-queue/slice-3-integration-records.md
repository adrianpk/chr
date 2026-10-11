# Slice 3: Durable integration records

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-3-native-integration-records`
PR: pending

## Purpose

Persist native integration identity and proof through the existing workflow owners while retaining historical receipts, invocation consumption and uncertain publication.

## Delivered Behavior

T3.1 adds distinct native candidate and receipt schemas with original S, published P, validated B, exact tree, method, owning invocation/unit and hashed capability, preparation, source/history and validation evidence. The candidate hash includes the complete preparation input and optional extensions. A native receipt requires the retained candidate, exact atomic transaction and original canonical merged-PR observations. An open or unconfirmed PR cannot create one.

Historical candidates retain their original reader. Historical receipt reads preserve their schema and exact bytes; a legacy schema cannot represent a native candidate or carry the new receipt extension. Unsupported versions, missing method proof, substituted evidence, foreign invocation references and symlinks block rather than inventing zero-value authorization.

T3.2 adds idempotent ticket and slice entries with a coordinator-wide sequence across linked namespaces. Admission journals the entry and exact invocation pointer together. Guarded transitions reject stale invocation/entry revisions, claim generations and candidate hashes. A blocked entry resumes only to its recorded state; pause requires an explicit guarded resume. Unknown or confirmed publication cannot be cancelled or requeued.

Persistent budgets, native journal extensions and cross-record migration/index validation remain the recorded T3.3 and T3.4 work.

## Implementation Notes

The new candidate reader verifies versioned source/history manifests against the existing host Git proof and binds current adapter/capability identity, full source/target refs and S/B/P. Validation evidence retains its owned input and required combined gate. Receipt creation checks the evidence before adding a durable record; receipt reads apply the same checks. Existing historical producers and immutable receipt rules remain in place.

Positive fixtures prepare all three Git methods with the host combined gate. The receipt fixture executes a real local atomic receive-pack transaction, independently fetches the target history and verifies the original-PR observation binding. Additional checks cover unsupported evidence versions with freshly sealed content hashes, stale authorization, ownership mismatches, missing proof, changed file bytes and symlink traversal.

T3.2 allocates sequences from retained authoritative entries under the existing repository lock, without another editable queue or database. The existing file transaction owns admission and phase replacement. Retained invocation bindings pin every entry revision and image; losing or changing an entry cannot manufacture another admission. Canonical repository aliases normalize before target binding.

T3.1 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/64369b81579c8b4136609383568731067bc7bdfd). Acta recorded the completed task as `01M4PG9MKE8EF2DZJB2HJ024TH`.

T3.2 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/9a004ba55c6b72ac69aa0322fba531eebcc257d7). Acta recorded the completed task as `01M4PJDG8Q7TEC5W7ZCGJ576RH`.

## Contracts Added or Changed

- `chr-integration-candidate/v2` and `chr-workflow-receipt/v2` distinguish native proof from historical schemas.
- Owned hashed evidence references, original source manifests and versioned method/history manifests.
- Native authorization binds the original source separately from the published result and rejects incompatible capability/observation versions.
- Existing store and preparation readers enforce the new proof before accepting native evidence.
- Versioned integration entries and additive invocation current/history bindings with joint journal replay and checked sequence/revision bounds.
- Explicit blockers and retained stop intent; native publication/finalization phases require their existing journal and receipt owners.

## Files of Interest

- `internal/core/workflow_native_records.go`: additive bindings, version dispatch and candidate hashing.
- `internal/core/workflow_native_record_evidence.go`: owned evidence, method, capability, gate, transaction and merged-PR verification.
- `internal/core/workflow_native_records_test.go` and `workflow_native_receipt_test.go`: real Git, unsupported proof, byte preservation and ownership checks.

- `internal/core/workflow_integration_entry*.go`: queue admission, guarded transitions, durable ownership and fault/concurrency tests.

## Validation

Development regression command passed:

- `go test ./internal/core ./internal/app -run '^TestNativeRecord|^TestNativeReceipt|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowFinalization|^TestWorkflowRoundTrip|^TestWorkflowRecord|^TestWorkflowSubmission' -count=1`

T3.1 checks passed:

- `go test -race ./internal/core -run '^TestNativeRecord|^TestNativeReceipt' -count=1`
- `make verify`
- `git diff --check`

T3.2 focused race check passed (51.501 seconds):

- `go test -race ./internal/core -run '^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt' -count=1`

The T3.2 required pre-commit `make verify` gate passed: strict lint reported zero issues, core tests passed in 833.592 seconds and app tests in 76.943 seconds. `git diff --check` also passed. Final exact-HEAD checks remain pending until T3.3, T3.4 and the PR-reference follow-up are complete.

Pinned strict lint reported zero issues; the full core suite passed in 878.901 seconds and app tests in 85.979 seconds. Final Slice 3 exact-HEAD evidence remains pending until all four tasks and PR-reference follow-up commits are complete.

Execution uses Go 1.27.1, the repository-pinned golangci-lint v2.14.0, the retained neutral Git fixture configuration, a thirty-minute Go test timeout and the supported absent default-agent command fixture. Host settings and positive explicit provider/executor fixtures are retained. These settings do not change repository assertions or publication policy.

## Risks and Follow-ups

T3.2 queue persistence is validated with focused race checks and `make verify`. T3.3 owns persistent attempts, budgets and publication subphases. T3.4 owns additive migration, strict cross-record validation and rebuildable queue queries. T3.3 and T3.4 remain pending; no canonical Slice 3 PR has been opened.
