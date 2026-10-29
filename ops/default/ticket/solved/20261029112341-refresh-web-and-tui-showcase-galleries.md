---
id: TKT-20261029112341
title: "Refresh Web and TUI showcase galleries"
status: solved
kind: task
severity: low
priority: normal
scope: docs
tags: gallery, web-ui, tui
source: chat
reported_at: 2026-10-29T11:23:41Z
commits: [0175216b0a6a34621ea021f7d38a143a1d4186a7]
branch: dev
resolution: fixed
closed_at: 2026-10-29T23:50:10Z
ready_at: 2026-10-29T20:54:51Z
started_at: 2026-10-29T20:54:51Z
---

# Refresh Web and TUI showcase galleries

## Context

The gallery linked from README combines mostly older TUI captures with a Web
Board preview. Recent improvements, including the Web/TUI parity delivery,
need a fresh showcase. The maintainer wants two separate, concise galleries
using the same reference use case so each interface has a coherent story.

The current reference is the tictactoe project and its `cli-gameplay` use case,
seeded under `qa/ref/003/tictactoe/`. How the previous screenshots were taken
does not constrain this work. Match or improve their clarity and presentation.

## Prerequisite and execution boundary

Execute after the `web-tui-parity` delivery set has completed and all its slice
PRs are integrated into `dev`. Its canonical plan and tracker are
`ops/default/plan/web-tui-parity.md` and
`ops/default/tracker/web-tui-parity.md` in the integration checkout.

This is a separate follow-up ticket and commit, outside that delivery set.
The maintainer authorized implementation after the parity set and dev-to-main
alignment completed. Keep capture and delivery separate from the closed set.

## Requested work

1. Use one integrated chr revision for both interfaces. Prepare an isolated
   copy of the tictactoe reference and the same `cli-gameplay` content and
   meaningful workflow states for both capture sequences.
2. Produce a Web gallery and a TUI gallery as separate documentation pages.
   Select representative views of Cases and wizard entry, Dump, Quiz mode
   and history, Spec, Plan and prepared tickets, Board, Dash, and execution
   results where supported. Keep each sequence concise and avoid redundant
   images of the same state.
3. Capture the actual running interfaces. Web can use headless Chromium;
   prepare a real terminal and virtual display for TUI capture when needed.
   The process must work without requiring the maintainer to log into an
   interactive desktop session on the development host.
4. Preserve each interface's visual language. Use consistent dimensions,
   legible text, clean framing and coherent content. Remove desktop chrome
   and incidental private information from published images. Do not fabricate
   UI states, delivery receipts or implementation results for the screenshots.
5. Link both galleries directly from README and update the documentation
   index/gallery entry points so readers can choose Web or TUI. Use short
   captions explaining the user-visible behavior in each image.
6. Record the application revision, reference case, state preparation,
   dimensions and capture commands needed to reproduce the set. Choose the
   capture tooling when execution starts based on what is available; this
   ticket does not prescribe a new capture framework or dependency.

## Acceptance criteria

- Two clearly separate Web and TUI galleries are discoverable from README.
- Both show the same tictactoe `cli-gameplay` use case at comparable milestones.
- Screenshots come from the revision containing the fully integrated parity
  set and visibly demonstrate its delivered interactions.
- Screenshots are readable, consistently framed within each gallery, and
  contain no incidental private host, account or network information.
- Any execution/review/completion shown agrees with actual retained evidence.
- Capture is possible without an interactive desktop login by the maintainer.
- Image paths, Markdown links and short English captions are verified, and
  each gallery is reviewed as a complete visual sequence.
- Reproduction details identify the revision and prepared reference states.

## Validation

For capture of this ticket, check metadata, UTC filename/id/timestamp,
status-directory alignment, required enums, tags and whitespace.

For later implementation, open both rendered galleries, inspect every image
at its published size, verify README/index links and asset references, and
confirm both sequences use the recorded revision and case. This documentation
work does not require an unrelated exhaustive test run.

## Delivery

- Added separate Web and TUI galleries, each with eight application views and
  the shared real gameplay capture. README links their shared index; the
  documentation index also lists each gallery.
- Captured the actual application at `7b093af7`, using the same CLI gameplay
  case. Kept the incomplete `qa/ref/003/tictactoe` seed unchanged.
- Preserved the six generated Go files from published commit `0f897632`, plus
  exact Dump input, two clarification answers, source/input hashes and capture
  commands under `qa/showcase/tictactoe` and the capture guide.
- The retained native result is completed: three validated tasks, three
  confirmed tickets, one integrated unit and zero remaining tickets. Delivery
  and set-closing receipt hashes are in the capture record.
- The run needed CLI continuation and maintainer-approved forge acknowledgement
  of its already-applied Git result. The QA manual-merge option was restored.
  The capture guide and record disclose this assistance; this is not proof of
  an unattended workflow.
- This ticket's delivery commit contains implementation and closure. The
  `commits` field preserves the original capture commit as provenance.

## Completed validation

- Inspected all 17 PNGs for legibility, native appearance and private details.
  Checked geometry: Web 1440 × 916, TUI 1720 × 1044, game 876 × 630.
- Opened both rendered galleries in Chromium: all nine images on each page
  loaded, local gallery links returned HTTP 200, and no page errors occurred.
- Verified 134 local Markdown links and exact hashes for the two retained
  inputs and six exported Go files against the published game commit.
- `node --check scripts/assets/capture-gallery.cjs` passed; both capture modes
  produced real screenshots. Capture dependencies remain outside the repo.
- Standalone example: `go test -timeout=30s ./...` passed both packages;
  `gofmt -l cmd internal` returned no files. Six bounded executable smokes
  passed: X win, O win, draw, invalid-input recovery, EOF and help.
- Validated ticket metadata, English content, source privacy and staged
  whitespace. No application source or completed parity-set artifacts changed.
