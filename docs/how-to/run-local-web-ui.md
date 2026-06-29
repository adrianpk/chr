# Run the Local Web UI

The Web UI is an explicit local process. It does not start from `chc`,
`chc tui`, ticket loops, agent work, or other normal commands.

## Preview Configuration

From a managed project:

```sh
chc web serve --dry-run
```

Expected default boundary:

- bind: `127.0.0.1:7331`
- token policy: `auto`
- token required: `false`
- origins: same-origin only
- assets: embedded

## Start The Server

```sh
chc web serve
```

Open:

```text
http://127.0.0.1:7331
```

Use a different project directory:

```sh
chc web serve --dir /path/to/project
```

Use a different local port:

```sh
chc web serve --port 7340
```

Stop the server with `ctrl+c` in the terminal that started it.

## Local Workflow

Use the Web UI as another view over the same project-local state:

- Project shows the resolved workspace and Web UI runtime boundary.
- Board and Dash show ticket, readiness, queue, and runtime status.
- Cases, Dumps, Quizzes, Specs, and Plans read the same Markdown artifacts used
  by CLI and TUI workflows.
- Supported forms write through the same application service paths as the CLI
  and TUI.
- Pages and fragments refresh from project state; browser state is presentation
  cache only.

## Smoke Validation

Run the repeatable Web UI smoke script from the repository root:

```sh
scripts/qa/web-ui-smoke.sh
```

The script creates a temporary managed project, starts `chc web serve`, checks
the main pages and fragments, posts a dump creation action, and captures
browser screenshots when Chromium or Chrome is available.
