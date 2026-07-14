---
title: "Project Context"
status: "curated"
type: "dump"
---

# Project Context Dump

Dialer is a small CRM phone workflow service. It is not a telephony platform and
does not place real calls. It tracks leads, agent queues, call outcomes, and the
business rules that decide what happens next.

The first implementation should stay in memory, use the Go standard library, and
keep HTTP handlers thin. The useful demo surface is the workflow: context is
captured, requirements are clarified, specs and plans are generated, and tickets
move across the board.

Important domain language:

- lead: a person or company that may become a prospect;
- prospect: a qualified lead accepted for sales follow-up;
- disposition: the outcome recorded after a call;
- queue: the ordered list of leads an agent should call;
- follow-up: a future callback created from a call outcome.
