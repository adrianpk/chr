# Work Request Routing

This document defines how incoming requests are classified for execution workflow.

## Objective

Route requests into one of these execution modes:
- discussion
- planning
- implementation
- investigation
- maintenance/meta

## Rules

- Questions are treated as discussion unless implementation is explicitly requested.
- Planning is produced only when requested.
- Implementation starts only with explicit instruction.
- Ambiguous requests should be clarified before code changes.
- During discussion or analysis, work inside the direction the user is exploring.
  Do not reflexively answer with phrases such as "I would not include this",
  "I would not implement this now", or "I would not attack this yet" when the
  user is evaluating an idea. Prefer mapping how the idea could fit, what shape
  it could take, and what concrete risks or tradeoffs would need attention.
  Push back only when there is a specific technical, safety, scope, or product
  risk to name.

## Non-goal

This document does not define product/runtime behavior.
