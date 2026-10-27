# Project

This package owns registered project identity, the SQLite registry, active-project
selection and registry diagnostics. Constructors consume directory capabilities;
diagnostics receive the configured workspace directory name. Host environment
resolution, workspace configuration policy and authoring transactions retain
their existing owners. The registry preserves schema, ordering, timestamps,
active-selection file contents and normal missing-project errors.

The only internal dependency is core/database. No root core, adapter or workflow
coordinator is imported. Callers use this package directly.

ResolveDirectory validates the existing shared project directory contract. It
returns the absolute path after the original existence/type checks; workspace
initialization and operational namespace callers use that same operation.

ResolvePath and ContainsPath own project-relative spelling and lexical containment.
ContainsPath does not resolve symlinks or replace a caller's authoritative path
guard. WriteFileAtomically retains the existing adjacent .tmp replacement and
errors; it does not replace bounded, synced workflow writes or authoring journals.
Those remain distinct durability contracts owned by their callers.

GitReader owns the existing Git command transport shared by guided checkpoints
and native workflow Git readers. It retains the same flags, directory/environment
selection, process-group cancellation, wait delay, output bounds and diagnostics.
Native integration keeps preparation, credentials, signing and proof policy;
authoring keeps checkpoint selection, hashes, references and mutation order.

ResolveRepositoryDirectory preserves the existing Git common-directory lookup,
caller context, symlink resolution and error contract. Runtime checkout mounts
and native workflow coordination share that path identity. Capacity admission,
slot ownership and lock conversion remain native workflow responsibilities.
