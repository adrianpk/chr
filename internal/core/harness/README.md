# Harness identity and selection

This owner contains worker and generation bindings, their exact sealing and
validation rules, role references, readiness and selection projections, and
bounded private profile loading. Consumers import this owner directly.

`LoadProfiles` reads only the explicit private harnesses.yml file, with the
original 64 KiB, regular-file, permission and strict field checks. Profile
revisions hash semantic references without reading credentials. Loading never
installs, launches or authenticates an adapter.

`SealBinding`, `ValidateBinding`, `SealGenerationBinding` and
`ValidateGenerationBinding` preserve JSON field order, capability evidence,
digest bytes, profile-name bounds and error meanings. GenerationSelection is a
value projection of the operation identity and binding without private launch
references. Executable pairs a public reference with the private frozen command;
that host preparation value is never serialized into Binding.

Adapter-neutral capabilities come from internal/harness. This owner has no
workflow or configuration-coordinator dependency. Complete WorkspaceConfig image
parsing and revision stay with their existing configuration owner. Native command
fencing, selection transactions and generation launch lifecycle retain their own
execution and configuration responsibilities.

The `host` child owns executable resolution, streaming fingerprints and retained
command admission. The `daemon` child owns lifecycle state, persistent heartbeat
and stop requests, start guards and project loading. These children consume the
parent's values; the parent cannot import either child.
