# Conventions

## Language

Use English for code, comments, documentation, commit messages, and ChrysoCoord artifacts.

## Go Code

- Prefer the Go standard library before adding dependencies.
- Keep packages small and behavior-oriented.
- Use explicit data flow over hidden global state.
- Return errors with useful context.
- Place tests beside the package behavior they validate.
- Prefer table-driven tests for input and output behavior.

## Change Discipline

- Keep changes scoped to the requested task.
- Avoid unrelated refactors.
- Preserve existing project decisions unless the task explicitly changes them.
