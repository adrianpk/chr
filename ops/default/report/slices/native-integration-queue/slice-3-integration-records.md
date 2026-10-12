# Slice 3: Durable integration records

Status: reviewing
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

T3.3 adds an invocation-wide sealed budget with per-unit counters and stable reservations. Whole-call input/output allowances share an optional finite token ceiling across executor, merger, reviewer and closing roles. Unknown usage retains its full allowance until reliable settlement. Merger dispatch reservations and accepted replacement bases remain consumed across restart and candidate generations; observation consumes no new submission. Positive operator grants extend ceilings with retained attribution and revision history.

The native common submission journal records the original S, expected B, exact P, capability/review binding, claim generation, dispatch identity, reconciliation deadline and sanitized advertisement, command and response evidence. Git-applied publication with the original PR still pending remains awaiting observation and keeps target exclusion. Confirmation requires the physically verified native receipt, fetched exact Git proof and original canonical merged PR. Unknown response evidence survives later reconciliation.

T3.4 adds bounded authoritative queue, submission, publication-barrier, capacity and closure queries. Reads report interrupted transitions and stale projections without replaying them. The existing SQLite index gains a disposable queue projection; lost or corrupt rows can be rebuilt after source validation. Explicit revision-guarded admission and budget migration retain prior allocation, usage, claims and publication history. Malformed, unsupported or unbound source images remain intact and produce typed reconciliation errors.

## Implementation Notes

The new candidate reader verifies versioned source/history manifests against the existing host Git proof and binds current adapter/capability identity, full source/target refs and S/B/P. Validation evidence retains its owned input and required combined gate. Receipt creation checks the evidence before adding a durable record; receipt reads apply the same checks. Existing historical producers and immutable receipt rules remain in place.

Positive fixtures prepare all three Git methods with the host combined gate. The receipt fixture executes a real local atomic receive-pack transaction, independently fetches the target history and verifies the original-PR observation binding. Additional checks cover unsupported evidence versions with freshly sealed content hashes, stale authorization, ownership mismatches, missing proof, changed file bytes and symlink traversal.

T3.2 allocates sequences from retained authoritative entries under the existing repository lock, without another editable queue or database. The existing file transaction owns admission and phase replacement. Retained invocation bindings pin every entry revision and image; losing or changing an entry cannot manufacture another admission. Canonical repository aliases normalize before target binding.

T3.3 admission journals the initial budget, entry and invocation seal together through the existing file transaction. Each budget update pins its exact image and revision; existing worker, preparation, review/repair, closing and legacy submission usage is retained rather than initialized to zero. Checked signed arithmetic and bounded histories protect reservations and grants. Common-journal intent follows durable reservation before transport IO; projection recovery cannot reserve or dispatch a second attempt. Candidate replacement is blocked while publication remains uncertain. Native publication facts cannot fall back to the legacy submission API.

T3.1 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/64369b81579c8b4136609383568731067bc7bdfd). Acta recorded the completed task as `01M4PG9MKE8EF2DZJB2HJ024TH`.

T3.2 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/9a004ba55c6b72ac69aa0322fba531eebcc257d7). Acta recorded the completed task as `01M4PJDG8Q7TEC5W7ZCGJ576RH`.

T3.3 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/dce7077f632e9d1d93dcdde5d35ba958f4695e0f). Acta recorded the completed task as `01M4PN8W5K2TA9P691MVGK08Y9`.

T3.4 keeps sequence and unit uniqueness in authoritative records. Queries and mutation guards verify ledger ownership, retained usage, reservation generations, original PR target, candidate proof, exact native journal projections and receipt ownership. Capacity observation checks real slot locks, live process groups and unexpired executor/review/closing owners without pruning process evidence. Set closure status verifies actual receipt fanout and retained final gate evidence without executing closure.

## Contracts Added or Changed

- `chr-integration-candidate/v2` and `chr-workflow-receipt/v2` distinguish native proof from historical schemas.
- Owned hashed evidence references, original source manifests and versioned method/history manifests.
- Native authorization binds the original source separately from the published result and rejects incompatible capability/observation versions.
- Existing store and preparation readers enforce the new proof before accepting native evidence.
- Versioned integration entries and additive invocation current/history bindings with joint journal replay and checked sequence/revision bounds.
- Explicit blockers and retained stop intent; native publication/finalization phases require their existing journal and receipt owners.
- Invocation-wide budget schema/version, sealed invocation binding, compatible whole-call allowance and immutable settlement/grant history.
- `chr-integration-journal/v2` extends the existing common journal with exact native intent, dispatch, application, observation and receipt facts; uncertain legacy attempts remain retained.

- Read-only integration query contracts and a rebuildable extension to the existing SQLite index.
- Explicit additive migration, typed source-record blockers and cross-record generation/receipt/publication guards.

## Files of Interest

- `internal/core/workflow_native_records.go`: additive bindings, version dispatch and candidate hashing.
- `internal/core/workflow_native_record_evidence.go`: owned evidence, method, capability, gate, transaction and merged-PR verification.
- `internal/core/workflow_native_records_test.go` and `workflow_native_receipt_test.go`: real Git, unsupported proof, byte preservation and ownership checks.

- `internal/core/workflow_integration_entry*.go`: queue admission, guarded transitions, durable ownership and fault/concurrency tests.
- `internal/core/workflow_integration_budget*.go`: guarded shared ledger, counters, grants, settlement, legacy consumption and race/recovery tests.
- `internal/core/workflow_submission_native*.go`: native common-journal extensions, transport/receipt proof checks and real Git pending/confirmed publication fixtures.

- `internal/core/workflow_integration_query.go`, `workflow_integration_snapshot.go` and `workflow_integration_status.go`: bounded observational queries.
- `internal/core/workflow_integration_index.go` and the additive SQLite migration: disposable coordinator-wide queue rows.
- `internal/core/workflow_integration_migration.go` and `workflow_integration_records_validate.go`: guarded migration and authoritative graph checks.

## Validation

Development regression command passed:

- `go test ./internal/core ./internal/app -run '^TestNativeRecord|^TestNativeReceipt|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowFinalization|^TestWorkflowRoundTrip|^TestWorkflowRecord|^TestWorkflowSubmission' -count=1`

T3.1 checks passed:

- `go test -race ./internal/core -run '^TestNativeRecord|^TestNativeReceipt' -count=1`
- `make verify`
- `git diff --check`

T3.2 focused race check passed (51.501 seconds):

- `go test -race ./internal/core -run '^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt' -count=1`

The T3.2 required pre-commit `make verify` gate passed: strict lint reported zero issues, core tests passed in 833.592 seconds and app tests in 76.943 seconds. `git diff --check` also passed. Final exact-HEAD evidence is recorded by the controller after all task and PR-reference commits.

Pinned strict lint reported zero issues; the full core suite passed in 878.901 seconds and app tests in 85.979 seconds. Final Slice 3 exact-HEAD evidence is recorded separately for the complete clean PR head.

T3.3 checks passed:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeBudget|^TestNativeSubmission|^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt|^TestWorkflowSubmission' -count=1`
- `git diff --check`

The focused race check passed in 202.917 seconds. The completed pre-commit gate reported zero strict-lint issues; core tests passed in 937.940 seconds and app tests in 78.072 seconds. Two earlier gate attempts stopped in strict lint before running tests; the conflicting whitespace condition was split before the successful gate. Tests cover concurrent finite-ceiling reservation, idempotent retry, unknown usage after restart, exactly-once settlement, counter/grant bounds, generation retention, interrupted budget/common-journal replacement, legacy byte/consumption preservation, actual Git application with pending PR metadata, exact receipt confirmation and response-history retention.

T3.4 corrected focused race check passed in 132.700 seconds:

- `go test -race ./internal/core -run '^TestNativeQueue|^TestNativeBudget|^TestNativeSubmission|^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt|^TestWorkflowIndex|^TestWorkflowCacheDeletion|^TestWorkflowSubmissionExclusion' -count=1`

Strict lint reported zero issues. Source formatting, affected Markdown links, ticket directory/status consistency and `git diff --check` passed.

T3.4 focused migration-status checks passed:

- `go test ./internal/core ./internal/app -run '^TestMaintenanceDBStatus|^TestMigrationStatus' -count=1`
- `go test ./internal/core -run '^TestWorkflowSubmissionExclusion|^TestNativeQueueCorruption' -count=1`

The first focused race run exposed an invocation-reader check that rejected linked repository aliases before observing their canonical publication barrier. The reader was corrected while retaining native cross-record scope checks. The first broad gate also exposed migration-status fixtures still expecting fourteen migrations; those expectations now include the fifteenth migration. The earlier broad run completed with these migration-count failures and alias-scope regressions across linked claim, capacity, execution, recovery and publication tests. Its corrected replacement `make verify` passed with zero strict-lint issues; core tests passed in 975.617 seconds and app tests in 103.692 seconds. No running tests were interrupted.

Execution uses Go 1.27.1, the repository-pinned golangci-lint v2.14.0, the retained neutral Git fixture configuration, a thirty-minute Go test timeout and the supported absent default-agent command fixture. Host settings and positive explicit provider/executor fixtures are retained. These settings do not change repository assertions or publication policy.

## Risks and Follow-ups

All four tasks passed their recorded pre-commit and focused checks. Canonical PR preparation and its reference commits precede the controller's ordered exact-HEAD validation. Verified integration and slice closure remain pending.
