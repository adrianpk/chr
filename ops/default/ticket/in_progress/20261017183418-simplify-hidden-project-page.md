---
id: TKT-20261017183418
title: "Keep a minimal Project page outside visible navigation"
status: in_progress
kind: task
severity: low
priority: normal
scope: ui
tags: web-ui, navigation, project-context
source: chat
reported_at: 2026-10-17T18:34:18Z
commits: []
ready_at: 2026-10-17T18:34:18Z
started_at: 2026-10-17T18:34:18Z
branch: dev
---

# Keep a minimal Project page outside visible navigation

## Context

Finding from [the Web UI review](../open/20261017105432-review-draft-web-ui-usability.md).
The Project page foregrounds internal identifiers, listener settings, polling
configuration and repeated absolute paths. Its ticket path is assembled from
an obsolete layout rather than resolved from the operational artifact owner.
The maintainer approved implementing this bounded finding during the review.

## Expected Outcome

1. Keep `/project` available through a manually entered URL. Remove its tab
   and all visible links from the shared header and navigation.
2. Show the project name and root directory. Keep Project ID as secondary
   diagnostic information on this hidden page.
3. Remove Bind, Refresh and internal workspace, ticket and runtime paths.
4. Stop automatic polling of the static Project page. Preserve polling on
   working screens that display changing information.
5. Keep the current project name visible in the shared header as plain context,
   without suggesting an interactive selector.
6. Preserve access policies, existing direct routes and unrelated work.

## Validation

- Passed: `go test -short ./internal/webui/... -count=1 -timeout=3m`.
- Passed: `go test -short ./internal/app -run '^TestWeb' -count=1 -timeout=3m`.
- Passed: focused golangci-lint checks for `internal/webui/...` and `internal/app/...`.
- Live installation and final delivery metadata are pending.
