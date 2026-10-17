---
id: TKT-20261017183627
title: "Remove the project name from Web UI page headings"
status: in_progress
kind: task
severity: low
priority: normal
scope: ui
tags: web-ui, headings, project-context
source: chat
reported_at: 2026-10-17T18:36:27Z
commits: []
ready_at: 2026-10-17T18:36:27Z
started_at: 2026-10-17T18:36:27Z
branch: dev
---

# Remove the project name from Web UI page headings

## Context

Finding from [the Web UI review](../open/20261017105432-review-draft-web-ui-usability.md).
Every page heading repeats the current project name before its section or
artifact title. The shared header already identifies the project.
The maintainer approved implementing this finding during the review.

## Expected Outcome

1. Use only the section title for Cases, Dumps, Quizzes, Specs, Plans, Board
   and Dash, consistently through the shared shell.
2. Preserve the actual artifact title on detail pages without adding the
   project name.
3. Keep the current project name visible in the shared header.
4. Preserve the document title, page actions, forms and navigation behavior.

## Validation

- Passed: `go test -short ./internal/webui/... -count=1 -timeout=3m`.
- Passed: focused golangci-lint checks for `internal/webui/...` and `internal/app/...`.
- Catalog, Board, Dash, Project and artifact-detail routes assert their exact
  heading text through the existing handler tests.
- Live installation and final delivery metadata are pending.
