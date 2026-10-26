# Quality dependency and naming ledger

Owner: T1.5; refined by T2.3/T2.4 before package moves.
Baseline: `3a35bdf3445120cad3c47874d59712ef47227531`.
Plan: [Approved delivery](code-quality-remediation.md).

## Observed production graph

`go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/... ./cmd/...`
confirms these module-local edges. Standard-library and external dependencies
are omitted here; tests do not define production ownership.

| Consumer | Existing dependencies |
| --- | --- |
| cmd/chr | app |
| app | core, harness, harness/codex, harness/pi, tui, webui, webui/ui |
| core | harness, harness/process |
| harness | none |
| harness/codex | harness, harness/process |
| harness/pi | harness, harness/process |
| harness/process | harness |
| tui | core |
| webui | webui/htmx, webui/render, webui/ui |
| webui/render | webui/htmx, webui/ui |
| webui/ui | webui/htmx |
| webui/htmx | none |

All names above are relative to internal/ except cmd/chr. The graph compiles
without cycles. The tooling command scripts/goquality depends only on the
standard library and has no application import edge.

## Shared values and move prerequisites

| Responsibility | Owner and allowed direction | Required before moving |
| --- | --- | --- |
| Adapter-neutral runner results and process execution | Existing internal/harness and harness/process; never import core workflow policy | Keep provider identity and project admission distinct in Slice 10 |
| Workspace configuration, user environment and neutral path/hash operations | Existing core declarations; retain only genuinely shared primitives at the core root | T2.3/T2.4 isolate dependencies before any leaf imports root core; root must not import that leaf or coordinator |
| Project identity and registry | core/project after Slice 6 | Move behavior with its config consumer ports; no import of a workflow stage |
| Ticket identity, indexing and status | core/ticket after Slice 11 | Separate readiness/execution from ticket values; preserve persisted spelling |
| Workflow invocation, claim and retained outcome values | core/workflow/state after Slice 12 | State contains values, not storage, dispatch, transport or provider clients |
| Durable filesystem operations, locks and journals | core/workflow persistence owner after Slice 12 | Store owns atomic operations over state; never expose a mutable untyped store or split a transaction across stage callers |
| Candidate preparation and merge policy | core/workflow/integrate and core/workflow/merge | Consumer-owned narrow ports use state values; store/coordinator supplies implementations |
| Provider publication and review | core/workflow/forge and core/workflow/review | Depend on state and provider contracts, not core/workflow coordinator |
| Closing decisions and receipts | core/workflow/closeout | Keep receipt/terminal distinctions and durable ownership |
| Sequencing and lifecycle | core/workflow | Imports stages and implements their ports; stages never import this coordinator |
| Reusable application operations | Existing central internal/app service | Adapter wrappers map input/results; service files stay free of TUI/Web types |

No move is authorized by a name alone. Each task first checks actual symbol
and import dependencies, resolves private helper ownership and updates callers
in the same compiling unit. The approved owner handles unresolved shared-value
cohesion before moving; exporting private internals or adding a generic common,
util or options package is not an acceptable cycle repair.

## Domain names

| Term | Governing meaning and constraint |
| --- | --- |
| ticket-native | Canonical ticket-owned authority in spec/stable/ticket-native-workflow.md; not a synonym for provider publication |
| forge | Provider publication and PR interaction; never confirmation of delivery merely from an open PR |
| adoption | The existing ticket-native adoption contract; retain original operation identity and authority |
| frontier / loose selection | Existing eligible ticket selection rules; define the exact selected record before renaming helpers |
| continuation / amendment | Existing retained-work transitions; preserve revision, budget and request identity |
| closing | Existing closing gate and receipt-backed terminal decision; validation is not integration |

The original specifications retain their paths and wire terms. Package names
use plain nouns; package qualification replaces redundant prefixes only after
the mechanical move. The state and package maps above are dependency
constraints, not permission to rename an ambiguous domain concept.

## Temporary bridges

None introduced in Slice 1. Any necessary bridge must identify its declaration,
consumer, owning removal task and validation before introduction. All bridges
expire by T15.4; permanent aliases or pass-through facades do not satisfy the
package split. Baseline debt reassociation requires explicit identity and
signature changes in the manifest, plus the original inventory reference.
