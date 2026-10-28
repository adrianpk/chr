# Production signature policy

Run `go run ./scripts/goquality` from the repository root. The check reads the
complete owned production tree on every branch, including dev. It does not use
a merge-base, changed-line filter or remote reference. `-inventory` emits the
declarations and syntactic call sites as JSON for explicit reconciliation.

Sources are Git-tracked and untracked non-ignored Go files at the module root
and under cmd, internal, pkg and scripts. Tests and testdata are excluded;
embedded example repositories, operational snapshots and QA reference modules
are outside these production roots. Generated sources and every build variant
are inspected, with their generated/build metadata present in the inventory.
Mutually exclusive variants must still obey the same policy.

Grouped parameters/results count once per name. Receivers are excluded and a
variadic parameter counts once. Function declarations, interface methods,
named function types and function literals are inspected. Local aliases of
bool/error resolve through same-package alias chains across files; defined
types remain distinct. Qualified imported types are not assumed to be builtin
bool/error: nonconforming results require explicit reconciliation. This is a
syntactic contract check, not interface satisfaction or whole-program call
graph analysis. Selector call sites in the inventory are candidates by method
name, not proof of a particular receiver's runtime reachability.

The permanent parameter maximum is five. The result policy permits zero or one
result, or two whose second result is error or bool. The checker has no phase
configuration, debt manifest or migration allowance path. Every production
source is checked against the same contract on feature branches and dev.

exceptions.json is permanent external-interface evidence. Each entry requires
its exact declaration, signature, external contract and rationale. Locally owned
interfaces obey the normal policy. Empty, duplicate, changed, stale and unnecessary
exceptions fail. Model.Update has the single imposed Bubble Tea result contract;
internal update helpers have no exception.

The pinned revive argument-limit (five) and function-result-limit (two) rules
provide additional declaration checks. The AST check still covers function types,
interface methods, literals and second-result meaning. dupl uses threshold 150 on
production code. These new production rules exclude test fixtures; all existing
lint protections keep their previous test scope.

The existing signature-layout check and all defect-oriented lint rules remain
unfiltered. `make verify` runs both checks, tests and the application build.


Extracted leaf packages also have exact module import bounds in boundaries.go.
The same production scan collects imports for every build variant and generated
source. Gate, project, notify and their shared dependencies cannot import root
core, application adapters or a workflow coordinator. Descendants inherit their
owner's contract. Standard-library and existing external dependencies remain
subject to the repository's ordinary module/lint rules. Tests are excluded; real
consumer integration fixtures remain free to exercise public boundaries.

The focused AST fixtures reject aliases, forbidden generated/build-variant imports
and nested-directory bypasses. They also preserve valid shared values, external
drivers and application assembly. A real new dependency requires updating the
code-backed owner contract in the same approved task.
