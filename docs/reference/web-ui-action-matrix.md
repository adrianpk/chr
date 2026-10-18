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
| Dumps | `/dumps` | `chr dump *` | `F3` | `matched` | Catalog, capture and one body editor use the shared dump service; append remains a transport operation. |
| Quizzes | `/quizzes` | `chr quiz *` | `F4` | `web_subset` | Start and resume are exposed; answer, skip, complete, cancel, and mode-only updates remain CLI/TUI. |
| Specs | `/specs` | `chr spec list`, `chr spec generate` | `F5` | `matched` | Catalog and generation action use spec service paths. |
| Plans | `/plans` | `chr plan *` | `F6` | `web_subset` | Concrete plan documents have fixed cards and exact full read-only details. Missing-scope placeholders and internal trackers are excluded. Generate/Regenerate stay in the selected case. |
| Board | `/board` | `chr ticket board`, `chr ticket list`, `chr ticket show` | `F7` | `read_only_matched` | Board columns, attention, and ticket cards read the shared Board model. |
| Dash | `/dash` | `chr work doctor`, `chr work run-all --dry-run`, `chr daemon stop` | `F8` | `web_subset` | Readiness, paged queue, independent current contexts and retained results refresh with stale recovery. Work preview and daemon stop retain separate confirmed controls. |

## Write Actions

| Domain | Web action | CLI command | TUI action | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Case metadata | `case.update` | `chr metadata case set` | Cases save | `matched` | Uses stale revision checks and writes project-local case metadata. |
| Dump create | `dump.create` | `chr dump new` | Dumps create | `matched` | One New form selects Project-wide or an existing case. Retained capture identities make retries safe; general captures extend the shared dump while case captures remain independent. |
| Dump save | `dump.save` | `chr dump save` | Dumps editor save | `matched` | One Body editor and Save use the canonical serialized revision guard, preserve metadata and retain input on failures. |
| Dump append | `dump.append` (transport only) | `chr dump append` | none | `web_subset` | Shared append remains available to capture adapters. The manual Web editor adds text in Body and Save. |
| Spec generate | `spec.generate` | `chr spec generate` | Specs generate | `matched` | Uses the same generation service and artifact contract. |
| Plan generate | `plan.generate` | `chr plan generate` | Plans generate | `matched` | Uses the same generation service and existing-plan guard. |
| Plan regenerate | `plan.regenerate` | `chr plan regenerate` | Plans regenerate | `matched` | Requires browser confirmation before replacement. |
| Plan tracking create | none | `chr plan tracking create` | Legacy Plans tracking action | `cli_tui_only` | Internal maps are retained during generation and bound through native approval. The former Web action is unsupported; direct tracker URLs cannot expose the files. |
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

## Plan Document Reads

The central operation service lists concrete canonical plan artifacts across
namespaces, independently of latest-in-scope generation lookup. Multiple
same-scope documents and retained archives have distinct references. Reads are
bounded to 4096 inventory entries, 8 MiB in aggregate and the existing 1 MiB
per-file limit. Corrupt envelopes, malformed frontmatter, pending publication
and symlink traversal fail explicitly; reads do not recover or migrate files.

Plan previews omit managed frontmatter and a duplicate opening title. Details
show the complete body through escaped text rendering. Only canonical plan
body references are accepted at `/plans/detail` and its fragment route; missing
or foreign references return HTTP 404. Retained tracker companions cannot be
selected through these routes. Internal tracker storage and CLI/TUI operations
remain under their existing owners.
