---
project_id: "blog"
case_slug: "comments"
scope: "comments"
status: "raw"
runtime_id: "seed-comments"
curated_path: "ops/default/quiz/<id>/artifact.mdcomments-quiz.md"
updated_at: "2026-05-30T00:00:00Z"
---

## Turn

- question_id: q1
- status: answered
- mode: deep
- topic: comments-api
- depth: 1
- tags: comments, api

Question:

What is the smallest useful comments feature for the blog?

Answer:

Readers can submit a display name and comment body for a published post. The
API should allow creating a comment and listing comments for that post.

## Turn

- question_id: q2
- status: answered
- mode: deep
- topic: comments-boundaries
- depth: 1
- tags: comments, constraints

Question:

Which comments capabilities should be excluded from the first implementation?

Answer:

Do not include moderation, persistence, authentication, rate limiting, or admin
operations. Keep the feature small and focused on validating the chr generation
and ticket loop flow.
