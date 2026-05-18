# Specification Workflow

Use this playbook when a topic is in active design/brainstorming, ready for stable specification, or needs model/SQL alignment before implementation.

## Intent

- Preserve design context across sessions.
- Keep brainstorming material in curated, editable form.
- Avoid losing decisions, tensions, and open questions.

## Artifact Pairing

For each draft in `ops/specs/drafts/`, maintain a paired quiz file in `ops/conv/quiz/`.

- Draft: `ops/specs/drafts/...-draft.md`
- Quiz notes: `ops/conv/quiz/...-quiz.md`

Both files are living artifacts until promoted.

Stable specifications may also be created directly under `ops/specs/specs/` when the behavior is already clear enough and no draft cycle is useful.

When a specification introduces business models, persistence, or durable data contracts, create a sibling model file with the same basename and `-model.md` suffix.

Examples:

- Spec: `ops/specs/specs/auth-service.md`
- Model: `ops/specs/specs/auth-service-model.md`
- Draft spec: `ops/specs/drafts/auth-service-draft.md`
- Draft model: `ops/specs/drafts/auth-service-model.md`

## Rules

1. Keep both files in English.
2. `...-draft.md` captures structured behavior/proposal.
3. `...-quiz.md` captures curated discussion, not raw transcript.
4. Include `Open Questions` explicitly in both artifacts.
5. Mark provisional decisions clearly (for example: `keep`, `unsure`, `discard`).
6. Remove noise, typos, and irrelevant detours from quiz notes.
7. Keep the quiz file editable for future expansion.
8. Do not use title or heading suffixes in parentheses such as `(Draft)`.
9. Use semantic fields for state metadata (for example `Status`) only when they add operational value.
10. Add explicit cross-references: each draft must reference its quiz counterpart and vice versa.
11. Before commit, run a full consistency pass across playbook, templates, and concrete artifacts.
12. For each iteration, prefer one cohesive commit over corrective micro-commits.
13. Review newly introduced business models with their database shape in the same spec cycle.
14. Do not approve a newly introduced business model without its SQL/table/query/migration counterpart when persistence is involved.
15. Cross-reference spec and model files in both directions.

## Model Files

Use a `-model.md` file when the spec introduces:

- business entities
- value objects
- status/lifecycle fields
- role or permission concepts
- persistence tables
- SQL queries
- migrations
- durable API payloads that mirror stored data

Model files should include only the models involved in that spec.

For easy review, keep attribute and column order aligned:

- list business model attributes in the same order used by the Go model when known
- list database columns in the same order used by the SQL table/migration
- when both are present, keep model attributes and table columns parallel enough to compare quickly

Model files should include, as applicable:

- business model definitions
- field list and semantics
- lifecycle/status rules
- invariants and validation rules
- SQL table shape
- indexes and uniqueness rules
- query/migration notes
- relationship to existing reference implementations

Keep purely behavioral or workflow-only specs model-free when no durable model is involved.

## Template Locations

- `ops/templates/spec-draft-template.md`
- `ops/templates/spec-quiz-template.md`
- `ops/templates/spec-model-template.md`

## Promotion Path

- Iterate with drafts in `ops/specs/drafts/` and quiz notes in `ops/conv/quiz/` until stable.
- Promote to `ops/specs/specs/*` once behavior and decisions are sufficiently converged.
- Promote paired `-model.md` files with the spec when the model is part of the stable contract.

## Quiz-Driven Direct Spec Path

When a topic is explored through the quiz playbook, the pre-spec artifacts may be:

- Raw capture: `ops/conv/quiz/...-quiz-raw.md`
- Curated quiz notes: `ops/conv/quiz/...-quiz.md`

In that path, a separate `ops/specs/drafts/...-draft.md` artifact is optional. A sufficiently complete curated quiz file can replace the draft and serve as the direct source for a formal spec under `ops/specs/specs/`.

Use this path when:

- the quiz was bounded and completed
- the raw capture preserves the question/answer flow
- the curated quiz notes preserve decisions, tensions, and open questions
- the user explicitly agrees that no separate spec draft is needed

Do not create a draft merely to satisfy ceremony when `...-quiz-raw.md` and `...-quiz.md` already provide stronger alignment context.
