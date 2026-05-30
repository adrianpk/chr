---
key: "CHC-1"
title: "Fix unknown post status"
synopsis: "Return 404 instead of 500 for unknown post slugs."
status: "ready_for_dev"
priority: "high"
severity: "high"
type: "bug"
tags: ["qa", "http", "posts"]
use_case: ""
project_area: "api"
source_kind: "manual"
source_ref: "qa/ref/001/blog"
slice: ""
created_at: "2026-05-30T00:00:00Z"
updated_at: "2026-05-30T00:00:00Z"
---

# Objective

Fix the existing post lookup behavior so `GET /posts/{slug}` returns `404` for
an unknown published post slug.

# Context

The seeded baseline intentionally fails:

```sh
go test ./...
```

The failing test is `TestServerPostEndpoints/unknown_post`. It currently
observes status `500` and expects status `404`.

# Tasks

- Run `go test ./...` and observe the existing failing test before changing
  code.
- Fix the unknown-post handler behavior.
- Keep the response JSON-shaped.
- Run `go test ./...` again.

# Acceptance

- Unknown post slug requests return HTTP `404`.
- Existing post listing and known post lookup behavior still work.
- `go test ./...` succeeds.

# Validation

```sh
go test ./...
```
