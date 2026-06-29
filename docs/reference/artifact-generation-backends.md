# Artifact Generation Backends

Chryso routes long artifact generation through Codex by default.

This applies to:

- dump curation;
- spec generation;
- plan generation;
- plan tracking and ticket payload synthesis.

Quiz question generation remains on the direct quiz model path because it is
interactive and one question at a time.

## Defaults

When no backend override is set, these artifact operations use Codex:

- `CHR_DUMP_GENERATION_BACKEND`
- `CHR_SPEC_GENERATION_BACKEND`
- `CHR_PLAN_GENERATION_BACKEND`
- `CHR_PLAN_TRACKING_GENERATION_BACKEND`

Set one variable to `openai` to use the direct OpenAI-backed path for that
artifact kind.

Example:

```sh
CHR_PLAN_GENERATION_BACKEND=openai chr plan generate --use-case blog
```

Any value other than `openai`, including an empty value, uses Codex.

## Codex Runtime Settings

Codex-backed artifact generation also reads:

- `CHR_CODEX_COMMAND`: command used to invoke Codex; default is `codex`.
- `CHR_CODEX_MODEL`: optional model passed to Codex.
- `CHR_CODEX_TIMEOUT_SECONDS`: optional per-run timeout in seconds.

Codex runs are read-only against the project checkout. The Go service owns file
writes, ticket keys, ticket paths, tracking materialization, and validation.
