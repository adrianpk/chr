# Separate Acceptance Feature Workflow

Status: draft

Date: 2026-06-06

## Summary

ChrysoCoord should separate acceptance feature generation from general
requirements discovery and specification generation.

The current workflow can bias quiz questions toward Gherkin-oriented slots and
can produce Gherkin content inside the generated specification body instead of
writing a dedicated `.feature` artifact. This makes the user answer low-level
actor/trigger questions too early and makes later tools, such as `godog` or
agent validation workflows, parse Markdown instead of consuming a conventional
feature file path.

The desired workflow is:

```text
dump -> quiz -> spec -> acceptance feature round -> .feature
```

The general quiz remains focused on requirements discovery. Acceptance feature
generation becomes a separate workflow that derives from the already curated
dump, quiz, and spec material.

## Goals

- Keep the general quiz centered on use-case discovery instead of Gherkin slots.
- Generate specifications without embedded `.feature` blocks.
- Add a separate acceptance feature generation operation.
- Write Gherkin feature files to a conventional path:

  ```text
  .chc/user/acceptance/<case>.feature
  ```

- Let the feature-generation workflow ask follow-up questions only when the
  spec plus quiz context still leaves meaningful scenario gaps.
- Prefer inferred defaults during the feature round, so the user confirms or
  adjusts behavior instead of answering actor/trigger questions from scratch.
- Keep `godog` execution out of scope for this iteration while preserving a
  file layout that can support it later.

## Non-Goals

- Do not execute `godog`.
- Do not generate Go step definitions.
- Do not add authentication or external systems to QA fixtures.
- Do not make every use case produce a feature file.
- Do not make the general quiz expose Gherkin terminology.

## Desired Behavior

### General Quiz

The quiz should gather requirements for the use case.

It may retain weak internal awareness that some topics could later support
acceptance scenarios, but it must not repeatedly ask actor, trigger, and
observable outcome questions as a primary path. When a topic is not
acceptance-testable, the quiz continues as normal dump/spec clarification.

### Specification Generation

Spec generation writes only the specification artifact.

The generated spec may contain acceptance criteria written in prose, but it must
not contain fenced Gherkin blocks or a full feature document. If the model
attempts to emit embedded Gherkin in the spec body, the generation path should
either reject that output or normalize it before writing, depending on the
slice's chosen implementation boundary.

### Acceptance Feature Generation

Acceptance feature generation is a separate operation that consumes:

- project metadata;
- use-case metadata;
- relevant dumps;
- curated quiz notes;
- the generated or selected spec.

The feature generator returns one of:

- `generated`: a valid `.feature` body can be written;
- `needs_more_info`: the context has acceptance potential but specific scenario
  gaps remain;
- `not_applicable`: the topic is not naturally suited to a feature artifact.

When information is missing, the feature round should ask default-oriented
questions such as:

```text
I will assume callback_requested requires a future callback_time and rejects the
request without changing history when it is missing. Confirm or adjust?
```

The feature round should avoid making the user manually translate the spec into
Gherkin concepts.

### Feature Artifact

Generated feature files live under:

```text
.chc/user/acceptance/<case>.feature
```

The feature must pass deterministic static checks before it is written.

The spec should reference the feature path when one exists, but it should not
embed the full feature body.

## Acceptance Criteria

- `spec generate` no longer writes `.feature` files.
- `spec generate` no longer embeds fenced Gherkin in generated spec Markdown.
- A dedicated acceptance feature command can generate
  `.chc/user/acceptance/<case>.feature`.
- Feature generation uses spec, quiz, dump, project, and use-case context.
- Feature generation reports `generated`, `needs_more_info`, or
  `not_applicable`.
- Generated feature files pass existing static feature checks before being
  written.
- The TUI exposes an acceptance feature action from the spec/use-case workflow.
- User-facing docs describe the separated workflow through Diataxis-aligned
  guide and reference material.

## Open Questions

- Whether the command namespace should be `chc acceptance feature generate` or
  `chc feature generate`.
- Whether feature-round notes should live at
  `.chc/user/acceptance/<case>-feature-notes.md` or under the quiz artifact
  directory.
- Whether embedded Gherkin in a spec body should fail generation immediately or
  be extracted into the structured feature field during a compatibility period.
