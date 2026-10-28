# Legacy ticket metadata

This package owns the existing frontmatter parser and review reference value.
Authoring sessions and transcripts consume its scalar fields directly; ticket
parsing consumes both scalars and review references. The opening/closing delimiter
rules, ignored malformed fields, commits lists, review block scanning, empty
collections and diagnostics retain their original behavior. This codec performs
no ticket validation, persistence, selection or authoring transitions.

OptionalInteger owns the existing optional integer codec shared by ticket
parsing and native status-transition metadata. Empty input remains zero and
malformed integers retain their original ticket parse error.

The shared edit contract owns required-field updates, UTC updated_at formatting,
scalar insertion and exact frontmatter line/block ranges. FrontmatterBounds and
ReplaceFrontmatterLines retain the original opening/closing delimiter and body
bytes; TextRange exposes the same half-open bounds to real editor consumers.
SetField and UpdateRequiredField remain distinct operations: the latter rejects a
missing field. Map iteration, preferred insertion keys and empty-value behavior
are unchanged. A zero-width ops/metadata.ReplaceLines performs the original line
insertion. Native and legacy operations share these pure edits while retaining
their own authorization, transitions, file locks and guarded persistence.
