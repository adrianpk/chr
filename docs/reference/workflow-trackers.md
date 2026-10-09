# Workflow Tracker Projections

Tickets, set-owned slice descriptors and immutable receipts own native progress.
A tracker is an optional generated Markdown view. It never supplies selection,
claims, validation success, approval, integration or completion authority.
`tracker_compat` applies only to explicitly selected unadopted legacy work; it
does not make an exported native tracker executable.

## Export and inspection

```sh
chr work tracker export --dir /path/to/project --json <slice-set-id>
chr work tracker status --dir /path/to/project --json <slice-set-id>
chr work tracker repair --dir /path/to/project --json <slice-set-id>
```

Use the exact adopted set ID from native plan materialization or selected
adoption. Export registers one view at `ops/<namespace>/tracker/<set-id>.md`.
Its checksummed provenance lives at
`ops/<namespace>/tracker/projections/<set-id>.json`. The common service exposes
the same export and inspection operations to adapters. Repair runs export alone;
it does not start an invocation, claim a task, launch a worker or contact a forge.

The view distinguishes task status and phase, slice phase, canonical PR and
delivery receipt. Host-validated work remains separate from receipt-finalized
integration. Set completion requires its own closing receipt. The provenance
retains source IDs, revisions and hashes, shared delivery/closure receipt hashes,
the source snapshot hash, output hash, generation time and current/stale state.
Each file is bounded to 1 MiB; a namespace supports at most 128 registered views.
Symlink traversal and nonregular source/output files are rejected.

Unchanged export retries retain the generated image and time. Inspection is
read-only and verifies current authoritative bytes and output bytes; it does not
trust the last stored state marker. Missing output or changed sources make the
view stale. Registered views are marked stale after relevant successful domain
writes and refreshed only by explicit export or repair.

## Edited or failed exports

Export replaces a file only when it is absent, matches the previously generated
image, matches an interrupted export's retained pending image, or has explicit
permission tied to its current hash. Existing unregistered files and operator
edits are preserved by default. Edits are never imported into tickets.

Inspect the stale view and its `observed_output_hash` before authorizing its
replacement:

```sh
chr work tracker status --json <slice-set-id>
chr work tracker repair --replace-hash <observed-output-hash> --json <slice-set-id>
```

The hash is permission to discard exactly those bytes. A different or newly
edited image rejects replacement. The CLI prints a stale result even when export
returns an error. `projection_stale` records failure separately from domain
progress; a successful task result, finalized integration or closure is never
rolled back because tracker output failed. Direct ticket-native execution can
continue. A manual runner must inspect native records and receipts, never
continue using a stale projection as authority.

Export writes a stale pending marker before replacing output and confirms current
state afterward. Retry rebuilds from authoritative records and recognizes known
generated images after interruption or a lost response. This state is separate
from the domain transition journal, so a failed export does not block native
workflow recovery. If the stale marker itself cannot persist, inspection still
detects changed source/output bytes. Edited or corrupted provenance is preserved
and blocks export rather than granting replacement; recover known provenance
before retrying. These operations do not repair unrelated SQLite or run records.

See [Workflow execution](workflow-execution.md) for authoritative progress and
[Workflow controls](workflow-controls.md) for native entrypoints and recovery.
