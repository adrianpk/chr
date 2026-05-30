# QA Agent Work Scenario

This runbook exercises the seeded blog QA scenario from reset through CHC
artifact generation and ticket execution.

## Reset The Scenario

From the ChrysoCoord repository:

```sh
make qa-reset-blog
```

This recreates `qa/project/blog` from `qa/ref/001/blog`, moves any previous
active copy to `qa/tmp/blog`, and initializes an independent Git repository in
the active project.

Run the baseline tests:

```sh
make qa-test-blog
```

The baseline is expected to fail. The seeded bug is that an unknown post returns
`500` instead of `404`.

Run the smoke check:

```sh
make qa-smoke-blog
```

The smoke check resets the scenario, confirms the active project is ignored by
the parent repository, confirms it has an independent Git repository, and
confirms the expected failing test is still present.

## Open CHC

Open CHC against the active scenario:

```sh
cd qa/project/blog
chc
```

The project contains seeded metadata, dumps, quiz notes, ADRs, and one manual
bug ticket. It intentionally does not contain generated specs, plans, tracking,
or generated comments tickets.

## Resolve The Seeded Bug Ticket

Use the TUI or CLI to run the manual ticket:

```sh
chc work ticket CHC-1
```

The agent should reproduce the failing unknown-post test, fix the status code,
and leave validation evidence in the ticket workflow.

Agent work initializes missing runtime assets automatically before execution.

## Generate Comments Work

After the bug ticket flow is validated, generate comments artifacts:

```sh
chc spec generate --use-case comments
chc plan generate --use-case comments
chc plan tracking create --use-case comments
```

The generated spec and plan should use the seeded dumps, quiz notes, project
metadata, and ADRs under `docs/explanation/adr/`.

## Run Generated Tickets

Preview generated ready work:

```sh
chc work run-all --dry-run --use-case comments --order key
```

Run the generated ticket loop:

```sh
chc work run-all --use-case comments --order key
```

Use this scenario to observe whether CHC can preserve context from seeded
inputs into specs, plans, tracking, tickets, and implementation work.
