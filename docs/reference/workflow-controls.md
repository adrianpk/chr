# Workflow Controls

`chr work` delegates ticket-native selection and lifecycle operations to the
shared workflow service. Explicit extents require `work.state_authority: tickets`
in `.chr/default/config.yml`. The configured integration branch and canonical
remote identify the target; the remote must have a public HTTPS repository URL.
Projects without `workflow.version` retain `tracker_compat`; version one selects
`tickets`. See [Workflow configuration](workflow-configuration.md) for policy
sources and `chr work config`.

For the operating sequence and manual harness protocol, see
[Run ticket-native workflows](../how-to/run-ticket-workflows.md). Optional
`chr work tracker export|status|repair` commands manage
[tracker projections](workflow-trackers.md) without execution authority.

## Explicit extents

| Command | Selection |
| --- | --- |
| `chr work ticket CHR-1` | One exact adopted ticket. |
| `chr work slice slice-1` | One set-owned slice in task order. |
| `chr work slice-set 00001-delivery` | One exact ordered set. |
| `chr work use-case delivery` | Ordered sets associated with one use case. |
| `chr work loose` | Eligible independent tickets with the supplied filters. |

Add `--preview` or its alias `--dry-run` to observe selection without admission,
claims, Git fetching or execution. `--json` emits the bounded selection, target
and effective publication policy. Preview pages accept `--offset` (0–4096),
`--limit` (1–128), and the exact `--snapshot` hash from the previous page.
Changed records, evidence or remote target reject stale pagination.

```sh
chr work slice-set --preview --json --limit 2 00001-delivery
chr work slice-set --preview --offset 2 --limit 2 --snapshot <hash> 00001-delivery
chr work slice --max-steps 1 slice-1
chr work loose --preview --priority high --tag backend --order priority_severity
```

Add `--human-review` when admitting an extent to retain a human approval
restriction. It cannot enable automatic review or change the restriction of an
existing invocation.

Execution requires offset zero. `--max-steps` accepts 0–128; zero uses the
retained continuation limit. The requested step bound cannot exceed that frozen
limit. An execution prints its retained ID and progress. `work loose` supports
`--status`, `--priority`, `--severity`, `--project-area`, `--tag`, `--type`,
`--use-case`, `--order`, `--min-priority` and `--min-severity`; hierarchy extents
accept no loose filters. Selection rules follow [Workflow selection](workflow-selection.md).

## Retained execution controls

| Command | Behavior |
| --- | --- |
| `chr work status execution-id` | Observe identity, revision, members, task evidence, integration queue/phase, actual role occupancy, budgets, questions, publication state and delivery references. |
| `chr work continue --max-steps 1 execution-id` | Continue bounded steps over the frozen selection and retained budgets. |
| `chr work pause execution-id` | Retain stop intent and checkpoint a running owned group. |
| `chr work cancel execution-id` | Retain terminal cancellation after owned work stops. |
| `chr work resume execution-id` | Restore a completed pause and its prior phase without resetting budgets. |
| `chr work approve --binding <hash> execution-id` | Grant the exact displayed current review binding; native integration also requires its displayed guards below. |
| `chr work grant execution-id` | Extend one retained allowance with a positive, attributable grant and displayed guards. |
| `chr work answer-decision execution-id` | Resolve or defer one retained question with its displayed revision, hash, choice and authority. |
| `chr work reconcile --kind ownership execution-id` | Recover only quiescent retained ownership through the common domain owner. |
| `chr work reconcile --kind membership execution-id` | Reconcile allowed member changes within the retained extent. |
| `chr work reconcile --kind submission execution-id` | Observe an uncertain retained publication without resubmitting it. |

All controls accept `--preview` and `--json`. A control preview observes retained
inputs without applying the action. `status` pages accept `--offset`, `--limit`
and `--revision`; a changed invocation revision rejects the next page. A pending
transition journal blocks observation until explicit reconciliation, rather than
being replayed by a preview or status read.

Status separates validated tasks, confirmed tickets and integrated units.
Native integration includes target queue position, source/base/result identities,
actual occupied roles, remaining attempts and token accounting. Serial capacity
shows one possible active role; parallel capacity keeps each retained lane and
target queue distinct. Unmetered tokens are reported as unmetered. A retained
`reviewing` queue phase does not become publication success through approval.
Candidate, PR, approval and delivery references identify their distinct evidence.
Recent attempts are capped at 128. Private claim tokens and owner process
identities are excluded. An approval rechecks the exact candidate, current PR
head, canonical target and effective review policy. Status or approval alone
cannot certify integrated delivery.

## Native budget grants and decisions

Start with `chr work status --json execution-id`. Copy its displayed
invocation revision and integration entry ID/revision, budget revision and
candidate hash unchanged. The entry and candidate may be absent before
integration preparation: omit those absent strings and use entry revision zero.
The positive invocation and budget revisions remain required.

Native approval requires those guards in addition to its exact review binding:

```sh
chr work approve --binding <review-binding> --revision <invocation-revision> \
  --entry-id <entry-id> --entry-revision <entry-revision> \
  --budget-revision <budget-revision> --candidate-hash <candidate-hash> execution-id
```

Legacy approval without native accounting keeps its existing binding-only
behavior. Stale native guards reject and return current status; refreshing status
does not approve a changed candidate.

An explicit positive grant extends one existing allowance:

```sh
chr work grant --revision <invocation-revision> \
  --entry-id <entry-id> --entry-revision <entry-revision> \
  --budget-revision <budget-revision> --candidate-hash <candidate-hash> \
  --record-id operator-extension-1 --kind merger --amount 1 \
  --actor operator --reason 'Authorize one additional conflict repair' execution-id
```

Kinds are `executor`, `reviewer`, `closing`, `merger`, `drift`, `submission`
and `tokens`. Amount is a positive signed 64-bit integer; identity, actor and
reason are 1–128 bytes. The domain owner enforces its retained bounds and
rejects overflow. Reusing a grant ID cannot add it twice. Consumption and
admission limits remain intact. A token grant requires an already finite ceiling.
A grant does not run an agent, resolve a decision or approve publication.

For a waiting question, also copy its revision/hash, exact choice and authority:

```sh
chr work answer-decision --revision <invocation-revision> \
  --entry-id <entry-id> --entry-revision <entry-revision> \
  --budget-revision <budget-revision> --candidate-hash <candidate-hash> \
  --record-id operator-answer-1 --decision-revision <decision-revision> \
  --decision-hash <decision-hash> --decision-id <question-id> \
  --choice <displayed-option> --answer-action resolve --authority <displayed-authority> \
  --actor operator --reason 'Resolve the displayed requirement' execution-id
```

Use `--answer-action defer` to retain the wait. A replay-ownership question may
require `--original-commit <original-source-commit>`. An explicitly selected
integration finding uses `--fix-category`, `--fix-title` and `--fix-detail`;
copy the supported category from the question or Web form. Text payloads are
bounded to 8192 bytes. The shared owner validates question meaning and
authority. An answer neither spends an agent attempt nor grants review or
publication.

Manual service callers use `ControlWorkExtent` with `WorkflowControlRequest`
actions `grant` and `answer_decision`, its public `Integration` guard, and an
explicit `BudgetGrant` or `DecisionAnswer`. They supply the displayed
`ExpectedRevision`; the host derives private claim fences. Operator inputs
cannot change scope or provide ownership credentials.

## Manual reconciliation

```sh
chr work reconcile --kind manual --harness ./manual-worker execution-id
chr work reconcile --kind manual --harness sh --harness -c --harness '<script>' execution-id
```

Repeat `--harness` once per argv element. Arguments, including commas and spaces,
remain verbatim. Supply 1–64 arguments within 4096 bytes. The shared host claims
one task, freezes its public inputs and supervises the harness in its isolated
lane. Inputs arrive on stdin and through the file named by `CHR_WORKFLOW_INPUT`.
The host retains private ownership credentials.

The harness must print one JSON object within 4096 bytes containing only
`input_hash`, `branch`, `commit` and `tree`. Compute `input_hash` from the exact
input file bytes; report the full Git branch ref, commit and tree. Harness exit
success does not validate a task. The host checks those identities against the
actual clean candidate and runs the frozen validation gate before accepting
progress. Malformed or mismatched reports retain a checkpoint without task
success. An unaccepted lane change remains fenced for explicit correction.
`--preview` validates argv and reports status without launching the harness.
CLI, TUI, Web and `daemon workflow status` expose the same accepted manual
checkpoint, revision and evidence. Status reads cannot advance task or integration
progress.

## TUI controls

With tickets authority enabled, the command palette offers native work for the
selected ticket, its slice, its slice set, the selected use case, or independent
tickets. In the Cases view, use-case work uses the selected case slug. In ticket
views, hierarchy commands resolve the selected native ticket's owning IDs.
Independent work uses the Dash priority filter and order and excludes grouped
work. The existing selected-ticket and ready-work shortcuts use these native
controls under tickets authority.

Press Enter to load the read-only preview. It shows the exact extent, selected
ticket revision, snapshot, target, policy, exclusions and ordered members. Press
Enter again to confirm that displayed input; Esc cancels the confirmation.
Admission checks the snapshot and selected revision under repository exclusion.
A changed record or competing owner produces a visible failure.

The Dash shows retained authorization in pages of 16 invocations, with eight
member and recent evidence rows per detail view. Use `Inspect retained workflow
status`, `Select next retained workflow`, and `Load next retained workflow page`
to choose an invocation. The next page command returns to the first page after
the last page. Member details include a bounded CLI pagination command when more
rows remain.

The selected invocation supports continue, pause, cancel, resume, exact-binding
approval and ownership reconciliation. Each mutating control first displays a
read-only confirmation and then sends the displayed invocation ID and revision
unchanged. Continue accepts at most one shared checkpoint per activation.
Dashboard refresh observes current progress; an older asynchronous read cannot
replace a newer control revision. Native details use the same integration,
queue, role, budget and question observations as CLI status. Budget grants and
question answers are available through work/daemon commands, Web forms and the
manual service; the TUI retains its existing lifecycle controls. Status and stop
controls remain available
while a native execution action runs. A second execution action is rejected
locally rather than queued.

The display separates host-validated tasks, confirmed tickets and integrated
units. It includes target, policy source, frozen limits, selected task/slice,
capacity resume phase, blocker and next-action owner, candidate/PR/approval
bindings, continuation evidence and delivery receipts. A rejected control keeps
current status visible alongside its failure. Inspect again to obtain a fresh
binding before retrying. Observation rejects pending journals across linked
worktrees and namespaces without recovery writes.

## Web controls

With tickets authority enabled, the Board lists native workflow tickets and their
current phases. Open a ticket to preview its exact ticket, owning slice, slice set
or use case. Case details offer a use-case preview using that case's identity.
The Dash links to `/workflows`, which lists retained invocations in pages of 16
and offers loose-work filters for project area, priority, severity and order.

A preview is read-only. It displays selection, exclusions, ordered members,
target, policy source, the selected ticket revision and a complete extent
snapshot. **Confirm and start one checkpoint** sends that binding unchanged.
Changing a selected record or admitting a competing owner rejects confirmation.
Native ticket details show workflow records; their status is changed through
workflow operations. The existing legacy ticket editor remains available for
legacy tickets.

Retained status displays eight members and eight recent evidence rows. Member
pagination binds the displayed invocation revision; reload after a revision
conflict. Visible status pages refresh while the document is visible. Admission
previews and write results retain their displayed bindings until the user acts
or explicitly refreshes them. Status remains observable if configuration changes
to tracker compatibility; new native admission still requires tickets authority.

The invocation view offers confirmed continue, pause, cancel, resume and ownership
reconciliation controls, plus exact-binding approval during a human review wait.
Each action sends the displayed invocation revision. Native approval forms also
retain the displayed entry, budget and candidate guards. Native allowance and
waiting-question forms submit those same guards plus an explicit actor/reason;
question forms retain the question revision/hash, option and authority. Token
extension appears only for an already finite ceiling. A rejected or stale form
renders current status without granting the refreshed candidate. Start and continue execute
at most one shared checkpoint; the Web adapter has no scheduler or execution
queue. An accepted action returns its retained status. A rejected action renders
a visible failure with current status when the invocation is known. Host task
validation, confirmed integration and remaining work remain separate counts.

`GET /workflows` and `/workflows/fragment` are read-only.
`POST /workflows/actions` requires explicit confirmation, a frozen admission
binding or invocation revision, and the existing origin/token access policy.
Forms are limited to 32 KiB with no duplicate fields; admission bindings are
limited to 8 KiB and reject unknown JSON fields. Expected failures render as
HTTP 200 fragments for htmx so that the browser can show them.

## Daemon and notification boundary

`chr daemon start` continues retained ticket-native authorization in enabled
registered projects. It never admits the ready backlog. Each project iteration
attempts at most 16 runnable invocations and one shared checkpoint per invocation.
Live owners, operator stops, blockers and human review waits retain their meaning;
dead-owner recovery uses shared process exclusion without resetting budgets.
See [Daemon coordination](daemon-coordination.md) for foreground lifecycle and
coordinator stop behavior.

`chr daemon workflow` reuses the retained `status`, `continue`, `pause`, `cancel`,
`resume`, `approve`, `reconcile`, `grant` and `answer-decision` controls above,
with the same displayed guards. For example:

```sh
chr daemon workflow status --dir /path/to/project --json execution-id
chr daemon workflow continue --dir /path/to/project --max-steps 1 execution-id
chr daemon workflow pause --dir /path/to/project execution-id
```

Existing Telegram notifications are informational. Remote command intake stores
bounded handoff records; it does not execute ticket-native admission,
continuation, stop, approval or reconciliation. Use the delivered CLI, TUI, Web or
daemon controls for those actions. A notification does not certify execution or
integration evidence.
