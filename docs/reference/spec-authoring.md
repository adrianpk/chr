# Spec Authoring

Specs use canonical `ops/<namespace>/spec/<id>/artifact.md` documents and their
existing `metadata.json` envelopes. Each manual create allocates an independent
identity, including when several specs share a scope. Generated specs retain
their allocated identities. Concrete-document listing is separate from the
latest-in-scope inventory used by existing generation adapters.

## Lifecycle and Scope

The document frontmatter owns lifecycle: writes use `draft` or `approved`.
Reads normalize `accepted` and `final` to `approved` without changing files.
A generic envelope status such as `created` does not imply approval. Shared
writes update the envelope's lifecycle projection from the document state.

Manual creation supports draft or approved directly without a model, dump or
quiz prerequisite. Generation starts in draft. A save preserves the selected
identity, project and case association. Its optional explicit status selects
draft or approved; omitting status retains the native state. Creation time,
unknown frontmatter/metadata and retained generation provenance survive editing.
The editable Body excludes managed frontmatter, generated title scaffolding
and retained provenance.

## Guarded Publication

`ReadSpecDocument` observes the exact body and metadata bytes and returns a
revision bound to the project root and document reference. `MutateSpec` requires
that revision for save and compares it under the shared authoring owner.
Metadata-only changes invalidate the same revision. Each operation ID binds
immutable arguments; an identical retry acknowledges its committed result,
and changed arguments conflict.

Body, envelope metadata and operation receipt publish through the existing
recoverable authoring journal. Reads fail closed during pending publication;
an explicit mutation recovers it before comparing its request. Each image is
bounded to 1 MiB and the receipt inventory to 4096 entries. Catalog reads are
bounded to 4096 inventory entries and 8 MiB of document/envelope images.

Generation observes its target or its absence before requesting output, then
compares that revision and retained scope under the shared writer. An intervening
edit or another generated publication rejects stale output. Both direct spec and
quiz-to-spec generation use that publication boundary. The target revision is
caller-owned and excluded from model serialization.

## Generation Context

Spec generation receives project context, the canonical general dump and
matching optional case discovery. Project generation excludes case-scoped
material; case generation excludes other cases and the project-wide quiz.
Raw capture siblings, archived material and unrelated spec/plan documents do
not feed initial spec generation. Ambiguous general dumps require explicit
reconciliation. Editing a spec changes its source hash without rewriting
historical plans or granting execution authority.
