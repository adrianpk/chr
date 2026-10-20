# Document Generation Model

chr treats project knowledge as Markdown-first material.

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
6. Materialize plan tracking and tickets when the plan is ready to execute.
7. Track execution through tickets, slices, tasks, commits, branches, and pull
   requests.

Not every project needs every step. The value is that each step has a durable
artifact when the project needs it.

`Draft` is not a first-class document kind in the active model. Draftness is a
status of a spec, not a separate artifact family.

## Dumps

A dump captures broad project or use-case context. It is useful when the user
has a lot of information that is not yet structured.

Dump creation produces a curated primary document plus a raw capture sibling.
The curated document feeds specifications, plans, tickets, and execution
context by default. The raw sibling preserves the original capture and
highlights, but it is not selected as regular dump context.

Dumps may declare a use-case slug. Dumps without a use-case slug are
project-wide.

Dump-style material can be written directly as Markdown today. The current CLI
already exposes the related session, quiz, legacy quiz-to-draft, and assisted
authoring boundaries.

## Quizzes

A quiz captures missing knowledge through a bounded question session.

Context-aware quizzes generate questions from a structured request containing
project context, authoring intent, target artifact, topic, and a response
contract. Assisted authoring sessions can ask one question at a time while
preserving transcript events and curated answers. In both cases, the useful
output is persistent project material, not unstructured chat history.

Quizzes may declare a use-case slug. Quizzes without a use-case slug are
project-wide.

## Legacy Quiz-To-Draft

The active document flow does not have a general promotion operation. Current
spec and plan generation read accumulated workspace material directly and write
the target artifact for the selected project or use case.

The CLI still exposes `author promote quiz-to-draft` as a compatibility path
for older authoring sessions. That command turns one quiz artifact into an
editable draft artifact and records provenance. New flows should generate or
edit specs instead of creating draft artifacts.

## Specs, Plans, and Tracking

Specs are generated from accumulated project and use-case material. They are the
place where expected behavior is made explicit before implementation planning.

Spec generation produces the Markdown spec only. For behavior-testable use
cases, the separate acceptance feature workflow derives a linked Gherkin feature
artifact from the spec, quiz, and dump context. If scenario details are missing,
it writes feature-round notes with inferred defaults before another generation
attempt. The feature-round material is acceptance-specific; it should complete
missing scenario details without moving Gherkin-shaped questioning into the
general quiz. This keeps general requirements discovery separate from
`.feature` formalization.

The `.feature` artifact is also intended to become source material for a later
agent workflow that generates executable acceptance tests. For Go projects, that
future workflow will likely target `godog` step definitions and validation runs.

Plans are generated from the same workspace material plus existing specs. A plan
is an implementation decomposition: slices, ordering, dependencies, tasks,
branch expectations, pull request expectations, and validation criteria.

Plan tracking is a separate materialization step. It creates a tracking
Markdown document and one ticket per generated slice. The tracking document is
the ordered execution map; tickets are the execution surface for agents and
operators.

Generated plan regeneration archives the previous generated plan before writing
the replacement. Archived generated plans remain on disk for recovery, but
normal document loading excludes archived generated documents so discarded
material does not influence later generation.

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

chr does not depend on implicit prior conversation for ticket execution
or authoring continuity. Relevant discussion should be captured as Markdown
artifacts, transcript events, dumps, quizzes, specs, plans, or tickets.
