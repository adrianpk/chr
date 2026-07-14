---
title: "Lead Queue Quiz"
status: "curated"
raw_path: "ops/default/quiz/<id>/artifact.mdlead-queue-quiz-raw.md"
case: "lead-queue"
---

# Lead Queue Quiz

## Summary

The queue must prioritize callable leads for one agent at a time. Leads are
callable when they are assigned to that agent, not marked do-not-call, and not
scheduled for a future callback.

## Curated Answers

- Highest score should come first.
- Do-not-call records must be excluded even if they have high score.
- Future callbacks should return to the queue only when due.
- A missing agent can return an empty queue for the first implementation.
