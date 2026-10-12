# Slice 5: Method preparation and provenance

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-5-native-method-preparation`
PR: pending

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

## Implementation Notes

The additive `chr-integration-provenance/v1` proof joins the existing history
and candidate authorization hashes. The native source-manifest owner requires
the same original commit list. Historical proof hashes and receipts remain
readable without rewriting their evidence. Preparation retains the original
4096-commit, 1 MiB output/evidence and three-minute operation bounds.

T5.1 implementation and canonical report introduction: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/d6532b26691da0df0002f3d29753a9ae0a52be11). Acta task record: `01M4QJN40VG8QKAKS12Z40FY3T`.

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

## Risks and Follow-ups

T5.1 task validation is complete. T5.2 through T5.4
retain the approved replay attribution, signing/no-op and isolation obligations.
The canonical PR and final exact-HEAD controller checks follow all four tasks.
