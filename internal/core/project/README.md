# Project

This package owns registered project identity, the SQLite registry, active-project
selection and registry diagnostics. Constructors consume directory capabilities;
diagnostics receive the configured workspace directory name. Host environment
resolution, workspace configuration policy and authoring transactions retain
their existing owners. The registry preserves schema, ordering, timestamps,
active-selection file contents and normal missing-project errors.

The only internal dependency is core/database. No root core, adapter or workflow
coordinator is imported. Callers use this package directly.
