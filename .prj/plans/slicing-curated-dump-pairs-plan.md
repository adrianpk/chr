# Curated Dump Pairs Slicing Plan

Status: planned

## Purpose

Change dump creation so a user-provided free-form dump is not stored only as
the raw text that was typed or pasted. A dump capture should produce a pair of
related Markdown documents:

- a raw capture document that preserves the original input and records extracted
  highlights;
- a curated dump document that presents the same user intent in cleaner,
  normalized prose for later use by specs, plans, tracking, and tickets.

This makes dumps useful as durable project memory while still preserving the
recoverable source material behind the curated version.

## Product Definition

Dumps are conversational knowledge captures. They often contain spelling
errors, repeated phrasing, incomplete sentences, pauses, and ordering that made
sense during conversation but is noisy when reused as context.

The dump workflow should keep two forms:

- Raw capture: the recoverable source record. It preserves the user's original
  text, includes capture metadata, and stores a short highlights section derived
  from the capture.
- Curated dump: the primary reusable artifact. It improves spelling, grammar,
  punctuation, paragraphing, and continuity while preserving the user's wording,
  intent, tensions, decisions, and open questions.

The curated dump must not become a spec, plan, ticket, or architecture essay. It
is still a dump: an improved knowledge capture that feeds later workflows.

## Source Material

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/dump.md`
- `internal/chrapp/service_dump.go`
- `internal/chrapp/dump_cli.go`
- `internal/chrapp/tui_adapters_dump_quiz.go`
- Current LLM tracing behavior for non-ticket-generation LLM calls
- Existing `.chr/dumps` canonical layout

Conversation decisions:

- Do not store only the raw user text as the main dump artifact.
- Preserve the original raw input in a sibling document.
- Generate a cleaned-up, readable sibling document for normal consumption.
- Generate highlights for the raw capture.
- Keep downstream context assembly focused on the curated dump by default.
- Keep this behavior behind the existing service boundary so CLI and TUI remain
  adapter surfaces over the same logic.
- The implementation must not reference external project names in product,
  runtime, or workflow documentation.

## Current State

`CreateDump` currently renders one Markdown file under `.chr/dumps`.

The generated file:

- uses the requested title;
- writes `status: "raw"` in frontmatter;
- optionally writes a use-case slug;
- stores the provided content directly in the body;
- has deterministic filename collision handling.

This is service-owned and already shared by CLI and TUI adapters, which is the
correct integration point.

Limitations:

- The raw user wording is the only durable dump content.
- There is no LLM-assisted polishing or curation step.
- There is no sibling raw/curated relationship.
- There is no highlights section derived from the source capture.
- Downstream workflows cannot distinguish curated dump text from raw capture
  text.

## Target Model

Dump creation should generate a sibling pair:

```text
.chr/dumps/<slug>.md
.chr/dumps/<slug>-raw.md
```

The curated dump is the primary artifact and keeps the stable path currently
expected by users and downstream workflows.

The raw capture uses the same slug with a `-raw` suffix. It preserves the source
input and points back to the curated sibling.

Both files are Markdown-first artifacts with frontmatter.

The curated file frontmatter should include:

```yaml
title: "..."
status: "curated"
type: "dump"
source: ".chr/dumps/<slug>-raw.md"
case: "..." # only when use-case scoped
```

The raw file frontmatter should include:

```yaml
title: "..."
status: "raw"
type: "dump_raw"
curated: ".chr/dumps/<slug>.md"
case: "..." # only when use-case scoped
```

Exact field names may change during implementation if existing indexing rules
require a narrower schema, but the relationship must remain explicit.

## LLM Responsibility

Dump curation should be one structured LLM operation owned by the application
service.

Input:

- project metadata when available;
- project-wide context when available;
- use-case metadata when scoped;
- use-case long-form context when scoped;
- dump title;
- raw user content;
- optional operation metadata.

Output:

```json
{
  "polished_body": "Cleaned Markdown body that preserves user intent.",
  "highlights": [
    "Short recoverable point from the raw capture."
  ],
  "warnings": [
    "Non-fatal concern about ambiguity, missing context, or redaction."
  ]
}
```

The prompt must instruct the model to:

- preserve the user's meaning and important wording;
- correct spelling, punctuation, grammar, and paragraph flow;
- remove conversational filler only when it does not carry meaning;
- avoid inventing requirements;
- avoid producing specs, plans, tickets, or implementation tasks;
- keep uncertainty as uncertainty;
- redact obvious secrets, credentials, tokens, passwords, and private keys.

## Fallback Behavior

Dump creation must remain usable when the LLM is unavailable.

Fallback behavior:

- create the raw capture file with the exact provided content;
- create the curated dump file using deterministic local cleanup only, such as
  trimming outer whitespace and preserving paragraphs;
- include a warning that LLM curation was unavailable;
- record the error through normal operation feedback and LLM tracing when the
  request reached the LLM boundary.

The fallback must not block basic dump capture.

## Downstream Consumption

By default, context assembly should consume curated dump files, not raw capture
files.

Raw capture files remain available for audit, recovery, and manual inspection,
but should not be injected into spec, plan, tracking, or ticket generation unless
a future workflow explicitly asks for source material.

The workspace index should distinguish raw dump captures from curated dumps so
the Dumps view and downstream context selection do not double-count both siblings
as equivalent context.

## Slice Map

Slicing name: Curated Dump Pairs

Active slice number: Slice 1

Active slice short name: dump pair service contract

Expected branch name: `feat/slice-1-dump-pair-service-contract`

Expected PR title: `feat(slice-1): dump pair service contract`

Expected report path:
`.prj/reports/slices/curated-dump-pairs/slice-1-dump-pair-service-contract.md`

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | dump pair service contract | `feat/slice-1-dump-pair-service-contract` | `feat(slice-1): dump pair service contract` | `.prj/reports/slices/curated-dump-pairs/slice-1-dump-pair-service-contract.md` |
| Slice 2 | dump pair consumption and tui | `feat/slice-2-dump-pair-consumption-tui` | `feat(slice-2): dump pair consumption and tui` | `.prj/reports/slices/curated-dump-pairs/slice-2-dump-pair-consumption-tui.md` |

## Slice 1: Dump Pair Service Contract

Objective:

Introduce service-owned dump pair creation without moving dump logic into CLI or
TUI adapters.

Tasks:

- Add a dump curation provider contract that accepts raw dump creation context
  and returns polished body, highlights, and warnings.
- Add a default LLM-backed implementation using the existing LLM adapter and
  trace sink.
- Add deterministic fallback behavior when LLM curation fails or is disabled.
- Change `CreateDump` so it writes the curated file and raw sibling as one
  operation.
- Preserve existing title, use-case scope, slug collision handling, and service
  error behavior.
- Update CLI tests for `chr dump new`.
- Add service tests for successful curation, fallback curation, raw sibling
  creation, frontmatter relationship fields, and filename collision handling.
- Add the slice implementation report and validation evidence.

Acceptance criteria:

- `chr dump new` creates both `.chr/dumps/<slug>.md` and
  `.chr/dumps/<slug>-raw.md`.
- The curated file is the primary path reported by the operation.
- The raw file preserves the original input.
- The raw file includes a highlights section.
- LLM failures do not prevent dump capture.
- LLM request and response payloads are traceable when tracing is enabled.

## Slice 2: Dump Pair Consumption And TUI

Objective:

Make the rest of the product treat curated dumps as the primary dump artifact
while keeping raw siblings discoverable but non-primary.

Tasks:

- Update workspace indexing or dump read-model mapping so raw dump siblings are
  distinguishable from curated dumps.
- Ensure context assembly uses curated dump content by default.
- Ensure raw dump siblings do not double-count as regular dumps in spec, plan,
  tracking, or ticket generation context.
- Update F3 Dumps to present the curated dump as the main item.
- Expose raw sibling metadata in the detail panel when available.
- Keep `alt+e` editing behavior focused on the selected curated dump.
- Add tests for index classification, context selection, and TUI rendering.
- Update user-facing documentation for dump pair behavior.
- Add the slice implementation report and validation evidence.

Acceptance criteria:

- Dumps view remains centered on normal curated dump documents.
- Raw siblings are visible as source metadata, not as duplicate primary dumps.
- Spec, plan, tracking, and ticket generation consume curated dump text by
  default.
- Existing direct-edit workflows still work for curated dumps.
