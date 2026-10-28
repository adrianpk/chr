# Shared workflow state

ErrConflict is the existing revision/image conflict identity shared by durable
workflow updates and conditional notification completion. All callers refer to
the same sentinel. This package owns no coordinator, store or transport behavior.

MutationBinding owns the unchanged opaque request/operator validation shared by
native owners and transport consumers. ValidRequestKey preserves workflow request
identity spelling. MutationReceiptPath derives the existing content-addressed
receipt reference; the workflow store still owns containment, IO, locks,
transactions and receipt verification. These pure values depend only on
core/contenthash, core/harness, ops/metadata and the standard library. Root callers use this owner directly.

FileChange owns the existing path, before hash and replacement bytes shared by
authoring and workflow journals. It keeps the original JSON keys and field order;
there is no IO, path permission, transaction or replay decision in this value.

Control requests/results, attributable integration grants/answers, review finding
values, guided control proof and hashed evidence references are shared retained
values used by authoring and native workflow consumers. Their original field
order and JSON tags remain intact. This owner adds no authority validation,
coordinator, store, IO or policy; each native owner retains those decisions.

RequireMarkdown owns the original native envelope-presence rule over the shared
ops/metadata splitter. ErrNotWorkflow retains the original not-adopted-record
identity/message; native schema recognition still belongs to its caller. Legacy
authoring selects the optional splitter directly. The dependency is one-way:
metadata cannot import workflow state or a coordinator.

Retained admission, selection and execution-policy values are shared with the
authoring guided record. These values keep the original field order, JSON tags,
enum spellings and zero/empty representation. Native owners still normalize and
validate requests, filter/order indexed tickets, observe policy and hold locks.
The shared values contain no host resolver, executable command or configuration
snapshot. Configuration defaults retain their original explicit enum values.

StartInput is the unchanged serializable native request retained by guided work.
It references the existing core/harness binding value; that owner cannot import
workflow state. The native caller embeds StartInput and supplies executable argv,
selection guard and resolver callback separately. Guided replay retains no host
callback, command or configuration snapshot, and native admission still owns all
normalization, authority, current-state checks, receipt identity and dispatch.
The frozen guided/session/start/invocation fixtures retain the original bytes.

The complete native set descriptors and retained plan-revision binding live here
because native persistence and concrete authoring document reads share their
original JSON representation. Closing attempt and amendment references move with
the set graph; field order, defaults and tags remain exact. Native set validation,
selection, mutation, materialization and publication still have their existing
owners. This value layer imports neither generation nor the coordinator.

StopStatus retains the original cancel-to-cancelled mapping and paused fallback.
TrackerMetadataPath retains the original set-identity check and projection path;
ErrProjectionStale keeps the same error identity for committed domain progress
with a failed derived export. These shared primitives perform no IO or recovery.


IntegrationCycleMode, IntegrationReviewDecision, AgenticReviewResult and finding
category values are shared by retained workflow review/repair records and the
legacy integration-cycle consumer. Their existing enum types remain distinct
from ReviewMode. The result keeps its original field order, JSON tags and
nil/empty slices. Pure value validation preserves category/text/decision checks
and error messages. Native callers still own approval, source/revision guards,
cosmetic blocking policy, transport, configuration, persistence and mutations.

Shared integration proof/candidate, review input/decision, merger protocol,
submission and closing input/result values now serve both durable records and
consumer-owned stage ports. All original public type and embedded-field names,
JSON keys/tags/order, enum strings and nil/empty representations remain intact.
ProcessIdentity is retained process identity data, not a process handle or
execution capability. EvidenceRefs enumerates the original ordered references
on merger input and native record-binding values; native owners still validate
containment, hashes, authorization and actual evidence. This data owner gains no
store, configuration, transport, claim resource, journal or recovery authority.

Repository identity canonicalization and lexical branch, commit, fingerprint,
evidence-reference and allowance checks are shared by native authority and stage
consumers. They preserve the original values, error identities and overflow
behavior. These checks perform no IO, admission, evidence reads or authorization;
those decisions remain with the native owner.
