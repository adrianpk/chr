---
id: TKT-20260522071859
title: Align CLI authoring and TUI operation surface
status: open
kind: task
severity: high
priority: high
scope: ux
tags: cli,tui,authoring,quiz,workspace-config,lab-feedback
source: lab-runbook
reported_at: 2026-05-22T07:18:59Z
ready_at:
started_at:
reviewed_at:
closed_at:
resolution:
owner:
branch:
pr:
commits: []
---

## Context

The full-system lab run against `chr-blog-work` exposed a coherent UX problem:
chr has a useful command foundation, but the human operation path is too
noisy for daily use.

The CLI must remain complete because it is valuable for automation, testing,
runbooks, and pro-user workflows. However, a normal operator should not need to
copy opaque IDs and absolute paths from one command into another to make
progress.

The user-facing operating model should become:

- CLI as the executable foundation;
- TUI as the primary local human workflow surface;
- `.chr/` as the durable project-local source of truth.

Related feedback capture:

- `.prj/drafts/chr-cli-output-polish-notes.md`

## Problem

The lab surfaced several linked issues:

- CLI success output prints long absolute paths by default.
- List output lacks headers, making fields such as `draft` and `spec`
  ambiguous.
- `auth-...` authoring session IDs are easy to confuse with authentication.
- The static quiz flow generates generic hardcoded questions that do not
  produce a strong spec.
- Workspace config exposes internal invariants such as `workspace_root`,
  `canonical_assets`, runtime workdir, and artifact paths.
- The workspace manifest repeats invariant/config-like facts instead of helping
  users understand the workspace.
- Generated protocol docs blur project documentation rules with chr
  operational artifact layout.
- The runbook originally overexposed developer-only flags such as `--dev-root`.
- The CLI is necessary but insufficient as the daily human workflow surface.

## Expected

- CLI output is compact, readable, and script-friendly.
- Human-readable output uses relative paths by default and reserves absolute
  paths for verbose or structured modes.
- List commands include headers or another unambiguous field presentation.
- Authoring session identifiers are understandable and recoverable through
  topic/status/target listings.
- Static quiz generation is clearly positioned as a fallback/template mode.
- Context-aware quiz generation exists for spec-quality knowledge extraction.
- `.chr/config.yml` exposes only meaningful user decisions.
- `.chr/manifest.md` is either useful orientation or removed from the required
  context path.
- chr operational artifact layout is enforced by the binary, not by
  generated Markdown convention text.
- A Bubble Tea TUI is specified as the main human workflow surface over the same
  command/domain operations.

## Actual

- CLI output currently favors exhaustive technical detail over scannability.
- `chr author quiz new` creates a session plus quiz with absolute paths and
  repeated session fields.
- `chr author session list` returns rows without headers.
- Static quiz questions are generic and hardcoded.
- Generated workspace config and manifest include questionable configurable or
  pseudo-configurable fields.
- The TUI is not specified yet.

## Fix Strategy

- Use a slicing plan because the work spans CLI formatting, bootstrap contracts,
  authoring, quiz generation, and a future TUI.
- Keep slices vertical and behavior-oriented.
- Treat observed lab issues as product feedback, not one-off formatting nits.
- First improve CLI readability and config/workspace contracts enough to reduce
  immediate confusion.
- Then introduce context-aware authoring/quiz foundations.
- Specify the TUI in granular UX/UI documents before implementation.

## Validation

- Full-system lab runbook remains executable by a final-user-style operator.
- CLI examples use normal user-facing commands from inside the managed project.
- `make verify` passes after implementation slices.
- New/changed command output has focused tests.
- Static and context-aware quiz modes are distinguishable in docs and CLI output.
- TUI specifications cover navigation, entity lists, detail views, actions,
  authoring, ticket execution, evidence inspection, and error states.

## Outcome

Pending slicing and implementation.
