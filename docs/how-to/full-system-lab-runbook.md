# Full-System Lab Runbook

This runbook defines a repeatable lab for exercising Chryso end to end.
It is intentionally broader than a quick smoke test. The goal is to create a
moderately realistic Go project, initialize it with Chryso, drive the
available workflows, record failures, and use those failures to decide what
needs to be fixed next.

The first execution is expected to fail in multiple places. A failure is useful
when it is captured with the command, output, artifact path, and next action.

## Lab Shape

Use a dedicated lab root outside the Chryso repository:

```sh
export CHR_LAB_ROOT="$HOME/Projects/labs/ghc"
export CHR_PROJECTS="$CHR_LAB_ROOT/projects"
export CHR_REF_PROJECT="$CHR_PROJECTS/chr-blog-ref"
export CHR_WORK_PROJECT="$CHR_PROJECTS/chr-blog-work"
export CHR_LOG_DIR="$CHR_LAB_ROOT/logs"
export CHR_ARTIFACT_DIR="$CHR_LAB_ROOT/artifacts"
export CHR_SCRIPTS="$CHR_LAB_ROOT/scripts"
```

Create the base directories:

```sh
mkdir -p "$CHR_PROJECTS" "$CHR_LOG_DIR" "$CHR_ARTIFACT_DIR" "$CHR_SCRIPTS"
```

The reference project is treated as immutable after it is created:

- `chr-blog-ref` is the clean reset source.
- `chr-blog-work` is the project Chryso operates on.
- `logs` stores copied command output for later diagnosis.
- `artifacts` stores snapshots, bundles, and exported evidence.
- `scripts` may later contain reset helpers.

Do not run destructive reset commands outside `$CHR_LAB_ROOT`.

The command path below is written like a final user would run it from inside the
managed project. It does not include developer-only flags such as `--dev-root`.
Use a separate shell environment when you need to isolate user-level state while
developing Chryso itself.

## Reference Project

Use a small but non-trivial Go CLI blog project as the fixture. A blog is useful
because it exercises file IO, Markdown content, metadata, indexing, searching,
commands, tests, and validation without requiring a web server.

The reference project should be named `chr-blog-ref`.

Recommended package shape:

```text
chr-blog-ref/
  go.mod
  Makefile
  cmd/blog/main.go
  internal/blog/
    post.go
    repository.go
    index.go
    search.go
    export.go
  internal/blogtest/
    fixtures.go
  testdata/
    posts/
      welcome.md
      draft.md
  README.md
```

The app should support these user-facing commands:

```sh
blog new "Title" --tag go --tag agent
blog list
blog show <slug>
blog search <query>
blog publish <slug>
blog export --format html --out dist
```

The app should have a `Makefile` with at least:

```makefile
.PHONY: fmt test verify

fmt:
	go fmt ./...

test:
	go test ./...

verify: fmt test
```

## Intentional Incompleteness

The fixture must be useful for agent work. Do not start with a perfect project.
Create a project that compiles only after the first fix, or that has a failing
test with an obvious but non-trivial correction.

Seed at least these issues:

- A bug ticket candidate: publishing a post does not update the index, so
  `blog list` still shows the old draft state.
- A feature ticket candidate: `blog search` should match tags and titles, but
  only title matching exists.
- A validation ticket candidate: one test should fail until the index update or
  search behavior is fixed.
- A documentation ticket candidate: the project README should describe the CLI
  commands after they work.

The goal is not to trick the agent. The goal is to give it a project that
requires reading code, changing behavior, running tests, and leaving evidence.

## Reset Model

Reset the work project by deleting only the work copy and copying the reference
project again:

```sh
rm -rf "$CHR_WORK_PROJECT"
cp -a "$CHR_REF_PROJECT" "$CHR_WORK_PROJECT"
```

Useful helper scripts under `$CHR_SCRIPTS`:

- `reset-project.sh`: recreates `chr-blog-work` from `chr-blog-ref`.
- `snapshot-project.sh`: copies the current work project to `$CHR_ARTIFACT_DIR`.
- `run-step.sh`: runs one command, captures stdout, stderr, exit code, and time.

Keep those scripts constrained to `$CHR_LAB_ROOT`.

## TUI Quiz Parity Reset

Use this reset before validating the TUI interactive quiz loop. It removes only
interactive quiz runtime and interactive quiz artifacts from the work project.
It does not remove project metadata, use-case metadata, dumps, tickets, specs,
or authoring-session quizzes.

```sh
rm -rf "$CHR_WORK_PROJECT/.chr/user/tmp/quiz-runtime"

find "$CHR_WORK_PROJECT/.chr/user/quizzes" -maxdepth 1 -type f \
  \( -name "project-quiz.md" \
     -o -name "project-quiz-raw.md" \
     -o -name "*-quiz.md" \
     -o -name "*-quiz-raw.md" \) \
  -delete
```

Confirm the reset:

```sh
find "$CHR_WORK_PROJECT/.chr/user/tmp/quiz-runtime" -maxdepth 1 -type f 2>/dev/null
find "$CHR_WORK_PROJECT/.chr/user/quizzes" -maxdepth 1 -type f \
  \( -name "project-quiz.md" \
     -o -name "project-quiz-raw.md" \
     -o -name "*-quiz.md" \
     -o -name "*-quiz-raw.md" \)
```

Both commands should print no interactive quiz runtime or interactive quiz
artifacts.

## TUI Quiz Parity Validation

Run this validation from the Chryso repository after building the current
branch:

```sh
make install
```

Open the TUI against the reset work project:

```sh
cd "$CHR_WORK_PROJECT"
chr tui
```

Project-level validation:

1. Open `F4 Quizzes`.
2. Select the project quiz row.
3. Start the quiz.
4. Confirm the right pane maximizes and shows exactly one active question.
5. Answer with one or two short sentences.
6. Confirm the UI shows model activity before the next question appears.
7. Confirm the next question replaces the previous one instead of adding a
   visible batch.

Use-case-level validation:

1. Return to `F4 Quizzes`.
2. Select `Article Authoring`.
3. Start the quiz.
4. Confirm exactly one use-case-specific question appears.
5. Answer with one or two short sentences.
6. Skip the next follow-up question.
7. Confirm the following question returns to broader coverage rather than
   continuing the skipped follow-up branch.

After validation, inspect the artifacts:

```sh
find "$CHR_WORK_PROJECT/.chr/user/tmp/quiz-runtime" -maxdepth 1 -type f -print
find "$CHR_WORK_PROJECT/.chr/user/quizzes" -maxdepth 1 -type f -print
```

Expected evidence:

- one runtime JSON file per started quiz scope;
- one raw quiz capture per answered or skipped scope;
- no large pre-generated default question batch;
- raw turns include `Prompt intent` when generated by the model;
- skipped follow-up branches do not force the next question to stay on the same
  branch.

## Install Chryso

From the Chryso repository:

```sh
make install
```

Confirm the installed binary is the one being tested:

```sh
command -v chr
chr --help
chr ticket --help
chr author --help
chr runtime --help
```

Do not rely on a stale `./bin/chr` from a previous build. Use `make install` or
run `make build` immediately before testing a repository-local binary.

## Initialize the Work Project

Initialize Chryso inside the resettable work project:

```sh
cd "$CHR_WORK_PROJECT"
chr init
```

Validate the runtime boundary without running Docker:

```sh
chr runtime validate --dry-run
```

Expected output uses project-relative paths for files inside the managed
project and groups runtime checks under `checks:`.

Register and select the project:

```sh
chr project add --name chr-blog-lab .
chr project list
chr project select chr-blog-lab
chr project current
```

Expected result:

- `.chr/` exists inside `chr-blog-work`.
- `.chr/user/config.yml` exists.
- `.chr/user/runtime/agent/Dockerfile` exists.
- The lab project is registered in the local user-level project registry.

## Authoring Workflow

Create an authoring session for the lab project:

```sh
chr author session new \
  --target spec \
  --topic "CLI blog project validation"
```

List and inspect sessions:

```sh
chr author session list
chr author session show <session-handle>
```

`chr author session list` prints headers so `status`, `target`, `updated_at`,
and `topic` are distinguishable. Use the session id, exact topic, or a unique
topic slug when a later command requires `<session-handle>`.

Generate a context-aware quiz:

```sh
chr author quiz new \
  --target spec \
  --topic "CLI blog project validation" \
  --question-count 8
```

The command sends structured project context to the quiz provider and prints the
generated quiz path relative to the project, for example
`.chr/user/quizzes/<quiz-file>.md`.

Edit the generated quiz answers in Markdown, then generate or edit the
corresponding spec from the Specs surface. Specs that still need review should
use `status: draft`; legacy `quiz-to-draft` remains only for older authoring
sessions.

Exercise assisted authoring with deterministic fake adapter inputs:

```sh
chr author assisted start \
  --target spec \
  --topic "CLI blog project validation"

chr author assisted resume \
  --session <session-handle> \
  --fake-question "What behavior should the CLI blog prove first?"

chr author assisted answer \
  --session <session-handle> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id <question-id> \
  --question "What behavior should the CLI blog prove first?" \
  --answer "Publishing a draft should update the persistent index and tests should pass."

chr author assisted status \
  --session <session-handle>

chr author assisted finalize \
  --session <session-handle> \
  --fake-summary "The lab should prove initialization, authoring, ticket execution, validation, reporting, maintenance, and remote command intake against the CLI blog project." \
  --target-kind spec \
  --target-name "cli blog validation spec"
```

Current boundary: there is no standalone `chr dump` command. For dump-style
context, create a Markdown file manually under the project workspace and use it
as source material for drafts, specs, plans, or tickets.

## Ticket Workflow

The examples below are intentionally explicit because this runbook is meant to
be repeatable from any directory and auditable after a failed run. In normal
interactive use, when the shell is already inside the managed project, the
ticket can be created with much less input:

```sh
chr ticket new "Fix publish index update" \
  --synopsis "Fix the publish index update path."
```

Metadata such as type, priority, severity, use case, project area, and tags can
be supplied at creation time when it is useful. They can also be edited later in
the generated Markdown ticket. The Markdown file remains the canonical project
artifact.

Create tickets from the seeded project issues:

```sh
chr ticket new \
  "Fix publish index update" \
  --synopsis "Fix the publish index update path." \
  --type bug \
  --priority high \
  --severity critical \
  --use-case cli-blog \
  --project-area indexing \
  --tag runtime

chr ticket new \
  "Search posts by tags" \
  --synopsis "Add tag-aware post search." \
  --type feature \
  --priority normal \
  --severity normal \
  --use-case cli-blog \
  --project-area search \
  --tag authoring
```

Inspect and prepare tickets:

```sh
chr ticket list
chr ticket show CHR-1
chr ticket comment CHR-1 "Lab ticket created from the seeded publish/index bug."
chr ticket status CHR-1 ready
chr ticket status CHR-2 ready
```

Rebuild and inspect the ticket index:

```sh
chr index rebuild
chr index status
```

## Runtime and Single-Ticket Execution

Preview one ticket execution:

```sh
chr ticket start CHR-1 --dry-run
```

Run it:

```sh
chr ticket start CHR-1
```

Capture the execution id from stdout or from reports:

```sh
chr report list --ticket CHR-1
```

Preview and apply the result:

```sh
chr ticket apply-result <execution-id> --dry-run
chr ticket apply-result <execution-id>
```

Inspect evidence:

```sh
chr report show --execution <execution-id>
chr work logs <execution-id>
make verify
git diff
```

Expected result:

- the runtime writes execution artifacts under `ops/default/run`;
- the ticket records progress, blockers, or completion evidence;
- validation output is inspectable;
- runtime logs and event streams are discoverable;
- changed files can be reviewed with `git diff`.

## Agent Work Ready Queue

Preview the ready queue through the agent-work facade:

```sh
chr work run-all \
  --order priority_severity \
  --dry-run
```

Run the ready queue:

```sh
chr work run-all \
  --order priority_severity
```

Observe the run from Dash with `F8`, or inspect evidence after it finishes:

```sh
chr report list --kind loop
chr report list --kind execution
chr report show --loop <loop-id>
chr work events ops/default/run/<loop-artifact-id>/evidence/events.jsonl
```

The run should stop on a blocked or failed ticket instead of spinning forever.
Use `chr work doctor` or Dash readiness diagnostics when work is blocked before
execution starts.

## Foreground Ticket Loop

Preview eligible work:

```sh
chr ticket loop \
  --status ready \
  --order priority_severity \
  --max-iterations 2 \
  --dry-run
```

Run a bounded loop:

```sh
chr ticket loop \
  --status ready \
  --order priority_severity \
  --max-iterations 2
```

Inspect loop evidence:

```sh
chr report list --kind loop
chr report show --loop <loop-id>
```

The loop should stop on a blocked or failed ticket instead of spinning forever.

## Daemon Coordinator

Exercise local daemon state:

```sh
chr daemon status
chr daemon start --max-iterations 1
chr daemon status
chr daemon stop
```

This validates local coordinator state. It does not create a hosted service.

## Notifications and Remote Command Intake

Inspect notification configuration:

```sh
chr notify status
```

Preview a Telegram payload without contacting Telegram:

```sh
chr notify telegram test --dry-run
```

Create a fixture file under `$CHR_ARTIFACT_DIR/telegram-updates.json` with test
Telegram updates, then parse it:

```sh
chr notify telegram poll \
  --fixture "$CHR_ARTIFACT_DIR/telegram-updates.json" \
  --dry-run

chr notify telegram poll \
  --fixture "$CHR_ARTIFACT_DIR/telegram-updates.json"
```

The expected result is a stored accepted or rejected remote command handoff
record. This stage validates intake, not live Telegram operation.

## Maintenance and Recovery

Write a diagnostic bundle:

```sh
chr maintenance bundle
```

Inspect local state:

```sh
chr maintenance db status
chr maintenance registry status
chr maintenance execution artifacts
chr maintenance execution cleanup-preview
```

Preview registry repair commands before using the write variants:

```sh
chr maintenance registry disable --dry-run chr-blog-lab
chr maintenance registry set-path --dry-run chr-blog-lab "$CHR_WORK_PROJECT"
```

Copy the diagnostic bundle into `$CHR_ARTIFACT_DIR` after each run that exposes
a failure.

## Proving-Ground Check

If the work project includes a proving objective, run the proving commands
against it:

```sh
chr proving validate \
  --objective .chr/proving/objectives/CLI-BLOG-1-fix-publish-index.md

chr proving run \
  --objective .chr/proving/objectives/CLI-BLOG-1-fix-publish-index.md \
  --dry-run

chr proving ralph \
  --objective .chr/proving/objectives/CLI-BLOG-1-fix-publish-index.md \
  --max-attempts 3
```

Use this only as a direct sandbox proof. The main workflow remains tickets.

## Failure Ledger

Create a run-specific ledger:

```sh
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
LEDGER="$CHR_LOG_DIR/$RUN_ID-failures.md"
touch "$LEDGER"
```

For every unexpected failure, record:

```md
## <short failure title>

- Command:
- Exit code:
- Expected:
- Actual:
- Blocking: yes/no
- Relevant files:
- Relevant reports:
- Runtime log:
- Follow-up ticket:
```

If the failure is in Chryso, create a Chryso ticket after the run. If
the failure is only in the lab fixture, fix the fixture or reset it from the
reference project.

## Run Completion Criteria

A useful full-system lab run produces:

- a resettable `chr-blog-work` project;
- a clean way to restore it from `chr-blog-ref`;
- initialized `.chr/` project-local state;
- authoring session, quiz, legacy quiz-to-draft, and assisted authoring artifacts;
- at least two Markdown tickets;
- ticket index rebuild and status output;
- one single-ticket execution attempt;
- one foreground loop attempt;
- report evidence for execution and loop records;
- daemon status evidence;
- notification and remote command fixture evidence;
- maintenance bundle and state inspection output;
- a failure ledger for anything that does not work.

After the run, decide fixes from evidence. Do not rely on memory of terminal
output.
