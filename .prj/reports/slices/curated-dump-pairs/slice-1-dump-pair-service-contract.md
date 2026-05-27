# Slice 1: Dump Pair Service Contract

Status: delivered
Slicing: Curated Dump Pairs
Plan: `.prj/plans/slicing-curated-dump-pairs-plan.md`
Tracker: `.prj/trackers/slicing-curated-dump-pairs-tracker.md`
Branch: `feat/slice-1-dump-pair-service-contract`
PR: #231

## Purpose

Introduce service-owned dump pair creation so each new dump capture produces a
curated Markdown artifact plus a raw sibling with highlights.

## Delivered Behavior

- `CreateDump` now creates `.chc/dumps/<slug>.md` as the curated primary dump.
- `CreateDump` also creates `.chc/dumps/<slug>-raw.md` as the raw capture.
- The curated file records `status: "curated"`, `type: "dump"`, and a `source`
  reference to the raw sibling.
- The raw file records `status: "raw"`, `type: "dump_raw"`, and a `curated`
  reference to the curated sibling.
- Raw files preserve the original dump content and include a generated
  highlights section.
- The service uses a dump-specific OpenAI curation provider when available.
- LLM failures or missing configuration fall back to deterministic local
  curation without blocking dump capture.
- CLI `chc dump new` uses the same service flow and reports the curated path.

## Implementation Notes

- Dump curation is modeled with `DumpCurationProvider`, not the quiz model.
- The OpenAI provider records `dump_curation` request and response events through
  the shared LLM trace sink.
- Tests use deterministic fallback or `httptest`; they do not call the network.
- Existing unique slug behavior is preserved for pairs: the service skips a
  candidate when either the curated or raw path already exists.

## Contracts Added or Changed

- Added `DumpCurationRequest`, `DumpCurationResult`, and
  `DumpCurationProvider`.
- Added `openAIDumpCurationProvider`.
- Added `fallbackDumpCurationProvider`.
- Changed `CreateDump` from single-file output to paired curated/raw output.
- Changed generated dump frontmatter for service-created dumps.

## Files of Interest

- `internal/chcapp/service_dump.go`
- `internal/chcapp/service_dump_curation.go`
- `internal/chcapp/service_dump_curation_openai.go`
- `internal/chcapp/service_dump_curation_test.go`
- `internal/chcapp/service_test.go`
- `internal/chcapp/cli_test.go`

## Validation

- `go test ./internal/chcapp` pass
- `go test ./internal/chc ./internal/chcapp` pass
- `make verify` fail
  - Fails on pre-existing strict lint issues in `internal/chc`,
    `internal/chctui`, and related tests.
  - The final failing output no longer reports new issues from the curated dump
    pair files.

## Risks and Follow-ups

- Slice 2 must prevent raw siblings from being treated as duplicate primary dump
  context.
- Slice 2 must tune F3 Dumps so raw siblings are exposed as source metadata, not
  as normal dump rows.
