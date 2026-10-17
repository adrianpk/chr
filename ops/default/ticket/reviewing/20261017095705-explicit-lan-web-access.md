---
id: TKT-20261017095705
title: "Add an explicit opt-in flag for token-free LAN Web UI access"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: ui
tags: web-ui, lan-access, navigation
source: chat
reported_at: 2026-10-17T09:57:05Z
commits: []
ready_at: 2026-10-17T10:14:56Z
started_at: 2026-10-17T10:14:56Z
branch: dev
reviewed_at: 2026-10-17T10:29:00Z
---

# Add an explicit opt-in flag for token-free LAN Web UI access

## Problem

A user needs to run Chryso on a development server and use its Web UI from a
different computer on the internal network. Opening an individual page with
`?token=...` can pass authentication, but subsequent navigation and HTMX
requests omit the token and return `web token required`.

The requested workflow is explicit, token-free access on a trusted internal
network. Enabling it must be a deliberate boolean choice that defaults to
false, with working navigation throughout the application.

## Existing Behavior

`internal/app/web.go` already exposes `--host` and
`--require-token=auto|true|false|always`. Auto mode requires a token for a
non-loopback bind. Explicitly selecting a non-loopback host together with
`--require-token=false` already resolves a token-free access policy.

The new flag should use this policy machinery, rather than introduce another
authentication implementation. The current default bind is `127.0.0.1:7331`.

## Requested Work

1. Add the boolean flag `chr web serve --allow-lan`, default `false`. With
   normal defaults and no flag, keep the existing loopback bind and token
   policy. Do not infer permission to expose the server from a hostname,
   missing token, private address, environment or request header.
2. When enabled, select a non-loopback bind (`0.0.0.0` by default) and disable
   the Web UI token requirement. Honor an explicit `--host` and `--port`.
   Opening `http://devbox:7331/` from another computer must require neither a
   token-bearing URL nor a token environment variable, browser extension or
   custom header. The hostname in this example is illustrative.
3. Resolve the flag through the existing web configuration and access-policy
   owners. Reject a conflicting explicit token requirement (`true` or
   `always`, from CLI or configuration) with an actionable error; do not
   silently weaken it. Auto mode is resolved to token-free operation when
   LAN access is explicitly enabled. Preserve the existing lower-level host
   and token options when the new flag is not enabled.
4. Apply the resolved policy consistently to full pages, direct deep links,
   browser refresh, static assets, HTMX fragments/polling and same-origin
   forms/actions. Navigation must work without carrying a token between
   pages. Keep existing origin checks, write confirmations, revision checks
   and domain authorization in place.
5. Show the effective LAN choice, bind and token requirement in startup and
   `--dry-run` output. Reuse the current exposure warnings. Document one
   launch command and how the browser computer uses the server's reachable
   hostname or address. The flag selects trusted-network operation; it does
   not install firewall rules or restrict connections to private CIDRs.
6. Add focused regressions and update CLI/reference guidance plus Unreleased
   release notes when the feature is implemented.
7. Support the matching project setting `web.allow_lan`, default `false`, so
   dogfooding can retain the operator's explicit choice. CLI `--allow-lan=false`
   must override a retained `true`. Enable it in the local dogfooding setup
   without changing the product default or publishing machine configuration.

## Acceptance and Validation

1. CLI tests prove the default flag is false; normal startup/dry-run remains
   loopback-only with repository defaults. `--allow-lan` resolves to the
   expected bind and `token_required: false`, with no token configured.
   `--allow-lan=false` preserves normal resolution. Explicit host/port and
   contradictory token requirements have deterministic tests.
2. A local HTTP/browser journey represents access through a remote request
   hostname. Open the root without credentials, follow the actual navigation
   links, open direct pages, refresh and load representative HTMX fragments.
   Verify the Project, Board, Dash and artifact navigation remains usable
   without query tokens, custom headers or authentication cookies.
3. Submit a valid same-origin action through the normal handler and verify
   token authentication does not block it. Keep the business confirmation and
   revision assertions. Verify a disallowed cross-origin write remains denied.
4. Existing token-required modes still reject missing or invalid credentials.
   New LAN behavior must require explicit opt-in; it must not become a silent
   fallback when token authentication fails.
5. During implementation, run the affected CLI/access-policy and navigation
   regressions and the repository development gate. Record exact commands and
   results. Keep exhaustive verification in the scheduled nightly. No service
   restart or external network test is required merely to capture this ticket.

## Implementation References

- `internal/app/web.go`: flags, config resolution, effective-policy output,
  `webTokenRequired` and `webAccessPolicy`.
- `internal/core/web_config.go`: existing bind/token validation.
- `internal/core/config.go`: persisted `web.allow_lan` setting and its disabled
  product default.
- `internal/webui/webui.go`: request authentication, origin checks and routes.
- `internal/app/cli_test.go` and `internal/webui/webui_test.go`: existing
  default, token-mode, origin and handler coverage.
- `docs/how-to/expose-local-web-ui.md`: current network-access guidance.

## Scope Boundary

This ticket defines one focused CLI/config/access-policy change and its
navigation regressions. Local configuration does not automatically start a
server. Repairing browser sessions for token-required operation is a separate
concern; token propagation is not a prerequisite for this explicitly token-free
LAN workflow.

## Authorized Execution

The maintainer authorized this implementation on 2026-10-17 after completion
of Telegram help ticket TKT-20261017094606. Help is delivered on dev before
LAN implementation starts.

The maintainer subsequently requested persistent LAN access for CHR's own
local dogfooding setup. This authorizes its local configuration separately
from the product's disabled default.

## Delivered Behavior

- `chr web serve --allow-lan` selects token-free operation and defaults to
  `0.0.0.0`, preserving an existing non-loopback interface or explicit host and
  port overrides.
- Project configuration can retain `web.allow_lan: true`; an explicit false
  flag overrides it. Missing configuration keeps the product default false.
  Required token modes are rejected as conflicts unless explicitly overridden.
- Startup and dry-run output show the effective LAN selection. Existing
  authentication, origin, confirmation and revision owners remain in use.
- User Guide and launch/exposure guides document the one-command workflow,
  persisted choice and existing operator-owned network boundary.

## Validation Evidence

- `go test ./internal/app ./internal/core -run '^(TestWebLAN|TestLoadWorkspaceConfig.*Web|TestLoadWorkspaceConfigUsesDefaults)' -count=1`
  passed. This covers the disabled default, persisted opt-in and explicit false
  CLI override alongside bind and token conflicts.
- `go test ./internal/app -run '^(TestWebLAN|TestWebServe)' -count=1`
  passed in 10.620 seconds, without short mode. It includes the actual HTTP
  navigation journey, valid and rejected writes, and existing token-required
  regressions.
- No new concurrency behavior was introduced; the exhaustive suite, complete
  race suite and live-forge verification remain scheduled nightly work.
- `make verify GOLANGCI_LINT=/tmp/chr-lint-serial-20261017.sh` passed:
  strict lint reported zero issues, short Go tests passed, all eight CI script
  regressions passed and the application built. The temporary wrapper invokes
  the same pinned linter with `--allow-serial-runners` to wait for its shared
  process lock; it does not change checks or repository configuration.
- `git diff --check` and focused ticket metadata/link checks passed.
