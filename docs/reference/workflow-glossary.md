# Workflow glossary

These names identify existing contracts. Internal package extraction preserves
public command flags, stored paths, JSON keys and enum spellings. The governing
contract is [ticket-native workflow](../../ops/default/spec/stable/ticket-native-workflow.md);
the [capability map](../explanation/capability-boundaries.md) locates the owners.

| Term | Exact meaning | Implementation owner |
| --- | --- | --- |
| Ticket-native work | Ticket-owned records, admission, execution and continuation. This names the workflow authority independently of the selected forge or worker. | [workflow](../../internal/core/workflow/README.md), [state](../../internal/core/workflow/state/README.md) |
| Forge-native publication | The explicitly selected `git_atomic_dual_ref` publication mechanism, bound to current provider policy, identity, transport capability and original source/base/prepared result. | [forge](../../internal/core/workflow/forge/README.md), guarded workflow host |
| Frontier | The confirmed prefix of the admitted order and its verified canonical delivery. Advancing releases the finalized unit and selects the next target while the original admission target remains frozen. Set boundaries also require confirmed closure. | [workflow_frontier.go](../../internal/core/workflow/workflow_frontier.go), `AdvanceWorkflowFrontier` |
| Adoption | Explicit quiescent binding of one approved legacy plan and exact selected tasks to native records. It checks current source, scope, identity and ownership. Reads and ordinary initialization do not adopt historical work. | [workflow_adoption.go](../../internal/core/workflow/workflow_adoption.go) |
| Loose selection | A queue of eligible independent tickets. Slice-owned tickets stay excluded. Fields combine with AND, lists with OR, and all requested tags must match. Ranking cannot reorder a slice. | [workflow_selection.go](../../internal/core/workflow/workflow_selection.go) |
| Continuation | Bounded accepted transitions on one retained invocation, reusing frozen selection, target, filters, policy and allowances. Zero requested steps selects the frozen bound; a barrier returns before another step starts. | [workflow_continuation.go](../../internal/core/workflow/workflow_continuation.go), `ContinueWorkflow` |
| Amendment | Explicit revision-guarded membership/order/contract change at a quiescent checkpoint. Integrated tickets, prior revision maps and closure receipts survive. Later-admission policy does not amend an existing slice. | [workflow_amendment.go](../../internal/core/workflow/workflow_amendment.go) |
| Closing | A separate decision after actual slices integrate. It freezes actual members, plan, validation and target, then validates an isolated checkout. Retry does not replay integrated tasks; waiver retains its reason and member proof. Host confirmation establishes closure. | [workflow_closing.go](../../internal/core/workflow/workflow_closing.go), [closeout](../../internal/core/workflow/closeout/README.md) |
| Submission intent | A journaled identity and exclusion barrier permitting one dispatch. Uncertain outcomes require observation before another send. | `SubmissionIntent` in [workflow/state](../../internal/core/workflow/state/README.md) |
| Canonical observation | Independently read PR and Git facts bound to exact source, target, method and history. Transport success, an open PR and worker success cannot substitute for receipt proof. | [forge](../../internal/core/workflow/forge/README.md), guarded host confirmation |

`Handled` and `Observed` are internal branch/barrier outcomes and preserve partial
errors. They do not claim completed delivery. The
[publication and review terms](../explanation/architecture.md#publication-and-review-terms)
define review approval, immutable review input and receipt-backed delivery.
