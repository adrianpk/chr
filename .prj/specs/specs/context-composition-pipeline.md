# Context Composition Pipeline

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chr_ticket_execution_context_and_ralph_loop.md`

## Purpose

Define how project knowledge moves from rough conversation artifacts into
execution-ready work.

Most durable knowledge in `chr` is Markdown. The database supports indexing,
querying, and optimization, but Markdown remains the human-readable and
versionable source of truth.

## Document Families

### Dump

A dump captures free-form user knowledge.

Dumps preserve breadth and phrasing. They may contain contradictions, repeated
ideas, incomplete thoughts, and context that is not ready for implementation.

### Quiz

A quiz converts rough context into clarified answers.

Quizzes are useful when the dump leaves important decisions implicit, when the
domain is ambiguous, or when a future spec needs sharper boundaries.

### Specification

A spec defines durable behavior, contracts, models, or policy.

Specs should be organized by intent. A large conversation may produce multiple
specs when it covers different concerns.

A spec may use `status: draft` while its behavior or policy still needs review.
That status is lifecycle metadata, not a separate document family.

### Plan

A plan translates a spec or feature area into implementation slices.

Plans describe delivery order, slice scope, expected branches, expected PR
titles, reports, validation, and non-goals.

### Tracker

A tracker records execution progress against a plan.

Trackers should not replace specs. They record delivery state, commits,
validation, PRs, and close-out state.

### Ticket

A ticket is the smallest operational work item that an agent can attempt.

Tickets may be generated from plans, created directly by a user, or refined from
existing context.

## Valid Transitions

Supported transitions include:

- dump to quiz to spec;
- dump plus quiz to spec;
- dump directly to spec when the dump is clear enough;
- quiz directly to spec when the quiz captures the required structure;
- spec with `status: draft` to accepted or final spec;
- spec to refined spec with or without quiz;
- spec to plan and tracker;
- refined spec to plan and tracker;
- plan to tracker;
- plan and tracker to tickets;
- ticket to execution context;
- ticket to refined ticket through quiz;
- ticket to code context.

Shortcuts are allowed when the target artifact remains useful and auditable.

## Candidate Commands

The product may eventually expose explicit transition commands.

Core transition candidates:

- `spec-from-dump-with-quiz`;
- `refine-spec-with-quiz`;
- `plan-from-spec`;
- `tickets-from-plan`;
- `execute-ticket`;
- `build-code-context`.

Shortcut candidates:

- `spec-from-dump`;
- `spec-from-quiz`;
- `refine-spec`.

These names are not final CLI contracts. They identify workflow capabilities
that may become commands, internal services, or agent playbooks.

## Readiness Checks

### Dump Readiness for Spec

A dump is ready to become a spec when it contains:

- a clear subject;
- enough stable vocabulary;
- identifiable decisions or candidate decisions;
- explicit unknowns;
- enough boundaries to avoid a vague mega-spec.

### Spec Readiness for Plan

A spec is ready for planning when it defines:

- goals;
- non-goals;
- expected behavior;
- relevant contracts;
- validation expectations;
- open questions that do not block slicing.

### Ticket Readiness for Execution

A ticket is ready for execution when it identifies:

- one concrete problem;
- project context;
- expected outcome;
- acceptance criteria or equivalent completion signal;
- known constraints;
- validation expectations.

## Provenance

Every generated artifact should preserve enough provenance to reconstruct why it
exists.

Useful provenance includes:

- source dumps, quizzes, or specs;
- originating plan or tracker;
- ticket key;
- execution ID;
- generation timestamp when relevant.

## Versioning and Conflict Handling

The pipeline should detect when a newer artifact contradicts an older one.

When conflict exists, the newer artifact does not automatically erase the older
one. The system should record the conflict and require either a clear decision
or an explicit supersession note.

## Open Questions

- What metadata fields should every `.prj` document carry?
- Which transitions require user approval?
- How should document families be indexed?
- Should insights be stored inside source documents or as separate derived
  artifacts?
- When should ad hoc tickets generate specs retroactively?
- Should transition commands live in `chr`, in agent playbooks, or in both?
