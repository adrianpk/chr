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
| Harness defaults | `/project/harnesses`, project context | none | none | `web_only` | Separate worker/generation choices, local readiness and retained worker bindings use the shared selection projection. |
| Cases | `/cases` | `chr metadata case set` for writes | `F2` | `matched` | Catalog and detail read project-local case metadata; detail form writes case metadata. |
| Dumps | `/dumps` | `chr dump *` | `F3` | `matched` | Catalog, capture and one body editor use the shared dump service; append remains a transport operation. |
| Quizzes | `/quizzes` | `chr quiz *` | `F4` | `matched` | Named pending scopes and retained answering support Start, Resume, Answer, Skip, Done, Cancel and Deep/Fast mode changes with question/revision guards. |
| Specs | `/specs` | `chr spec list`, `chr spec generate` | `F5` | `matched` | Individual manual/generated documents share New/Edit, initial Draft/approved state and revision-safe Save through the authoring owner. |
| Plans | `/plans` | `chr plan *` | `F6` | `web_subset` | Concrete plan documents have fixed cards and exact full read-only details. Missing-scope placeholders and internal trackers are excluded. Generate/Regenerate stay in the selected case. |
| Board | `/board` | `chr ticket board`, `chr ticket list`, `chr ticket show` | `F7` | `matched` | Six lanes, attention and paging share one New/Edit form, guarded moves and independent comments. Plan-bound titles and execution-owned fields show their guards. |
| Dash | `/dash` | `chr work doctor`, `chr work run-all --dry-run`, `chr daemon stop` | `F8` | `web_subset` | Readiness, paged queue, independent current contexts and retained results refresh with stale recovery. Work preview and daemon stop retain separate confirmed controls. |

## Write Actions

| Domain | Web action | CLI command | TUI action | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Harness defaults | `harnesses.change` at `/project/harnesses` | none | none | `web_only` | Named profile selection and clear use the shared configuration owner, exact project/revision/request identity and existing login/origin guards. |
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
| Quiz answering | `quiz.answer`, `quiz.skip`, `quiz.done`, `quiz.cancel`, `quiz.mode` | `chr quiz answer`, `skip`, `complete`, `cancel`, `mode` | Focused quiz actions | `matched` | Retains question/session binding, idempotent actions, confirmed history and real Deep/Fast queues. |
| Manual spec authoring | `spec.create`, `spec.save` | none | none | `web_only` | Manual and generated documents use shared revision-safe authoring operations; generation remains a separate action. |
| Ticket editing | `ticket.create`, `ticket.update`, `ticket.move`, `ticket.comment` | `chr ticket *` | Board ticket actions | `matched` | Shared document ownership and revision guards preserve native contracts, comments and delivery references. |
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
