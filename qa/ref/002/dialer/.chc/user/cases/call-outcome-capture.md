---
schema_version: 1
name: "Call Outcome Capture"
description: "Record a sales call result and update lead state from an agent action."
---

Agents need to submit the outcome of a call immediately after speaking with a
lead. The system should make the result visible in the lead history, update the
lead status when the outcome is decisive, and reject ambiguous submissions that
would make follow-up work unreliable.
