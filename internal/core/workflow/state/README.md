# Shared workflow state

ErrConflict is the existing revision/image conflict identity shared by durable
workflow updates and conditional notification completion. All callers refer to
the same sentinel. This package owns no coordinator, store or transport behavior.

MutationBinding owns the unchanged opaque request/operator validation shared by
native owners and transport consumers. ValidRequestKey preserves workflow request
identity spelling. MutationReceiptPath derives the existing content-addressed
receipt reference; the workflow store still owns containment, IO, locks,
transactions and receipt verification. These pure values depend only on
core/contenthash and the standard library. Root callers use this owner directly.
