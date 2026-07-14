---
title: "Merger Agent Workflow Plan"
project_id: "chryso"
case_slug: "ticket-execution-loop"
status: "planned"
source_request_id: "chat-2026-06-02-merger-agent-role"
source_response_id: "codex-plan"
created_at: "2026-06-02T00:00:00Z"
mirror: ".prj/plans/slicing-merger-agent-workflow-plan.md"
---

# Merger Agent Workflow Plan

This plan mirrors `.prj/plans/slicing-merger-agent-workflow-plan.md` during
Chryso self-adoption. The `.prj` plan remains the operational source for
slice branch names, PR titles, and implementation report paths.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | merger handoff contract | `feat/slice-1-merger-handoff-contract` | `feat(slice-1): merger handoff contract` | `.prj/reports/slices/merger-agent-workflow/slice-1-merger-handoff-contract.md` |
| Slice 2 | merger preflight | `feat/slice-2-merger-preflight` | `feat(slice-2): merger preflight` | `.prj/reports/slices/merger-agent-workflow/slice-2-merger-preflight.md` |
| Slice 3 | merge execution command | `feat/slice-3-merge-execution-command` | `feat(slice-3): merge execution command` | `.prj/reports/slices/merger-agent-workflow/slice-3-merge-execution-command.md` |
| Slice 4 | merger workflow docs | `feat/slice-4-merger-workflow-docs` | `feat(slice-4): merger workflow docs` | `.prj/reports/slices/merger-agent-workflow/slice-4-merger-workflow-docs.md` |

## Source Material

- Spec: `.chr/user/specs/merger-agent-workflow.md`
- Dump: `.chr/user/dumps/curated/2026-06-02-merger-agent-role.md`
- Tracker mirror: `.chr/user/trackers/merger-agent-workflow-tracker.md`
- Operational plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
- Operational tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`

## Delivery Rule

The implementation follows serialized slice delivery. The next slice starts only
after the previous slice has been reviewed, merged into `dev`, and local `dev`
has been updated.
