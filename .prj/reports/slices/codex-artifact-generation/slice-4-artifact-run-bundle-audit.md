# Slice 4: Artifact Run Bundle Audit

Status: reviewing
Slicing: Codex Artifact Generation
Plan: `.prj/plans/slicing-codex-artifact-generation-plan.md`
Tracker: `.prj/trackers/slicing-codex-artifact-generation-tracker.md`
Branch: `feat/slice-4-artifact-run-bundle-audit`
PR: pending

## Purpose

Validate the default Codex artifact generation path end to end on a copied
reference workspace and record trace and run-bundle evidence.

## Probe

Script:

```sh
.prj/scripts/probe-codex-artifact-generation.sh
```

Default source workspace:

```text
/home/adrian/Projects/labs/chr/projects/chr-blog-work
```

The script copies the source workspace to a temporary directory, seeds the
`artifact-backend-audit` use case in the copy, clears trace and agent run
bundles in the copy, and runs:

- `chr dump new`
- `chr spec generate`
- `chr plan generate`
- `chr plan tracking create`

## Run Result

Successful run workspace:

```text
/tmp/chr-codex-artifacts-93G0Fu/chr-blog-work
```

Generated artifacts:

- `.chr/default/dumps/artifact-backend-audit-notes.md`
- `.chr/default/dumps/artifact-backend-audit-notes-raw.md`
- `.chr/default/specs/artifact-backend-audit-spec.md`
- `.chr/default/plans/artifact-backend-audit-plan.md`
- `.chr/default/plans/artifact-backend-audit-tracking.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-21-Copied-workspace-safety-guard.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-22-Default-Codex-artifact-generation-path.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-23-Go-owned-artifact-materialization.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-24-Ticket-and-tracking-materialization-boundary.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-25-Trace-and-run-bundle-evidence.md`
- `.chr/default/tickets/artifact-backend-audit/CHR-26-Audit-validation-and-reporting.md`

Command output summary:

- Dump creation completed.
- Spec generation completed.
- Plan generation completed.
- Plan tracking created 6 tickets.
- Trace file contained 8 events.
- Agent run directory contained 4 run bundles.

## Trace Evidence

Trace file:

```text
/tmp/chr-codex-artifacts-93G0Fu/chr-blog-work/.chr/default/log/llm/llm.ndjson
```

Observed events:

| Event | Action | Provider | Status |
| --- | --- | --- | --- |
| `llm_request` | `dump_curation` | `codex` | |
| `llm_response` | `dump_curation` | `codex` | `ready` |
| `llm_request` | `spec_generation` | `codex` | |
| `llm_response` | `spec_generation` | `codex` | `ready` |
| `llm_request` | `plan_generation` | `codex` | |
| `llm_response` | `plan_generation` | `codex` | `ready` |
| `llm_request` | `plan_tracking_materialization` | `codex` | |
| `llm_response` | `plan_tracking_materialization` | `codex` | `ready` |

## Bundle Evidence

Agent run directory:

```text
/tmp/chr-codex-artifacts-93G0Fu/chr-blog-work/.chr/default/tmp/agent-runs
```

Observed bundles:

- `dump-curation-artifact-backend-audit-notes`
- `spec-generation-artifact-backend-audit`
- `plan-generation-artifact-backend-audit`
- `plan-tracking-artifact-backend-audit`

Each bundle contained:

- `command.json`
- `output-schema.json`
- `output.json`
- `prompt.md`
- `stderr.log`
- `stdout.log`

## Source Workspace Check

The source reference workspace did not contain `artifact-backend-audit` files
after the probe. Generated artifacts were present only in the temporary copy.

## Initial Failed Probe

The first probe run failed during spec validation because the generated spec
referenced `.chr/default/user/log/llm/llm.ndjson`. The probe fixture was adjusted
to avoid feeding concrete trace paths into generated spec content. The
successful run above completed after that adjustment.

## Validation

- `.prj/scripts/probe-codex-artifact-generation.sh`
- `make verify`

## Follow-ups

- No required implementation follow-up was found in this slice.
