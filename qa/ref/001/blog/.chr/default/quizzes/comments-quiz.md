---
project_id: "blog"
case_slug: "comments"
scope: "comments"
status: "curated"
raw_path: ".chr/default/quizzes/comments-quiz-raw.md"
updated_at: "2026-05-30T00:00:00Z"
---

# Quiz Notes

## Decisions

- Add comments through HTTP endpoints scoped to published post slugs.
- Store comments in memory for the first implementation.
- Accept a display name and body for each comment.
- Return JSON for successful responses and errors.

## Constraints

- Do not add persistence in the first implementation.
- Do not add moderation, authentication, spam controls, notifications, or admin
  tooling.
- Preserve the standard-library routing approach.

## Preferences

- Keep handlers small and move comment behavior into internal package code.
- Use table-driven tests for endpoint behavior.

## Open Questions

- None recorded.

## Implementation Implications

- The generated spec should include `POST /posts/{slug}/comments` and
  `GET /posts/{slug}/comments`.
- Unknown post slugs should return JSON `404` responses.
- Comments should be associated with post slugs rather than a global thread
  model.
- The seeded base post `404` bug is covered by manual ticket `CHR-1`; generated
  comments tickets should not duplicate that preexisting bug ticket.

## Provenance

- raw quiz: .chr/default/quizzes/comments-quiz-raw.md
