# Web UI Action Matrix

This matrix tracks the first Web UI surface against the existing CLI and TUI
operator surfaces.

Status values:

- `matched`: the Web UI exposes equivalent operator behavior through the same
  service path or read model.
- `read_only_matched`: the Web UI exposes equivalent inspection behavior but no
  mutation.
- `web_subset`: the Web UI intentionally exposes a narrower first action set.
- `cli_tui_only`: the action remains outside the Web UI.
- `web_only`: the action is specific to starting or using the Web UI.

## Navigation And Views

| Domain | Web UI entrypoint | CLI entrypoint | TUI entrypoint | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Web server | `chr web serve` | `chr web serve` | none | `web_only` | Explicit foreground process; no automatic startup from CLI or TUI workflows. |
| Project | `/`, `/project` (direct URL only) | `chr project current`, metadata commands | `F1` | `read_only_matched` | Shows the project name, root directory and secondary diagnostic ID; no visible navigation link or automatic polling. |
| Cases | `/cases` | `chr metadata case set` for writes | `F2` | `matched` | Catalog and detail read project-local case metadata; detail form writes case metadata. |
| Dumps | `/dumps` | `chr dump *` | `F3` | `matched` | Catalog, detail, create, save, and append use dump service paths. |
| Quizzes | `/quizzes` | `chr quiz *` | `F4` | `web_subset` | Start and resume are exposed; answer, skip, complete, cancel, and mode-only updates remain CLI/TUI. |
| Specs | `/specs` | `chr spec list`, `chr spec generate` | `F5` | `matched` | Catalog and generation action use spec service paths. |
| Plans | `/plans` | `chr plan *` | `F6` | `matched` | Catalog, generate, regenerate, and tracking creation use plan service paths. |
| Board | `/board` | `chr ticket board`, `chr ticket list`, `chr ticket show` | `F7` | `read_only_matched` | Board columns, attention, and ticket cards read the shared Board model. |
| Dash | `/dash` | `chr work doctor`, `chr work run-all --dry-run`, `chr daemon stop` | `F8` | `web_subset` | Readiness, queue, current, recent events, work preview, and daemon stop are exposed. |

## Write Actions

| Domain | Web action | CLI command | TUI action | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Case metadata | `case.update` | `chr metadata case set` | Cases save | `matched` | Uses stale revision checks and writes project-local case metadata. |
| Dump create | `dump.create` | `chr dump new` | Dumps create | `matched` | One New form selects Project-wide or an existing case. Retained capture identities make retries safe; general captures extend the shared dump while case captures remain independent. |
| Dump save | `dump.save` | `chr dump save` | Dumps editor save | `matched` | Uses revision checks before replacing content. |
| Dump append | `dump.append` | `chr dump append` | Dump append action | `matched` | Appends to the selected dump through the dump service. |
| Spec generate | `spec.generate` | `chr spec generate` | Specs generate | `matched` | Uses the same generation service and artifact contract. |
| Plan generate | `plan.generate` | `chr plan generate` | Plans generate | `matched` | Uses the same generation service and existing-plan guard. |
| Plan regenerate | `plan.regenerate` | `chr plan regenerate` | Plans regenerate | `matched` | Requires browser confirmation before replacement. |
| Plan tracking create | `plan.tracking.create` | `chr plan tracking create` | Plans tracking action | `matched` | Materializes tracking artifacts and tickets from an existing plan. |
| Quiz start | `quiz.start` | `chr quiz start` | Quizzes start | `matched` | Supports project or case scope and deep or fast mode. |
| Quiz resume | `quiz.resume` | `chr quiz resume` | Quizzes resume | `matched` | Resumes the selected project or case quiz scope. |
| Work preview | `work.preview` | `chr work run-all --dry-run` | Dash work preview | `matched` | Previews queue selection without starting agent execution. |
| Daemon stop | `work.daemon.stop` | `chr daemon stop` | Dash stop request | `matched` | Requests stop through the shared daemon stop path. |

## Boundaries

- Web UI state is never authoritative. It reloads from project state after CLI,
  TUI, or Web UI changes.
- Web UI writes use the same service layer as the CLI and TUI where the action
  is marked `matched`.
- Browser confirmation is required for overwrite-prone and work-control forms.
- Stale writes return reload and discard choices instead of overwriting newer
  project state.
- Administrative commands, project registry management, maintenance,
  notification setup, assisted authoring, proving workflows, and full ticket
  execution remain CLI/TUI-only in this rollout.
