# Document Generation Model

ChrysoCoord treats project knowledge as Markdown-first material.

Commands can create and connect documents, but the documents remain normal files
that a technical user can inspect and edit. The command layer is there to reduce
ceremony, preserve provenance, and keep related artifacts linked. It is not a
replacement for reading or editing the files directly.

## Authoring Flow

The intended progression is:

1. Capture rough knowledge through a dump or quiz.
2. Use dump and quiz material to create a specification.
3. Keep the specification in `status: draft` while it needs review.
4. Mark or treat the specification as accepted/final when behavior is clear
   enough.
5. Turn the specification into a plan when implementation work is ready.
6. Track the plan through slices, tasks, commits, branches, and pull requests.

Not every project needs every step. The value is that each step has a durable
artifact when the project needs it.

`Draft` is not a first-class document kind in the active model. Draftness is a
status of a spec, not a separate artifact family.

## Dumps

A dump captures broad project or use-case context. It is useful when the user
has a lot of information that is not yet structured.

Dump material can later feed a specification, plan, ticket, or execution
context. The dump is not expected to be polished; it preserves knowledge before
the agent or user shapes it into a stricter document.

Dumps may declare a use-case slug. Dumps without a use-case slug are
project-wide.

Dump-style material can be written directly as Markdown today. The current CLI
already exposes the related session, quiz, promotion, and assisted authoring
boundaries.

## Quizzes

A quiz captures missing knowledge through a bounded question session.

Context-aware quizzes generate questions from a structured request containing
project context, authoring intent, target artifact, topic, and a response
contract. Assisted authoring sessions can ask one question at a time while
preserving transcript events and curated answers. In both cases, the useful
output is persistent project material, not unstructured chat history.

Quizzes may declare a use-case slug. Quizzes without a use-case slug are
project-wide.

## Promotion

Promotion records how one artifact became another. Answered quiz material and
dumps can be promoted into a specification with `status: draft` when the result
needs review.

Promotion should keep enough source context for later review without making the
generated document feel locked. Users can edit promoted Markdown directly after
it is created.

## Workspace Index

The workspace index is the local read model used to find and connect documents.
It is derived from Markdown files and can be rebuilt from them.

Documents such as dumps, quizzes, specs, and plans are associated with a use
case when they declare a canonical case slug. Documents without a use-case slug
belong directly to the project.

The structural index stores metadata, source paths, relationships, diagnostics,
source hashes, and modified times. It should not store full document bodies by
default. Bodies are read lazily when a detail view, context assembler, writer,
validator, or semantic worker needs them.

An optional semantic digest cache can store compact retrieval hints generated
from document bodies. These hints are for agents and search, not for human
display. They are cache-backed by source hash and can be regenerated when
stale.

## Conversation History

ChrysoCoord does not depend on implicit prior conversation for ticket execution
or authoring continuity. Relevant discussion should be captured as Markdown
artifacts, transcript events, dumps, quizzes, specs, plans, or tickets.
