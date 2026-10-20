# API Boundaries

Status: accepted

## Context

The blog reference project is intentionally small. It exists to validate
chr artifact generation and agent-work execution, not to model a
production publishing system.

## Decision

The initial API uses the Go standard library HTTP stack and in-memory stores.
Handlers should stay thin and delegate behavior to internal package types.

HTTP responses should be JSON for both successful responses and expected client
errors. Unknown post slugs should use `404` responses, not generic internal
server errors.

## Consequences

- New features should avoid external router dependencies.
- The comments feature should use `http.ServeMux` method/path patterns.
- Comments should be associated with post slugs in memory.
- Persistence, authentication, moderation, and notification concerns are
  outside the first comments implementation.
