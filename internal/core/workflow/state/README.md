# Shared workflow state

ErrConflict is the existing revision/image conflict identity shared by durable
workflow updates and conditional notification completion. All callers refer to
the same sentinel. This package owns no coordinator, store or transport behavior.
