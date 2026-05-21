# Full-System Lab Runbook

This runbook defines a repeatable lab for exercising ChrysoCoord end to end.
It is intentionally broader than a quick smoke test. The goal is to create a
moderately realistic Go project, initialize it with ChrysoCoord, drive the
available workflows, record failures, and use those failures to decide what
needs to be fixed next.

The first execution is expected to fail in multiple places. A failure is useful
when it is captured with the command, output, artifact path, and next action.

## Lab Shape

Use a dedicated lab root outside the ChrysoCoord repository:

```sh
export GHC_LAB_ROOT="$HOME/Projects/labs/ghc"
export GHC_PROJECTS="$GHC_LAB_ROOT/projects"
export GHC_REF_PROJECT="$GHC_PROJECTS/chc-blog-ref"
export GHC_WORK_PROJECT="$GHC_PROJECTS/chc-blog-work"
export GHC_DEV_ROOT="$GHC_LAB_ROOT/dev-root"
export GHC_LOG_DIR="$GHC_LAB_ROOT/logs"
export GHC_ARTIFACT_DIR="$GHC_LAB_ROOT/artifacts"
export GHC_SCRIPTS="$GHC_LAB_ROOT/scripts"
```

Create the base directories:

```sh
mkdir -p "$GHC_PROJECTS" "$GHC_DEV_ROOT" "$GHC_LOG_DIR" "$GHC_ARTIFACT_DIR" "$GHC_SCRIPTS"
```

The reference project is treated as immutable after it is created:

- `chc-blog-ref` is the clean reset source.
- `chc-blog-work` is the project ChrysoCoord operates on.
- `dev-root` isolates ChrysoCoord user-level state from the real user account.
- `logs` stores copied command output for later diagnosis.
- `artifacts` stores snapshots, bundles, and exported evidence.
- `scripts` may later contain reset helpers.

Do not run destructive reset commands outside `$GHC_LAB_ROOT`.

## Reference Project

Use a small but non-trivial Go CLI blog project as the fixture. A blog is useful
because it exercises file IO, Markdown content, metadata, indexing, searching,
commands, tests, and validation without requiring a web server.

The reference project should be named `chc-blog-ref`.

Recommended package shape:

```text
chc-blog-ref/
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
rm -rf "$GHC_WORK_PROJECT"
cp -a "$GHC_REF_PROJECT" "$GHC_WORK_PROJECT"
```

Reset ChrysoCoord user-level lab state separately:

```sh
rm -rf "$GHC_DEV_ROOT"
mkdir -p "$GHC_DEV_ROOT"
```

Useful future scripts under `$GHC_SCRIPTS`:

- `reset-project.sh`: recreates `chc-blog-work` from `chc-blog-ref`.
- `reset-dev-root.sh`: clears and recreates `$GHC_DEV_ROOT`.
- `snapshot-project.sh`: copies the current work project to `$GHC_ARTIFACT_DIR`.
- `run-step.sh`: runs one command, captures stdout, stderr, exit code, and time.

Keep those scripts constrained to `$GHC_LAB_ROOT`.

## Install ChrysoCoord

From the ChrysoCoord repository:

```sh
make install
```

Confirm the installed binary is the one being tested:

```sh
command -v chc
chc --help
chc ticket --help
chc author --help
chc runtime --help
```

Do not rely on a stale `./bin/chc` from a previous build. Use `make install` or
run `make build` immediately before testing a repository-local binary.

## Initialize the Work Project

Initialize ChrysoCoord inside the resettable work project:

```sh
chc init --dir "$GHC_WORK_PROJECT"
```

Validate the runtime boundary without running Docker:

```sh
chc runtime validate --dir "$GHC_WORK_PROJECT" --dry-run
```

Register and select the project using isolated lab state:

```sh
chc project add --dev-root "$GHC_DEV_ROOT" --name chc-blog-lab "$GHC_WORK_PROJECT"
chc project list --dev-root "$GHC_DEV_ROOT"
chc project select --dev-root "$GHC_DEV_ROOT" chc-blog-lab
chc project current --dev-root "$GHC_DEV_ROOT"
```

Expected result:

- `.chc/` exists inside `chc-blog-work`.
- `.chc/config.yml` exists.
- `.chc/runtime/agent/Dockerfile` exists.
- The lab project is registered under `$GHC_DEV_ROOT`.

## Authoring Workflow

Create an authoring session for the lab project:

```sh
chc author session new \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --target spec \
  --topic "CLI blog project validation"
```

List and inspect sessions:

```sh
chc author session list --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
chc author session show --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" <session-id>
```

Generate a static quiz:

```sh
chc author quiz new \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --target spec \
  --topic "CLI blog project validation" \
  --question-count 8
```

Edit the generated quiz answers in Markdown, then promote the answered quiz to a
draft:

```sh
chc author promote quiz-to-draft \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --session <session-id> \
  --quiz "$GHC_WORK_PROJECT/.chc/quizzes/<quiz-file>.md"
```

Exercise assisted authoring with deterministic fake adapter inputs:

```sh
chc author assisted start \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --target spec \
  --topic "CLI blog project validation"

chc author assisted resume \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --session <session-id> \
  --fake-question "What behavior should the CLI blog prove first?"

chc author assisted answer \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --session <session-id> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id <question-id> \
  --question "What behavior should the CLI blog prove first?" \
  --answer "Publishing a draft should update the persistent index and tests should pass."

chc author assisted status \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --session <session-id>

chc author assisted finalize \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --session <session-id> \
  --fake-summary "The lab should prove initialization, authoring, ticket execution, validation, reporting, maintenance, and remote command intake against the CLI blog project." \
  --target-kind spec \
  --target-name "cli blog validation spec"
```

Current boundary: there is no standalone `chc dump` command. For dump-style
context, create a Markdown file manually under the project workspace and use it
as source material for drafts, specs, plans, or tickets.

## Ticket Workflow

Create tickets from the seeded project issues:

```sh
chc ticket new \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  "Fix publish index update" \
  --type bug \
  --priority high \
  --severity critical \
  --use-case cli-blog \
  --project-area indexing \
  --tag runtime

chc ticket new \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  "Search posts by tags" \
  --type feature \
  --priority medium \
  --severity major \
  --use-case cli-blog \
  --project-area search \
  --tag authoring
```

Inspect and prepare tickets:

```sh
chc ticket list --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
chc ticket show --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-1
chc ticket comment --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-1 "Lab ticket created from the seeded publish/index bug."
chc ticket status --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-1 ready
chc ticket status --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-2 ready
```

Rebuild and inspect the ticket index:

```sh
chc index rebuild --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
chc index status --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
```

## Runtime and Single-Ticket Execution

Preview one ticket execution:

```sh
chc ticket start --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-1 --dry-run
```

Run it:

```sh
chc ticket start --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" CHC-1
```

Capture the execution id from stdout or from reports:

```sh
chc report list --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" --ticket CHC-1
```

Preview and apply the result:

```sh
chc ticket apply-result --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" <execution-id> --dry-run
chc ticket apply-result --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" <execution-id>
```

Inspect evidence:

```sh
chc report show --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" --execution <execution-id>
make -C "$GHC_WORK_PROJECT" verify
git -C "$GHC_WORK_PROJECT" diff
```

Expected result:

- the runtime writes execution artifacts under `.chc/runtime/runs/`;
- the ticket records progress, blockers, or completion evidence;
- validation output is inspectable;
- changed files can be reviewed with `git diff`.

## Foreground Ticket Loop

Preview eligible work:

```sh
chc ticket loop \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --status ready \
  --order priority_severity \
  --max-iterations 2 \
  --dry-run
```

Run a bounded loop:

```sh
chc ticket loop \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --status ready \
  --order priority_severity \
  --max-iterations 2
```

Inspect loop evidence:

```sh
chc report list --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" --kind loop
chc report show --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" --loop <loop-id>
```

The loop should stop on a blocked or failed ticket instead of spinning forever.

## Daemon Coordinator

Exercise daemon state with the isolated lab root:

```sh
chc daemon status --dev-root "$GHC_DEV_ROOT"
chc daemon start --dev-root "$GHC_DEV_ROOT" --max-iterations 1
chc daemon status --dev-root "$GHC_DEV_ROOT"
chc daemon stop --dev-root "$GHC_DEV_ROOT"
```

This validates local coordinator state. It does not create a hosted service.

## Notifications and Remote Command Intake

Inspect notification configuration:

```sh
chc notify status --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
```

Preview a Telegram payload without contacting Telegram:

```sh
chc notify telegram test --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT" --dry-run
```

Create a fixture file under `$GHC_ARTIFACT_DIR/telegram-updates.json` with test
Telegram updates, then parse it:

```sh
chc notify telegram poll \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --fixture "$GHC_ARTIFACT_DIR/telegram-updates.json" \
  --dry-run

chc notify telegram poll \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --fixture "$GHC_ARTIFACT_DIR/telegram-updates.json"
```

The expected result is a stored accepted or rejected remote command handoff
record. This stage validates intake, not live Telegram operation.

## Maintenance and Recovery

Write a diagnostic bundle:

```sh
chc maintenance bundle --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
```

Inspect local state:

```sh
chc maintenance db status --dev-root "$GHC_DEV_ROOT"
chc maintenance registry status --dev-root "$GHC_DEV_ROOT"
chc maintenance execution artifacts --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
chc maintenance execution cleanup-preview --dev-root "$GHC_DEV_ROOT" --dir "$GHC_WORK_PROJECT"
```

Preview registry repair commands before using the write variants:

```sh
chc maintenance registry disable --dev-root "$GHC_DEV_ROOT" --dry-run chc-blog-lab
chc maintenance registry set-path --dev-root "$GHC_DEV_ROOT" --dry-run chc-blog-lab "$GHC_WORK_PROJECT"
```

Copy the diagnostic bundle into `$GHC_ARTIFACT_DIR` after each run that exposes
a failure.

## Proving-Ground Check

If the work project includes a proving objective, run the proving commands
against it:

```sh
chc proving validate \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --objective .chc/proving/objectives/CLI-BLOG-1-fix-publish-index.md

chc proving run \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --objective .chc/proving/objectives/CLI-BLOG-1-fix-publish-index.md \
  --dry-run

chc proving ralph \
  --dev-root "$GHC_DEV_ROOT" \
  --dir "$GHC_WORK_PROJECT" \
  --objective .chc/proving/objectives/CLI-BLOG-1-fix-publish-index.md \
  --max-attempts 3
```

Use this only as a direct sandbox proof. The main workflow remains tickets.

## Failure Ledger

Create a run-specific ledger:

```sh
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
LEDGER="$GHC_LOG_DIR/$RUN_ID-failures.md"
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

If the failure is in ChrysoCoord, create a ChrysoCoord ticket after the run. If
the failure is only in the lab fixture, fix the fixture or reset it from the
reference project.

## Run Completion Criteria

A useful full-system lab run produces:

- a resettable `chc-blog-work` project;
- a clean way to restore it from `chc-blog-ref`;
- initialized `.chc/` project-local state;
- isolated user-level state under `$GHC_DEV_ROOT`;
- authoring session, quiz, promotion, and assisted authoring artifacts;
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
