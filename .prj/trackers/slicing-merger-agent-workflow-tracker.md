# Merger Agent Workflow Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
Spec: `.prj/specs/specs/merger-agent-workflow.md`
Dump: `.prj/dumps/curated/2026-06-02-merger-agent-role.md`
Mirror: `.chc/user/trackers/merger-agent-workflow-tracker.md`

## Slicing Checklist

- [x] Capture merger role raw and curated dumps.
- [x] Create accepted spec.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`. Commit: `8f8171d`
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with the appropriate repository gate.

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | merger handoff contract | delivered | `feat/slice-1-merger-handoff-contract` | `#283` | `.prj/reports/slices/merger-agent-workflow/slice-1-merger-handoff-contract.md` |
| Slice 2 | merger preflight | planned | `feat/slice-2-merger-preflight` | pending | `.prj/reports/slices/merger-agent-workflow/slice-2-merger-preflight.md` |
| Slice 3 | merge execution command | planned | `feat/slice-3-merge-execution-command` | pending | `.prj/reports/slices/merger-agent-workflow/slice-3-merge-execution-command.md` |
| Slice 4 | merger workflow docs | planned | `feat/slice-4-merger-workflow-docs` | pending | `.prj/reports/slices/merger-agent-workflow/slice-4-merger-workflow-docs.md` |

## Slice 1: Merger Handoff Contract

Pre-flight:

- Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
- Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
- Slicing: `Merger Agent Workflow`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-merger-handoff-contract`
- Expected PR title: `feat(slice-1): merger handoff contract`
- Expected report: `.prj/reports/slices/merger-agent-workflow/slice-1-merger-handoff-contract.md`

Tasks:

- [x] T1.1 Add or normalize ticket metadata for merge handoff and merge outcome. Commit: this commit
- [x] T1.2 Update ticket parsing, rendering, and indexing preservation behavior. Commit: this commit
- [x] T1.3 Add service-level handoff checks for `ready_for_merge` tickets. Commit: this commit
- [x] T1.4 Add tests for passed-review handoff and legacy ticket preservation. Commit: this commit
- [x] T1.5 Add slice report and tracker updates. Commit: this commit

## Slice 2: Merger Preflight

Pre-flight:

- Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
- Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
- Slicing: `Merger Agent Workflow`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-merger-preflight`
- Expected PR title: `feat(slice-2): merger preflight`
- Expected report: `.prj/reports/slices/merger-agent-workflow/slice-2-merger-preflight.md`

Tasks:

- [ ] T2.1 Add merger preflight service boundary and result model.
- [ ] T2.2 Reuse provider diagnostics for PR readability and branch/base checks.
- [ ] T2.3 Add serial repair blocker checks.
- [ ] T2.4 Add blocker classification for operational versus repair blockers.
- [ ] T2.5 Add tests for pass, missing metadata, wrong base, stale branch,
  conflict, credentials, and serial blockers.
- [ ] T2.6 Add slice report and tracker updates.

## Slice 3: Merge Execution Command

Pre-flight:

- Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
- Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
- Slicing: `Merger Agent Workflow`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-merge-execution-command`
- Expected PR title: `feat(slice-3): merge execution command`
- Expected report: `.prj/reports/slices/merger-agent-workflow/slice-3-merge-execution-command.md`

Tasks:

- [ ] T3.1 Add CLI/service command for merging a reviewed ticket.
- [ ] T3.2 Implement provider merge execution for Forgejo/Codeberg first.
- [ ] T3.3 Add GitHub merge execution support when provider config is available.
- [ ] T3.4 Apply merge result to ticket metadata and status.
- [ ] T3.5 Add tests with provider fakes for dry-run, success, provider failure,
  conflict, and blocked merge.
- [ ] T3.6 Add slice report and tracker updates.

## Slice 4: Merger Workflow Docs

Pre-flight:

- Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
- Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
- Slicing: `Merger Agent Workflow`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-merger-workflow-docs`
- Expected PR title: `feat(slice-4): merger workflow docs`
- Expected report: `.prj/reports/slices/merger-agent-workflow/slice-4-merger-workflow-docs.md`

Tasks:

- [ ] T4.1 Update how-to documentation for reviewed ticket merge.
- [ ] T4.2 Add or update reference documentation for merger metadata and
  commands.
- [ ] T4.3 Update explanation docs for `@dev`, `@reviewer`, and `@merger`
  roles.
- [ ] T4.4 Link ticket review workflow docs to merger handoff docs.
- [ ] T4.5 Run docs-only validation.
- [ ] T4.6 Add slice report and tracker closure updates.

## Risks

- The merger role must stay narrow and objective.
- Provider mergeability semantics differ and should be normalized only where
  behavior is required.
- Blocked merge handling must avoid creating Board column noise.
