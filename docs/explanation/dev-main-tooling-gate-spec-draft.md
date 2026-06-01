# Dev to Main Tooling Gate Draft

Status: draft

This is an extensive pre-spec draft. It is meant to be refined through later
review and quiz sessions before becoming an implementation spec.

## Intent

ChrysoCoord should eventually provide a way to decide whether a `dev` to `main`
promotion has enough objective evidence to proceed.

The gate should focus on facts produced by tools, not linguistic judgment about
whether code "looks good". AI may summarize, organize, or explain evidence, but
it should not be the source of truth for pass/fail decisions.

The first concrete target should be a Forgejo repository using Woodpecker CI.
That fits the current operating model: Forgejo as canonical forge, `dev` as the
agent integration branch, `main` as the promoted branch, and CI as the place
where promotion checks can become visible and enforceable.

## Problem

ChrysoCoord can generate specs, plans, trackers, tickets, and agent execution
evidence. It can also run slices that move through `dev` before the user
promotes the accumulated work to `main`.

The missing piece is a deterministic promotion check.

Without that check, `dev` to `main` promotion relies on a mix of:

- whatever validation the agent ran inside each ticket;
- whatever the user remembers to inspect;
- whatever the PR body says;
- whatever CI happens to run;
- manual judgment over a potentially large accumulated diff.

That is workable for early development, but it does not scale well when many
small slices accumulate before promotion.

The system needs a way to answer:

- What exactly changed between `main` and `dev`?
- Which tickets and slices are included?
- Which validation commands ran?
- Which packages and contracts changed?
- Did the change touch runtime, persistence, configuration, or public docs?
- Is there execution evidence for completed tickets?
- Are tests, coverage, complexity, or risk signals weaker than before?
- Is the review report fresh for the exact base/head being promoted?
- Can CI enforce that the report exists and is current?

## Non-Goals

This feature should not:

- replicate CodeRabbit, Greptile, or a generic AI reviewer;
- require GitHub or GitLab;
- require a hosted code review service;
- use an LLM as the authority for pass/fail;
- require every project to use the same strictness profile;
- make every slice wait for a human review;
- require ChrysoCoord to understand every Go code quality problem at launch.

## Principles

### Tooling First

Facts should come from deterministic tools:

- git;
- Go tooling;
- configured linters;
- coverage tooling;
- complexity tooling;
- ChrysoCoord ticket and execution artifacts;
- CI environment metadata.

AI can summarize the report later, but the gate must be useful without AI.

### Project Portable

The gate must work for a generic Go project managed by ChrysoCoord. It should
not assume a particular repository shape beyond a Git repository and a
`.chc/user/` workspace when ChrysoCoord is initialized.

### Forgejo First

Forgejo should be a first-class target. The gate should work with Forgejo pull
requests and branches without requiring a mirror to GitHub or GitLab.

### Woodpecker CI First

Woodpecker CI should be the first CI target because it fits Forgejo and open
self-hosted workflows. The initial design should be runnable locally and in CI.

### Evidence Before Judgment

The system should collect evidence first. Human or AI interpretation can happen
afterward.

### Configurable Strictness

Different projects have different risk profiles. The first version should
support a conservative default and allow stricter policies over time.

## Expected User Workflow

### Local Preflight

Before opening or merging a `dev` to `main` PR:

```sh
chc gate dev-main --base main --head dev
```

or:

```sh
chc review collect --base main --head dev
```

The final command name is open. `gate dev-main` is direct and CI-friendly.
`review collect` is more general if future review surfaces expand.

The command produces:

```text
.chc/user/reports/gates/YYYYMMDD-HHMMSS-dev-main.json
.chc/user/reports/gates/YYYYMMDD-HHMMSS-dev-main.md
```

The Markdown report is for humans. The JSON report is for automation and CI.

### PR Body

The `dev` to `main` PR should reference the gate report:

```md
## Validation

- `make verify`
- `chc gate dev-main --base main --head dev`

## Gate

- `.chc/user/reports/gates/20260601-090000-dev-main.md`
```

### CI Enforcement

Woodpecker should be able to run:

```sh
chc gate dev-main --base origin/main --head origin/dev --ci
```

The CI step should fail when configured hard gates fail.

The step should upload or preserve the Markdown/JSON report as a CI artifact
when Woodpecker artifact handling is configured.

## Evidence Model

The gate report should be based on facts. It should not rely on prose generated
by an agent.

### Git Evidence

Collect:

- base ref;
- head ref;
- resolved base SHA;
- resolved head SHA;
- merge base SHA;
- commit list;
- changed file list;
- rename/delete/add/change status;
- diff stat;
- touched top-level areas;
- touched Go packages;
- touched docs;
- touched config/runtime/persistence paths.

Possible commands:

```sh
git merge-base <base> <head>
git log --oneline <base>..<head>
git diff --name-status <base>..<head>
git diff --stat <base>..<head>
git diff --numstat <base>..<head>
```

### ChrysoCoord Evidence

Collect from `.chc/user/`:

- tickets included in the promoted range if they can be mapped;
- ticket status;
- execution reports;
- loop reports;
- changed-file summaries;
- handoff/review summaries;
- blockers or unresolved comments;
- generated spec/plan/tracker references when available.

Open questions:

- How should commits map to tickets?
- Should ticket IDs be required in branch names, commit messages, PR bodies, or
  result artifacts?
- Should the gate warn or fail when changed code cannot be mapped to a ticket?

### Go Validation Evidence

Collect:

- `go test ./...`;
- `go test -race ./...` when enabled;
- `go test -coverprofile`;
- coverage summary;
- `go vet ./...`;
- configured `golangci-lint`;
- `govulncheck` when available and enabled.

The command should detect missing tools and report them as skipped, warning, or
failure according to policy.

### Complexity and CRAP-Like Evidence

Collect complexity when configured tooling is available:

- cyclomatic complexity;
- cognitive complexity;
- functions above threshold;
- changed functions above threshold;
- complexity deltas if baseline exists.

A CRAP-like signal needs both complexity and coverage. Early implementation can
avoid exact CRAP scoring and still report:

- complex changed functions;
- packages with low coverage;
- changed packages with low or declining coverage;
- complex functions without nearby tests.

Later implementation can calculate CRAP when coverage mapping is reliable.

### Doctrine Evidence

The gate should eventually check whether the change touched areas that imply
stronger evidence:

- runtime execution;
- ticket lifecycle;
- persistence or migration code;
- config parsing;
- external integrations;
- security or auth-related paths;
- generated artifact formats;
- public CLI/TUI behavior;
- public docs/reference contracts.

For those areas, the report can ask for or require:

- compatibility notes;
- explicit validation commands;
- docs/reference updates;
- migration notes;
- rollback or recovery notes;
- higher coverage or targeted tests.

## Report Schema

The JSON report should be stable enough for CI.

Sketch:

```json
{
  "schema_version": "0.1",
  "kind": "dev_main_gate_report",
  "created_at": "2026-06-01T09:00:00Z",
  "project_root": "/workspace",
  "base": {
    "ref": "origin/main",
    "sha": "..."
  },
  "head": {
    "ref": "origin/dev",
    "sha": "..."
  },
  "merge_base": "...",
  "status": "pass",
  "checks": [
    {
      "id": "go_test",
      "status": "pass",
      "command": "go test ./...",
      "summary": "all packages passed"
    }
  ],
  "warnings": [],
  "failures": [],
  "changed_files": [],
  "tickets": [],
  "artifacts": []
}
```

Status values:

- `pass`;
- `warn`;
- `fail`;
- `error`.

Check status values:

- `pass`;
- `warn`;
- `fail`;
- `skip`;
- `error`.

## Markdown Report Shape

The Markdown report should be concise but complete.

Suggested sections:

```md
# Dev to Main Gate Report

## Summary

## Base and Head

## Included Commits

## Included Tickets

## Changed Areas

## Validation Commands

## Coverage and Complexity

## Doctrine Checks

## Warnings

## Failures

## Human Review Focus
```

The "Human Review Focus" section should be generated from tool facts:

- high-risk paths changed;
- contracts changed;
- tests missing or weak;
- coverage declined;
- complexity increased;
- unmapped changes;
- tickets without evidence.

## Policy Model

The gate needs configurable policy.

Possible location:

```text
.chc/user/config/gates.yml
```

or project config:

```text
.chc/user/config.yml
```

Open design point: ChrysoCoord recently moved project data under user
directories. Gate config should follow that rule unless there is a deliberate
project-wide config story.

Policy sketch:

```yaml
gates:
  dev_main:
    required_checks:
      - git_range
      - go_test
      - lint
    optional_checks:
      - coverage
      - complexity
      - govulncheck
    fail_on:
      missing_execution_evidence: true
      stale_report: true
      validation_failure: true
    warn_on:
      unmapped_changes: true
      changed_go_without_tests: true
      changed_reference_contract_without_docs: true
      complexity_increase_without_coverage: true
    high_risk_paths:
      - internal/chcapp/runtime/**
      - internal/chcapp/ticket/**
      - docs/reference/**
```

## Woodpecker CI Direction

Initial Woodpecker integration should be simple.

Example conceptual step:

```yaml
steps:
  gate-dev-main:
    image: golang:1.25
    commands:
      - go install ./cmd/chc
      - git fetch origin main dev
      - chc gate dev-main --base origin/main --head origin/dev --ci
```

Open questions:

- Should CI build ChrysoCoord from the repository under test or use an installed
  binary?
- How should artifacts be uploaded in Woodpecker?
- Should the command emit Woodpecker-friendly annotations?
- How should private Forgejo authentication be handled for API enrichment?

First implementation can avoid Forgejo API calls and rely only on git refs.

## Forgejo Direction

Forgejo support should not require GitHub-compatible review bots.

The initial design can work entirely from local git data and ChrysoCoord
artifacts. Forgejo API integration can come later for:

- PR metadata;
- PR comments;
- status checks;
- artifact links;
- posting gate summaries;
- fetching changed files when running outside a full checkout.

The first version should not depend on Forgejo API access.

## CLI Surface

Candidate commands:

```sh
chc gate dev-main --base main --head dev
chc gate dev-main --base origin/main --head origin/dev --ci
chc gate dev-main --base main --head dev --policy .chc/user/config/gates.yml
chc gate show latest
```

Possible broader namespace:

```sh
chc review collect --base main --head dev
chc review gate --report .chc/user/reports/gates/...
```

The `gate` namespace is clearer for CI. The `review` namespace is clearer for
humans. This needs a later naming decision.

## TUI Surface

The TUI can eventually show:

- latest gate status;
- base/head SHAs;
- checks run;
- warnings/failures;
- linked Markdown report;
- human review focus.

This should not be required for the first CLI/CI implementation.

## Minimal First Slice

A minimal useful version could:

1. Accept `--base` and `--head`.
2. Resolve SHAs and merge base.
3. Collect commits and changed files.
4. Run `go test ./...`.
5. Run configured lint command if available.
6. Detect `.chc/user/` reports/tickets presence.
7. Write JSON and Markdown reports.
8. Fail if validation fails.
9. Fail if report cannot resolve base/head.
10. Return stable exit codes for CI.

This would already let Woodpecker enforce a real promotion gate.

## Later Slices

Possible follow-up slices:

- ticket mapping;
- execution evidence checks;
- coverage collection;
- complexity collection;
- CRAP-like risk report;
- high-risk path policy;
- docs/reference contract checks;
- Forgejo PR metadata integration;
- Woodpecker artifact/annotation support;
- TUI gate status view;
- AI summary over the generated factual report.

## Exit Codes

Draft exit code model:

- `0`: pass, no hard failures;
- `1`: gate failed;
- `2`: usage or configuration error;
- `3`: required tool missing;
- `4`: internal collection error;
- `5`: stale or mismatched report when verifying an existing report.

## Open Quiz Questions

These questions should be answered before turning this draft into a formal spec:

1. Should the first command be named `chc gate dev-main` or live under
   `chc review`?
2. Should the gate always write reports, even when it fails early?
3. Which checks are hard failures in the first version?
4. Should the first version require `.chc/user/` ticket evidence, or only report
   its absence?
5. How should commits map to tickets?
6. Should gate reports be committed, kept as local artifacts, or uploaded only
   as CI artifacts?
7. Should the PR body reference the report manually or should ChrysoCoord update
   it through Forgejo later?
8. What Woodpecker artifact mechanism should be the first supported path?
9. Should coverage be part of the first implementation or deferred?
10. Which complexity tool should be preferred for Go?
11. How strict should high-risk path checks be?
12. Should the gate be project-neutral or Go-first in its first version?
13. Should CI use the ChrysoCoord binary from the repo under test or from a
    trusted prebuilt image?
14. Should AI summarization be completely out of scope for the first spec?

## Success Criteria

The eventual implementation should be considered successful when:

- a Forgejo/Woodpecker project can run the gate in CI;
- the gate fails on validation failure;
- the gate produces a Markdown report useful to a human;
- the gate produces JSON usable by automation;
- the report records exact base/head SHAs;
- the report can be linked from a `dev` to `main` PR;
- the gate does not require GitHub, GitLab, CodeRabbit, Greptile, or any hosted
  review provider;
- the design leaves room for stronger project-specific policies later.
