# Retained harness identity

This owner contains executable references, worker bindings and their pure sealing
and validation rules. Exact JSON field order, capability evidence, digest bytes,
profile-name bounds and error meanings remain unchanged. Consumers use this
owner directly; there is no root compatibility facade.

The adapter-neutral capability value comes from internal/harness. Binding checks
have no workflow, provider client, filesystem or configuration dependency.
Host executable checks, selection transactions and launch lifecycle remain with
their current owners while Slice 10 resolves their narrower boundaries.
