# Slice 11: Defaults and authority transition

Status: drafting

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-11-workflow-defaults`

PR: pending

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
- Frozen execution limits cannot be refilled by project edits; current
  publication policy remains a separate decision boundary.

## Files of Interest

- `internal/core/config_workflow.go` and `config_workflow_parse.go`: bounded
  version resolution, approved keys and effective source reporting.
- `internal/core/config_workflow_execution.go` and `workflow_entry.go`: frozen
  admission/retry limits and retained entry attachment.
- `internal/core/workflow_review_policy.go`: current owned project policy.
- `internal/app/work_workflow_config.go` and `service_workflow_controls.go`:
  shared configuration inspection and human restriction admission.
- [Workflow configuration](../../../../../docs/reference/workflow-configuration.md)
  and the [User Guide](../../../../../docs/how-to/user-guide/index.md): delivered
  configuration, command, timeout and precedence guidance.

## Validation

T11.1 development checks passed:

- `make verify` (pre-commit repository gate; passed)
- `go test ./internal/core ./internal/app -run '^(TestWorkAuthority|TestWorkflowCapacityLater|TestWorkflowSubmissionPolicy|TestWorkflowTicketDelivery|TestWorkflowEntry.*|TestWorkflowConfig.*|TestWorkflowInvocation.*|TestServiceWorkNative|TestTicketReviewMode.*)$' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkAuthority|TestWorkflowCapacityLater|TestWorkflowSubmissionPolicy|TestWorkflowTicketDelivery|TestWorkflowEntry.*|TestWorkflowConfig.*|TestWorkflowInvocation.*|TestServiceWorkNative|TestTicketReviewMode.*)$' -count=1`
- `python3 /tmp/chr-slice-11-docs-check.py`
- `git diff --check`

The repository gate used Go 1.27.1, the retained curation fixture on `PATH`,
and scoped uncached verbose tests with a 30-minute timeout. Its successful log
is `/tmp/chr-slice-11-t11-1-verify-final.log`. Final Slice 11 candidate validation
remains pending until all four tasks and PR references are complete.

Real Git and owned subprocess checks cover readonly CLI observation, strict
bounds, configured admission, unchanged retry budgets, target-preserving entry
attachment, human restrictions and task/slice/closing deadline failures.
Fixtures stop through their owning validation supervisors.

## Risks and Follow-ups

T11.2 is active for new-project initialization. T11.3–T11.4 remain ready for
explicit selected upgrade
with prior-policy snapshots, and tracker export/repair with complete guidance.
