# Isolated Worktree Reconciliation

Status: documentation reconciled; legacy proving-cycle QA remains incomplete

Tracker: [Isolated Ticket Worktrees](../../trackers/2026-06-19-isolated-ticket-worktrees-tracker.md)

## Delivered Behavior and Default Decision

The original cycle delivered Slices 1–7 through PRs #328–#334. Its Slice 8
still lists seven pending documentation and QA tasks. Later ticket-native
delivery independently completed prospective worktree defaults in
[PR #374](https://forge.adrianpk.com/adrianpk/chr/pulls/374), merged as
`750378ec9b7ee05399395184def0954ab47cf1d7`. The canonical API confirms the
merge, and that commit is an ancestor of the inspected `dev` candidate.

`internal/core/config_workflow.go` and `config_workflow_test.go` establish:

- New workspace initialization selects nested `workflow.version: 1`, whose
  default execution mode is `worktree`.
- Missing markers retain legacy `project` mode. An old root `version: 1`
  does not opt into the new defaults.
- Explicit workspace policy overrides version defaults. Reinitialization
  preserves existing config and absent legacy config; deliberate adoption
  belongs to the guarded workflow upgrade operation.
- The canonical worktree root is `ops/<namespace>/worktree`; metadata is
  advisory under its `.metadata/` directory. Git remains authoritative.
- Work readiness reports dirty user state as informational in worktree mode,
  checks the selected base separately, and blocks dirty or unsafe execution
  worktrees. Runtime execution mounts the selected ticket worktree.

This resolves the old default-switch decision without changing runtime policy
or silently upgrading a project. Older draft proposals using
`.chr/default/runtime/worktrees` do not describe the accepted canonical root.

## Documentation Corrections

The work-readiness, runtime-config and project-workspace references still
presented `project` as an unconditional default. They now distinguish legacy,
version-one and explicit policy. Work readiness also distinguishes user checkout
state from ticket workspace and remote-base checks. Ticket execution, Dash/TUI
guidance and the User Guide now explain where the agent writes and how to
inspect the selected workspace with `chr work doctor --ticket <ticket-key>`.

## Focused Verification

On source commit `a66b9e5bebf9`, this non-short selection passed:

```sh
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -json ./internal/core -run '^Test(NativePreparationIsolation|WorkflowConfigDefaults|CheckWorkReadinessReportsDirtyUserCheckoutInWorktreeMode|CheckWorkReadinessBlocksDirtyTicketWorktree|WorkSelectionPreflightWorktreeModeAllowsDirtyUserCheckout)$' -count=1 -timeout=3m
```

The package completed in 0.587 seconds. The preparation isolation case uses
real temporary Git repositories and exercises merge, rebase and squash. It
compares staged, unstaged and untracked content, checkout metadata and retained
evidence before/after; the owned validation lane contains a clean exact result
and canonical refs remain unchanged. Other cases cover version precedence and
informational user dirtiness versus blocking ticket dirtiness.

Raw local events are retained temporarily at
`/tmp/chr-worktree-reconciliation-20261020.jsonl`. This focused evidence is not
a Docker worker end-to-end journey or complete cycle certification. The prior
development gate passed on the unchanged production/test source; this change
adds documentation only, validated through diff and local-link checks.

## Remaining Legacy Acceptance

| Task | Disposition |
| --- | --- |
| T8.1: runtime/readiness/execution/User Guide docs | Addressed by this documentation correction. |
| T8.2: Dash/TUI workspace visibility | Addressed by current source inspection and the TUI/readiness guidance. |
| T8.3: resettable isolated-mode QA reference | Still pending. Existing `qa/ref` fixtures contain no explicit workspace-mode scenario; general reset support is not this fixture. |
| T8.4: worktree QA smoke command | Still pending. Existing smoke commands do not prove a worktree-specific scenario. |
| T8.5: dirty user checkout plus real worker journey | Local Git preparation isolation is proven above; an actual Docker worker journey and its result/application evidence remain pending. Current session lacks Docker socket access. |
| T8.6: default-switch decision | Superseded by delivered PR #374 and verified version/legacy precedence; no new default change needed. |
| T8.7: Slice 8 and final cycle reports | Still pending full QA. This reconciliation report is not a delivered Slice 8 or final cycle report. |

The original tracker remains open with these acceptance items explicit. No
worktree was removed, reset or migrated, and no old report is relabeled as
having run a Docker journey or passed the current nightly.
