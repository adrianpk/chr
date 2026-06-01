# Slice 4: Forge Provider Diagnostics

Status: reviewing
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-4-forge-provider-diagnostics`
PR: pending

## Purpose

Verify pull request provider access before later agentic review execution reads
or writes remote review state.

## Delivered Behavior

- Review workflow config now includes Forgejo-compatible and GitHub provider
  settings.
- Pull request URLs are detected as Forgejo-compatible when they use
  `/pulls/<number>` and as GitHub when they use `github.com/.../pull/<number>`.
- `chc ticket review-mode diagnostics --url <pr-url>` prints provider mapping,
  CLI availability, token availability, and pull request readability checks.
- `--dry-run` skips remote pull request readability while still reporting local
  provider, CLI, and token evidence.
- Unsupported providers and missing capabilities return machine-readable
  blocked or unsupported reasons.

## Implementation Notes

- Forgejo-compatible providers are the first-class path and use `tea`.
- GitHub support is included through `gh` without changing the Forgejo model.
- SSH remotes are documented as Git access evidence only; they do not satisfy
  API review capability.
- Diagnostics use an injected runner in tests so provider behavior is covered
  without requiring real local credentials.

## Contracts Added or Changed

- `.chc/user/config.yml` review settings may define provider fields:
  `forgejo_enabled`, `forgejo_cli`, `forgejo_token_env`, `github_enabled`,
  `github_cli`, and `github_token_env`.
- Default provider config enables Forgejo-compatible providers with `tea` and
  `FORGEJO_TOKEN`, and GitHub with `gh` and `GITHUB_TOKEN`.
- Diagnostic check statuses are `ok`, `fail`, and `skipped`.
- Diagnostic overall statuses are `ready`, `blocked`, and `unsupported`.

## Files of Interest

- `internal/chc/review_provider_diagnostics.go`
- `internal/chc/config.go`
- `internal/chcapp/ticket.go`
- `cmd/chc/assets/workspace/profiles/standard/config.yml`
- `docs/reference/review-workflow-config.md`

## Validation

- `git diff --check`
- `go test ./internal/chc -run 'Config|ReviewProvider'`
- `go test ./internal/chcapp -run 'TicketReviewMode|Workspace|Init'`
- `go test ./internal/chc ./internal/chcapp`

## Risks and Follow-ups

- Remote comment reading and writing are not implemented in this slice.
- `make verify` was not used as the final gate after validation scope was
  clarified; final validation used focused tests and diff validation.
