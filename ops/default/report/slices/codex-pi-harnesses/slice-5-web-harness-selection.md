# Slice 5: Web Harness Selection

Status: drafting
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-5-web-harness-selection`
PR: pending

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
- Profile inspection exposes no private executable or resource references and
  performs no model call or driver launch.

## Implementation Notes

T5.1 maps the shared application selection operation into presentation-only
values. The existing workflow status projection now includes its public frozen
harness binding, which the chooser uses without another progress store.
Profile/readiness inspection and retained status are bounded by a ten-second
request context. No setup, service activation or credential editor is added.

T5.2 authenticated revision-bound saves, rejection recovery and the focused
browser journey remain in progress.

## Contracts Added or Changed

- Optional Web harness selection projection and project-context GET control.
- Public retained binding in the existing application workflow status.
- Separate default-role and retained-invocation presentation.

## Files of Interest

- `internal/app/web_harness.go`, `service_workflow_status.go`
- `internal/webui/harness.go`, `assets/templates/harness.html`
- `internal/webui/assets/templates/shell.html`, `assets/static/chr.css`
- `internal/app/web_harness_test.go`, `internal/webui/harness_test.go`

## Validation

Passed focused T5.1 checks:

- `go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --timeout=5m ./internal/webui ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The application fixture inspects real private profiles and a claimed invocation.
It preserves the invocation revision/binding and does not launch the owned
driver. Presentation fixtures cover unsupported providers and sanitized setup
failure. Browser, save/authentication and final exact-head checks are pending.

## Risks and Follow-ups

Complete T5.2 before opening the canonical PR. The chooser currently reads
selections; it does not yet publish changes.
