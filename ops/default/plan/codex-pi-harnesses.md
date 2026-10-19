# Selectable Harness Delivery

Status: Approved

Date: 2026-10-19

Delivery set: codex-pi-harnesses

Scope approval: maintainer, 2026-10-19; the specification is approved and this
slicing preparation was explicitly requested.

Approved: maintainer, 2026-10-19; explicit instruction to execute this set through Looper.

Spec: [Selectable Codex and Pi Harnesses](../spec/stable/codex-pi-harnesses.md).

Model: [Harness Selection and Retained Bindings](../spec/stable/codex-pi-harnesses-model.md).

Tracker: [Selectable Harness Delivery Order](../tracker/codex-pi-harnesses.md).

Planning baseline: `79c789acc842a4e71757203df275ece5296875b1`.

Implementation authorization: granted on 2026-10-19 for these eight slices
through Looper, including canonical slice PRs and controller-owned Rebase +
Fast-forward merges into dev. Use one active development slice and one open
canonical PR for this set. The live acceptance prerequisites below remain in
force. Existing controller configurations and user services remain separately
owned.

First planned slice: Slice 1, Neutral contracts and Codex extraction; T1.1.
Actual processing state belongs to the source tickets and slice reports.

Slice strategy: behavior-first

Reason: each slice delivers a usable adapter, configuration boundary or complete
interface journey while preserving Codex. The package extraction is a working
Codex migration, not empty scaffolding; later slices include their own tests and
documentation rather than accumulating a separate test-only implementation gap.

Development gate: `make verify`, the affected focused regressions and
`git diff --check`. Full Go/race, provider/model and architecture matrices stay
in the existing scheduled nightly.

## 1. Delivery Ownership and Boundaries

1. Deliver eight slices sequentially, one ticket, branch and PR targeting dev
   per slice. Integrate the preceding slice before starting the next. Keep at
   most one canonical open PR for this set.
2. The tracker defines deterministic order and commit-sized tasks only.
   Tickets and slice reports own processing state, commits, PRs, blockers and
   validation. Task IDs do not create additional tickets. Reuse the original
   feature ticket for Slice 8 acceptance; create no duplicate parent ledger.
3. Commit and publish these approved design/planning inputs before creating an
   implementation branch. On execution activation, verify current upstream,
   worktree/controller/validation ownership and the installed tool's help.
   Do not replace another controller or resume a different delivery set.
4. Follow the approved sibling package/import boundary. Core owns workflow,
   prompts and semantic validation; app owns assembly and the shared operation
   service; interfaces map requests and responses. No second scheduler, global
   mutable Pi session or interface-specific configuration store is introduced.
5. Keep product concurrency supported. Configuration changes and admissions
   are ordered through persisted revisions across processes; retries, pending
   tickets and resumed operations retain their original bindings.
6. Live model requests and external repository writes require explicit
   operator authorization and disposable fixtures. Do not touch live project
   data, bot delivery, account credentials or running user services merely to
   demonstrate the feature. Readiness does not prove remote authentication.
7. Missing private readiness does not block deterministic implementation tests.
   Before activating Slice 8 live acceptance, verify the approved fixture,
   exact configured model, budget/driver capability and authorization. Record
   any missing prerequisite as a blocker without claiming full acceptance.

## 2. Resolved Implementation Bounds

These are initial hard ceilings, not target payload sizes. Reuse any stricter
existing caller limit; reject an oversize value before accepting partial output.

| Boundary | Maximum |
| --- | --- |
| Structured generation context/input | Existing 131,072 bytes |
| Structured instructions | Existing 8,192 bytes |
| Structured final JSON | Existing 65,536 bytes |
| Prepared worker stdin/context envelope | 4,194,304 bytes |
| One RPC JSONL record, excluding LF | 1,048,576 bytes |
| Aggregate protocol stdout per owned call | 67,108,864 bytes |
| Aggregate stderr per owned call | 2,097,152 bytes |
| Retained sanitized diagnostic excerpt | 262,144 bytes |
| Normalized worker final payload | 1,048,576 bytes |

Count stream bytes as they arrive and spool within the ceiling rather than
accumulating unbounded buffers. Exceeding a ceiling cancels and reaps the owned
process and records the specific failure; it never truncates accepted JSON.
Transport wait/retry/compaction fits the existing caller deadline and retained
attempt/token policy. No independent Pi retry budget or extra capacity pool is
created. Focused fixtures may inject smaller ceilings to exercise failures.

The selection operation changes one or both role references atomically using
the existing configuration owner's whole-document revision and serialized
publication. Reuse the existing mutation identity/receipt mechanism for bounded
replay evidence. Check profile readiness before commit, then compare the
current configuration again inside cross-process exclusion. Every admission
reads one complete persisted revision; it does not depend on cache invalidation
events. Freeze generation selection once for all calls in an operation.

## 3. Pinned Pi Runtime Inputs

Resolved from upstream release/registry metadata on 2026-10-19. Apply these pins
to the new Pi runtime; retain existing Codex runtime paths and installation
behavior. Pinning is not a claim that an image was already built or certified.

| Input | Selected identity |
| --- | --- |
| Go/base image, multi-platform manifest | `golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437` |
| Node.js | `22.23.3`, upstream Linux archive |
| Node Linux x64 SHA-256 | `df450af89261115ef9f9e3830c3eeb2cc9213b63c720b1af623cb5dcbe2e02de` |
| Node Linux arm64 SHA-256 | `a44aeb94849a299b22df10b9e622ec2f605c2183501bc40590705131de7c740f` |
| Pi package | `@earendil-works/pi-coding-agent@1.3.0` |
| Pi package integrity | `sha512-V/WoCjA2hQSe3QLWIUB24cTRdtEdeLiIzkDqU5RMXV6VWuD+KGZnzVBMhglwdnQTySkehZsjt8a3g6Jd2g4nnw==` |

Slice 4 supplies a minimal Pi installation manifest and committed npm lockfile
with resolved transitive versions and integrity metadata, then installs through
`npm ci`. Do not copy unrelated upstream applications or describe a global
unpinned install as reproducible. Map the existing selected Linux platform to
its pinned archive; an unsupported platform produces an explicit readiness
error. Capture the built image identity in the frozen runtime binding.

Sources: [Node release metadata](https://nodejs.org/dist/index.json),
[Node archive checksums](https://nodejs.org/dist/v22.23.3/SHASUMS256.txt),
[Pi package metadata](https://registry.npmjs.org/@earendil-works%2fpi-coding-agent/1.3.0)
and [Go official image registry manifest](https://registry-1.docker.io/v2/library/golang/manifests/1.25-bookworm).
The specification links the pinned Pi CLI/RPC contract.

## 4. Canonical Slice Map

| Slice | Short name | Ticket | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | Neutral contracts and Codex extraction | [TKT-20261019103827](../ticket/solved/20261019103827-harness-codex.md) | `refactor/slice-1-harness-codex` | `refactor(slice-1): extract Codex behind neutral harness contracts` | `ops/default/report/slices/codex-pi-harnesses/slice-1-harness-codex.md` |
| Slice 2 | Live role selection and retained bindings | [TKT-20261019103828](../ticket/solved/20261019103828-harness-selection.md) | `feat/slice-2-harness-selection` | `feat(slice-2): retain live harness selections and invocation bindings` | `ops/default/report/slices/codex-pi-harnesses/slice-2-harness-selection.md` |
| Slice 3 | Pi RPC and structured generation | [TKT-20261019103829](../ticket/in_progress/20261019103829-pi-generation.md) | `feat/slice-3-pi-generation` | `feat(slice-3): generate structured content through the Pi adapter` | `ops/default/report/slices/codex-pi-harnesses/slice-3-pi-generation.md` |
| Slice 4 | Pi Docker workers and private runtime setup | [TKT-20261019103830](../ticket/ready/20261019103830-pi-worker-runtime.md) | `feat/slice-4-pi-worker-runtime` | `feat(slice-4): execute Pi workers in their selected Docker runtime` | `ops/default/report/slices/codex-pi-harnesses/slice-4-pi-worker-runtime.md` |
| Slice 5 | Web harness selection | [TKT-20261019103831](../ticket/ready/20261019103831-web-harness-selection.md) | `feat/slice-5-web-harness-selection` | `feat(slice-5): change project harness selections from the Web UI` | `ops/default/report/slices/codex-pi-harnesses/slice-5-web-harness-selection.md` |
| Slice 6 | TUI harness selection | [TKT-20261019103832](../ticket/ready/20261019103832-tui-harness-selection.md) | `feat/slice-6-tui-harness-selection` | `feat(slice-6): change project harness selections from the TUI` | `ops/default/report/slices/codex-pi-harnesses/slice-6-tui-harness-selection.md` |
| Slice 7 | Telegram harness selection | [TKT-20261019103833](../ticket/ready/20261019103833-telegram-harness-selection.md) | `feat/slice-7-telegram-harness-selection` | `feat(slice-7): change project harness selections through Telegram` | `ops/default/report/slices/codex-pi-harnesses/slice-7-telegram-harness-selection.md` |
| Slice 8 | Cross-interface acceptance and operator guidance | [TKT-20261017203525](../ticket/ready/20261017203525-support-pi-alongside-codex.md) | `test/slice-8-harness-acceptance` | `test(slice-8): verify selectable harnesses and live configuration end to end` | `ops/default/report/slices/codex-pi-harnesses/slice-8-harness-acceptance.md` |

## 5. Slice Contracts

### Slice 1: Neutral contracts and Codex extraction

Ticket: [TKT-20261019103827](../ticket/solved/20261019103827-harness-codex.md).

Dependency: committed planning inputs and execution activation.

Outcome: Existing generation and worker behavior runs through injected neutral contracts, with a complete Codex adapter and no Pi prerequisite.

Owners: internal/harness, internal/harness/process, internal/harness/codex; affected core and application assembly callers.

Acceptance: HP-01, HP-04, HP-05, HP-06, HP-07.

Evidence: Use owned subprocess and filesystem fixtures for generation and worker success, schema failure, cancellation, redaction and bounded output. Check the complete caller/import inventory and a Codex-only installation.

Focused regression (planned): `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestHarnessCodex|TestCodex|TestAgentRun|TestStructured)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/harness/... -run '^TestHarnessCodex' -count=1 -timeout=3m`.

### Slice 2: Live role selection and retained bindings

Ticket: [TKT-20261019103828](../ticket/solved/20261019103828-harness-selection.md).

Dependency: Slice 1 integrated into dev.

Outcome: A shared service and CLI select prepared profiles per role without restart, while admitted work retains its binding across processes, retries and recovery.

Owners: Existing workspace configuration and invocation owners in core; the central application operation service and CLI.

Acceptance: HP-02, HP-04, HP-12, HP-14, HP-15, HP-19, HP-20, HP-21.

Evidence: Exercise two independent processes, simultaneous admissions and switches, stale/unauthorized/invalid updates, clear-to-legacy, crash/replay, unchanged unrelated fields, active pending tickets and multi-call generation. An old v1-only reader must reject v2 before spawning.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m`.

### Slice 3: Pi RPC and structured generation

Ticket: [TKT-20261019103829](../ticket/in_progress/20261019103829-pi-generation.md).

Dependency: Slice 2 integrated into dev.

Outcome: A prepared native Pi profile can author validated content through all generation families while Codex and direct API dispatch keep working.

Owners: internal/harness/pi; existing structured model, authoring, contextual Quiz and application assembly owners.

Acceptance: HP-02, HP-03, HP-05, HP-09, HP-10, HP-11, HP-12, HP-16, HP-20.

Evidence: Use deterministic RPC subprocess fixtures for fragmented LF framing, Unicode separators, interleaved commands, retry/compaction before agent_settled, backpressure and cancellation. Exercise every authoring family through its real shared service with schema-valid and rejected outputs; a real paid Pi call remains a final opt-in acceptance item.

Focused regression (planned): `go test ./internal/harness/... ./internal/core ./internal/app -run '^(TestPiTransport|TestPiGeneration|TestHarnessGenerationBinding)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/harness/pi ./internal/app -run '^(TestPiTransport|TestPiGeneration)' -count=1 -timeout=3m`.

### Slice 4: Pi Docker workers and private runtime setup

Ticket: [TKT-20261019103830](../ticket/ready/20261019103830-pi-worker-runtime.md).

Dependency: Slice 3 integrated into dev.

Outcome: Pi worker RPC runs inside the chosen container with pinned runtime inputs and isolated authentication/session state; existing Codex runtimes remain usable.

Owners: Bundled standard/minimal/personal runtime assets and explicit upgrade path; core runtime assembly, worker bridge and app readiness.

Acceptance: HP-01, HP-02, HP-07, HP-08, HP-12, HP-13, HP-14, HP-16, HP-17.

Evidence: Build and inspect one selected Pi image, exercise a real container with a deterministic local protocol/model endpoint, and run two isolated calls against the shared capacity boundary. Verify only the intended mounts/flags and credential owner. An authorized real-model disposable task supplies HP-08 during final acceptance.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^(TestPiWorker|TestPiRuntime|TestHarnessBinding)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/core ./internal/app -run '^(TestPiWorkerIsolation|TestPiCredentialRefresh)' -count=1 -timeout=3m`.

### Slice 5: Web harness selection

Ticket: [TKT-20261019103831](../ticket/ready/20261019103831-web-harness-selection.md).

Dependency: Slice 4 integrated into dev.

Outcome: An authenticated user can inspect and change both role defaults in a compact project-context control, with clear effects on new work.

Owners: Existing Web projections/actions/templates and the shared application selection operation.

Acceptance: HP-19, HP-20, HP-21.

Evidence: Exercise real HTTP action/authentication boundaries and one focused browser journey on wide and narrow layouts, including a switch while an invocation remains active. Inspect the persisted result from an independent process.

Focused regression (planned): `go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`.

### Slice 6: TUI harness selection

Ticket: [TKT-20261019103832](../ticket/ready/20261019103832-tui-harness-selection.md).

Dependency: Slice 5 integrated into dev.

Outcome: Local users can inspect and change the same per-project role defaults from the TUI without restarting the TUI or service.

Owners: Existing TUI project controls and thin application adapters.

Acceptance: HP-19, HP-20, HP-21.

Evidence: Exercise keyboard/event-driven update behavior through the real TUI adapter with temporary workspace configuration, cancellation, stale revision and independent process changes; retain existing compact layout behavior.

Focused regression (planned): `go test ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`.

### Slice 7: Telegram harness selection

Ticket: [TKT-20261019103833](../ticket/ready/20261019103833-telegram-harness-selection.md).

Dependency: Slice 6 integrated into dev.

Outcome: An authorized mapped user can inspect and confirm role/profile changes for the selected project through Telegram.

Owners: Existing Telegram deterministic commands, callback/confirmation binding, help and central application operations.

Acceptance: HP-03, HP-19, HP-20, HP-21.

Evidence: Use deterministic Telegram updates/callbacks through the existing receiver and shared service. Cover stale confirmation, changed selected project, actor mismatch and replay without sending messages to the live bot.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`.

### Slice 8: Cross-interface acceptance and operator guidance

Ticket: [TKT-20261017203525](../ticket/ready/20261017203525-support-pi-alongside-codex.md).

Dependency: Slice 7 integrated into dev.

Outcome: The complete approved contract has exact-candidate evidence, usable setup/recovery guidance and explicitly bounded live Pi proof.

Owners: Shared integration fixtures, User Guide/how-to/reference documentation and final slice report; original feature ticket.

Acceptance: HP-01, HP-02, HP-03, HP-04, HP-05, HP-06, HP-07, HP-08, HP-09, HP-10, HP-11, HP-12, HP-13, HP-14, HP-15, HP-16, HP-17, HP-18, HP-19, HP-20, HP-21.

Evidence: Use deterministic fixtures for routine regressions. The live fixture must be private, disposable and explicitly authorized, with no unrelated remote write or live bot message. Missing credentials/access or an unsupported capability is a real blocker, not a simulated success. Full provider/model, full Go/race and architecture matrices remain nightly.

Focused regression (planned): `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`.

## 6. Validation Cost and Closure

The commands above name planned test roots; they are not evidence that tests
already exist or passed. Verify selected tests actually execute rather than
accepting an empty regex match. Keep regressions focused on failure modes and
shared contracts, with named table cases where useful; do not duplicate the
same transport suite for each adapter.

Run focused checks during the owning task. Before each slice review/integration,
run the development gate once for the actual candidate and retain its evidence;
reuse the same successful result while that candidate is unchanged. The gate
contains short tests/build/lint and normally costs minutes. Do not accumulate
whole-suite, browser, race or live matrices into every task or final gate.
Selected concurrency/process checks use focused race mode. The one selected
image build may need several minutes and network access on its first run;
record its actual timing rather than inventing a completion estimate.

Web acceptance includes its one named browser journey when the browser fixture
is available. The final deterministic journey covers all four role combinations.
Only Slice 8's explicitly authorized live acceptance calls the real model:
one structured-generation call and one bounded worker task. Use a private
prepared fixture/profile and existing finite budgets; no unlimited role
certification, automatic fallback or unrestricted merger assignment is allowed.

After Slice 8 merges, run `make verify` on the exact integrated dev candidate,
record its hash and all HP-01 through HP-21 evidence in the final slice report,
and close the source ticket only when required evidence is complete. If the
gate requires a correction, follow the existing validation-fix PR workflow;
do not add an unplanned slice automatically. Missing live proof remains an
explicit unresolved acceptance item. Do not treat a short development pass as
nightly/exhaustive certification. Main promotion, release and live service
installation require their own explicit request.
