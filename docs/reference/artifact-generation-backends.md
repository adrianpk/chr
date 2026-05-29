# Artifact Generation Backends

ChrysoCoord routes long artifact generation through Codex by default.

This applies to:

- dump curation;
- spec generation;
- plan generation;
- plan tracking and ticket payload synthesis.

Quiz question generation remains on the direct quiz model path because it is
interactive and one question at a time.

## Defaults

When no backend override is set, these artifact operations use Codex:

- `CHC_DUMP_GENERATION_BACKEND`
- `CHC_SPEC_GENERATION_BACKEND`
- `CHC_PLAN_GENERATION_BACKEND`
- `CHC_PLAN_TRACKING_GENERATION_BACKEND`

Set one variable to `openai` to use the direct OpenAI-backed path for that
artifact kind.

Example:

```sh
CHC_PLAN_GENERATION_BACKEND=openai chc plan generate --use-case blog
```

Any value other than `openai`, including an empty value, uses Codex.

## Codex Runtime Settings

Codex-backed artifact generation also reads:

- `CHC_CODEX_COMMAND`: command used to invoke Codex; default is `codex`.
- `CHC_CODEX_MODEL`: optional model passed to Codex.
- `CHC_CODEX_TIMEOUT_SECONDS`: optional per-run timeout in seconds.

Codex runs are read-only against the project checkout. The Go service owns file
writes, ticket keys, ticket paths, tracking materialization, and validation.
