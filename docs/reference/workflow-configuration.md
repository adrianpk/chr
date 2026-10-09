# Workflow Configuration

Project workflow policy lives in `.chr/default/config.yml`. Operational namespaces
share this project policy; `ops/<namespace>/config.yml` is not a policy source.

## Version and effective values

The nested `workflow.version: 1` marker selects the version-one defaults below.
An absent marker retains legacy defaults. The older root `version` field and
artifact filenames do not select workflow defaults. Explicit project values
apply after defaults, independent of their order in the file.

```yaml
workflow:
  version: 1
review:
  mode: human_review
work:
  state_authority: tickets
  workspace_mode: worktree
  merge_strategy: rebase
  max_active_workers: 1
  ready_queue_capacity: 64
  worker_timeout_seconds: 3600
  validation_timeout_seconds: 3600
  checkpoint_step_limit: 128
  include_later_tickets: false
  max_worker_attempts: 3
  max_integration_attempts: 3
```

The explicit human-review mode in this example overrides the version-one
trusted-auto-merge default. Adding the marker to an existing project changes
absent defaults; preserve its effective review and authority choices explicitly
when making that policy decision. Reading configuration never writes the marker
or converts historical records.

| Key | Version-one default | Accepted values |
| --- | --- | --- |
| `workflow.version` | `1` | `1`; absence selects legacy defaults |
| `work.state_authority` | `tickets` | `tickets`, `tracker_compat` |
| `review.mode` | `trusted_auto_merge` | `trusted_auto_merge`, `human_review`, `agentic_review` |
| `work.workspace_mode` | `worktree` | `worktree`; `project` requires explicit serial `tracker_compat` work |
| `work.merge_strategy` | `rebase` | `rebase`, `merge`, `squash`; no fallback method |
| `work.max_active_workers` | `1` | 1–8 |
| `work.ready_queue_capacity` | `64` | 1–1024 |
| `work.worker_timeout_seconds` | `3600` | 1–86400 |
| `work.validation_timeout_seconds` | `3600` | 1–86400 |
| `work.checkpoint_step_limit` | `128` | 1–128 |
| `work.include_later_tickets` | `false` | `true`, `false` |
| `work.max_worker_attempts` | `3` | 1–128 |
| `work.max_integration_attempts` | `3` | 1–10 |

Legacy absence means human review, tracker compatibility and project workspace.
The numerical defaults otherwise match the table, except that absent legacy
validation timeout adds no deadline beyond the approved validation contract.
Effective reporting represents that legacy absence as `0`; an explicit timeout
of `0` is invalid.

Configuration is bounded to 1 MiB. The `workflow`, `work` and `review` sections
use unique scalar keys indented by two spaces. Duplicate sections or keys,
unknown keys, invalid bounds, tab indentation, nested mappings and conflicting
workspace policy block execution and publication instead of falling back.
Existing readiness and provider keys remain supported; see
[Work readiness](work-readiness.md) and
[Review workflow configuration](review-workflow-config.md).

## Observation and invocation restrictions

```sh
chr work config
chr work config --json
chr work config --json --review-restriction human_review
chr work slice --human-review slice-1
```

`work config` is read-only. It reports the thirteen policy keys, their effective
values and sources, plus a hash of the project configuration bytes. Sources are
`project`, `version_default`, `legacy_default` or `invocation_restriction`.
Provider credentials and host-private configuration are excluded.

The effective review precedence is a human invocation restriction, then an
explicit project value, then the selected version default. `--human-review`
on native extent admission retains that restriction for the invocation. The
configuration command can show its effect without admitting work. Invocation
restrictions cannot enable automatic review or relax a project review mode.
A conflicting restriction on an existing entry rejects reuse of that invocation.

## Frozen execution and current publication

Admission freezes execution limits, later-ticket membership policy and the
configuration revision. Retries and continuation use those retained values;
editing config does not refill attempts, extend worker deadlines or change the
retained checkpoint limit. An explicit project execution limit cannot be
expanded by an invocation.

Validation has one deadline across its ordered commands. The smaller approved
contract timeout and configured timeout wins. Task and slice inputs retain the
frozen timeout separately from the original contract; set closing inputs retain
the effective closing timeout and original contract as distinct evidence.

Publication reads the current project review policy and merge strategy again.
The invocation's human restriction remains in force. A changed config revision,
candidate, target or provider capability invalidates a previously displayed
approval binding. See [Workflow controls](workflow-controls.md) for admission,
status and approval commands.
