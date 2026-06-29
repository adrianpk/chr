# Dialer QA Gallery Scenario

This project is a small Go CRM dialer used to capture Chryso gallery
screenshots.

It models leads, call attempts, agent queues, conversion decisions, and follow-up
work without external dependencies. The code is intentionally compact; the
Chryso workspace is pre-seeded with realistic context, use cases, specs,
plans, trackers, and tickets so the TUI has useful material immediately after a
reset.

The `call-outcome-capture` use case is intentionally seeded with only use-case
metadata and a raw dump. It is meant to exercise context-aware quiz generation
against a behavior-oriented topic without an existing quiz, spec, or plan.

Reset from the parent repository with:

```sh
make qa-reset QA_CASE=dialer QA_REF=002
```
