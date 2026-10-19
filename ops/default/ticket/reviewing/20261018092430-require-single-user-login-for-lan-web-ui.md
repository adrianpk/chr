---
id: TKT-20261018092430
title: "Protect LAN Web UI access with single-user login"
status: reviewing
kind: task
severity: unclassified
priority: unclassified
scope: api
tags: web-ui, authentication, lan
source: chat
reported_at: 2026-10-18T09:24:30Z
commits: [22ec35740e8294468f7c8c48f6b7d8defe6febc0]
ready_at: 2026-10-19T09:06:23Z
started_at: 2026-10-19T09:06:23Z
branch: feat/ticket-20261018092430-web-login
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/411
reviewed_at: 2026-10-19T09:31:25Z
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

## Authorized Delivery

On 2026-10-19, the maintainer requested this implementation in its own branch
and pull request after the backlog audit. The earlier local-only capture
boundary ended with the completed Web UI delivery cycle. This work is separate
from that closed delivery set and from the later optional Pi harness.

## Implementation

Use one private account verifier loaded at startup, masked terminal setup and
bounded server-side sessions. Require login across project routes, fragments
and actions; retain data-free health/static access. Sessions expire after 12
hours, logout revokes the session and restart invalidates all sessions.
Configured login supersedes legacy tokens. Preserve existing access modes when
no account is configured and provide explicit fail-closed startup.

See [Configure Single-user Web Login](../../../../docs/how-to/configure-web-login.md)
for setup, bounds, network behavior and password replacement.

## Validation

- `make verify`: passed; strict lint reported zero issues, short Go tests,
  eight CI script regressions and build passed.
- `go test -race ./internal/webui -run '^TestLogin' -count=1 -timeout=3m`:
  passed, including HTTP route guards, credential-file boundaries, redirects,
  cookie/session expiry and revocation, concurrent reads and resource bounds.
- `CHR_WEB_BROWSER=/path/to/chromium go test ./internal/app -run '^TestWebLoginBrowser$' -count=1 -timeout=3m`:
  passed in an owned Chromium instance; login, destination retention, navigation,
  expired HTMX navigation and narrow-screen layout work.
- Actual CLI setup passed in an isolated pseudo-terminal: masked password and
  confirmation, private verifier storage, implicit replacement refusal and
  explicit replacement. No real operator credential or running service changed.
- Metadata, Markdown links and `git diff --check`: passed.

The nightly browser selection now includes the login journey. This records
local focused validation, not a completed exhaustive scheduled nightly.

## Review Handoff

[PR #411](https://forge.adrianpk.com/adrianpk/chr/pulls/411) targets `dev`.
Implementation and local validation are complete. Keep this ticket `reviewing`
until integration is verified; do not infer deployment from the PR or its checks.
The separately configured Web preview is active. On 2026-10-19, the maintainer
confirmed that login works and accepted the ticket's functional result.
Configured-account login, logout and subsequent anonymous-access denial also
passed against that running preview. Credentials remain in private local
storage and are not part of this ticket or pull request.

Only PR integration and the post-merge ticket transition remain for delivery
closure. The default Web service still uses its existing installed build;
acceptance of the preview does not claim that service has been upgraded.
