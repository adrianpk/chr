# Dev to Main Agent Evidence Gate

Status: draft

## Summary

ChrysoCoord needs a `dev` to `main` promotion gate that checks objective
evidence before accumulated work is promoted.

The gate must be tooling-first and agent-aware. It must not behave like a
generic PR reviewer, and it must not treat every project change as work that
should be tracked through the ChrysoCoord board.

The board is primarily for agent-managed work. Human or external changes may
exist outside ChrysoCoord and must not produce warnings or failures by default.

## Problem

The current workflow can already force compile, lint, and tests during slice
execution. Re-running those checks at `dev` to `main` is useful as a final
sanity check, but it is not enough to justify a new gate.

The missing value is promotion evidence:

- which agent-managed work is included;
- whether ChrysoCoord's own recorded runs have evidence;
- whether completed agent tickets have result artifacts;
- whether the report matches the exact base/head being promoted;
- what manual or external changes are present for context;
- which validation commands passed or failed.

The gate should verify ChrysoCoord's own claims about agent work. It should not
accuse manual work of being missing from the board.

## Goals

- Provide a repeatable `dev` to `main` evidence report.
- Keep the report useful in local workflows and Woodpecker CI.
- Support Forgejo-first workflows without requiring GitHub, GitLab, CodeRabbit,
  Greptile, or hosted review services.
- Separate agent-managed changes from manual or external changes.
- Avoid false positives by default.
- Fail only on objective issues in default policy.
- Leave room for stricter project-specific policies later.

## Non-Goals

- Do not implement an AI code reviewer.
- Do not require every project change to have a ChrysoCoord ticket.
- Do not require humans to create ChrysoCoord tickets for their own manual work.
- Do not fail on manual or external changes by default.
- Do not require Forgejo API integration in the first version.
- Do not require coverage, CRAP, mutation testing, or fuzzing in the first
  version.
- Do not treat this gate as a replacement for human promotion judgment.

## Core Rule

Only ChrysoCoord-managed work is required to have ChrysoCoord evidence.

Manual or external changes are listed for context only unless project policy
explicitly configures stricter behavior.

Default behavior:

```text
agent-managed work without evidence -> fail
manual or external work without ticket -> info
validation command failure -> fail
stale base/head report -> fail
```

## Definitions

### Agent-Managed Work

Work is agent-managed when ChrysoCoord has recorded it as part of ticket
execution or loop execution.

Signals may include:

- ticket ID;
- run ID;
- execution report;
- loop report;
- result artifact;
- changed-file summary;
- commit SHA recorded by ChrysoCoord;
- branch created or managed by ChrysoCoord.

### Manual or External Work

Work is manual or external when it appears in the Git diff but ChrysoCoord has
no evidence that it produced it.

This is not an error by default.

The gate should report it as context:

```text
Manual or external changes:
- README.md
- docs/explanation/foo.md
```

It should not emit a warning unless explicit policy asks for it.

### Evidence

Evidence is structured or semi-structured data produced by ChrysoCoord or
configured tools.

Examples:

- gate report;
- ticket execution result;
- loop report;
- runtime log path;
- validation command output summary;
- changed-file summary;
- ticket status transition with run ID.

## CLI

Initial command:

```sh
chc gate dev-main --base main --head dev
```

CI-oriented form:

```sh
chc gate dev-main --base origin/main --head origin/dev --ci
```

Optional output flags:

```sh
chc gate dev-main \
  --base origin/main \
  --head origin/dev \
  --json .chc/user/reports/gates/dev-main.json \
  --markdown .chc/user/reports/gates/dev-main.md
```

## Outputs

The gate writes reports under the user workspace:

```text
.chc/user/reports/gates/YYYYMMDD-HHMMSS-dev-main.json
.chc/user/reports/gates/YYYYMMDD-HHMMSS-dev-main.md
```

The JSON report is for automation. The Markdown report is for humans.

## JSON Report

Draft schema:

```json
{
  "schema_version": "0.1",
  "kind": "dev_main_gate_report",
  "created_at": "2026-06-01T00:00:00Z",
  "base": {
    "ref": "origin/main",
    "sha": ""
  },
  "head": {
    "ref": "origin/dev",
    "sha": ""
  },
  "merge_base": "",
  "status": "pass",
  "checks": [],
  "agent_managed_work": [],
  "manual_or_external_changes": [],
  "failures": [],
  "warnings": [],
  "info": []
}
```

Status values:

- `pass`;
- `warn`;
- `fail`;
- `error`.

Default policy should rarely produce `warn`. Informational findings should stay
under `info`.

## Markdown Report

Suggested shape:

```md
# Dev to Main Gate Report

## Summary

## Base and Head

## Validation

## Agent-Managed Work

## Manual or External Changes

## Failures

## Warnings

## Human Review Context
```

The "Manual or External Changes" section must be neutral. It exists to help the
human understand the promotion set, not to imply that the user did something
wrong by working outside the board.

## Evidence Collection

### Git

Collect:

- base ref and SHA;
- head ref and SHA;
- merge base;
- commits in `base..head`;
- changed files;
- diff stat;
- file status: added, modified, deleted, renamed.

### ChrysoCoord

Collect:

- tickets touched by ChrysoCoord runs;
- execution reports;
- loop reports;
- run IDs;
- result artifacts;
- changed-file summaries;
- ticket status transitions caused by ChrysoCoord.

The first version can be best-effort. If ChrysoCoord cannot map a Git change to
agent-managed work, it should classify it as manual or external.

### Validation

Run configured validation.

Initial default:

- use `make verify` when available;
- otherwise use `go test ./...`.

This rule can be refined later. The important point is that validation failure
is objective and can fail the gate.

## Default Failure Conditions

The default gate fails when:

- base or head cannot be resolved;
- validation command fails;
- report cannot be written;
- a ChrysoCoord-recorded run references missing result artifacts;
- a ChrysoCoord-managed ticket is marked completed by the system without
  evidence;
- an existing report is being verified and its base/head SHAs are stale.

The default gate does not fail when:

- files changed without ChrysoCoord tickets;
- commits lack ticket IDs;
- docs changed manually;
- code changed manually;
- a human did not use the ChrysoCoord board.

## Policy

Future project policy can opt into stricter behavior.

Example:

```yaml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: true
      warn_on_manual_changes: false
    manual_changes:
      list: true
      warn_on_high_risk_paths: false
    high_risk_paths:
      - internal/runtime/**
      - internal/persistence/**
      - docs/reference/**
```

Initial policy mode should be equivalent to:

```yaml
traceability:
  mode: agent_only
```

## Woodpecker CI

The first Woodpecker path should not require Forgejo API access.

Conceptual step:

```yaml
steps:
  dev-main-gate:
    image: golang:1.25
    commands:
      - go install ./cmd/chc
      - git fetch origin main dev
      - chc gate dev-main --base origin/main --head origin/dev --ci
```

The gate should exit non-zero on default failure conditions.

The Markdown and JSON reports should be stored under `.chc/user/reports/gates/`
so Woodpecker can preserve them as artifacts when configured.

## Forgejo

Forgejo is the first canonical forge target.

Initial implementation should work entirely from Git refs and local checkout
data. Later Forgejo integration can add:

- PR metadata;
- PR comments;
- status posting;
- artifact links;
- richer branch/ref resolution.

Forgejo API integration is not required for the first implementation.

## Human Review

The gate should help the human decide promotion.

It should present:

- objective failures;
- validation results;
- ChrysoCoord-managed work and its evidence;
- manual or external changes for context;
- exact base/head SHAs.

It should not attempt to decide whether the whole release is semantically good.

## First Implementation Slice

The first implementation should include:

1. `chc gate dev-main --base --head`.
2. Git base/head/merge-base resolution.
3. Commit and changed-file collection.
4. Validation command execution.
5. Basic ChrysoCoord report discovery under `.chc/user/reports`.
6. Classification into agent-managed vs manual/external when evidence is clear.
7. JSON and Markdown report generation.
8. Default failure conditions.
9. CI-friendly exit codes.

## Later Implementation Slices

Later slices can add:

- better run-to-commit mapping;
- ticket status evidence checks;
- loop report aggregation;
- configurable policy file;
- high-risk path policy;
- coverage summary;
- complexity tooling;
- CRAP-like risk radar;
- Woodpecker artifact examples;
- Forgejo PR metadata and status integration;
- optional AI summary over factual reports.

## Exit Codes

Draft model:

- `0`: pass;
- `1`: gate failed;
- `2`: usage or configuration error;
- `3`: required tool missing;
- `4`: collection or report write error;
- `5`: stale report when verifying an existing report.

## Acceptance Criteria

- The command works in a generic Go project initialized with ChrysoCoord.
- The command does not require GitHub or GitLab.
- The command does not require Forgejo API access.
- The command can run in Woodpecker CI from a normal checkout.
- The command writes Markdown and JSON reports under `.chc/user/reports/gates/`.
- The report records base/head SHAs.
- The report lists validation results.
- The report lists agent-managed work when ChrysoCoord evidence exists.
- The report lists manual or external changes neutrally.
- The default policy does not warn or fail merely because a change lacks a
  ChrysoCoord ticket.
- The default policy fails when ChrysoCoord's own recorded work lacks required
  evidence.
