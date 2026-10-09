# Slice 11: Defaults and authority transition

Status: reviewing

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-11-workflow-defaults`

PR: [#374](https://forge.adrianpk.com/adrianpk/chr/pulls/374)

## Purpose

Make new-project workflow defaults consistent, preserve existing effective policy
through deliberate adoption, and keep exported trackers repairable projections.

## Delivered Behavior

Task commit: `d081c3daa795065cf8c635c3270aa6d26a7dabe8`. Acta: `01M4H1KP8V3MM1RVM5KSDQRSQN`.

T11.1 resolves the nested workflow version marker and thirteen effective policy
keys with their sources. Missing markers retain legacy defaults; explicit values
override version defaults. Invalid bounds, unsupported values, duplicate keys
and conflicting workspace policy block execution and publication.

Read-only `chr work config` reports policy without creating or upgrading files.
Native CLI extent admission can retain a human-review restriction. Admission
freezes execution budgets, validation timeout and later-ticket policy; retries
and entry attachment preserve those inputs across configuration changes.
Publication rechecks current project review policy and merge strategy, retaining
any human restriction and invalidating changed approval bindings.

T11.2 task commit: `6a4b3f019fe3910b5c93662cb947ee4d2b7caf52`. Acta: `01M4H34ACREHJNR3NZBY5ZDHG5`.

T11.2 initializes every bundled profile with version-one defaults: ticket
authority, isolated worktrees, trusted automatic integration and one worker.
Dry-run and repeated initialization preserve existing configuration bytes,
including invalid policy, and preserve config absence in an existing workspace.
Runtime bootstrap leaves both new and legacy project policy unchanged.

T11.3 task commit: `7783c4e0ac08a64e27d680cc1861123f3f429244`. Acta: `01M4H59ZFMXVFDYYZ0RTC5TM3W`.

T11.3 adds explicit selected-project upgrade with an observed config revision,
required authority choice and optional new review choice. Absent and explicit
legacy review policy is preserved by default. A credential-free prior-policy
snapshot is durable before the version marker. Optional selected plan adoption
preserves keys, creation times, metadata and history without converting unrelated
records. Active native or legacy ownership and ambiguous selections block it.

T11.4 task commit: `408a622d1074edfbeefe9c54c2f13e85c41747f9`. Acta: `01M4H7DGCVJ662A7ATXFFG5NPC`.

T11.4 adds shared tracker export, read-only status and explicit repair. Generated
views retain source revisions/hashes, output hash, generation time and shared
delivery/closure receipts. Relevant committed domain writes mark registered
views stale. Failed export preserves native progress and operator edits; only
known generated images or exact observed-hash permission allow replacement.
Operation/configuration guidance and manual harness instructions distinguish
delivered CLI/TUI/Web/daemon controls from deferred execution surfaces.

## Implementation Notes

The project config loader and publication-policy resolver share strict bounded
parsing of the owned `.chr/default/config.yml`. Operational namespaces cannot
supply a competing policy file. Effective reporting exposes only approved values,
sources and a configuration hash, excluding provider credentials.

Frozen policy fields enter immutable task and slice inputs. Ordered validation
uses the smaller approved contract and configured deadline. Closing retains its
original approved contract separately from the effective gate deadline. Legacy
absence imposes no extra validation deadline. Entry lookup fills omitted limits
from the retained invocation and still rejects explicitly changed policy. Legacy
omitted integration limits retain their zero representation while the finite
default remains three. Review fixtures preserve unrelated project limits in the
shared policy file.

Initialization observes whether `.chr/` existed before creating directories.
Existing config absence is reported separately as preserved policy. Exclusive
file creation prevents initialization from replacing a file created by another
writer between inspection and creation. The bundled profiles rely on the
versioned resolver for their missing policy fields.

Upgrade acquires exclusive repository capacity and namespace operation locks
before replaying workflow journals. The project configuration owner keeps its
own bounded checksummed pending record and snapshot paths under `.chr/default`;
workflow record paths retain their original namespace boundary. Pending upgrades
block policy reads, publication and ordinary materialization/adoption until the
original explicit request resumes. Torn domain writes and lost responses replay
under the original selection; manual config or snapshot changes are preserved
for reconciliation.

Tracker pending/current provenance is separate from the domain transition
journal. Source and output observation detects staleness even when a marker
cannot persist. Restart recognizes the pending generated image after process
loss. Export/repair never imports edits, claims work or contacts a forge. Regular
file and namespace path guards, 1 MiB files and 128 registered-view bounds apply.

## Contracts Added or Changed

- `workflow.version: 1` selects prospective defaults; the old root version does
  not select them. Absent workflow markers keep legacy policy.
- Effective sources are `project`, `version_default`, `legacy_default` and
  `invocation_restriction`. The revision hashes the project configuration bytes.
- `chr work config [--json] [--review-restriction human_review]` observes policy.
  Native extent admission accepts `--human-review`; restrictions cannot enable
  automatic review or replace a retained invocation's restriction.
- Workers, queues, execution timeouts, checkpoint limits and attempt budgets
  enforce their approved finite bounds before admission.
- `chr init` selects version one only for a new workspace; existing config
  bytes and absent legacy configuration survive every profile and dry-run.
  `chr runtime init` installs assets without writing project policy.
- Frozen execution limits cannot be refilled by project edits; current
  publication policy remains a separate decision boundary.
- `chr work upgrade --config-revision <revision> --state-authority
  tickets|tracker_compat [--review-mode MODE] [--adopt-file PATH] [--json]`
  upgrades one quiescent legacy project and preserves review unless explicitly
changed. No read, initialization or upgrade converts historical work in bulk.
- `chr work tracker export|status|repair [--dir PATH] [--json] SET_ID` manages
  an optional view; export/repair accepts `--replace-hash` for explicit permission
  tied to current operator-edited bytes. `projection_stale` never rolls back a
  successful authoritative transition or becomes native execution authority.

## Files of Interest

- `internal/core/config_workflow.go` and `config_workflow_parse.go`: bounded
  version resolution, approved keys and effective source reporting.
- `internal/core/config_workflow_execution.go` and `workflow_entry.go`: frozen
  admission/retry limits and retained entry attachment.
- `internal/core/workflow_review_policy.go`: current owned project policy.
- `internal/app/work_workflow_config.go` and `service_workflow_controls.go`:
  shared configuration inspection and human restriction admission.
- `internal/core/workspace_init.go`, `internal/app/init.go` and bundled profile
  configs: prospective defaults and explicit preservation of absent policy.
- [Workflow configuration](../../../../../docs/reference/workflow-configuration.md)
  and the [User Guide](../../../../../docs/how-to/user-guide/index.md): delivered
  configuration, command, timeout and precedence guidance.
- `internal/core/workflow_upgrade*.go` and
  `internal/app/service_workflow_upgrade.go`: guarded selected transition,
  prior-policy snapshots and restart recovery shared across adapters.
- `internal/core/workflow_tracker*.go` and the shared service/CLI tracker
  adapters: guarded one-way export, provenance and stale-image repair.
- [Tracker projections](../../../../../docs/reference/workflow-trackers.md) and
  [Run ticket-native workflows](../../../../../docs/how-to/run-ticket-workflows.md):
  delivered operations, manual reconciliation and explicit adapter boundaries.

## Validation

T11.1 development checks passed:

- `make verify` (pre-commit repository gate; passed)
- `go test ./internal/core ./internal/app -run '^(TestWorkAuthority|TestWorkflowCapacityLater|TestWorkflowSubmissionPolicy|TestWorkflowTicketDelivery|TestWorkflowEntry.*|TestWorkflowConfig.*|TestWorkflowInvocation.*|TestServiceWorkNative|TestTicketReviewMode.*)$' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkAuthority|TestWorkflowCapacityLater|TestWorkflowSubmissionPolicy|TestWorkflowTicketDelivery|TestWorkflowEntry.*|TestWorkflowConfig.*|TestWorkflowInvocation.*|TestServiceWorkNative|TestTicketReviewMode.*)$' -count=1`
- `python3 /tmp/chr-slice-11-docs-check.py`
- `git diff --check`

The repository gate used Go 1.27.1, the retained curation fixture on `PATH`,
and scoped uncached verbose tests with a 30-minute timeout. Its successful log
is `/tmp/chr-slice-11-t11-1-verify-final.log`. Canonical Slice 11 candidate validation is recorded by the configured
controller against the exact clean PR head after PR references are committed.

T11.2 development checks passed:

- `make verify` (pre-commit repository gate; passed)
- `go test ./internal/core ./internal/app -run '^(TestInit|TestRuntimeInit|TestWorkflowConfigDefaults)' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestInit|TestRuntimeInit|TestWorkflowConfigDefaults)' -count=1`
- `python3 /tmp/chr-slice-11-docs-check.py`
- `git diff --check`

The T11.2 repository gate used the same Go 1.27.1 toolchain, retained curation
fixture and scoped uncached 30-minute timeout. Its successful log is
`/tmp/chr-slice-11-t11-2-verify.log`. The actual bundled profiles, root CLI,
effective-value/source observation, runtime initialization and absent/explicit/
invalid existing policies passed the new/existing/reinitialization matrix.
Initialization guidance passed 27 local documentation-link checks.

T11.3 development checks passed:

- `make verify` (pre-commit repository gate; passed)
- `go test ./internal/core ./internal/app -run '^(TestWorkflowUpgrade|TestWorkflowPlanAdoption|TestWorkflowConfig|TestInit|TestRuntimeInit)' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowUpgrade|TestWorkflowPlanAdoption|TestWorkflowConfig|TestInit|TestRuntimeInit)' -count=1`
- `python3 /tmp/chr-slice-11-docs-check.py`
- `git diff --check`

The successful repository log is `/tmp/chr-slice-11-t11-3-verify-final.log`,
using Go 1.27.1, the retained curation fixture and scoped uncached 30-minute
tests. Strict lint reported zero issues and every package passed. Initial lint
spacing findings were corrected before the full successful gate. The focused
matrix covers absent/explicit review preservation, deliberate review changes,
authority choices, unselected history, stale and ambiguous adoption, live native
capacity and namespace locks, retained invocations, actual legacy execution
rows, torn domain writes, all policy write boundaries, manual-edit preservation,
symlink/size rejection and a real owned subprocess crash/restart. Actual CLI
upgrade/config/retry commands and 27 guidance links passed. Final slice-head
controller evidence is stored separately from these development checks.

Real Git and owned subprocess checks cover readonly CLI observation, strict
bounds, configured admission, unchanged retry budgets, target-preserving entry
attachment, human restrictions and task/slice/closing deadline failures.
Fixtures stop through their owning validation supervisors.

T11.4 development checks passed:

- `make verify` (pre-commit repository gate; passed)
- `go test ./internal/core ./internal/app -run '^(TestWorkflowTracker|TestWorkflowUpgrade|TestWorkflowConfig|TestInit|TestRuntimeInit)' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowTracker|TestWorkflowUpgrade|TestWorkflowConfig|TestInit|TestRuntimeInit)' -count=1`
- `python3 /tmp/chr-slice-11-docs-check.py`
- `git diff --check`

The successful repository log is `/tmp/chr-slice-11-t11-4-verify-final.log`,
using Go 1.27.1, the retained curation fixture and scoped uncached 30-minute
tests. Strict lint reported zero issues; core passed in 846.591 seconds and app
in 92.696 seconds. Initial complexity/error-checking findings and a test path
assertion were corrected before the successful final development checks. The
new tracker cases cover actual CLI/service agreement, read-only status,
source/output changes, edited-file permission, torn exports, a real owned child
crash/restart, shared finalized receipt provenance and continued native progress
through derived failures. Five guidance documents passed 41 local link checks;
12 operational documents passed 102 local link checks.

Canonical PR #374 was opened and recorded in Acta `01M4H7GG3DTK7PYP382R6RD9H3`. The configured
controller records the exact clean PR head and ordered candidate commands:
`make verify`, `go test ./internal/core ./internal/app -count=1`,
`go test -race ./internal/core ./internal/app -count=1`, then `git diff --check`.
The controller's checks ledger and canonical PR validation list carry candidate
results separately from the local development evidence above.

Candidate attempt `181dd85835a7f55224e1d4ea42ddb2d39e402d6c` passed controller
`make verify` and `go test ./internal/core ./internal/app -count=1`. The exact
`go test -race ./internal/core ./internal/app -count=1` command failed because
the core suite exhausted the scoped 30-minute Go runner deadline near its end;
`TestWorkflowTrackerReceipt` had been running for one second. No test-case
failure or data-race report appeared; app race passed in 290.770 seconds.
Diagnostics are retained in the controller's checks ledger directory under
`181dd85835a7f55224e1d4ea42ddb2d39e402d6c.log`. Failure diagnosis Acta:
`01M4HAYM69MX5HQWHXSKN77MQ5`. No receipt fixture process remained afterward.
The retry uses a scoped 60-minute Go runner timeout with the same ordered
commands; workflow execution policy and source code are unchanged.

## Risks and Follow-ups

All four tasks are validated locally. Integration through canonical PR #374
is pending. Exact clean-head candidate checks are owned by the controller.
