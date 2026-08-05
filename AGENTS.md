# Agent Operating Contract

Read this document before starting work in this workspace.

This contract defines agent execution, planning, communication, workspace safety, and documentation boundaries.
It is not a product or runtime specification.

## 1) Execution

- Execute only after a clear implementation request.
- Treat questions and discussion as alignment, not authorization to change files.
- For high-impact changes, present the approach and wait for explicit approval.
- Complete the requested scope, report the outcome, and stop.
- Do not add unsolicited follow-up work.

## 2) Planning Protocol

- Before creating a plan, ask whether the user wants a plan.
- If requested, make the plan concrete and execution-oriented.
- Do not implement while a plan is pending confirmation.
- Keep implementation inside the approved scope.

## 3) Communication Protocol

- Use direct, neutral language.
- Keep updates factual and concise.
- Explain reasoning before proposing significant changes.
- Report validation results explicitly.
- Trust user-provided facts unless technical evidence proves otherwise.

## 4) State Alignment Before Action

- When the user asks a status question, expresses confusion, challenges the
  current state, or asks what comes next, answer with the verified state first
  and do not change files, branches, tickets, trackers, or PRs in that response.
- Only execute changes after the user gives a clear follow-up instruction such
  as "continue", "do it", "close it", "activate it", or "start the next slice".
- Before closing, activating, or starting a slice, verify the current branch,
  remote `dev`, tracker active slice, and PR or merge state. Report mismatches
  before making changes.

## 5) Workspace and Git Safety

- When this workspace uses git, `origin` is the default canonical remote unless local instructions say otherwise.
- Never perform destructive Git operations without explicit approval.

## 6) Language and Documentation Rules

- All code and workspace documentation must be written in English.
- Keep naming, comments, and workspace text in English.

## 7) Scope Boundary

- Keep this file focused on working rules.
- Product and runtime specifications belong in `docs/`.

## 8) Referenced Agent Docs

- `.agents/DOCUMENTING.md` — documentation model and placement rules.
- `.agents/CONVENTIONS.md` — coding standards and codebase boundaries.
- `.agents/ARCHITECTURE.md` — workspace structure map.
- `.agents/LIFECYCLE.md` — delivery lifecycle and handoff rules.
- `.agents/VERSIONING.md` — commit, branch, and merge rules.
- `.agents/SLICES.md` — task decomposition and delivery strategy.
- `.agents/ROUTING.md` — request classification for execution workflow.
- `.agents/PLAYBOOKS.md` — operational procedures.
- `.agents/REPORTING.md` — persistent implementation reporting rules.
- `.agents/LINTING.md` — linting policy and quality-gate setup.
- `.agents/INVARIANTS.md` — non-negotiable collaboration invariants.

## Acta Activity Record

- After creating a commit for completed work that is not hypertrivial, run
  `acta record [flags] [note]` immediately while that commit is still `HEAD`.
- Record features, fixes, refactors, meaningful documentation, tests,
  maintenance, decisions, specifications, plans, and operational milestones.
  During initial adoption, record the commit when uncertain.
- Omit only abandoned or work-in-progress commits, generated-only changes,
  punctuation-only changes, isolated formatting, or similarly trivial edits.
- Add a concise note when the commit subject lacks useful retrieval terms or
  does not explain why the change matters. Supply provenance flags only when
  their values are known reliably.
- Never record credentials, tokens, personal data, or hidden reasoning.
