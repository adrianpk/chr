# Shared workflow state

ErrConflict is the existing revision/image conflict identity shared by durable
workflow updates and conditional notification completion. All callers refer to
the same sentinel. This package owns no coordinator, store or transport behavior.

MutationBinding owns the unchanged opaque request/operator validation shared by
native owners and transport consumers. ValidRequestKey preserves workflow request
identity spelling. MutationReceiptPath derives the existing content-addressed
receipt reference; the workflow store still owns containment, IO, locks,
transactions and receipt verification. These pure values depend only on
core/contenthash, ops/metadata and the standard library. Root callers use this owner directly.

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
