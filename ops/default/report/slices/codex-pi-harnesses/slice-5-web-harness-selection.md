# Slice 5: Web Harness Selection

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-5-web-harness-selection`
PR: [#416](https://forge.adrianpk.com/adrianpk/chr/pulls/416)

## Purpose

Inspect and change project worker and generation defaults from the existing Web
project context while preserving retained operation bindings.

## Delivered Behavior

- The project context opens a compact harness chooser using the existing modal
  flow, with a normal-page fallback and responsive role columns.
- The chooser reads one shared persisted selection snapshot and displays named
  choices with role-specific local readiness. Clearing and keeping a role are
  explicit choices; generation legacy dispatch remains action-specific.
- Retained worker invocations show their original profile, harness and model
  separately from defaults for new operations. Their list is bounded to eight,
  with the existing Workflows view available for further inspection.
- Authenticated saves change either or both defaults through the shared atomic
  configuration owner. Keep leaves a role untouched; clear restores legacy
  dispatch. Existing handlers and other processes see the persisted result
  without a restart.
- Forms retain the exact project, configuration revision and request identity.
  Stale, unprepared or invalid choices remain selected after rejection, including
  when fresh private-profile inspection fails. Receipt replay recovers a lost
  acknowledgement; closing the chooser does not save unsent choices.
- The dedicated bounded form rejects duplicate, unknown and private setup fields.
  Existing authentication, confirmation and origin guards protect writes.
  Inspection and saves expose no private executable or resource references and
  perform no model call or driver launch.

## Implementation Notes

T5.1 maps the shared application selection operation into presentation-only
values. The existing workflow status projection now includes its public frozen
harness binding, which the chooser uses without another progress store.
Profile/readiness inspection and retained status are bounded by a ten-second
request context. No setup, service activation or credential editor is added.

T5.2 translates the closed form into the existing application write path and
shared selection service. The active operator comes from the server's local
userspace; the browser supplies no actor or private setup. Rejections retain the
original expected revision rather than silently rebasing. Full-page forms use
normal POST while the project-context control uses the existing modal flow.

The User Guide, project-workspace reference, action matrix and existing
Unreleased capability describe the usable control.

## Contracts Added or Changed

- Optional Web harness selection projection and project-context GET/POST control.
- Public retained binding in the existing application workflow status.
- Separate default-role and retained-invocation presentation.
- Exact selection envelope through shared authority, revision and receipt guards.

## Files of Interest

- `internal/app/web_harness.go`, `service_workflow_status.go`
- `internal/webui/harness.go`, `assets/templates/harness.html`
- `internal/webui/assets/templates/shell.html`, `assets/static/chr.css`
- `internal/app/web_harness*_test.go`, `internal/webui/harness_test.go`

## Validation

Passed focused task checks:

- `go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/webui ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

HTTP fixtures cover login, missing/foreign origins, project scope, confirmation,
duplicate/unknown/oversized forms, generic-action bypass, unavailable profiles,
clearing each role, normal-page saves and exact receipt replay. A separate OS
process loads the persisted image after the HTTP save.

One owned Chromium journey logs in, switches both defaults while an invocation
remains claimed, drops a successful acknowledgement and retries the same form.
It proves the existing invocation's revision/status/binding stay unchanged while
new work retains the new worker profile. An independent save while the form is open makes
the browser form stale; rejection retains both its input and original revision.
The same journey checks 1280 px and 390 px layouts, narrow-screen clear, cancel
and private-field exclusion. Both screenshots were inspected. The owned browser,
HTTP server, claim and subprocess are stopped by fixture cleanup.

The controller records the final ordered `make verify`, focused HTTP/application,
focused browser and whitespace checks against the clean PR head after reference
follow-up commits. That exact-head evidence is separate from these task checks.

## Risks and Follow-ups

None recorded.
