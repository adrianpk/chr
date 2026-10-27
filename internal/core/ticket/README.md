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

The ticket index owns SQLite ticket rows/tags, discovery, rebuilding, reconciliation,
index status and durable key reservation. UpsertInTransaction shares the original
row/tag bindings with native projection writes in the caller's transaction; it does
not validate, invalidate snapshots or commit. Public index writes retain their
existing validation and snapshot invalidation order. Native composition retains
one shared SQLite handle and source/cache authority.

NewIndexFromDB constructs Index, which provides Upsert,
ListProject, GetProjectByKey and DeleteProject. Discoverer.Discover,
Rebuilder.Rebuild, Reconciler.ReconcileProject/ReconcileTicket and
IndexStatusChecker.CheckIndex retain the original operation boundaries.
KeySequenceRepository.ReserveKey retains durable reservation semantics.
Directory resolves the existing legacy document path.

Repository.Create owns ticket file admission, allocation, rendering and creation.
NewRepository requires a key allocator and passes its configured prefix to that
allocator after validation. WithKeyPrefix, WithRenderer and WithRepositoryClock
configure the existing writer dependencies. Native namespace/envelope observation
remains outside this package. ReadNumber returns NumberObservation plus error;
NumberFromText and DefaultKeyPrefix share the original lexical key contract.

LoopQuery owns eligible-status validation, normalized filters and deterministic
ordering over the ticket index. BoardQuery, BuildBoard and ReadBoardDocument own
status/review attention and the bounded synopsis projection. App and adapters use
the same Board values; native workflow reads use the same selection and projection
rules while retaining their inspection locks, paging and source-hash guards.
Shared workflow/state filter/order values do not introduce a coordinator import.
CompareKeys preserves numeric same-prefix keys and lexical fallback; priority,
severity and timestamp ties retain their previous order and nil handling.
