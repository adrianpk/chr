# Run Ticket-Native Workflows

## Select policy and work

Initialize a new project with `chr init`, then inspect `chr work config --json`.
Version one selects ticket authority, worktrees, trusted automatic integration
and one worker. Configure `review.mode: human_review` when human approval is
required. Existing projects retain legacy policy until an explicit selected
upgrade; follow [Workflow configuration](../reference/workflow-configuration.md)
to observe the revision, preserve prior review and adopt only approved work.

Materialize or deliberately adopt an approved native plan through the shared
service before selecting its set or slices; the explicit contract is documented
in [Plan workflow materialization](../reference/plan-workflow-materialization.md).
The legacy plan-tracking command without native input keeps its legacy behavior.
Historical Markdown and tracker files do not become executable
through naming or configuration alone. Preview the intended extent first:

```sh
chr work slice-set --preview --json <slice-set-id>
chr work slice-set --human-review --max-steps 1 <slice-set-id>
chr work status --json <execution-id>
```

The admitted invocation retains the selected records, target, policy restriction
and finite budgets. Continue that ID with `chr work continue --max-steps 1
<execution-id>`. Smaller step limits give bounded checkpoints; they do not reset
attempts or reselect the ready backlog. Task validation, slice integration and
set closing validation are distinct barriers. Inspect candidate, approval and
delivery references before treating a unit as delivered. See
[Workflow selection](../reference/workflow-selection.md) and
[Workflow execution](../reference/workflow-execution.md).

Use `pause`, `cancel` and `resume` on the retained ID. For dead ownership,
membership changes or uncertain publication, use the corresponding explicit
`reconcile` operation. A live owner or descendant blocks replacement. Human
review waits require `approve --binding <displayed-binding>` for the exact
current candidate. See [Workflow controls](../reference/workflow-controls.md)
for arguments and read-only previews.

## Run a manual harness

Admit the intended native invocation, then reconcile one task through a
supervised harness:

```sh
chr work reconcile --kind manual --harness ./manual-worker <execution-id>
```

The harness runs in the owned isolated lane. It receives the frozen public input
on stdin and in `CHR_WORKFLOW_INPUT`; the host retains private claim credentials.
It must edit, commit and leave a clean candidate, then print one JSON object with
the exact input-file SHA-256, full branch ref, commit and tree:

```json
{"input_hash":"<sha256>","branch":"refs/heads/<lane>","commit":"<full-commit>","tree":"<full-tree>"}
```

Repeat `--harness` for each argv element, preserving spaces and commas. The
report limit is 4096 bytes. The host checks actual Git state and the frozen gate
before accepting task progress. Harness success alone and direct ticket edits
cannot advance the cursor. Failed reports retain a checkpoint for correction;
the approved attempt budget remains consumed. Integration and set closing still
use their shared owners after manual task validation.

Core callers can use `ClaimWorkflow`, `PrepareWorkflowManualTask`,
`ExecuteWorkflowManualTask` and `ReconcileWorkflowTask` while retaining the same
claim. An empty preparation argv permits a human candidate without launching a
worker; the caller must retain ownership and reconcile before the lease expires.
The CLI requires a supervised harness and does not offer an unattended human
completion flag or import status edits. Ownership credentials are host-private.

## Inspect projections and available controls

Export an optional readable tracker with `chr work tracker export <slice-set-id>`.
After progress changes, inspect and repair it using
[Workflow tracker projections](../reference/workflow-trackers.md). Preserve
operator edits unless explicit replacement is authorized for their observed
hash. Work continues through native tickets and receipts when a projection fails.

CLI, TUI and Web deliver explicit native extent admission and retained controls.
The daemon continues retained authorization without admitting the ready backlog.
Their previews, checkpoints, approval and reconciliation use the same service.
Dedicated tracker controls are currently CLI/service operations. Telegram
notifications remain informational; remote intake records handoffs without
executing native lifecycle commands. Notification receipt is not execution or
integration evidence. Do not infer a Telegram execution adapter or additional
manual completion surfaces from the delivered protocol.
