# Ticket records and document rules

This package owns ticket and indexed-ticket records, exact legacy document
rendering, filesystem parsing, field validation and board status classification.
It preserves record field order, timestamps, optional metadata, text/JSON bytes,
legacy status aliases and error meanings. Consumers import this owner directly.

Native configuration/workflow code retains guarded file mutations, authorizations,
review/merge admission and execution/readiness orchestration. The ticket owner
cannot import the root coordinator. Shared rendering and validation operations
serve the existing native metadata paths without compatibility re-exports.

The metadata child owns frontmatter/review-reference parsing and optional integer
decoding. It cannot depend on ticket records or their coordinator consumers.

Ticket and Indexed are document/projection values. Renderer.Render, Parser.Parse,
ValidateIndexed and the status/classification functions retain their original
contracts. Native mutation paths consume the same WriteReviewRefsField and
ValidateReviewRef implementation as document rendering and validation.
