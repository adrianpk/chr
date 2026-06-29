# Reports

Reports are Markdown summaries of execution and loop evidence.

They are project-local artifacts under the managed project core. Reports
summarize existing records and artifacts; they do not replace tickets,
execution records, loop records, runtime logs, validation artifacts, or result
envelopes.

Related report formats:

- `docs/reference/execution-reports.md`;
- `docs/reference/loop-reports.md`;
- `docs/reference/review-handoff-summaries.md`.

## List Reports

List reports for the resolved project:

```sh
chr report list
```

Project resolution follows the same rules as ticket commands. Use `--dir` to
select a project explicitly:

```sh
chr report list --dir /path/to/project
```

## Output

Each report is printed as one line:

```text
kind: execution path: .chr/user/runtime/runs/CHR-1/exec-1/execution-report.md ticket: CHR-1 execution: exec-1 loop: none status: succeeded
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
chr report list --kind execution
chr report list --kind loop
chr report list --ticket CHR-1
chr report list --execution exec-1
chr report list --loop loop-1
chr report list --status succeeded
```

Filters can be combined. All provided filters must match.

## Show One Report

Show a report body by report id or lookup filter:

```sh
chr report show .chr/user/runtime/runs/CHR-1/exec-1/execution-report.md
chr report show --execution exec-1
chr report show --loop loop-1
chr report show --ticket CHR-1 --kind execution
```

The report id can be a project-relative report path, an absolute report path,
a ticket key, an execution id, or a loop id.

Lookup must resolve to exactly one report. If no report matches, the command
fails. If multiple reports match, the command fails with an ambiguity error.
Use narrower filters such as `--kind`, `--execution`, or `--loop` to select a
single report.

To print metadata instead of the Markdown body:

```sh
chr report show --execution exec-1 --metadata
```

Metadata output uses the same stable line format as `chr report list`.

## Discovery Boundary

Report listing discovers Markdown report artifacts under:

```text
.chr/user/runtime/runs/
```

Current discovered report filenames:

- `execution-report.md`;
- `loop-report.md`.

The listing command reads only lightweight report metadata from Markdown list
fields. It does not inspect runtime logs or validation output.

Direct report lookup reads the selected Markdown report artifact and prints it
without rewriting it.
