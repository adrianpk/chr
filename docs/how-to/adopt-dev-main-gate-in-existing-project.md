# Adopt The Dev To Main Gate In An Existing Project

This guide starts from a Go project that already has history, tests, and
development habits before ChrysoCoord is introduced.

The goal is gradual adoption. Existing work should become visible as baseline
context, not fail because it was created before ChrysoCoord managed the
project.

## Initialize The Workspace

From the project root:

```sh
chc init
```

This creates the project-local ChrysoCoord workspace under:

```text
.chc/user/
```

The dev-to-main gate reads and writes under that workspace. It does not read
`.prj/` or any repository-specific planning directory.

## Start With Informational Traceability

Create:

```text
.chc/user/config/gates.yml
```

Use a permissive policy for the first adoption pass:

```yml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: false
      warn_on_manual_changes: true
    manual_changes:
      list: true
      warn_on_high_risk_paths: false
```

This runs validation, writes reports, and lists manual or external changes
without blocking the project because older work has no ChrysoCoord execution
evidence.

## Run The First Gate

Fetch the remote refs:

```sh
git fetch origin
```

Run the gate against the current promotion range:

```sh
chc gate dev-main --base origin/main --head origin/dev
```

Use local refs that match the project's branch model. The command requires a
Git repository and a ChrysoCoord workspace.

## Read The Baseline Evidence

Reports are written under:

```text
.chc/user/reports/gates/
```

Read the Markdown report first. It shows:

- commits in the range;
- changed files;
- validation result;
- manual or external changes;
- missing agent evidence;
- high-risk manual changes when configured.

The first report for an existing project is an adoption baseline. It records
what ChrysoCoord can observe at the point where the project starts using the
gate.

## Add Agent-Managed Work Over Time

After adoption, create or generate tickets and let ChrysoCoord execute new work.
Those executions create evidence under `.chc/user/` that later gate runs can
classify as ChrysoCoord-managed.

Manual work can continue to exist. The default recommendation is to list it as
context first, then tighten policy only when the project has enough agent-run
history to make stricter checks useful.

## Tighten Policy Later

When new agent-managed work consistently leaves evidence, make missing evidence
blocking:

```yml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: true
      warn_on_manual_changes: true
    manual_changes:
      list: true
      warn_on_high_risk_paths: true
    high_risk_paths:
      - internal/runtime/**
      - migrations/**
      - .woodpecker.yml
```

Use high-risk paths for areas where manual changes deserve extra attention.
They are warnings unless the project adds a stricter policy in a later version.

## Add Woodpecker When Ready

Preview the workflow:

```sh
chc gate woodpecker init --dry-run
```

Write it when the generated workflow matches the project:

```sh
chc gate woodpecker init
```

Existing `.woodpecker.yml` files are left unchanged unless they were generated
by ChrysoCoord and `--write` is provided.
