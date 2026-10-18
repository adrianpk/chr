# Workflow Tracker Projections

Tickets, set-owned slice descriptors and immutable receipts own native progress.
A tracker supplies generated Markdown order and references. It never supplies
selection, claims, validation success, approval, integration or completion
authority. Revision-bound sets use the order-only `chr-tracker/v2` contract;
existing `chr-tracker/v1` views retain their optional historical contract.
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

The v1 view distinguishes task status and phase, slice phase, canonical PR and
delivery receipt. Host-validated work remains separate from receipt-finalized
integration. Set completion requires its own closing receipt. The provenance
retains source IDs, revisions and hashes, shared delivery/closure receipt hashes,
the source snapshot hash, output hash, generation time and current/stale state.
Each file is bounded to 1 MiB; a namespace supports at most 128 registered views.
Symlink traversal and nonregular source/output files are rejected.

Unchanged export retries retain the generated image and time. Inspection is
read-only and verifies current authoritative bytes and output bytes; it does not
trust the last stored state marker. Missing output or changed sources make the
view stale. Registered v1 views are marked stale after relevant successful domain
writes and refreshed only by explicit export or repair. Existing registrations
retain their version; inspection never upgrades a v1 document.

## Revision-bound order-only maps

Each successful shared plan generation or regeneration retains its rendered
body, structured intended order, tracker and binding under the plan artifact's
`files/plan-revisions/<revision>-<body-hash>/` directory. Its manifest registers
the retained files. Fresh generation advances the revision even when rendered
body bytes match; retrying an explicit committed operation reuses its revision.
Publication uses the existing recoverable authoring journal. The scoped current
tracker keeps its existing `tracker/00000-<scope>-tracking/artifact.md` path.

Before approved native materialization, the companion names stable planned
slice/task identities and positive order but has no allocated ticket references.
It is explicitly non-executable. Immutable `planned-order.json`,
`planned-tracker.md` and `planned-binding.json` retain that original intent.
The plan body hash is independent of companion metadata. Retained source hashes
protect the original planned images from reinterpretation during export repair.

A native set carrying `plan_revision` exports v2 by default when no older
registration exists. V2 keeps the tracker family, registered set path and table
layout, with blank legacy progress cells. Its structural source hash and bytes
contain project/namespace, plan identity/revision/body hash, source observations,
set/materialization identity, ordered slice/task IDs and titles, and canonical
ticket keys/references. They exclude complete record revisions/hashes, status,
processing phase, claims, attempts, review, delivery and closure state.

V2 inspection reads native set/task owners without receipt payloads. A claim,
checkpoint, validated result or ticket progress does not stale or rewrite the
map. Structural changes and edited output are detected independently. Repair of
unchanged structure retains its generation time; structural replacement updates
it. A runner reads live tickets and receipts to determine remaining work and
uses the map only for declared order. Neither version imports processing state
from Markdown.

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
rolled back because tracker output failed. Optional v1 projection failure does
not prevent direct native execution. A manual runner must inspect native records
and receipts, never continue using a stale projection as authority.

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
