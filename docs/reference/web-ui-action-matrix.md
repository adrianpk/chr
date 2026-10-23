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
| Harness defaults | `/project/harnesses`, project context | none | Project `h` | `matched` | Separate worker/generation choices, local readiness and retained worker bindings use the shared selection projection. |
| Cases | `/cases` | `chr metadata case set` for writes | `F2` | `matched` | New creates a case from its name, description and context; detail edits existing metadata. Creation generates the slug automatically. |
| Dumps | `/dumps` | `chr dump *` | `F3` | `matched` | Catalog, capture and one body editor use the shared dump service; append remains a transport operation. |
| Quizzes | `/quizzes` | `chr quiz *` | `F4` | `matched` | Opening an unstarted scope prepares the first question through a guarded browser POST, with visible progress and retry on failure. Active sessions retain their question across reloads. The panel groups Deep/Fast, the answer editor and Skip/Answer/Cancel/Done; history appears after answers exist. |
| Specs | `/specs` | `chr spec list`, `chr spec generate` | `F5` | `matched` | Individual manual/generated documents share New/Edit, initial Draft/approved state and revision-safe Save through the authoring owner. |
| Plans | `/plans` | `chr plan *` | `F6` | `web_subset` | Concrete plan documents have fixed cards and exact full read-only details. Missing-scope placeholders and internal trackers are excluded. Generate Plan is available in Plans and creates the plan, native task tickets and internal order map together. Cases remain metadata editors. |
| Board | `/board` | `chr ticket board`, `chr ticket list`, `chr ticket show` | `F7` | `matched` | Six lanes, attention and paging share one New/Edit form, guarded moves and independent comments. Plan-bound titles and execution-owned fields show their guards. |
| Dash | `/dash` | `chr work doctor`, `chr work run-all --dry-run`, `chr daemon stop` | `F8` | `web_subset` | Readiness, paged queue, independent current contexts and retained results refresh with stale recovery. Work preview and daemon stop retain separate confirmed controls. |

## Write Actions

| Domain | Web action | CLI command | TUI action | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Harness defaults | `harnesses.change` at `/project/harnesses` | none | Project `h`, Enter | `matched` | Named profile selection and clear use the shared configuration owner, exact project/revision/request identity and existing login/origin guards. |
| Guided development | `guided.start`, `guided.input` at `/guided/actions` | shared application operations | shared application operations | `web_subset` | A new guided run retains development intent. After explicit capture completion and sufficient answers it approves the exact generated specification, prepares and seals native work, and admits/activates only that materialized set under native policy. Retried operations reconcile original identities. Standalone authoring remains independent. |
| Guided progress | GET `/guided` | native workflow status | native workflow status | `web_subset` | Read-only committed document and native progress queries distinguish validated tasks, confirmed delivery, remaining work, admitted limits and review waits. Rendering never activates execution. |
| Guided controls | `guided.input` with `pause`, `resume` or `cancel` | shared application operations | shared application operations | `web_subset` | Fresh controls bind displayed authoring/native revisions. Retained native receipts reconcile uncertain results; pending stage input and admitted limits survive stops. Requested stop and native quiescence are distinct. Managed host startup performs recovery independently of queries and client connections. |
| Case create | `case.create` | `chr metadata case set` | Cases New | `matched` | Uses the shared metadata service to generate an available slug and atomically save the new case. The case becomes available in dump scope choices. |
| Case metadata | `case.update` | `chr metadata case set` | Cases save | `matched` | Uses stale revision checks and writes project-local case metadata. |
| Dump create | `dump.create` | `chr dump new` | Dumps create | `matched` | One New form selects Project-wide or an existing case. Retained capture identities make retries safe; general captures extend the shared dump while case captures remain independent. |
| Dump save | `dump.save` | `chr dump save` | Dumps editor save | `matched` | One Body editor and Save use the canonical serialized revision guard, preserve metadata and retain input on failures. |
| Dump append | `dump.append` (transport only) | `chr dump append` | none | `web_subset` | Shared append remains available to capture adapters. The manual Web editor adds text in Body and Save. |
| Spec generate | `spec.generate` | `chr spec generate` | Specs generate | `web_subset` | Generate selects an existing named use case; shared project context accompanies its saved dumps and quiz answers. Project-wide/unknown targets are rejected before model invocation. Pending feedback and disabled controls prevent repeated submission. Success opens the saved document in its editor and adds its catalog card; generation results and errors require explicit dismissal. |
| Plan generate | `plan.generate` in Plans | `chr plan generate` | Plans generate | `web_subset` | The shared delivery operation prepares a named case plan and its native task tickets in one submission. A saved spec and sufficient discovery are required. Existing plans are kept; retries reuse committed publication and ticket identities. No worker starts. |
| Plan regenerate | no visible control | `chr plan regenerate` | Plans regenerate | `cli_tui_only` | The first Web delivery flow preserves existing plans. Replacement remains a distinct existing operation. |
| Plan tracking create | none | `chr plan tracking create` | Legacy Plans tracking action | `cli_tui_only` | Generate Plan retains and binds internal order maps while preparing native tickets, without a second tracking or approval step. The former standalone Web tracking action is unsupported; direct tracker URLs cannot expose the files. |
| Quiz start | `quiz.start` | `chr quiz start` | Quizzes start | `matched` | Supports project or case scope and deep or fast mode. |
| Quiz resume | `quiz.resume` | `chr quiz resume` | Quizzes resume | `matched` | Resumes the selected project or case quiz scope. |
| Quiz answering | `quiz.answer`, `quiz.skip`, `quiz.done`, `quiz.cancel`, `quiz.mode` | `chr quiz answer`, `skip`, `complete`, `cancel`, `mode` | Focused quiz actions | `matched` | Retains question/session binding, idempotent actions, confirmed history and real Deep/Fast queues. Deep uses full scoped discovery history, honors completion recommendations and closes at the safety budget; repeated answered prompts block continuation without losing answers. |
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
  notification setup, proving workflows, and general native-work administration
  remain CLI/TUI-only in this rollout. Guided automatic execution uses the
  bounded path above.

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
