---
title: "Comments context"
scope: "comments"
status: "curated"
---

# Comments context

The blog currently exposes published posts but has no way for readers to react
or leave comments.

The first comments feature should add a small HTTP surface:

- `POST /posts/{slug}/comments` creates a comment for an existing published
  post.
- `GET /posts/{slug}/comments` lists comments for an existing published post.

Comments should be short plain-text messages. The first version can accept a
display name and body, assign IDs in memory, and return JSON responses. It does
not need persistence, moderation, authentication, spam controls, email
notifications, or admin tooling.

Unknown post slugs should use the same JSON error response style as the rest of
the API.
