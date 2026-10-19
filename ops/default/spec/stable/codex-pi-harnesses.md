# Selectable Codex and Pi Harnesses

Status: Approved

Date: 2026-10-19

Approved: maintainer, 2026-10-19; explicit design approval and request for slicing.

Source: [Optional Pi execution and generation](../../ticket/ready/20261017203525-support-pi-alongside-codex.md)

Model: [Harness selection and retained bindings](codex-pi-harnesses-model.md)

Design request: support Pi alongside Codex, with equally placed modules and clear ownership.

Plan: [Selectable Harness Delivery](../../plan/codex-pi-harnesses.md).

Tracker: [Selectable Harness Delivery Order](../../tracker/codex-pi-harnesses.md).

Implementation authorization: granted by the maintainer on 2026-10-19 for the
linked eight-slice plan through Looper, including its canonical PRs and
controller-owned Rebase + Fast-forward merges into dev. Live acceptance retains
the plan's fixture, access and explicit-authorization requirements.

## 1. Outcome

Allow one project to choose Codex or Pi independently for ticket execution and
model-backed conversation or authoring. The same harness can serve both roles;
different profiles can choose different models. Change these role selections
from Web, TUI, Telegram or CLI without restarting the service. Already admitted
work keeps its retained selection. Pi exposes the configured
providers and models that its installed release supports, including compatible
local and open-weight models. Required tool or structured-output capabilities
still apply to the selected role; model availability is not proof of suitability.

Keep Codex usable without installing Pi. Preserve existing direct OpenAI API
generation paths and their established defaults. Harness identity, model-provider
identity and forge-provider identity are distinct concepts.

CHR continues to own ticket state, deterministic order, claims, capacity,
validation, review, publication and delivery receipts. Harnesses perform bounded
calls under that authority. This is product functionality, independent of the
controller or harness used to develop CHR.

## 2. Verified Starting Point

The current implementation has no self-contained Codex module:

| Owner | Current responsibility to separate |
| --- | --- |
| `internal/core/agent_run.go` | Neutral request/result types, local artifacts, process execution and Codex flags are mixed. |
| `internal/core/codex_worker.go` | Codex worker commands and shell rendering. |
| `internal/core/codex_worker_context.go` | Mostly CHR ticket, context and evidence rules with Codex names. |
| `internal/core/quiz_codex_model.go` | Artifact prompts/parsing mixed with harness invocation. |
| `internal/core/structured_model.go` | Shared limits and capacity mixed with Codex/OpenAI dispatch. |
| `internal/app/service_dump_curation_codex.go` | Dump business preparation mixed with a Codex-specific provider. |
| `internal/app/service.go` | Different authoring actions resolve Codex or direct OpenAI independently. |
| `internal/app/service_work_authority.go`, `service_workflow_controls.go` | Native start and continuation default directly to Codex. |
| `internal/core/harness_invocation.go`, `runtime_plan.go` | Docker invocation and credential mounts assume Codex/OpenAI. |
| Bundled `runtime/agent/Dockerfile` files | Install Codex; the current Node installation is not a Pi runtime guarantee. |

Contextual quiz generation, structured Telegram intent/conversation calls and
agentic review also reach the current shared runner. They must be covered by
the call-site inventory, not left behind by changing only Spec generation.

The inspected Pi executable reports 1.3.0. The pinned upstream documentation
supports subprocess RPC, JSON event streams and one-shot execution. Initial
compatibility must target that exact release, not an unpinned latest download.

## 3. Package Ownership

Use sibling Go packages under one technical boundary:

```text
internal/
  harness/
    contracts.go          # Neutral requests, results, capabilities and identity.
    process/              # Shared bounded process and stream supervision.
    codex/                # Codex commands, transport, events and runtime requirements.
    pi/                   # Pi commands, transport, events and runtime requirements.
  core/                   # CHR workflow, prompts, schemas and domain validation.
  app/                    # Resolve profiles and assemble the selected adapters.
```

File names inside each adapter follow its real responsibilities. Do not create
empty files to complete a symmetrical tree. Neither adapter imports the other.
The root `harness` package defines contracts and never imports its child packages.
Adapters import those contracts and the shared process owner, not `core` or
`app`. Core may depend on the neutral contracts; app is the composition root.
This dependency direction must compile without import cycles.

Move technical `AgentRunRequest`, `AgentRunResult`, runner and command contracts
to the neutral boundary. Update internal callers directly; do not leave a second
API, alias collection or permanent forwarding layer in core. Provider-specific
tests move with their implementation. Common contract tests exercise both real
adapters through the same boundary.

Keep task-input construction, worker instructions, evidence requirements,
authoring prompts, output schemas and semantic validators in their existing
CHR owners. Rename generic Codex-labelled context/provider types when removing
the coupling. Do not copy those rules into Pi. Core callers receive a runner;
they no longer instantiate a hidden default Codex runner. Existing direct API
behavior stays available without an unrelated API-client redesign.

## 4. Call Contracts and Assembly

Expose two small capabilities rather than one method with unrelated modes:

1. Structured generation: explicit input/context, instructions, schema, output
   limit, deadline and request identity; returns normalized output and observed
   harness/model metadata. CHR parses and validates the result.
2. Worker execution: host-frozen task input, execution lane, selected runtime,
   result locations and bounded lifetime. It preserves the existing native and
   Docker worker contracts, including durable result validation.

Shared technical values cover command arguments, environment references,
process outcome, normalized events, usage availability and capability errors.
Domain types such as ticket transitions or integration receipts do not move to
the harness package. Reuse the existing domain ports where their semantics fit.

App resolves and injects adapters once for a service operation. A fresh process
is used for each independently owned call. No project-wide Pi session or mutable
global model selection is shared by concurrent work.
Profile-file IO and executable/runtime checks belong to explicit setup or
resolution; adapter constructors only receive already resolved dependencies.

## 5. Selection and Compatibility

Use role selections in workspace configuration and named profiles in private
user configuration. A project may select a profile but cannot supply a new
executable or credential path through model input or a Telegram message.

Illustrative workspace configuration:

```yaml
harnesses:
  worker: pi-development
  generation: pi-authoring
```

Illustrative private profile definitions:

```yaml
profiles:
  pi-development:
    harness: pi
    command: pi
    model_provider: example-provider
    model: example-coding-model
  pi-authoring:
    harness: pi
    command: pi
    model_provider: example-provider
    model: example-writing-model
```

Credentials remain outside both examples. Profile names and non-secret
selection fields can appear in diagnostic projections; private locations and
credential values cannot appear in versioned artifacts.

Resolution order is explicit workspace role selection, existing action-specific
backend configuration, then the current action's default. An explicit generation
selection applies to Dumps, Quiz questions, Specs, acceptance generation, Plans,
tracking materialization and structured conversational calls. Preserve fast/deep
quiz behavior and existing explicit model overrides. Inventory every shared
runner caller and map it to its existing worker or generation authority.

Keep the existing `CHR_*_GENERATION_BACKEND`, `CHR_CODEX_MODEL` and timeout
behavior for configurations without new role selections. A legacy `openai`
selection still means the direct API adapter. Unknown harness/profile values
produce an actionable error; they must not fall through to Codex or OpenAI.

Resolve selection before admission and freeze it for the invocation. Changing
configuration affects new work, not retries or resumed work. Distinguish an
explicit model from a CLI-managed default; record an actual model only when it
was observed. Never fabricate a historical provider/model value.

### 5.1. Live Selection Changes

Provide one application-service operation to inspect and change a project's
worker and generation selections. Web, TUI, Telegram and CLI are adapters to
that operation, not separate configuration stores. Offer the configured profile
choices and the effective selection for each role, including legacy dispatch
when a role has no explicit selection. Allow clearing an explicit selection to
restore that legacy behavior. Changing one or both roles is one atomic update.

Persist selections through the existing workspace configuration owner, with a
revision guard and cross-process write exclusion. A change carries the expected
revision, exact project/role/profile choices, an authorized actor and a request
identity. Reject stale updates instead of overwriting another actor's choices;
an identical retried request returns the original result. Preserve unrelated
configuration fields. Validate profile resolution, role capabilities and local
runtime readiness before publishing the new revision. A failed change leaves
the previous selections intact. Do not make a paid model call during switching
or represent local readiness as proof of remote authentication.

Every newly admitted operation reads one complete persisted selection revision.
Long-running services and separate CLI processes observe that revision without
restart; a process-local cache or notification alone is insufficient. Admission
and a concurrent switch must have a defined ordering: a new operation binds
either the complete old selection or the complete new selection. Never resolve
the two roles from different revisions of one update.

| Work boundary | Effect of changing the selection |
| --- | --- |
| New workflow invocation | Uses the current worker selection at admission. |
| Running, paused, retrying or resumed invocation | Retains its worker binding, including tickets not yet dispatched within that invocation. |
| New authoring or conversational request | Uses the current generation selection when that operation starts. |
| Multiple model calls or retries within one generation operation | Keep the operation's original binding, including nested/deep questions or multi-stage artifact generation. |
| A later independent user message or Quiz answer request | Resolves the current generation selection; an interrupted operation being resumed keeps its original binding. |

No switch cancels existing work, resets budgets, replaces an active session or
silently falls back to another harness. Present the selected defaults separately
from the binding of active work, with readable profile/harness/model names and
an explicit indication that the change applies to new operations. Use compact
controls within each interface's existing project context, not a new general
settings subsystem.

Web changes use the existing authenticated mutation and origin protections.
TUI and CLI use the existing local actor authority. Telegram changes require an
authorized mapped actor, an explicit selected project and confirmation bound to
the exact choices and observed revision; a model's interpretation alone cannot
authorize the change. No interface accepts arbitrary executable, credential or
runtime paths through this selection operation.

Switching selects already configured profiles. Credential provisioning,
installing executables, building runtime images and changing service listener
configuration remain explicit setup operations. Their existing activation rules
still apply; switching does not perform those operations implicitly.

## 6. Pi Transport

Use Pi 1.3.0 subprocess RPC for the controlled runner. Keep its framing and
event interpretation inside `harness/pi`. A thin host-owned worker bridge may
adapt the existing stdin/stdout worker contract to this runner; it must delegate
transport rather than duplicate it or implement a second workflow engine.

Start with an ephemeral session and an exact provider/model selection. Send
the prepared context through stdin, never prompt or credential command arguments.
Correlate protocol commands, drain stdout and stderr concurrently, and consume
the run through its settled terminal event. Command acceptance, an individual
message end or `agent_end` alone is not successful completion. Provider failures,
abort, malformed framing, deadline and early process exit have distinct outcomes.
Close stdin for orderly exit and retain the existing owned-process stop policy.

Use one bounded stream decoder. Normalize only the final completed assistant
output into the existing caller-owned output artifact. Intermediate reasoning,
tool output or a partial message cannot become the generated JSON result. Apply
the shared schema and domain validators regardless of a provider's structured
output features; invalid output is a failed call, not a committed artifact.

Pi generation uses a tool-disabled profile and no discovered extensions, MCP
servers, skills, prompt templates, themes or context files. CHR assembles the
required context explicitly. Where an existing operation needs read-only file
inspection, supply an explicit read-only capability profile; never enable shell
or write tools merely to obtain structured output.

Worker tools are explicitly enabled inside the owned execution environment.
Project instructions are supplied through the existing CHR context path;
automatic local Pi resources cannot change the model, credentials, tools or
authority of a frozen run. Do not import an operator's unrelated extensions.

## 7. Runtime and Credentials

Provide sibling Codex and Pi runtime assets under the existing agent runtime
boundary. Update all bundled profiles, initialization and explicit runtime
upgrade paths. Keep existing configured Dockerfile paths/images valid for
Codex. Selecting Pi must not silently reuse an image that contains only Codex.

The Pi image needs Node.js 22.19 or newer and an exact Pi 1.3.0 installation
with locked dependencies. Build and inspect that image as part of focused
runtime validation. Codex-only users incur no Pi install, probe or image build.
Run the adapter's protocol process inside the selected container through owned
stdin/stdout pipes; a host bridge must not move the model's worker tools outside
that container. Keep explicit existing native execution modes explicit.

Mount or supply only the selected profile's authentication and model settings.
Do not forward all environment variables or mount an entire developer home.
Each concurrent call has private writable runtime/session state. OAuth refresh,
where used, must have an explicit credential owner and atomic update/exclusion
rules; a read-only mount cannot be described as refresh support. Reuse provider
refresh behavior only after verifying that boundary under concurrent calls.

Custom provider/model definitions stay private. Local endpoints must be resolved
from inside Docker: container loopback is not the host. Support an explicit,
validated container-reachable endpoint without silently enabling host networking.
Never put API keys in arguments, frozen task inputs, command artifacts or logs.
Reuse sanitization before recording diagnostics or normalized events.

## 8. Lifecycle, Capacity and Bounds

Keep one shared native capacity policy across Codex and Pi. Selecting another
harness does not create another worker pool or bypass source/target exclusions.
Preserve the existing shared structured-model capacity and caller deadlines;
do not multiply slots by provider or profile.

Calls declare finite input, result, protocol-record, diagnostic and total-stream
limits before spawn. Reuse stricter existing caller limits, including structured
model input/output limits. Add explicit transport ceilings where none exist;
oversized records or output fail rather than produce truncated accepted JSON.
Spool bounded diagnostics instead of accumulating unbounded byte buffers.

The host owns retry attempts, stop intent, claims, validation and publication.
Configure Pi internal retries to fit the retained CHR deadline/attempt policy;
they cannot reset a budget or create an unobserved infinite continuation.
Exit zero, a model's success claim or a written result file alone never proves
validation, integration or ticket completion.

Retain current certified whole-call token requirements. Report missing usage
as unknown, never zero. Aggregate observed call usage without claiming it proves
hard enforcement. A finite-budget role requiring a certified driver must reject
an adapter that cannot establish that capability. An unrestricted Pi worker is
not a certified merger broker and cannot receive that role by renaming it.

## 9. Persistence and Recovery

Apply the companion model. New invocations bind role, harness, selection revision,
profile revision, provider/model selection, command/runtime identity and protocol
revision.
Retries and resumes validate this binding before spawning anything. Do not
resolve a new backend from current environment variables halfway through a run.

Version the invocation format when making this identity mandatory so an old
binary rejects Pi-bound work instead of ignoring a new field and running Codex.
Continue to read existing `chr-invocation/v1` records under their historical
rules; no automatic rewrite of active checkpoints or their hashes. Preserve
existing request-key idempotency, attempt counts and interrupted work ownership.

## 10. Validation and Acceptance

| ID | Required evidence |
| --- | --- |
| HP-01 | Codex-only installation and existing default actions work without Pi installed. |
| HP-02 | Worker and generation roles select Codex/Pi independently in all four combinations. |
| HP-03 | All generation call families use the resolved role, including contextual Quiz and Telegram paths. |
| HP-04 | Existing explicit OpenAI API action selections retain their behavior. |
| HP-05 | Both adapters satisfy common process, cancellation, output, redaction and failure contracts. |
| HP-06 | Core has no concrete Codex/Pi construction or provider-specific subprocess flags; imports follow section 3. |
| HP-07 | Worker input/result validation, artifacts, review and delivery receipts remain host-owned. |
| HP-08 | A real selected Docker Pi worker executes a bounded task in a disposable repository. |
| HP-09 | A real Pi generation call produces a schema-valid artifact; invalid output leaves previous artifacts intact. |
| HP-10 | Split/oversized JSONL, interleaved events, provider failure, early exit, abort and late completion are handled correctly. |
| HP-11 | Pi generation cannot acquire discovered tools, extensions or mutable project settings; existing Codex generation restrictions remain covered. |
| HP-12 | Missing executable, profile, credentials, model or runtime gives an actionable error without fallback. |
| HP-13 | Two independent calls retain separate sessions, credentials and output paths within shared capacity. |
| HP-14 | Resume uses the frozen identity after configuration changes; legacy checkpoints remain intact. |
| HP-15 | Old binaries reject the new mandatory invocation format before executing a worker. |
| HP-16 | Usage stays unknown when absent; finite-budget certification and merger-broker restrictions remain enforced. |
| HP-17 | A configured local/custom endpoint works from its actual execution environment without leaking credentials. |
| HP-18 | User Guide, configuration reference, examples and Unreleased notes describe both harness choices. |
| HP-19 | Web, TUI, Telegram and CLI read and change the same per-project role selections without restarting the service; another process sees the change on its next operation. |
| HP-20 | A switch during active work changes new admissions only; pending tickets, nested generation calls, retries and resumes retain their owning operation's binding. |
| HP-21 | Concurrent/stale, unauthorized and invalid changes preserve prior settings; atomic two-role updates, lost-response retries and unrelated configuration preservation are verified. |

Use real temporary files and owned subprocess fixtures for deterministic
transport tests. Share the common contract suite rather than copying it per
harness or testing every private helper. Use focused race tests for shared
capacity and cancellation. Run `make verify` at delivery integration. Keep full
normal/race campaigns and exhaustive live matrices in scheduled verification.
Live model calls and external repository writes require designated disposable
fixtures and explicit operator authorization; missing access is not simulated
success. Do not claim every Pi model/provider was certified from one live test.

## 11. Delivery Shape and Review

The source ticket records the parent need and final integrated acceptance,
not a duplicate execution tracker. The linked delivery set covers shared
contracts/Codex extraction, Pi transport, role wiring, live selection through all
four interfaces, runtime/credentials and end-to-end proof. Each implementation
unit must leave Codex usable. Choose slice boundaries from that behavior; do not
assign one ticket per file or one slice merely for empty package scaffolding.

The accepted decisions are the sibling package boundary, private named profiles
with two dynamically changeable role selections, one
shared configuration mutation service, Pi RPC 1.3.0 as the initial transport,
and mandatory frozen identity for new invocations and generation operations.
The plan records exact transport byte ceilings and runtime image pins before
its first runtime implementation slice. No broad backend-specific live matrix
is part of ordinary interactive validation.

## 12. Upstream References

- [Pi 1.3.0 CLI and resource controls](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/cli.md)
- [Pi 1.3.0 RPC lifecycle](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/rpc.md)
- [Pi 1.3.0 event stream](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/json.md)
- [Pi 1.3.0 model selection](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/models.md)
- [Pi 1.3.0 installation requirements](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/README.md)

These references describe verified upstream capabilities. Package boundaries,
selection rules, persistence changes and acceptance criteria above are CHR's
accepted design, not claims that Pi implements CHR's workflow.
