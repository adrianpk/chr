# Find Execution and Loop Evidence

Use reporting commands when you need to inspect what happened during ticket
execution or a foreground ticket loop.

Reports summarize evidence that already exists. They do not replace execution
records, result artifacts, runtime logs, validation output, ticket comments, or
loop records.

## List Available Reports

List all reports for the resolved project:

```sh
chc report list
```

Select a project explicitly:

```sh
chc report list --dir /path/to/project
```

Each line includes the report kind, project-relative path, ticket key,
execution id, loop id, and status when those fields are known.

## Narrow the List

List only execution reports:

```sh
chc report list --kind execution
```

List only loop reports:

```sh
chc report list --kind loop
```

Find reports for one ticket:

```sh
chc report list --ticket CHC-1
```

Find one execution:

```sh
chc report list --execution exec-1
```

Find one loop:

```sh
chc report list --loop loop-1
```

Filters can be combined:

```sh
chc report list --kind execution --ticket CHC-1 --status succeeded
```

## Show One Report

Show the Markdown body of a report by execution id:

```sh
chc report show --execution exec-1
```

Show a loop report by loop id:

```sh
chc report show --loop loop-1
```

Show a report by project-relative path:

```sh
chc report show .chc/runtime/runs/CHC-1/exec-1/execution-report.md
```

If a lookup matches more than one report, the command fails with an ambiguity
error. Add a narrower filter such as `--kind`, `--execution`, or `--loop`.

## Print Metadata Instead of the Body

Use `--metadata` when you only need the selected report identity:

```sh
chc report show --execution exec-1 --metadata
```

Metadata output uses the same stable line format as `chc report list`.

## Inspect Source Artifacts

Execution reports include paths for:

- result artifact;
- runtime log;
- validation output;
- execution report.

Loop reports include loop identity, status, counts, related executions when
available, and failure information when recorded.

Use the report first to choose the source artifact to inspect next. The source
artifacts remain the detailed evidence.

## Review Handoff

Review handoff summaries are concise Markdown summaries for review surfaces.
They are derived from execution reports and focus on status, validation,
changed files, blockers, artifacts, and next action.

Use the execution report when you need the complete execution evidence. Use the
handoff summary when you need a compact review view.
