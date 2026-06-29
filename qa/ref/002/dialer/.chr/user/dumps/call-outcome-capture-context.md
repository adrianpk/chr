---
title: "Call Outcome Capture Context"
status: "raw"
type: "dump"
case: "call-outcome-capture"
---

# Call Outcome Capture Context Dump

The dialer needs a clearer workflow for what happens after an agent completes a
call. Right now the important product question is not how the storage layer is
implemented, but what the agent can do and what the lead record should show
after the action.

The likely actor is a sales agent working from a lead detail screen or an agent
queue. The trigger is the agent submitting the outcome of a just-completed call.
The initial state should include an existing lead assigned to the agent and no
new call outcome recorded for the current attempt.

Useful outcomes include `qualified`, `not_interested`, `no_answer`, and
`callback_requested`. A qualified outcome should make the lead ready for
conversion follow-up. A not-interested outcome should close the lead for normal
calling. A no-answer outcome should leave the lead callable later. A callback
request probably needs a future callback time, but the exact validation rules
are still open.

The behavior should be observable through the HTTP API and through the lead
history returned to the UI. A good quiz should confirm the actor, endpoint,
required inputs, lead-state transitions, validation errors, and whether callback
requests belong in this use case or should remain part of the existing follow-up
workflow.

This use case is intentionally suitable for acceptance-style scenarios because
it has a clear actor, trigger, initial state, action, and visible result. It
should let the quiz decide whether enough information exists for Gherkin-ready
questions without requiring a prewritten quiz artifact.
