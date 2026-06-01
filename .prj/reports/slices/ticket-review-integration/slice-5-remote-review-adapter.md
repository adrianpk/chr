# Slice 5: Remote Review Adapter

Status: reviewing
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-5-remote-review-adapter`
PR: #277

## Purpose

Provide a provider-neutral adapter for reading pull request review evidence and
writing review findings before the ticket review command is introduced.

## Delivered Behavior

- A remote review adapter reads pull request metadata, diffs, issue comments,
  review comments, and checks where the provider path exposes them.
- Forgejo-compatible providers use `tea api` paths for pull request metadata,
  diffs, issue comments, commit statuses, inline review attempts, and fallback
  issue comments.
- GitHub uses `gh pr` and `gh api` paths for pull request metadata, diffs,
  issue comments, review comments, checks, inline comments, and fallback issue
  comments.
- Review comments attempt inline placement first when path and line metadata
  are available.
- If inline placement fails, the adapter writes a general comment that carries
  the finding body and source location.

## Implementation Notes

- The adapter is in `internal/chc` and has no CLI surface in this slice.
- Command execution is injected through `ReviewRemoteCommandRunner`, which
  keeps tests deterministic and avoids requiring real forge credentials.
- The adapter reuses provider URL detection and provider config from Slice 4.
- Clean passes still do not write remote comments; this adapter only writes
  explicit findings supplied by later review execution.

## Contracts Added or Changed

- Added `RemoteReviewAdapter` with `Read` and `WriteComments` operations.
- Added `RemoteReviewSnapshot` for provider evidence gathered from a pull
  request.
- Added `RemoteReviewComment` and `RemoteReviewWriteResult` for blocking and
  non-blocking finding comments.
- Added comment placement constants for `inline`, `general`, and `skipped`.

## Files of Interest

- `internal/chc/review_remote_adapter.go`
- `internal/chc/review_remote_adapter_test.go`
- `internal/chc/review_provider_diagnostics_test.go`

## Validation

- `git diff --check`
- `go test ./internal/chc -run 'Review(Remote|Provider)|Config'`
- `go test ./internal/chc`
- `golangci-lint run ./internal/chc`

## Risks and Follow-ups

- The next slice must decide how much of the raw snapshot becomes reviewer
  prompt evidence.
- Real provider command compatibility should be exercised during the final QA
  scenario once the review command exists.
