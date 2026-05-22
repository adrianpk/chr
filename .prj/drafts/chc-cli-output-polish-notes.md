# ChrysoCoord CLI Output Polish Notes

Status: Draft note

## Context

These notes come from manually running the full-system lab runbook against the
`chc-blog-work` fixture.

They are not an implementation plan yet. They collect CLI output observations
that should be reviewed later and converted into tickets or a small polish
slice if they still hold after more lab runs.

## Output Style

Avoid emojis in CLI and script output.

The current lab scripts do not use emojis. Keep that convention. CLI output
should remain readable in terminals, copied logs, CI output, and Markdown
failure ledgers.

Preferred output tools:

- stable labels;
- indentation;
- aligned columns when the shape is tabular;
- short sections;
- consistent ordering;
- optional ANSI color only when it does not break logs or automation.

ANSI color should not be required for comprehension. If color is added later, it
should be disabled automatically for non-TTY output or through a flag/env
setting.

## Runtime Validation Checks

Current observed output:

```text
check: ok workspace /path/.chc ok
check: ok dockerfile /path/.chc/runtime/agent/Dockerfile ok
check: ok artifact_dir /path/.chc/runtime/runs ok
```

The repeated `ok` at the start and end of each line feels redundant.

Preferred direction:

```text
checks:
  ok  workspace     /path/.chc
  ok  dockerfile    /path/.chc/runtime/agent/Dockerfile
  ok  artifact_dir  /path/.chc/runtime/runs
```

This is compact, readable, easy to copy into logs, and avoids adding a heavy
table format.

An alternative table format is acceptable if later output surfaces need more
columns:

```text
check  status  target        detail
-----  ------  ------------  -----------------------------------------------
1      ok      workspace     /path/.chc
2      ok      dockerfile    /path/.chc/runtime/agent/Dockerfile
3      ok      artifact_dir  /path/.chc/runtime/runs
```

For the current runtime validation output, the indented section is probably the
better default.

## Successful Command Output Verbosity

Observed output:

```text
session: created
id: auth-20260522-063839
target: spec
topic: CLI blog project validation
path: /home/adrian/Projects/labs/ghc/projects/chc-blog-work/.chc/authoring/sessions/auth-20260522-063839/session.md
transcript: /home/adrian/Projects/labs/ghc/projects/chc-blog-work/.chc/authoring/sessions/auth-20260522-063839/transcript.md
artifacts: /home/adrian/Projects/labs/ghc/projects/chc-blog-work/.chc/authoring/sessions/auth-20260522-063839/artifacts.md
```

This is technically useful, but too verbose for the default success path. The
absolute paths dominate the screen and make the important result harder to see.

Potential direction:

- default output should show the created thing and the next useful identifier;
- prefer project-relative paths by default;
- use absolute paths only when required to disambiguate;
- expose full paths through `--verbose`, `--json`, or a detail command;
- keep output stable enough for scripts.

Possible compact default:

```text
session: created auth-20260522-063839
target: spec
topic: CLI blog project validation
files:
  session     .chc/authoring/sessions/auth-20260522-063839/session.md
  transcript  .chc/authoring/sessions/auth-20260522-063839/transcript.md
  artifacts   .chc/authoring/sessions/auth-20260522-063839/artifacts.md
```

Even this may be more than needed for a default. A shorter variant could be:

```text
session: created auth-20260522-063839
target: spec
path: .chc/authoring/sessions/auth-20260522-063839/session.md
```

Then `chc author session show auth-20260522-063839` can expose full detail.

## List Output Without Headers

Observed output:

```text
session: auth-20260522-063839   draft   spec    CLI blog project validation
```

Once `chc author session list` is run, the session model becomes easier to
understand because the topic is shown next to the session id. However, without
headers the values are still ambiguous. A user can reasonably read `draft spec`
as "this session has both a draft and a spec" instead of "status=draft,
target=spec."

Potential direction:

```text
id                       status  target  topic
auth-20260522-063839     draft   spec    CLI blog project validation
```

or, if preserving the `session:` prefix is important:

```text
session  id                    status  target  topic
-------  --------------------  ------  ------  ---------------------------
session  auth-20260522-063839  draft   spec    CLI blog project validation
```

For human CLI output, headers are more important than saving one line.

Future TUI note:

A future TUI can reduce this friction further. What is now a command with a
session id can become selecting a row with arrow keys from a session list. That
does not remove the need for readable CLI output, but it supports the same
direction: expose sessions as identifiable records with clear status, target,
topic, and updated time.

## Demo Runbook Should Use User-Facing Commands

The full-system lab runbook currently includes `--dev-root` in many commands.
That is useful for developer isolation, but it is not representative of the
final user workflow.

For a demo or UX validation run, commands should look like what a user would
actually type:

```sh
chc author session show <session-id>
chc ticket list
chc runtime validate --dry-run
```

Developer-only isolation should be handled outside the command examples where
possible, for example through `direnv` and isolated XDG directories, or moved to
a dedicated "developer isolation" section.

Principle:

The main runbook path should test the product UX, not the developer harness
mechanics. `--dev-root` should appear only where the runbook explicitly explains
developer isolation.

## Runtime Config Smell

Observed generated config:

```yml
version: 1
workspace_root: .chc
canonical_assets: markdown
tickets:
  key_prefix: CHC
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
indexes:
  sqlite:
    enabled: false
notifications:
  telegram:
    enabled: false
  email:
    enabled: false
```

This is only an observation from the lab run, not a conclusion that the current
shape is wrong.

Questions to revisit:

- `version`: useful if config migrations exist and failed updates need a
  previous-version reference.
- `workspace_root`: do we actually intend to support any root other than
  `.chc`?
- `canonical_assets`: do we intend to process anything other than Markdown?
  Current product direction says no.
- `runtime.workdir`: what useful decision does this expose if `.chc` and the
  Docker mount already define the project workspace?
- `runtime.artifact_dir`: should this be configurable, or is it an invariant
  path that should stay hardcoded to avoid broken layouts?
- `runtime.validation_command`: should `make verify` be configurable per
  project, or is the intended contract that managed Go projects expose
  `make verify`?

Potential direction:

- keep only configuration that represents a real supported choice;
- avoid exposing paths that are product invariants;
- keep defaults predictable enough that users do not need to understand runtime
  internals during project bootstrap;
- if a value remains configurable, document why a user would safely change it.

Working principle:

Do not make the user configure values that are not meaningful user decisions.
Configuration should simplify project operation, not expose internal mechanics
or create extra ways to break a managed workspace.

`workspace_root` should probably not be configurable even if alternate paths
could be made to work technically. `.chc/` is the canonical project-local
workspace directory and should be treated as a product invariant.

## Workspace Manifest Smell

Observed generated manifest content:

```md
# ChrysoCoord Workspace

## Identity

- system: ChrysoCoord
- command: chc
- workspace root: .chc

## Status

This directory stores project-local assets for agent-coordinated development workflows.
Markdown files are the canonical assets. Indexes and runtime records are derived or operational state.
```

This repeats the same questionable surface as config:

- `workspace root: .chc` reads like a value, but it is an invariant;
- `Markdown files are the canonical assets` reads like a configurable mode, but
  Markdown is the product contract;
- repeating this in both YAML and Markdown blurs what is operational config and
  what is human-facing workspace description.

Config direction:

- Use YAML for machine-readable configuration that the binary actually consumes.
- Use Markdown for human-readable manifest, protocol, notes, guides, tickets,
  specs, plans, and evidence.
- Do not use Markdown as a second configuration surface.
- Do not duplicate invariant runtime/product facts in both YAML and Markdown
  unless the Markdown is clearly explanatory and not authoritative config.

The manifest should probably describe the workspace in plain language and point
to docs or local folders. It should not look like a parallel config file.

## Protocol Document Self-Reference Smell

Observed generated protocol content:

```md
# Documenting

## Durable Documentation

Use `docs/` for stable project documentation.

Suggested categories:

- `docs/tutorials/`
- `docs/how-to/`
- `docs/reference/`
- `docs/explanation/`

## ChrysoCoord Artifacts

Use `.chc/` for project-local operational artifacts such as tickets, plans, trackers, reports, context, and runtime traces.

## Rules

- Keep one primary intent per document.
- Prefer clear filenames and headings.
- Keep generated operational records separate from durable user-facing docs.
- Update documentation when behavior or contracts change.
```

Potential issue:

`DOCUMENTING.md` has historically meant "how this project documents itself."
Inside a managed project, it currently also describes where ChrysoCoord stores
its own operational artifacts. That may be useful, but it creates a mixed
purpose:

- project documentation rules;
- ChrysoCoord workspace/artifact placement rules;
- agent protocol hints.

The `ChrysoCoord Artifacts` section is especially self-referential. It explains
ChrysoCoord's internal workspace model inside a document that otherwise reads as
general documentation guidance for the project.

Stronger concern:

The layout and documentation shape of `.chc/` operational artifacts should not
be project-influencable convention. Tickets, drafts, plans, trackers, reports,
context packages, indexes, and runtime traces are ChrysoCoord operational
contracts. Their placement and canonical structure should be enforced by the
binary, not negotiated through a generated Markdown convention document.

Project documentation conventions can guide `docs/` and other user-authored
material. They should not define or override ChrysoCoord's operational artifact
layout.

Questions to revisit:

- Should project documentation rules and ChrysoCoord artifact rules live in the
  same protocol document?
- Should `.chc/` artifact placement be explained in the workspace manifest or a
  dedicated workspace guide instead?
- Should `DOCUMENTING.md` stay focused on durable project docs only?
- Should protocol docs avoid naming ChrysoCoord unless the section is explicitly
  about the `.chc/` workspace?

Potential direction:

- keep `DOCUMENTING.md` focused on user/project documentation conventions;
- move `.chc/` operational artifact placement into a workspace guide, manifest,
  or dedicated protocol document;
- reduce self-reference in files that are meant to guide the agent inside an
  arbitrary managed project.

## Follow-up

Do not implement from this note directly. Revisit after additional lab runs and
decide whether this belongs in a small CLI output polish slice or in a broader
diagnostics/readability pass.
