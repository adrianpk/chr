# Workflow Controls

`chr work` delegates ticket-native selection and lifecycle operations to the
shared workflow service. Explicit extents require `work.state_authority: tickets`
in `.chr/default/config.yml`. The configured integration branch and canonical
remote identify the target; the remote must have a public HTTPS repository URL.
The compatibility default remains `tracker_compat` until explicitly changed.

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

Execution requires offset zero. `--max-steps` accepts 0–128; zero uses the
retained continuation limit. The requested step bound cannot exceed that frozen
limit. An execution prints its retained ID and progress. `work loose` supports
`--status`, `--priority`, `--severity`, `--project-area`, `--tag`, `--type`,
`--use-case`, `--order`, `--min-priority` and `--min-severity`; hierarchy extents
accept no loose filters. Selection rules follow [Workflow selection](workflow-selection.md).

## Retained execution controls

| Command | Behavior |
| --- | --- |
| `chr work status execution-id` | Observe identity, revision, members, task evidence, publication policy, waits and delivery references. |
| `chr work continue --max-steps 1 execution-id` | Continue bounded steps over the frozen selection and retained budgets. |
| `chr work pause execution-id` | Retain stop intent and checkpoint a running owned group. |
| `chr work cancel execution-id` | Retain terminal cancellation after owned work stops. |
| `chr work resume execution-id` | Restore a completed pause and its prior phase without resetting budgets. |
| `chr work approve --binding <hash> execution-id` | Grant the exact displayed current review binding. |
| `chr work reconcile --kind ownership execution-id` | Recover only quiescent retained ownership through the common domain owner. |
| `chr work reconcile --kind membership execution-id` | Reconcile allowed member changes within the retained extent. |
| `chr work reconcile --kind submission execution-id` | Observe an uncertain retained publication without resubmitting it. |

All controls accept `--preview` and `--json`. A control preview observes retained
inputs without applying the action. `status` pages accept `--offset`, `--limit`
and `--revision`; a changed invocation revision rejects the next page. A pending
transition journal blocks observation until explicit reconciliation, rather than
being replayed by a preview or status read.

Status separates validated tasks, confirmed tickets and integrated units.
Candidate, PR, approval and delivery references identify their distinct evidence.
Recent attempts are capped at 128. Private claim tokens and owner process
identities are excluded. An approval rechecks the exact candidate, current PR
head, canonical target and effective review policy. Status or approval alone
cannot certify integrated delivery.

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
