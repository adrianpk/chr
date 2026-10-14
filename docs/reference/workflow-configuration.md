# Workflow Configuration

Project workflow policy lives in `.chr/default/config.yml`. Operational namespaces
share this project policy; `ops/<namespace>/config.yml` is not a policy source.

## Version and effective values

The nested `workflow.version: 1` marker selects the version-one defaults below.
An absent marker retains legacy defaults. The older root `version` field and
artifact filenames do not select workflow defaults. Explicit project values
apply after defaults, independent of their order in the file.

`chr init` writes this marker for new workspaces in the default, standard,
minimal and personal profiles. The profiles omit the policy keys that use the
version defaults. Runtime bootstrap loads the same project policy and does not
rewrite it. Reinitialization preserves existing config bytes and leaves a missing
config absent when `.chr/` already exists; it does not upgrade that workspace.

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
  integration:
    publication_profile: disabled
    max_prepared_per_target: 2
```

The explicit human-review mode in this example overrides the version-one
trusted-auto-merge default. Use the explicit upgrade operation below for an
existing project so its prior policy is recorded before the marker is written.
Reading configuration never writes the marker or converts historical records.
Native publication remains `disabled` unless the project explicitly selects
`work.integration.publication_profile: git_atomic_dual_ref`. Version-one defaults
and trusted review do not authorize source rewriting. Follow
[Configure native integration](../how-to/configure-native-integration.md) before
admitting work; [Native integration](native-integration.md) lists the executable
profile, policy support and service-only token/attempt options.

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
| `work.integration.publication_profile` | `disabled` | `disabled`, explicit `git_atomic_dual_ref`; independent of version and review mode |
| `work.integration.max_prepared_per_target` | `2` | 1–8, bounded by ready queue capacity |

Legacy absence means human review, tracker compatibility and project workspace.
The numerical defaults otherwise match the table, except that absent legacy
validation timeout adds no deadline beyond the approved validation contract.
Effective reporting represents that legacy absence as `0`; an explicit timeout
of `0` is invalid.

Configuration is bounded to 1 MiB. The `workflow`, `work` and `review` sections
use unique scalar keys indented by two spaces. The supported `work.integration`
mapping has unique scalar keys indented by four spaces. Duplicate sections or
keys, unknown keys, invalid bounds, tab indentation, other nested mappings and
conflicting workspace policy block execution and publication instead of falling
back.
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

`work config` is read-only. It reports the fifteen policy keys, their effective
values and sources, plus a hash of the project configuration bytes. Sources are
`project`, `version_default`, `legacy_default` or `invocation_restriction`.
Provider credentials and host-private configuration are excluded.

The effective review precedence is a human invocation restriction, then an
explicit project value, then the selected version default. `--human-review`
on native extent admission retains that restriction for the invocation. The
configuration command can show its effect without admitting work. Invocation
restrictions cannot enable automatic review or relax a project review mode.
A conflicting restriction on an existing entry rejects reuse of that invocation.

## Selected legacy project upgrade

Observe the selected project's revision, then supply that exact hash and an
explicit authority choice:

```sh
chr work config --dir /path/to/project --json
chr work upgrade --dir /path/to/project --config-revision <revision> --state-authority tickets --json
```

`--state-authority` accepts `tickets` or `tracker_compat` and is required. Ticket
authority selects isolated worktrees. Tracker compatibility retains the previous
workspace mode. Other explicit configuration and unrelated sections remain in
place. Missing version-one execution keys resolve to their version defaults.

Omitting `--review-mode` preserves the prior effective review mode, including
legacy absent human review and explicit agentic or trusted modes. An explicit
new `--review-mode human_review|agentic_review|trusted_auto_merge` is required
to change that choice. The version marker cannot silently enable automatic
integration during this operation. Upgrade preserves an explicit publication
profile and prepared-candidate limit; absent native settings retain their disabled
publication default and finite preparation default. It never adds source-rewrite
authority.

The credential-free snapshot at
`.chr/default/workflow-upgrades/<request-hash>.json` records prior effective
values, their sources and the observed revision before publishing the marker.
The operation requires unchanged legacy policy, repository-wide quiescence,
exclusive worker capacity and legacy operation locks across linked worktrees
and namespaces. Unfinished native invocations or pending/running legacy
executions block it. Existing version-one projects are not reconfigured by
upgrade; an exact completed request can be retried without a new snapshot.

By default, upgrade changes only configuration and leaves legacy records intact.
To adopt selected active work in the same transition, supply `--adopt-file` with
one bounded JSON `WorkflowPlanAdoptionOptions` envelope. Its `Plan` is the full
approved generated plan contract, `PlanPath` is its namespace-relative metadata
path, `SourceHash` binds the observed metadata and body, and `Tasks` is the exact
map of `task_id`, namespace-relative `path` and `source_hash` for every approved
task. This requires ticket authority. Unknown JSON fields, multiple objects,
stale images, duplicate identities, ambiguous ownership and active attempts
block adoption before the pending policy transition is written. Original keys,
creation times, compatible metadata, bodies and history are preserved; unrelated
historical tickets are not converted.

An interrupted upgrade retains a checksummed pending record at
`.chr/default/tmp/workflow-upgrade.json`. Config inspection, execution policy,
publication, ordinary materialization and ordinary adoption report the pending
transition. Retry the original command with the original observed revision,
choices and selection envelope to resume it. Changed config, snapshots or a
foreign request block recovery and preserve the conflicting images for explicit
reconciliation. Reading a project never finishes an upgrade automatically.

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
