# Dev To Main Gate Model

The dev-to-main gate adds an evidence layer before promoting `dev` into
`main`.

It does not replace tests, PR review, or project judgement. It records what can
be observed about the promotion range and turns that evidence into a local
Markdown and JSON report.

## What The Gate Provides

The gate provides:

- `chr gate dev-main`;
- Git range evidence collection;
- validation command execution;
- Markdown and JSON gate reports;
- Chryso execution evidence classification;
- manual or external change listing;
- missing agent evidence failures;
- `.chr/default/config/gates.yml` policy configuration;
- high-risk manual change warnings;
- `chr gate woodpecker init`;
- Woodpecker status as the first Forgejo feedback path;
- reference documentation, Woodpecker guidance, and a repeatable QA scenario.

The gate is usable from the CLI and from Woodpecker. It writes durable evidence
inside the managed project core.

## Workspace Boundary

The gate is a Chryso product feature. It uses:

```text
.chr/default/
```

The current paths are:

- `.chr/default/config/gates.yml`;
- `ops/default/report/<id>/ (Markdown in artifact.md, JSON in evidence/gate-report.json)`;
- execution and loop evidence under `.chr/default/` when agent work runs.

Project notes, issue trackers, or planning directories outside `.chr/default/` can
remain useful to the team, but gate commands do not read them as runtime state.

## Evidence Model

The gate combines three evidence types:

- Git evidence: refs, SHAs, commits, changed files, statuses, and diff stats.
- Validation evidence: configured command, exit code, duration, stdout, and
  stderr.
- Chryso evidence: execution reports and result paths that connect changed
  files to agent-managed work.

Manual or external changes are not errors by default. They are listed as
context because projects often have legitimate human changes, especially during
adoption.

Missing agent evidence is different. When Chryso recorded work but the
expected result evidence is unavailable, the gate can fail because the project
has lost traceability for work it claims was agent-managed.

## Adoption Model

Greenfield projects can start strict because all new work can be shaped through
Chryso from the beginning.

Existing projects need a baseline period. Their history and current branch diffs
were often created before Chryso existed in the project, so strict missing
evidence checks would produce noise.

The recommended adoption path is:

1. initialize `.chr/default/`;
2. use a permissive gate policy;
3. run the gate to capture baseline evidence;
4. process new work through Chryso tickets;
5. tighten traceability after the project has enough agent-managed evidence.

That keeps the gate useful immediately without pretending that older work has
execution evidence it cannot have.

## Forgejo And Woodpecker

The first Forgejo integration path uses Woodpecker pipeline status.

`chr gate woodpecker init` generates a workflow that runs the gate. A non-zero
gate exit fails the pipeline, which gives Forgejo a visible promotion signal
without requiring API comments or tokens.

The detailed evidence remains local to the project under:

```text
ops/default/report/<id>/ (Markdown in artifact.md, JSON in evidence/gate-report.json)
```

Future Forgejo-specific enrichment can add comments, artifact links, or status
details, but the first version keeps the trust boundary simple: CI status is the
signal, local reports are the evidence.

## Current Limits

The current implementation does not create an adoption baseline object with a
dedicated command. Baseline adoption is a documented workflow using the gate
report and permissive policy.

Projects that want Chryso-managed execution evidence need a real
`.chr/default/` workspace created by `chr init`.

The current Woodpecker integration does not write Forgejo comments or call
Forgejo APIs. It relies on Woodpecker status and local reports.
