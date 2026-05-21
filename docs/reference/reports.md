# Reports

Reports are Markdown summaries of execution and loop evidence.

They are project-local artifacts under the managed project workspace. Reports
summarize existing records and artifacts; they do not replace tickets,
execution records, loop records, runtime logs, validation artifacts, or result
envelopes.

## List Reports

List reports for the resolved project:

```sh
chc report list
```

Project resolution follows the same rules as ticket commands. Use `--dir` to
select a project explicitly:

```sh
chc report list --dir /path/to/project
```

## Output

Each report is printed as one line:

```text
kind: execution path: .chc/runtime/runs/CHC-1/exec-1/execution-report.md ticket: CHC-1 execution: exec-1 loop: none status: succeeded
```

Fields:

- `kind`: `execution` or `loop`;
- `path`: project-relative report path;
- `ticket`: ticket key when known;
- `execution`: execution id when known;
- `loop`: loop id when known;
- `status`: report status when known.

Missing fields render as `none`.

## Filters

Supported filters:

```sh
chc report list --kind execution
chc report list --kind loop
chc report list --ticket CHC-1
chc report list --execution exec-1
chc report list --loop loop-1
chc report list --status succeeded
```

Filters can be combined. All provided filters must match.

## Discovery Boundary

Report listing discovers Markdown report artifacts under:

```text
.chc/runtime/runs/
```

Current discovered report filenames:

- `execution-report.md`;
- `loop-report.md`.

The listing command reads only lightweight report metadata from Markdown list
fields. It does not inspect runtime logs or validation output.

Direct report lookup and body rendering belong to `chc report show`.
