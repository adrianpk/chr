# Commands

This page lists the main command surfaces. Use `chc --help` and command-specific
help for exact flags.

Running `chc` without a subcommand opens the TUI for the resolved project.

## Workspace

Initialize a managed project:

```sh
chc init
```

Preview initialization:

```sh
chc init --dry-run
```

## TUI

Open the local operational console:

```sh
chc
chc tui
chc tui --dir /path/to/project
```

See [TUI reference](tui.md) for views, keybindings, command palette behavior,
and confirmation modes.

See [CLI TUI action matrix](cli-tui-action-matrix.md) for current parity
between command-line and interactive operations.

## Runtime

Install missing runtime defaults:

```sh
chc runtime init
```

Validate the runtime boundary:

```sh
chc runtime validate
```

Preview validation:

```sh
chc runtime validate --dry-run
```

## Projects

Register and select local projects:

```sh
chc project add /path/to/project
chc project select /path/to/project
chc project current
chc project list
```

## Tickets

Create and inspect tickets:

```sh
chc ticket new "Add completed task support" --type feature --priority high
chc ticket list
chc ticket show CHC-1
```

Update ticket state:

```sh
chc ticket comment CHC-1 "Validation fails in the completion path."
chc ticket status CHC-1 ready
```

Start one ticket:

```sh
chc ticket start CHC-1
```

Preview one ticket execution:

```sh
chc ticket start CHC-1 --dry-run
```

Apply an execution result:

```sh
chc ticket apply-result <execution-id>
```

Run a filtered foreground loop:

```sh
chc ticket loop --status ready --priority high --max-iterations 3
```

Preview loop selection:

```sh
chc ticket loop --dry-run
```

## Index

Rebuild and inspect the ticket index:

```sh
chc index rebuild
chc index status
```

## Maintenance

Inspect local maintenance state:

```sh
chc maintenance bundle
chc maintenance db status
chc maintenance execution artifacts
chc maintenance execution cleanup-preview
chc maintenance registry status
chc maintenance registry disable example
chc maintenance registry set-path example /path/to/project
```

## Authoring

Create an authoring session:

```sh
chc author session new --target spec --topic "Billing sync"
```

Generate a context-aware quiz:

```sh
chc author quiz new --target spec --topic "Billing sync"
```

Generate or edit a spec with `status: draft` when quiz answers still need
review. Legacy `quiz-to-draft` exists only for compatibility with older
authoring sessions.

Run the current assisted authoring flow:

```sh
chc author assisted start --target spec --topic "Billing sync"
chc author assisted resume --session <session-handle>
chc author assisted answer --session <session-handle> --turn <turn-id> --answer "Retry writes must be idempotent."
chc author assisted status --session <session-handle>
chc author assisted finalize --session <session-handle> --target-kind spec --target-name "billing sync spec"
```

The current assisted commands use explicit adapter inputs for deterministic
local behavior.

The legacy quiz-to-draft command remains available for older authoring sessions:

```sh
chc author promote quiz-to-draft --session <session-id> --quiz <quiz-path>
```

## Daemon

Run local foreground coordination:

```sh
chc daemon status
chc daemon start --max-iterations 1
chc daemon stop
```

## Notify

Inspect notification configuration:

```sh
chc notify status
```

Preview the Telegram setup-check payload:

```sh
chc notify telegram test --dry-run
```

Send the Telegram setup-check message:

```sh
chc notify telegram test
```

The dry-run command does not contact Telegram. Non-dry-run dispatch reads the
bot token from the configured environment variable and records a notification
audit event.

Preview Telegram update parsing from a fixture:

```sh
chc notify telegram poll --fixture updates.json --dry-run
```

Persist accepted and rejected remote command handoff records from a fixture:

```sh
chc notify telegram poll --fixture updates.json
```

## Proving

Run direct proving-ground workflows against controlled objectives:

```sh
chc proving validate --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md
chc proving run --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md --dry-run
chc proving ralph --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md --max-attempts 3
```

Proving commands are for sandbox proofing and validation, not the normal ticket
execution path.
