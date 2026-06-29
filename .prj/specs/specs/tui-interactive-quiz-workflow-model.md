# TUI Interactive Quiz Workflow Model

Status: accepted

Spec: `.prj/specs/specs/tui-interactive-quiz-workflow.md`

## Purpose

Define the durable models, file contracts, and lifecycle invariants required by
the TUI interactive quiz workflow.

This model is file-backed. It must not depend on a database.

## Scope Identity

### QuizScope

Fields:

- `project_id`: resolved project identity.
- `case_slug`: optional use-case slug.

Semantics:

- Empty `case_slug` means project-wide quiz.
- Non-empty `case_slug` means use-case-scoped quiz.
- Scope determines canonical artifact paths.
- Scope determines runtime state cache path.

Derived scope ids:

- project scope: `project`;
- use-case scope: `<case-slug>`.

## Runtime State

### QuizRuntimeState

Fields:

- `id`: stable runtime id.
- `project_id`: resolved project identity.
- `case_slug`: optional use-case slug.
- `status`: `idle`, `active`, `done`, `cancelled`, or `blocked`.
- `mode`: `deep` or `fast`.
- `question_budget`: maximum question count.
- `questions_asked`: count of activated questions.
- `answers_recorded`: count of answered questions.
- `skips_recorded`: count of skipped questions.
- `active_topic`: current topic or coverage group.
- `coverage`: list of topic coverage records.
- `active_question`: current question, nullable.
- `trail`: current Ariadne trail from trunk to active question.
- `candidates`: internal fast candidates.
- `raw_path`: relative path to raw quiz artifact.
- `curated_path`: relative path to curated quiz artifact.
- `updated_at`: last runtime update timestamp.

Invariants:

- `question_budget` defaults to 50.
- `questions_asked` must not exceed `question_budget`.
- `mode` defaults to `deep`.
- `active_question` is required when `status` is `active`.
- `active_question` must be empty when `status` is `idle`, `done`,
  `cancelled`, or `blocked`, unless the blocked state needs to expose the
  failed question for recovery.
- `candidates` are internal runtime data and are not curated artifacts.

### QuizStatus

Values:

- `idle`: quiz exists or is discoverable, but no active runtime loop is open.
- `active`: a question is active and awaits answer, skip, done, or cancel.
- `done`: the user ended the loop with sufficient context for now.
- `cancelled`: the user stopped the loop without claiming sufficiency.
- `blocked`: the service cannot safely continue.

`done` and `cancelled` do not permanently close the quiz.

### QuizMode

Values:

- `deep`: model-backed next-question generation after each routing point.
- `fast`: service may use internal candidates for lower perceived latency.

Mode changes apply only to future generation.

## Question Model

### QuizQuestion

Fields:

- `id`: stable question id.
- `prompt`: visible question text.
- `required`: whether an answer is required to proceed.
- `topic`: coverage topic.
- `depth`: follow-up depth, starting at 1.
- `parent_id`: parent question id for follow-ups.
- `trunk_id`: root trunk question id for the current branch.
- `route_id`: optional deterministic route id.
- `rationale`: optional internal explanation for why this question is useful.
- `created_at`: generation or activation timestamp.

Invariants:

- `prompt` must be non-empty.
- `depth` must be 1, 2, or 3.
- `parent_id` is empty only for depth 1.
- `trunk_id` is the question id when depth is 1.

## Answer Model

### QuizEntry

Fields:

- `id`: stable entry id.
- `question_id`: source question id.
- `question`: question text at activation time.
- `answer`: answer text, empty when skipped.
- `status`: `answered` or `skipped`.
- `topic`: coverage topic.
- `depth`: follow-up depth.
- `parent_id`: parent question id when present.
- `mode`: mode used when the question was activated.
- `tags`: final retrieval tags, optional.
- `created_at`: activation timestamp.
- `answered_at`: answer or skip timestamp.

Invariants:

- `status` must be `answered` or `skipped`.
- `answer` is required when `status` is `answered`.
- `answer` must be empty when `status` is `skipped`.
- `tags` must be derived after activation and answer interpretation.
- Speculative candidates must not become `QuizEntry` records.

## Coverage Model

### QuizCoverageTopic

Fields:

- `id`: stable topic id.
- `label`: concise topic label.
- `status`: `pending`, `active`, `covered`, or `deferred`.
- `question_count`: activated questions in this topic.
- `answer_count`: answered questions in this topic.
- `skip_count`: skipped questions in this topic.

Semantics:

- Coverage prevents one topic from consuming the entire quiz budget.
- The runtime may force a topic switch after excessive local depth.

## Ariadne Trail

### QuizTrailNode

Fields:

- `question_id`: question id.
- `prompt`: short prompt or prompt summary.
- `topic`: coverage topic.
- `depth`: follow-up depth.

Semantics:

- The trail shows how the active question relates to the trunk question.
- The trail must be recoverable from runtime state.
- The TUI may render the trail compactly.

## Fast Candidate Model

### QuizCandidate

Fields:

- `id`: candidate id.
- `prompt`: possible future question text.
- `topic`: coverage topic.
- `depth`: proposed depth.
- `parent_id`: parent question id when applicable.
- `route_id`: deterministic route id when applicable.
- `hints`: internal non-retrieval hints.
- `created_at`: generation timestamp.

Invariants:

- A candidate is not a quiz entry.
- A candidate is not raw capture material.
- A candidate is not curated material.
- A candidate must not receive final retrieval tags.
- A candidate may be discarded without user-visible artifact changes.

## File Layout

### Canonical Quiz Files

Project-wide quiz:

```text
.chr/quizzes/project-quiz-raw.md
.chr/quizzes/project-quiz.md
```

Use-case-scoped quiz:

```text
.chr/quizzes/<case-slug>-quiz-raw.md
.chr/quizzes/<case-slug>-quiz.md
```

Runtime state cache:

```text
.chr/tmp/quiz-runtime/project.json
.chr/tmp/quiz-runtime/<case-slug>.json
```

### Raw Artifact Format

Required frontmatter:

```yaml
project_id: <project-id>
case_slug: <case-slug-or-empty>
scope: project | use_case
status: raw
runtime_id: <runtime-id>
curated_path: .chr/quizzes/<scope>-quiz.md
updated_at: <timestamp>
```

Body sections:

```md
# Quiz Raw Capture

## Scope

## Turns

## Runtime Notes
```

Each turn records the activated question and the answer or skip outcome.

### Curated Artifact Format

Required frontmatter:

```yaml
project_id: <project-id>
case_slug: <case-slug-or-empty>
scope: project | use_case
status: curated
raw_path: .chr/quizzes/<scope>-quiz-raw.md
updated_at: <timestamp>
```

Body sections:

```md
# Quiz Notes

## Decisions

## Constraints

## Preferences

## Open Questions

## Implementation Implications

## Provenance
```

The curated artifact is not a transcript. It contains distilled quiz outcomes.

### Runtime State Cache Format

The runtime state cache is JSON. It stores `QuizRuntimeState`.

Rules:

- It is a cache and working state file, not the human-facing source of truth.
- It may be regenerated or repaired from raw and curated artifacts only when
  enough information exists.
- It must be ignored by curated context retrieval unless a future spec says
  otherwise.
- It must be safe to delete without losing answered quiz knowledge already
  written to raw or curated artifacts.

## Service Interfaces

Candidate operation shape:

```go
type QuizOperation interface {
    ListQuizzes() (QuizListResult, error)
    StartQuiz(QuizStartRequest) (QuizOperationResult, error)
    ResumeQuiz(QuizResumeRequest) (QuizOperationResult, error)
    AnswerQuiz(QuizAnswerRequest) (QuizOperationResult, error)
    SkipQuiz(QuizSkipRequest) (QuizOperationResult, error)
    CompleteQuiz(QuizCompleteRequest) (QuizOperationResult, error)
    CancelQuiz(QuizCancelRequest) (QuizOperationResult, error)
    SwitchQuizMode(QuizModeRequest) (QuizOperationResult, error)
    RefreshQuizzes() (QuizListResult, error)
}
```

The concrete implementation remains the existing application service.

The exact Go type names may change during implementation, but the ownership
boundary must not change.

## Model Boundary Payloads

Quiz model calls must use formal request and response payloads.

The runtime must not introduce feature-local shellouts or untyped prompt
strings as the integration contract.

Candidate request families:

- `quiz_next_question`;
- `quiz_fast_candidates`;
- `quiz_answer_interpretation`;
- `quiz_curated_notes`;
- `quiz_to_draft_spec`.

The implemented draft-spec promotion path uses `quiz_to_draft_spec`. Its
request carries the fresh curated quiz body as a `curated_quiz` context section
after the service refreshes curation on done.

Each request must carry:

- request id;
- action;
- strategy;
- scope;
- mode;
- project context;
- agent protocol context;
- use-case context when scoped;
- relevant dump context;
- prior quiz entries;
- active runtime state;
- output contract;
- created time.

Each response must carry:

- response id;
- request id;
- provider;
- status;
- structured payload;
- raw text when available;
- created time.

For `quiz_next_question`, the structured payload is a question and optional
candidate list. For `quiz_to_draft_spec`, the structured payload is a draft
spec title and body that the service writes into `.chr/specs/drafts/` with
`status: draft`.

Responses must be validated before they update runtime state or write durable
artifacts.

## Validation Expectations

- Model validation rejects invalid status, mode, depth, and answer state.
- Repository tests cover project and use-case file paths.
- Runtime state tests cover resume after cache reload.
- Candidate tests prove discarded candidates do not appear in raw or curated
  artifacts.
- Service tests prove answer, skip, done, cancel, and mode switch semantics.
- TUI tests prove the UI dispatches intents instead of owning routing logic.

## Open Questions

- Whether runtime JSON should include a schema version from the first slice.
- Whether curated artifacts should include a compact tag block in frontmatter or
  keep tags only per derived context entry later.
- Whether archived historical quiz snapshots should live under
  `.chr/quizzes/archive/`.
