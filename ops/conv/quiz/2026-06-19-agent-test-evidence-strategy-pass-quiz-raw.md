# Agent Test/Evidence Strategy Pass Quiz Raw Capture

Status: complete

Date: 2026-06-19

Topic: Agent test and evidence strategy pass.

Related spec: `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

Question cap: 8

Completion criteria: enough clarity to specify how a development worker decides,
persists, follows, and reconciles testing and evidence strategy before moving a
ticket to review.

## Raw Capture

### Question 1

Question: Where should the Test/Evidence Strategy be persisted?

Answer: The primary operational location should be the ticket tasklist/checklist,
not a loose comment. Comments are acceptable only if chr can recognize
their semantics later. The execution report may duplicate more detailed
information, but that report is closer to a trace or log and should not be the
primary operational surface.

### Question 2

Question: When should that strategy be written into the ticket?

Answer: The strategy should be written before the agent touches code, with the
condition that the worker prompt explicitly instructs the agent to pay attention
to it. The agent must read, obey, and later update the strategy rather than
treating it as passive decoration.

### Question 3

Question: When a ticket has no user-written strategy, should the agent create a
minimal strategy before implementing, or only decide internally and report at the
end?

Answer: Strategy should not normally be a user-authored workflow step. The user
may edit or intervene, but the agent should take ownership by default. To avoid
decision branching, every ticket should have an initial mandatory strategy
stage: the agent evaluates the work, writes a checklist task in the ticket, and
then follows prompt instructions that require attention to that note during
implementation and closure.

### Question 4

Question: Should the checklist have a fixed minimum shape so chr can
recognize it later?

Answer: The checklist should be as parseable as possible to avoid ambiguity for
the agent, while retaining enough natural cadence that tickets remain pleasant
and readable for humans.

### Question 5

Question: If Gherkin is active and applies to observable behavior, must the
agent create or update a `.feature` file?

Answer: Yes. If Gherkin is active and applies, the agent must create or update
the Gherkin artifact.

### Question 6

Question: What should happen if the agent cannot execute required validation
because infrastructure is missing or broken?

Answer: Missing Docker or Codex runtime is a severe blocker. If there is no
Codex worker runtime, none of this workflow is possible. The ticket must not
move to `ready_for_review`; it should be blocked with a clear operational
cause.

### Question 7

Question: If validation runs but fails, should the agent fix the problem inside
the same ticket or stop with a failure report?

Answer: This is already decided by the dockerized Codex task contract: linting,
compilation, runtime errors, and required validation failures must not remain
when a ticket moves to `ready_for_review`.

### Question 8

Question: Should this contract be configurable from the beginning, or should the
first version use fixed defaults and become configurable later?

Answer: Configuration settings may exist from the beginning, set to the
preferred sane defaults. Documentation must explain how a project can change the
system's behavior.
