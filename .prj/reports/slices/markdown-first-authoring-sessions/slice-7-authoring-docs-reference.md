# Slice 7: Authoring Docs Reference

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-7-authoring-docs-reference`
PR: #97

## Purpose

Document the Markdown-first authoring session workflow delivered by this
slicing.

## Delivered Behavior

The durable documentation now covers:

- authoring session commands;
- session storage layout;
- generated artifact paths;
- session fields;
- provenance links;
- static quiz behavior;
- quiz-to-draft promotion;
- current boundaries for assisted authoring and TUI work.

## Implementation Notes

Documentation follows the repository Diataxis structure:

- exact command and layout contracts live in `docs/reference/`;
- task flow guidance lives in the user guide;
- the README links to the authoring reference.

No runtime behavior changed in this slice.

## Files of Interest

- `docs/reference/authoring-sessions.md`
- `docs/how-to/user-guide/index.md`
- `docs/reference/project-workspace.md`
- `docs/index.md`
- `README.md`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `make verify`: pass

## Risks and Follow-ups

Agent-assisted authoring and TUI authoring remain future surfaces over the same
session model. This slice only documents the Markdown-first workflow currently
implemented.
