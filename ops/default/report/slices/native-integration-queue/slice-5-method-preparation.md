# Slice 5: Method preparation and provenance

Status: reviewing
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-5-native-method-preparation`
PR: [#379](https://forge.adrianpk.com/adrianpk/chr/pulls/379)

## Purpose

Prepare exact configured integration histories with host-verified attribution,
method proof and isolation from original worker and user checkouts.

## Delivered Behavior

T5.1 extends merge and squash preparation with the complete original source
commit DAG, author identities and dates, message hashes, parent/tree identities,
prepared committer metadata and deterministic combined diff hash. Merge retains
parents B/S in that order; squash retains one parent B and uses host authorship
with the full original source attribution manifest. Local host guards compare
that evidence with the actual Git objects before validation or publication.

T5.2 preserves the complete ordered linear replay sequence, original author
identities/dates, exact message bytes and encoding. Empty source commits and
already-applied changes remain as attributed empty replays. Separate integration
fix commits retain actionable findings and host attribution after the source
sequence; host proofs reject undeclared fixes and cosmetic repair claims.
Merge-containing histories and wholly empty integration units are blocked before
an owned preparation lane is attached.

T5.3 implements bounded source-identity retries with recorded committer dates
and preparation identity. It adds the trusted host SSH signing path and binds
required signatures to current capability. Every rewritten object and declared
repair is signed and cryptographically verified against the public profile.
Missing or mismatched signing capability blocks before validation or publication.

T5.4 fixtures preserve distinct dirty worker and user checkouts byte for byte,
including staged and unstaged files, indexes, checkout heads, retained fetch
metadata and original evidence. Importing the prepared result no longer overwrites
the original worker's `FETCH_HEAD`. Real altered-result objects and recomputed
proofs are rejected across merge, rebase and squash before canonical ref updates.

## Implementation Notes

The additive `chr-integration-provenance/v1` proof joins the existing history
and candidate authorization hashes. The native source-manifest owner requires
the same original commit list. Historical proof hashes and receipts remain
readable without rewriting their evidence. Preparation retains the original
4096-commit, 1 MiB output/evidence and three-minute operation bounds.

T5.1 implementation and canonical report introduction: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/d6532b26691da0df0002f3d29753a9ae0a52be11). Acta task record: `01M4QJN40VG8QKAKS12Z40FY3T`.

Rebase proof records each original/prepared mapping and an explicit empty replay
decision. Rewritten objects retain original messages without newline normalization
and exclude old signature headers. Integration fixes extend the prepared sequence
while leaving original source mappings intact.

T5.2 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/ed92e270edd489332191667c1c80201536363ba2). Acta task record: `01M4QRQTNWK814YQMJG4Q5F53K`.

T5.3 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/134ebaf9c8041ca92bdd184ee48f6e521a293e07). Acta task record: `01M4QW3D3M4DHF1MHP5B7X11WR`.

T5.4 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/a6a4d6ec83d8dc15f4bdf05712417567dc6145fc). Acta task record: `01M4QY05BSSYEYMEPHE06JK3VR`.

## Contracts Added or Changed

`WorkflowIntegrationProof.Provenance` retains bounded original and prepared
commit attribution. `MessageHash` covers original message bytes, including
trailing newlines. `CombinedDiffHash` covers a fixed binary-capable B/P diff.
The host verifier rejects altered metadata even with recomputed enclosing
hashes. See [Workflow execution](../../../../../docs/reference/workflow-execution.md#combined-result-preparation).

## Files of Interest

- `internal/core/workflow_integration_attribution.go`
- `internal/core/workflow_integration_provenance.go`
- `internal/core/workflow_preparation_provenance_test.go`
- `internal/core/workflow_native_record_evidence.go`
- `internal/core/workflow_integration_replay.go`
- `internal/core/workflow_integration_replay_verify.go`
- `internal/core/workflow_preparation_replay.go`
- `internal/core/workflow_preparation_replay_test.go`
- `internal/core/review_workflow_native_observation.go`
- `internal/core/workflow_submission_observe.go`
- `internal/core/workflow_preparation_identity.go`
- `internal/core/workflow_preparation_signing.go`
- `internal/core/workflow_preparation_policy_test.go`
- `internal/core/workflow_preparation_isolation_test.go`
- `internal/core/workflow_preparation_method_test.go`

## Validation

Initial development checks passed:

- `go test ./internal/core -run '^TestNativeCombined|^TestWorkflowPreparationMethods$|^TestNativeRecordMethods$' -count=1`

Core passed in 39.603 seconds. The initial compile check found an existing
validation-function name collision; the new commit-identity validator now uses
a distinct name. Initial strict lint found complexity, spacing and an unchecked
test hash error; cohesive date/header/mapping helpers and explicit error handling
address those findings. Final T5.1 task checks passed:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeCombined|^TestNativeRecord|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit' -count=1`
- `git diff --check`

Strict lint reported zero issues. The required pre-commit gate passed with core
in 1269.073 seconds, app in 105.618 seconds and all other packages passing. The
expanded focused race regression passed in 262.157 seconds. Tests exercise real
merge/squash preparation and source DAGs, isolated worktrees, exact parent/tree
proof, host source/prepared attribution checks, unchanged refs after rejection,
native evidence bindings, original hash compatibility and existing preparation,
integration and Git transport regressions. Owned fixture processes completed
and their disposable directories were cleaned through the test lifecycle.

T5.2 development checks:

- `go test ./internal/core -run '^TestWorkflowGitAtomic$|^TestWorkflowPreparationMethods$|^TestNativeRecordMethods$' -count=1` — pass, 23.025 seconds.
- `go test ./internal/core -run '^TestNativeReplay' -count=1` — pass, 4.584 seconds after correcting the empty-unit fixture.
- `go test -race ./internal/core -run '^TestNativeReplay|^TestNativeCombined|^TestNativeRecord|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit' -count=1` — pass, 260.652 seconds after correction.

The initial expanded race run failed in 270.432 seconds because the empty-unit
fixture updated its disposable remote ref before importing the new commit
object. Fixture setup now fetches that object first. Initial lint found six
complexity and fifteen spacing issues; cohesive validation helpers and formatting
corrected those findings, followed by zero strict lint issues. The first
pre-commit `make verify` finished with that same fixture error after core ran for
1402.078 seconds; all other packages passed. The next `make verify` failed with native publication-observation hash mismatches
after core ran for 1403.022 seconds. Observation had regenerated rich metadata for
a retained legacy rebase proof. The explicit prepared-history host port now
re-verifies the retained schema, attribution and declared fixes from canonical
objects without rewriting historical hashes. Both legacy observation cases and
new rich replay/fix cases through Forgejo and GitHub pass:

- `go test ./internal/core -run '^TestNativeReplay|^TestNativeObservation|^TestNativeLaterSource|^TestWorkflowIntegrationMethods$' -count=1` — pass, 35.316 seconds.
- `make lint-strict` — pass, zero issues.

The subsequent broad native race run failed in 556.789 seconds at
`TestNativeSubmissionApplied`: its receipt fixture observed rich prepared history
through the legacy query. The fixture and exact-result submission observation now
use the retained-proof port. Strict lint passes and the expanded submission,
receipt, replay and observation checks pass:

- `go test ./internal/core -run '^TestNativeSubmission|^TestNativeReceipt|^TestNativeReplay|^TestNativeObservation|^TestNativeLaterSource|^TestWorkflowSubmission' -count=1` — pass, 167.016 seconds.

The earlier pre-commit gate completed with core in 1445.030 seconds and failures
in native receipt confirmation and generic publication intent matching. These
paths used the legacy observation query for the same rich prepared result. The
corrected exact-result submission path preserves the retained hash; focused
publication regression passes:

- `go test ./internal/core -run '^TestWorkflowPublication' -count=1` — pass, 165.846 seconds.

The final broader race regression passes:

- `go test -race ./internal/core -run '^TestNative|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit|^TestWorkflowSubmission' -count=1` — pass, 729.202 seconds.

Final T5.2 pre-commit checks passed:

- `make verify` — zero lint issues, core 1352.893 seconds, app 124.414 seconds and all other packages passed.
- `git diff --check` — pass.

The final code tree passed both normal repository checks and the expanded race
regression after the fixture and retained-proof corrections. Owned fixture
processes completed naturally and their disposable data was cleaned by the test
lifecycle.

T5.3 development checks:

- `go test ./internal/core -run '^TestNativePreparation|^TestNativeSigning|^TestNativeCollision|^TestNativeReplay|^TestNativeCombined' -count=1` — pass, 18.903 seconds.

The initial compile check found a collision with an existing SSH fixture helper;
the new host signer command helper now has its own name. Initial strict lint
reported complexity, nested branches and formatting; cohesive policy/object
validators, early returns and formatting corrected those findings. The added signed-fix fixture also passes:

- `go test ./internal/core -run '^TestNativePreparation|^TestNativeSigning|^TestNativeSignedFix|^TestNativeCollision|^TestNativeReplay|^TestNativeCombined' -count=1` — pass, 26.664 seconds.

The expanded native/workflow race regression passed:

- `go test -race ./internal/core -run '^TestNative|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit|^TestWorkflowSubmission' -count=1` — pass, 705.730 seconds.

Final T5.3 pre-commit checks passed:

- `make verify` — zero lint issues, core 1345.889 seconds, app 114.726 seconds and all other packages passed.
- `git diff --check` — pass.

The signing fixtures use owned local repositories and temporary Ed25519 keys.
Each prepared signature is independently checked by Git before atomic
publication and retained-proof observation. Required signing on declared repairs
is covered without modifying the original replay manifest. Owned fixtures
completed naturally and their temporary data was cleaned by the test lifecycle.

T5.4 focused checks:

- `go test ./internal/core -run '^TestNativePreparationIsolation$|^TestNativeMethod|^TestNativeCombined|^TestNativeReplay|^TestNativePreparationCollision$|^TestNativeCollisionBudget$' -count=1` — pass, 11.638 seconds.
- `go test -race ./internal/core -run '^TestNativePreparationIsolation$|^TestNativeMethod|^TestNativeCombined|^TestNativeReplay|^TestNativePreparationCollision$|^TestNativeCollisionBudget$' -count=1` — pass, 17.664 seconds.

The initial isolation regression failed for all methods because attachment
overwrote the original worker's `FETCH_HEAD`. The object import now disables
that metadata write. An altered-object fixture initially normalized replay
messages through Git's display format; it now reads exact commit-object message
bytes. The corrected real-object fixtures pass with truthful attribution and
recomputed history/candidate hashes, and prove that false combined trees are
blocked by host verification. Existing ordered replay, parent, empty-replay and
source-identity retry fixtures pass in the same focused run.

Final T5.4 pre-commit checks passed:

- `make verify` — zero lint issues; core 1260.604 seconds, app 104.874 seconds and all other packages passed.
- `git diff --check` — pass.

Owned test fixtures completed naturally and their disposable data was cleaned
through the test lifecycle.

Required final candidate commands, in canonical order:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

The controller ledger binds these commands to the final clean PR head after
PR-reference commits and pushes. Its exact-HEAD evidence is authoritative; the
canonical PR Validation section lists these commands only after they pass.

## Risks and Follow-ups

All four Slice 5 tasks are locally validated. Signed preparation is locally
certified; live
provider signing-profile certification remains in the recorded certification
slice.
Canonical PR #379 is open against `dev`; the controller owns integration.
Required final candidate evidence uses the controller ledger after PR-reference
updates. All task tickets remain reviewing until verified integration.
