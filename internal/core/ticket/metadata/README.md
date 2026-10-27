# Legacy ticket metadata

This package owns the existing frontmatter parser and review reference value.
Authoring sessions and transcripts consume its scalar fields directly; ticket
parsing consumes both scalars and review references. The opening/closing delimiter
rules, ignored malformed fields, commits lists, review block scanning, empty
collections and diagnostics retain their original behavior. This codec performs
no ticket validation, persistence, selection or authoring transitions.
