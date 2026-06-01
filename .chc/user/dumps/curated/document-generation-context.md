---
title: "Document Generation Context"
status: "curated"
type: "dump"
case_slug: "spec-to-tickets"
---

# Document Generation Context

ChrysoCoord treats project knowledge as Markdown-first material.

Dumps capture broad project or use-case context. Quizzes capture missing knowledge through bounded questions. Specs formalize expected behavior. Plans decompose specs into slices, tasks, dependencies, validation criteria, branches, and PR expectations. Plan tracking creates the execution map and tickets.

The workflow should preserve continuity across artifacts:

- context informs specs;
- specs inform plans;
- plans inform trackers and tickets;
- tickets inform execution context;
- execution produces reports and evidence.

Generated artifacts must remain editable by the user. The system should never behave as an irreversible generation pipeline.

Primary source:

- `docs/explanation/document-generation-model.md`
