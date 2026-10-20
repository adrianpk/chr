# Run the Local Web UI

The Web UI is an explicit local process. It does not start from `chr`,
`chr tui`, ticket loops, agent work, or other normal commands.

## Preview Configuration

From a managed project:

```sh
chr web serve --dry-run
```

Expected default boundary:

- bind: `127.0.0.1:7331`
- token policy: `auto`
- token required: `false`
- origins: same-origin only
- assets: embedded

## Start The Server

```sh
chr web serve
```

Open:

```text
http://127.0.0.1:7331
```

Use a different project directory:

```sh
chr web serve --dir /path/to/project
```

Use a different local port:

```sh
chr web serve --port 7340
```

Stop the server with `ctrl+c` in the terminal that started it.

For direct access from another computer on a trusted internal network, run
`chr web serve --allow-lan`. This explicit mode requires no token and defaults
to disabled. See [Expose the Local Web UI Deliberately](expose-local-web-ui.md)
for the browser URL and effective bind policy.

## Local Workflow

To protect the UI with one local account, follow
[Configure Single-user Web Login](configure-web-login.md). A configured account
applies to both local and LAN access. Use `--require-login` to refuse startup
without it.

The shared header identifies the current project. Page headings use only the
section or artifact title.

Use the Web UI as another view over the same project-local state:

- Project is available by entering `/project` directly. It shows only the
  project name, root directory and diagnostic project ID, and has no navigation
  link or automatic refresh.
- Board and Dash show ticket, readiness, queue, and runtime status.
- Cases, Dumps, Quizzes, Specs, and Plans read the same Markdown artifacts used
  by CLI and TUI workflows.
- Supported forms write through the same application service paths as the CLI
  and TUI.
- Pages and fragments refresh from project state; browser state is presentation
  cache only.

The Board opens ticket details in a dialog where you can edit the title and
description or add a comment. Ticket writes require the revision loaded by
that dialog. If another edit changes the ticket, reload the board and reopen
the ticket before submitting again.

## Smoke Validation

Run the repeatable Web UI smoke script from the repository root:

```sh
scripts/qa/web-ui-smoke.sh
```

The script creates a temporary managed project, starts `chr web serve`, checks
the main pages and fragments, posts a dump creation action, and captures
browser screenshots when Chromium or Chrome is available.
