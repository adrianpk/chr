# Artifact Generation Backends

A project's explicit `harnesses.generation` profile selects the generation
transport for Dumps, Quiz questions, Specs, acceptance features, Plans, tracking,
structured conversation, explanations and review generation. The shared owner
freezes that selection once for an operation; nested calls, retries and resumes
keep it. New independent operations read persisted defaults.

Profiles live in private user `harnesses.yml`; project configuration stores their
names. See [Select worker and generation harnesses](../how-to/select-harnesses.md)
for preparation, switching and recovery. Worker selection is independent.

## Legacy dispatch

When `harnesses.generation` is absent or cleared, action-specific legacy dispatch
continues. Long artifact generation uses Codex by default. These variables accept
`codex` or `openai`; unset/empty values preserve the Codex default:

- `CHR_DUMP_GENERATION_BACKEND`
- `CHR_SPEC_GENERATION_BACKEND`
- `CHR_ACCEPTANCE_FEATURE_GENERATION_BACKEND`
- `CHR_PLAN_GENERATION_BACKEND`
- `CHR_PLAN_TRACKING_GENERATION_BACKEND`
- `CHR_CONVERSATION_GENERATION_BACKEND`

`openai` means the existing direct API adapter, not a Pi provider alias. Other
values fail instead of choosing a fallback. Ordinary Quiz question generation
keeps its direct API default; contextual Quiz and agentic review keep their
existing Codex path. An explicit project generation profile takes precedence.

```sh
CHR_PLAN_GENERATION_BACKEND=openai chr plan generate --use-case blog
```

## Codex settings

Legacy Codex generation reads:

- `CHR_CODEX_COMMAND`: executable; default `codex` (contextual Quiz keeps its existing command semantics).
- `CHR_CODEX_MODEL`: optional generation model.
- `CHR_CODEX_TIMEOUT_SECONDS`: optional per-run timeout, bounded by the caller.

Named Codex profiles retain their explicit executable/model instead. Native
workers keep their historical CLI-default model when no worker profile is set;
a generation-only environment model is not retroactively attributed to workers.

## Pi generation

Native generation requires Pi 1.3.0, an explicit provider/model and a private
static API-key credential reference. Optional custom model definitions are
credential-free private resources. OAuth generation and dynamic credentials are
unsupported. The separate Docker worker bridge is not a generation executable.

Each call uses isolated private state with discovered tools, settings,
extensions, skills and project-local model overrides excluded. Generation cannot
edit the project or run tools. The host owns typed/schema validation and artifact
publication; malformed or unsupported output fails without a fallback or partial
replacement. Caller deadlines, bounded streams and shared capacity still apply.

## Retained evidence

The host records configured selection, executable/protocol identity, configuration
revision, observed model when supplied, output validation and available usage in
`harness-run.json`. Missing usage stays unknown. An observed model or token count
is not a whole-call budget certificate. Direct API operations retain their own
API identity; neither transport selects worker, merge or publication authority.
