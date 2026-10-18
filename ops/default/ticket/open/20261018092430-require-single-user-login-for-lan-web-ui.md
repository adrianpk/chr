---
id: TKT-20261018092430
title: "Protect LAN Web UI access with single-user login"
status: open
kind: task
severity: unclassified
priority: unclassified
scope: api
tags: web-ui, authentication, lan
source: chat
reported_at: 2026-10-18T09:24:30Z
commits: []
---

# Protect LAN Web UI access with single-user login

## Need

Provide a simple, effective login for the Web UI when it is exposed across a
local network. Only the one configured user may access project pages, data
and actions. Network reachability alone must not grant application access.

## Expected Outcome

1. Support one configured user, without a public registration flow or a
   multi-user administration system.
2. Require authentication before serving protected pages, document details,
   fragment refreshes or action endpoints. Direct URLs must not bypass login.
3. After login, allow normal navigation and refresh throughout the Web UI
   without manually generating or carrying tokens in page URLs.
4. Provide logout and a defined session lifetime. Invalid or expired sessions
   cannot read project information or submit changes.
5. Keep credentials and session secrets in private configuration/storage,
   outside versioned artifacts, URLs and logs. Keep setup practical for one
   operator.

This ticket captures the need. Authentication, credential setup and session
implementation details remain for later design; no implementation or service
change is authorized by this capture.

## Versioning Boundary

Keep this ticket local and untracked until the current
web-ui-workflow-consolidation delivery cycle finishes. Do not stage, commit or
push it during that cycle, or add it to the active plan/tracker or its slices.
It records separate future work.
