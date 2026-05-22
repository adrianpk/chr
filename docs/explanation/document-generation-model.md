# Document Generation Model

ChrysoCoord treats project knowledge as Markdown-first material.

Commands can create and connect documents, but the documents remain normal files
that a technical user can inspect and edit. The command layer is there to reduce
ceremony, preserve provenance, and keep related artifacts linked. It is not a
replacement for reading or editing the files directly.

## Authoring Flow

The intended progression is:

1. Capture rough knowledge through a dump or quiz.
2. Promote useful material into a draft.
3. Turn the draft into a specification when behavior is clear enough.
4. Turn the specification into a plan when implementation work is ready.
5. Track the plan through slices, tasks, commits, branches, and pull requests.

Not every project needs every step. The value is that each step has a durable
artifact when the project needs it.

## Dumps

A dump captures broad project or use-case context. It is useful when the user
has a lot of information that is not yet structured.

Dump material can later feed a draft, specification, plan, ticket, or execution
context. The dump is not expected to be polished; it preserves knowledge before
the agent or user shapes it into a stricter document.

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

## Promotion

Promotion records how one artifact became another. For example, answered quiz
material can be promoted into an editable draft.

Promotion should keep enough source context for later review without making the
generated document feel locked. Users can edit promoted Markdown directly after
it is created.

## Conversation History

ChrysoCoord does not depend on implicit prior conversation for ticket execution
or authoring continuity. Relevant discussion should be captured as Markdown
artifacts, transcript events, dumps, quizzes, drafts, specs, plans, or tickets.
