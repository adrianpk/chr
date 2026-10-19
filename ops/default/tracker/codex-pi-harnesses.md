# Selectable Harness Delivery Order

Status: Approved

Date: 2026-10-19

Delivery set: codex-pi-harnesses

Scope approval: maintainer, 2026-10-19; approved design and requested slicing.

Approved: maintainer, 2026-10-19; explicit instruction to execute this set through Looper.

Plan: [Selectable Harness Delivery](../plan/codex-pi-harnesses.md).

Spec: [Selectable Codex and Pi Harnesses](../spec/stable/codex-pi-harnesses.md).

Model: [Harness Selection and Retained Bindings](../spec/stable/codex-pi-harnesses-model.md).

Implementation authorization: granted on 2026-10-19 for this exact eight-slice
set through Looper, including controller-owned Rebase + Fast-forward merges
into dev. This document remains an order map; tickets and reports own progress.

First planned slice: Slice 1, Neutral contracts and Codex extraction; T1.1.

Slice strategy: behavior-first; the plan records the reason.

## Ordering Contract

Read slices in ascending numeric order and tasks in their listed order.
Each slice uses exactly one ticket, branch, PR targeting dev and report.
Integrate the preceding slice before starting its successor. Task IDs are
commit-sized instructions inside that ticket, not separate tickets.

The original feature ticket is Slice 8's acceptance unit; Slices 1-7 have
separate bounded tickets. This map has no progress checkboxes, task-state
columns, implementation commits, PR status or mutable completion counters.
Those facts belong to the ticket and report. Update this map only when scope,
order, dependency or planned delivery identity is deliberately amended.

New configuration applies to new operations. An admitted invocation retains
its worker binding for all pending tickets/retries/resumes, and one generation
operation retains its binding across nested calls. All four interfaces use the
same shared persisted revision; no interface invents its own live setting.

## Canonical Slice Map

| Slice | Short name | Ticket | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | Neutral contracts and Codex extraction | [TKT-20261019103827](../ticket/solved/20261019103827-harness-codex.md) | `refactor/slice-1-harness-codex` | `refactor(slice-1): extract Codex behind neutral harness contracts` | `ops/default/report/slices/codex-pi-harnesses/slice-1-harness-codex.md` |
| Slice 2 | Live role selection and retained bindings | [TKT-20261019103828](../ticket/solved/20261019103828-harness-selection.md) | `feat/slice-2-harness-selection` | `feat(slice-2): retain live harness selections and invocation bindings` | `ops/default/report/slices/codex-pi-harnesses/slice-2-harness-selection.md` |
| Slice 3 | Pi RPC and structured generation | [TKT-20261019103829](../ticket/solved/20261019103829-pi-generation.md) | `feat/slice-3-pi-generation` | `feat(slice-3): generate structured content through the Pi adapter` | `ops/default/report/slices/codex-pi-harnesses/slice-3-pi-generation.md` |
| Slice 4 | Pi Docker workers and private runtime setup | [TKT-20261019103830](../ticket/solved/20261019103830-pi-worker-runtime.md) | `feat/slice-4-pi-worker-runtime` | `feat(slice-4): execute Pi workers in their selected Docker runtime` | `ops/default/report/slices/codex-pi-harnesses/slice-4-pi-worker-runtime.md` |
| Slice 5 | Web harness selection | [TKT-20261019103831](../ticket/reviewing/20261019103831-web-harness-selection.md) | `feat/slice-5-web-harness-selection` | `feat(slice-5): change project harness selections from the Web UI` | `ops/default/report/slices/codex-pi-harnesses/slice-5-web-harness-selection.md` |
| Slice 6 | TUI harness selection | [TKT-20261019103832](../ticket/ready/20261019103832-tui-harness-selection.md) | `feat/slice-6-tui-harness-selection` | `feat(slice-6): change project harness selections from the TUI` | `ops/default/report/slices/codex-pi-harnesses/slice-6-tui-harness-selection.md` |
| Slice 7 | Telegram harness selection | [TKT-20261019103833](../ticket/ready/20261019103833-telegram-harness-selection.md) | `feat/slice-7-telegram-harness-selection` | `feat(slice-7): change project harness selections through Telegram` | `ops/default/report/slices/codex-pi-harnesses/slice-7-telegram-harness-selection.md` |
| Slice 8 | Cross-interface acceptance and operator guidance | [TKT-20261017203525](../ticket/ready/20261017203525-support-pi-alongside-codex.md) | `test/slice-8-harness-acceptance` | `test(slice-8): verify selectable harnesses and live configuration end to end` | `ops/default/report/slices/codex-pi-harnesses/slice-8-harness-acceptance.md` |

## Ordered Ticket and Task Map

### Slice 1: Neutral contracts and Codex extraction

Ticket: [TKT-20261019103827](../ticket/solved/20261019103827-harness-codex.md).

Prerequisite: committed planning inputs and execution activation.

| Task | Ordered work |
| --- | --- |
| T1.1 | Extract the technical request/result/capability and bounded process contracts into internal/harness and process; implement the selected byte ceilings, cancellation, owned-process cleanup and sanitized diagnostics without moving domain rules. |
| T1.2 | Move Codex generation and worker command/transport behavior into the sibling codex package; keep generic workflow input, prompts, schemas and validators in core and remove permanent compatibility aliases. |
| T1.3 | Inject Codex through app into every existing caller, including contextual Quiz, structured conversation, Dump curation, review and worker invocation; prove unchanged direct API defaults and the common adapter contract. |

Validation: Use owned subprocess and filesystem fixtures for generation and worker success, schema failure, cancellation, redaction and bounded output. Check the complete caller/import inventory and a Codex-only installation.

### Slice 2: Live role selection and retained bindings

Ticket: [TKT-20261019103828](../ticket/solved/20261019103828-harness-selection.md).

Prerequisite: Slice 1 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T2.1 | Resolve private named profiles and workspace role choices with the documented legacy precedence; return readable readiness and effective selections, preserve omitted/direct-API behavior and reject invalid or unsupported explicit choices. |
| T2.2 | Implement shared inspect/change operations and CLI status/select/clear adapters with authorized actors, whole-document revision comparison, cross-process exclusion, atomic two-role updates and bounded request-id replay; preserve unrelated configuration and unknown fields. |
| T2.3 | Add chr-invocation/v2 binding and validation before worker admission, preserving byte-identical v1 checkpoints and their historical command semantics; retain selected profile, model, runtime/executable and revision evidence for all pending tickets and retries. |
| T2.4 | Freeze generation selection once per authoring/conversation operation and thread it through nested calls, later stages and retries; fresh operations read persisted settings, active operations never rebind, and existing capacity/budget authorities stay shared. |

Validation: Exercise two independent processes, simultaneous admissions and switches, stale/unauthorized/invalid updates, clear-to-legacy, crash/replay, unchanged unrelated fields, active pending tickets and multi-call generation. An old v1-only reader must reject v2 before spawning.

### Slice 3: Pi RPC and structured generation

Ticket: [TKT-20261019103829](../ticket/solved/20261019103829-pi-generation.md).

Prerequisite: Slice 2 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T3.1 | Implement pinned Pi RPC framing, correlated requests and settled completion using the common process owner; distinguish prompt acknowledgement, intermediate events, provider failure, abort, malformed/oversized frames and early exit. |
| T3.2 | Map normalized final output and available usage to existing domain validators; wire every generation family and retain operation-level selection, schemas, deadlines, fast/deep Quiz behavior and shared model capacity. |
| T3.3 | Enforce generation resource/tool discovery restrictions and private per-call state; cover unsupported capability and uncertain usage without fallback, false token certification or artifact replacement after invalid output. |

Validation: Use deterministic RPC subprocess fixtures for fragmented LF framing, Unicode separators, interleaved commands, retry/compaction before agent_settled, backpressure and cancellation. Exercise every authoring family through its real shared service with schema-valid and rejected outputs; a real paid Pi call remains a final opt-in acceptance item.

### Slice 4: Pi Docker workers and private runtime setup

Ticket: [TKT-20261019103830](../ticket/solved/20261019103830-pi-worker-runtime.md).

Prerequisite: Slice 3 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T4.1 | Add sibling Pi runtime assets using the resolved pins below and a committed npm dependency lock; update all bundled profile initialization and explicit upgrade/readiness paths while preserving existing Codex Dockerfile/image references. |
| T4.2 | Execute the selected worker bridge and Pi RPC inside the selected container through owned pipes; mount only explicit profile credentials/model settings, allocate private writable call state and implement tested OAuth refresh exclusion without importing host home or arbitrary resources. |
| T4.3 | Retain invocation/claim/retry/result-validation authority and shared capacity; enforce finite-budget and merger-broker capability restrictions, verify container-reachable custom endpoints and provide actionable setup failures with no automatic fallback. |

Validation: Build and inspect one selected Pi image, exercise a real container with a deterministic local protocol/model endpoint, and run two isolated calls against the shared capacity boundary. Verify only the intended mounts/flags and credential owner. An authorized real-model disposable task supplies HP-08 during final acceptance.

### Slice 5: Web harness selection

Ticket: [TKT-20261019103831](../ticket/reviewing/20261019103831-web-harness-selection.md).

Prerequisite: Slice 4 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T5.1 | Add the compact project-context profile chooser using shared readable selection/readiness projections; show role defaults separately from active invocation bindings, support clearing a role and state that changes apply to new operations. |
| T5.2 | Submit exact revision-bound changes through existing authenticated mutation/origin guards; preserve input on rejection and verify login, stale updates, project scope and the absence of a service restart or credential/path editor. |

Validation: Exercise real HTTP action/authentication boundaries and one focused browser journey on wide and narrow layouts, including a switch while an invocation remains active. Inspect the persisted result from an independent process.

### Slice 6: TUI harness selection

Ticket: [TKT-20261019103832](../ticket/ready/20261019103832-tui-harness-selection.md).

Prerequisite: Slice 5 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T6.1 | Add the project-context role/profile picker with readable current defaults, readiness and separately retained active bindings; reuse the central operation service and local actor mapping. |
| T6.2 | Handle keyboard selection, clear, cancel and stale/conflicting saves without losing the visible choice; verify that Web/CLI changes become visible on the next refresh/admission and no TUI-owned configuration store exists. |

Validation: Exercise keyboard/event-driven update behavior through the real TUI adapter with temporary workspace configuration, cancellation, stale revision and independent process changes; retain existing compact layout behavior.

### Slice 7: Telegram harness selection

Ticket: [TKT-20261019103833](../ticket/ready/20261019103833-telegram-harness-selection.md).

Prerequisite: Slice 6 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T7.1 | Add a deterministic /harness inspection/selection entrypoint and concise help with readable role/profile/model names and active-versus-new-work semantics; use the selected registered project. |
| T7.2 | Bind confirmation to the mapped actor, exact project and role/profile choices, observed revision and request identity; reject stale, foreign, replayed-with-changed-arguments or unauthorized requests through the existing shared mutation owner. |
| T7.3 | Verify transport redelivery/lost responses and cross-interface visibility while active generation and workers keep their bindings; no model-only instruction may authorize changing configuration or private setup. |

Validation: Use deterministic Telegram updates/callbacks through the existing receiver and shared service. Cover stale confirmation, changed selected project, actor mismatch and replay without sending messages to the live bot.

### Slice 8: Cross-interface acceptance and operator guidance

Ticket: [TKT-20261017203525](../ticket/ready/20261017203525-support-pi-alongside-codex.md).

Prerequisite: Slice 7 integrated into dev; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T8.1 | Add one bounded cross-interface journey that switches defaults while prior worker/generation work remains active, exercises all four role combinations, resumes retained work and verifies a Codex-only setup and direct-API legacy selection. |
| T8.2 | Complete profile/runtime/credential setup and troubleshooting guidance, operator-visible switching semantics, model limitations and observed-versus-certified usage documentation; verify each HP criterion against its owning slice evidence and close any in-scope gap. |
| T8.3 | After explicit operator authorization, run one real Pi structured generation and one disposable Docker worker task with a fixed profile/model and bounded existing budgets; record exact evidence, run the final development gate on integrated dev and close the original feature ticket only after its PR is merged. |

Validation: Use deterministic fixtures for routine regressions. The live fixture must be private, disposable and explicitly authorized, with no unrelated remote write or live bot message. Missing credentials/access or an unsupported capability is a real blocker, not a simulated success. Full provider/model, full Go/race and architecture matrices remain nightly.
