# Slice 2: Blog Reference Project

Status: reviewing
Slicing: QA Agent Work Scenario
Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`
Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`
Branch: `feat/slice-2-blog-reference-project`
PR: pending

## Purpose

Replace the placeholder QA reference with a small but realistic Go blog service
that can exercise the manual agent-work flow.

## Delivered Behavior

- `qa/ref/001/blog` is now a Go module named `blog`.
- The reference project has an executable entrypoint under `cmd/blog`.
- The blog package has an in-memory post store and HTTP server.
- `GET /posts` returns published posts.
- `GET /posts/{slug}` returns a published post by slug.
- Unknown post lookup is intentionally wrong: the handler returns `500` instead
  of the expected `404`.
- The table-driven HTTP test documents the expected `404` behavior and fails
  after reset.
- The blog README explains the seeded failing baseline.

## Implementation Notes

The service uses only the Go standard library and modern `http.ServeMux`
method/path patterns. It avoids router dependencies so the scenario tests chr
workflow behavior rather than framework setup.

The failing test is intentional. `make qa-test-blog` must fail until a later
agent-work run resolves the seeded bug ticket that Slice 3 will add.

## Contracts Added or Changed

- `make qa-reset-blog` now creates an active Go project at `qa/project/blog`.
- `make qa-test-blog` now runs `go test ./...` in that active project.
- The active baseline failure is:
  - `TestServerPostEndpoints/unknown_post`
  - actual status `500`
  - expected status `404`

## Files of Interest

- `qa/ref/001/blog/go.mod`
- `qa/ref/001/blog/cmd/blog/main.go`
- `qa/ref/001/blog/internal/blog/post.go`
- `qa/ref/001/blog/internal/blog/server.go`
- `qa/ref/001/blog/internal/blog/server_test.go`
- `qa/ref/001/blog/README.md`
- `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

## Validation

- `make qa-reset-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test
  expects `404`
- `make verify`
- `git diff --check`

## Risks and Follow-ups

- Comments remain unimplemented until later generated work.
- Seeded `.chr` context and ADRs are added in Slice 3.
