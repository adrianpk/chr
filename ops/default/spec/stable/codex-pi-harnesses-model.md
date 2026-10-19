# Harness Selection and Retained Bindings

Status: Approved

Date: 2026-10-19

Approved: maintainer, 2026-10-19; promoted with the approved specification.

Specification: [Selectable Codex and Pi Harnesses](codex-pi-harnesses.md)

## 1. Authority

Existing workflow records remain authoritative for task state, claims,
attempts, validation and integration. Add technical harness identity to those
records; do not create another ticket database, execution scheduler or parallel
progress ledger. Private account/model settings remain user-owned configuration.

Harness and model-provider strings are not interchangeable. Existing fields
named `provider` retain their documented historical meaning; introduce explicit
`harness` and `model_provider` fields where both facts are needed. Forge-provider
fields remain unrelated.

## 2. Harness Profile

Stored in a private user configuration document, `harnesses.yml`.

| Field | Representation and rule |
| --- | --- |
| name | Unique non-secret profile key; 1-128 bytes. |
| harness | Closed enum `codex` or `pi`. |
| command | Host-configured executable name/path; never selected by model output. |
| model_provider | Pi provider ID; required for a Pi profile. |
| model | Exact Pi model ID; Codex may retain an explicit CLI-default mode. |
| thinking | Optional supported reasoning selection, validated by the adapter. |
| credential_ref | Private local reference, not a credential value or versioned path. |
| model_config_ref | Optional private custom-provider/model configuration reference. |
| runtime_ref | Selected approved image/runtime profile for a container worker. |
| revision | Digest of the effective non-secret semantic fields. |

The executable/runtime identity is resolved before dispatch. A revision never
hashes raw credential values. Files containing authentication material have
private permissions; diagnostic projections omit private references. Credential
refresh has one explicit owner and does not modify a frozen semantic selection.

## 3. Role Selection

Workspace configuration adds `harnesses.worker` and `harnesses.generation`, each
an optional named-profile reference. A role has one resolved selection per owned
operation, shared by all calls and retries belonging to that operation.
The same profile may serve both roles. Explicit generation selection covers the
shared authoring and conversation call families; omitted selection preserves
the documented legacy per-action dispatch.

Use the existing workspace configuration persistence owner for both role fields
and their revision, not an interface-specific settings file or database. The
application service exposes one read/change contract to Web, TUI, Telegram and
CLI. A persisted selection snapshot contains:

| Field | Stored rule |
| --- | --- |
| project_ref | Existing canonical project identity; changes cannot cross project scope. |
| worker, generation | Optional profile references; clearing a role restores legacy dispatch. |
| revision | Opaque configuration revision; changes when either selection changes. |
| changed_by, changed_at | Existing authorized actor reference and UTC time of the accepted change. |
| request_id, request_digest | Idempotency identity and digest of exact project, expected revision and requested role changes. |

Reuse an existing configuration revision or mutation receipt owner where it
satisfies this contract. Do not add a parallel progress ledger. Store only the
bounded idempotency evidence needed by the existing mutation mechanism, without
secrets or raw conversation text.

A change includes the expected revision and at least one explicit role change;
an omitted role remains unchanged. A null selection clears that role. Validate
authorization, project binding, selected profiles, supported capabilities and
local readiness before an atomic commit under cross-process exclusion. Re-read
and compare the revision inside that exclusion, preserve unrelated fields and
publish both role changes together. A failed or stale change writes nothing;
retrying the same request returns its original result, while reuse of its ID
with different arguments conflicts.

Operation admission obtains a coherent snapshot ordered against configuration
changes. Each process reads the persisted current revision before admitting new
work; do not make correctness depend on service restart, an in-memory cache or
an interface notification. A workflow invocation retains its worker binding for
all its pending tickets, attempts and resumes. A generation operation freezes
its generation binding before its first model call, including all nested calls
and retries. A later independent user request is a new operation; resuming an
interrupted request is not. Changing selections never rewrites those bindings.

An unknown non-null profile reference is an error, not an implicit default.
Readiness distinguishes executable availability, local credential presence,
runtime compatibility and actually tested remote authentication. Default
inspection performs no paid model call.

## 4. Frozen Invocation Binding

New invocations carrying mandatory selection use `chr-invocation/v2`. Extend
the existing invocation representation rather than creating a side database.

| Field | Stored rule |
| --- | --- |
| harness_binding.schema_version | Integer 1 for this nested contract. |
| harness_binding.role | Worker authority for this invocation; generation calls retain their own binding. |
| harness_binding.harness | `codex` or `pi`. |
| harness_binding.selection_revision | Workspace selection revision read at admission; explicit legacy provenance where no new selection document exists. |
| harness_binding.profile | Non-secret profile name. |
| harness_binding.profile_revision | Frozen semantic digest. |
| harness_binding.model_provider | Configured provider ID, or explicitly unresolved legacy/default provenance. |
| harness_binding.model | Requested exact model or explicit `cli_default`; do not invent an observed model. |
| harness_binding.thinking | Frozen reasoning selection when configured. |
| harness_binding.executable_ref | Existing host-owned executable identity/hash evidence. |
| harness_binding.runtime_ref | Existing immutable runtime/image identity evidence where containerized. |
| harness_binding.protocol_revision | Adapter protocol identifier, initially including the Pi 1.3.0 contract. |
| harness_binding.capabilities_ref | Host-validated capabilities, including a certificate only where genuinely established. |
| harness_binding.binding_hash | Hash of canonical non-secret binding data. |

Bind these fields into the admission/request image and worker-input hash before
claiming work. Existing attempt records refer to that binding and cannot replace
it from current configuration. Retries attach to the same request/binding;
reusing a request key with a changed binding conflicts instead of launching a
different harness. Reconcile a lost response before another attempt.

`cli_default` preserves Codex's existing configuration behavior and explicitly
does not claim an exact underlying model was frozen. Exact model reproducibility
requires configuring a model. Record the observed model separately when the
transport supplies reliable evidence; an absent observation remains unknown.

Executable fingerprinting streams a regular executable file with a 512 MiB
ceiling. It rejects larger or non-regular files before admission and rejects
changed bytes before dispatch. This accommodates supported CLI distribution
sizes without increasing the retained metadata or in-memory stream bounds.

## 5. Generation Run Metadata

Extend existing local agent-run artifacts with a bounded host-written metadata
record, `chr-harness-run/v1`. Keep existing output names and caller
parsers stable. This record explains a model call, not a task completion.

Retain the owning selection in a private `chr-harness-operation/v1` artifact
under `.chr/default/tmp/agent-runs/operations/<operation-id-sha256>.json` before
the first call. It contains the operation ID and purpose, sealed non-secret
selection, private resolved program reference, registered project reference,
private configuration/execution roots and a checksum. Project-local calls use
that project as their artifact owner. An existing host conversation retains its
artifact owner across project routing changes; a review can use its host-owned
execution worktree. The artifact owner excludes duplicate first writes, then
the selected configuration owner serializes resolution with live changes. A retained
operation is read before live configuration or profiles; nested calls use the
same operation, and a retry cannot exchange its purpose or selection. This is
identity retention within the existing run artifacts, not another task or retry
authority. Limit each operation/metadata record to 16 KiB and retain at most
4,096 operation identities without eviction; capacity exhaustion fails closed.
Operation and purpose identities are bounded to 256 bytes. The corresponding
per-call record is `harness-run.json` in the existing caller-owned run bundle.
Caller request IDs are namespaced by operation for ordinary run bundles, so
independent operations cannot overwrite each other's call evidence. Retried
calls may refresh their input under the original owning request's argument and
attempt guards; run metadata does not add another replay or retry authority.
Quiz state retains its pending generation operation ID, allowing a resume to
keep that binding while a later independent answer resolves current settings.

| Field | Stored rule |
| --- | --- |
| request_id, purpose | Existing caller-owned identity and operation purpose. |
| operation_id | Owning authoring/conversation operation; shared by its nested model calls and retries. |
| selection | The frozen non-secret harness/profile/provider/model binding. |
| selection_revision | Workspace revision resolved once for the owning operation. |
| observed_model | Optional verified actual provider/model; never guessed. |
| input_ref, schema_ref | Existing private run artifact references/digests. |
| output_ref, output_hash | Host-normalized final output only after complete framing. |
| process_outcome | Started, completed, failed, cancelled, timed_out or protocol_error. |
| output_validation | Unknown, valid or invalid; separate from process exit. |
| usage | Optional measured input/output/cache counts with accounting revision. |
| usage_status | Known, partial or unknown; unknown is not numeric zero. |
| started_at, finished_at | UTC timestamps for this owned call. |
| diagnostics_ref | Bounded sanitized private diagnostic artifact. |

Measured counts use host normalization revision `chr-harness-observed/v1`.
This observation does not certify whole-call enforcement. Unmeasured counters
remain absent.

No token, password, environment value, authorization header or complete private
profile is serialized into this record or into command/event traces. Semantic
validation remains with the original operation owner. A successful transport
cannot independently mark an artifact approved or a ticket delivered.

## 6. Compatibility and Migration

1. Retain an explicit reader/executor path for existing `chr-invocation/v1`
   checkpoints. Keep original bytes, request hashes, ownership and budgets;
   loading a record does not migrate or resume it.
2. A legacy invocation without new binding follows its previous worker command
   and Codex-default semantics. New global Pi selections cannot change it.
   Missing historical model metadata remains unknown.
3. New bound work uses v2 so the existing v1-only reader rejects it before
   execution. An unsupported future schema also fails closed.
4. Update new task-input identity and validators together with the binding.
   Other existing ticket, tracker, receipt and integration-journal schemas do
   not change merely to organize harness code.
5. Do not add SQLite migrations unless implementation identifies an existing
   durable call owner that needs an additive field. Any such change belongs in
   this model before it is implemented.

## 7. Consistency Checks

Profile lookup is unique, role selection is deterministic and binding hashes
exclude secrets. Configuration changes are atomic, authorized and revision
guarded across interfaces and processes. Readable current defaults do not
misrepresent an active operation's older retained binding. Adapter upgrades
cannot impersonate the retained executable identity. Shared capacity includes
both harnesses. Resumes do not import live
configuration changes. Bad schema, partial output or uncertain usage cannot
create successful workflow evidence. Import-direction checks and integration
fixtures cover these contracts alongside the behavioral acceptance criteria.
