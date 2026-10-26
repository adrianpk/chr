# Slice 1: Inventory and enforceable quality policy

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-1-quality-policy`
PR: `#438`

## Purpose

Establish full-tree, enforceable signature guardrails and exact baseline
ownership before the approved capability refactors.

## Delivered Behavior

Normal lint, strict lint and style checks inspect every owned production Go
signature, including generated files and inactive build variants. Grouped
names, variadics, named results, callbacks, interfaces and same-package/local
type aliases have explicit handling. Tests and reference fixtures stay outside
the production policy.

The active maximum is seven parameters. The manifest retains 532 exact
declaration/signature debts against the final five-parameter target; an
independent exception documents the installed Bubble Tea Model.Update contract.
New violations, changed/moved/renamed debt, resolved allowances, duplicate
entries and stale exceptions fail. Existing defect-oriented lint stays
unfiltered. Running on dev does not reduce the scan to an empty Git diff.

All 771 source inventory rows have exact declaration matches and caller
inspection entry points. Six additional category entries have assigned owners.
The secondary inventory records the 14 large files, eight error names, actual
bare-suppression locations and broader flag/compatibility candidates. Source
counts that differ from the review are recorded explicitly; the two original
review files retain their approved SHA-256 hashes.

## Implementation Notes

The checker uses only the Go standard library and the repository's Git file
list. Identity is a repository-relative path plus named declaration or a
scoped callback ordinal; signatures are AST-normalized. There is no automatic
baseline rewrite or merge-base allowance. Exact manifest reassociation belongs
to the owning migration task.

The dependency/naming ledger records the current package graph, shared-value
prerequisites, consumer-owned stage ports, persistence authority and bridge
expiry. No application behavior or package move is part of this slice.

## Contracts Added or Changed

The seven/six/five migration phases and zero/one/(T,error)/(T,bool) result rule
are executable. Temporary debt expires in Slice 16. External exceptions require
an exact signature, rationale and named external interface; local interfaces
receive no blanket exemption.

## Files of Interest

- `scripts/goquality/`: scanner, policy, exact manifests, fixtures and usage.
- `Makefile`: full-tree policy beside existing signature-layout checks.
- `ops/default/plan/code-quality-remediation-reconciled.csv`: all baseline rows.
- `ops/default/plan/code-quality-remediation-secondary.csv`: source locations.
- `ops/default/plan/code-quality-remediation-additional.json`: expanded coverage.
- [Inventory evidence](../../../plan/code-quality-remediation-evidence.md).
- [Dependency and naming ledger](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./scripts/goquality` — passed; includes real temporary Git repositories,
  dev/origin-dev self-comparison, exact debt, aliases, interfaces and build variants.
- `go run ./scripts/goquality` — passed with 532 exact debt entries and one exception.
- `make check-go-style` — passed with the existing layout check retained.
- `GOMAXPROCS=4 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./scripts/goquality/...`
  — passed with zero issues after the scanner's initial implementation.

Final clean-head handoff uses `make verify`, then `git diff --check`, through
the controller's validate command after the PR-reference commit. The controller
retains the exact HEAD, ordered commands, outcomes and execution timings;
that evidence and the PR Validation section own final-candidate gate results.
The short development gate includes checker fixtures, existing CI-script
regressions and the production scan. It does not establish the delivery set's
later integrated acceptance or dev integration.

## Risks and Follow-ups

Caller entries are syntactic candidates, not a receiver-resolved call graph.
Imported qualified result types fail closed rather than being assumed to be
builtin bool/error. Their owning slices must preserve exact debt identity
through moves and verify concrete callers. The six added detections remain
within their existing approved capability owners.
