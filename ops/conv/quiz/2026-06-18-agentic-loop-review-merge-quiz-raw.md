# Agentic Loop Review and Merge Quiz Raw Capture

Status: complete

Date: 2026-06-18

Topic: Agentic loop review and merge workflow.

Related spec: `ops/specs/2026-06-18-agentic-loop-review-merge.md`

Question cap: 50

Completion criteria: enough clarity to refine the spec around loop mode defaults,
integration unit policy, review evidence, review round limits, merge behavior,
and artifact ownership without overfitting to one provider.

## Raw Capture

### Question 1

Question: For a new managed project, which automatic loop mode should
chr use by default: development only, development plus review, or
development plus review plus merge?

Answer: The default should be the natural full cycle, `dev_review_merge`.
Project configuration may allow automatic merge. When a feature has been
developed, tested, and validated successfully, chr should be able to
merge the pull request automatically instead of leaving extra manual review and
merge interventions in the path.

### Question 2

Question: For the first implementation, what should be the default integration
unit that owns the pull request, review, and merge handoff: ticket, slice, or a
configurable policy with only one unit implemented first?

Answer: This is already defined by the current workflow shape: one ticket maps
to one branch, and one branch maps to one pull request. That pull request should
be reviewed and merged. Slice naming may appear naturally in the ticket, branch,
or pull request names, but the first workflow does not need a separate slice or
batch integration unit.

### Question 3

Question: Should automatic merge be enabled by default for new projects when
review passes and validation succeeds, or should the default require the user to
opt in through configuration?

Answer: Automatic merge is repository configuration. In this discussion,
`automerge` means chr does not launch separate reviewer and merger agent
roles for that path. A pull request is still desirable as evidence so a human can
inspect the development history and resulting change. When the feature is
developed, tested, and validated, the repository policy may allow the pull
request to merge automatically.

### Question 4

Question: For new projects, should this automatic merge policy default to off
until the project explicitly enables it, or should chr initialize new
projects with it enabled?

Answer: Automatic merge should be the default method for new projects. The
documentation must be clear that a more formal reviewer/merger role approach is
available. Even if it is not the default, the more protocol-heavy mode must have
intensive test coverage because it is an important supported workflow.

### Question 5

Question: When automatic merge is enabled, what minimum evidence must exist
before chr may merge a ticket pull request?

Answer: The development agent must be strict about implementation quality. The
implementation includes extensive tests: unit tests, table-driven tests,
fuzz/property-style checks where applicable, and stronger quality techniques
such as CRAP and mutation-oriented checks when feasible for the project. Linting
must be part of the gate. The application must also be runnable. Automatic merge
is allowed only when the implementation, tests, linting, validation, and
runability conditions are satisfied.

### Question 6

Question: Should chr encode those quality requirements as a configurable
validation profile, or should it only rely on the project's existing validation
command to enforce them?

Answer: The immediate enforcement mechanism should remain the project
`validation_command`. This deserves review, but chr should use the
initial setup path to influence the project baseline toward the strongest
practical code-quality gate the project can support. When the setup context
allows it, chr should be exhaustive and help establish tests, linting,
runability checks, and stronger quality practices as part of the baseline
validation command.

### Question 7

Question: Should the formal reviewer/merger mode use the same validation command
as auto-merge, or should it require extra review-specific checks before approval?

Answer: Code validation should not imply that the reviewer modifies code created
by the development role. There is a concern about crossing role boundaries if the
reviewer changes the branch. At the same time, review may reasonably need to
"try the thing" in some cases. Human review usually inspects the pull request and
evidence rather than always checking out the branch and running it, although that
does happen for some changes. The formal reviewer mode should distinguish
between read-only validation/evidence inspection and any code mutation.

### Question 8

Question: In formal reviewer mode, should the reviewer be allowed to run
read-only validation in a clean runtime, or should it only inspect the evidence
produced by the development run unless a human asks for deeper verification?

Answer: In the current sequential process, the remote branch is normally the same
branch currently checked out locally, unless a user or another actor is changing
state in parallel. If the reviewer verifies that it is really on the expected
branch, it should be able to run tests and validation. It must not mutate code.
This raises a stronger future design consideration: chr may eventually
need a git worktree model so the user remains free to interact with the original
repository while agent roles run in isolated workspaces. This is not required for
the immediate implementation but should be considered for the future.

### Question 9

Question: When the reviewer finds problems, should it write findings only to the
local chr artifact at first, or should it also publish comments to the
pull request by default?

Answer: Findings should be local first. This keeps the workflow simple, avoids
API integration requirements, and lets chr leave messages, update local
state, and reassign or return the ticket without depending on a forge provider.
Publishing comments to the pull request, and even obtaining the diff from the
remote provider, would make the workflow more similar to traditional review and
can be convenient. The product still needs a position on provider scope: remain
platform-agnostic, initially support open-source forge providers such as
Forgejo, and eventually account for GitHub because it remains heavily used in
industry despite current weaknesses.

### Question 10

Question: For the first implementation of formal review, should provider
integration be limited to local PR metadata and local artifacts, or should it
include one real provider integration such as Forgejo from the start?

Answer: Ideally the formal mode should support both GitHub through `gh` and
Forgejo through `tea`. When possible, it should operate in a traditional review
style and delegate the review dynamics to the forge provider, while still
recording the relevant state and handoff in local tickets. Review comments may
use explicit `@` mentions.

### Question 11

Question: Should provider-backed formal review be required for the automatic
full-cycle default, or should the default still work local-first when provider
credentials or CLIs are unavailable?

Answer: The complexity of provider integration is not fully clear. The API work
itself may not be hard, but handling project-specific forge setup, credentials,
and conventions can become heavy. A local-only path is easier and should remain
available. Local ticket comments should not become long review-status dumps. A
better shape is a reference mechanism: compact ticket metadata or comments point
to a Markdown review artifact stored in a specific review directory associated
with the pull request.

### Question 12

Question: Should local review artifacts be stored under the project workspace
runtime area, under `ops/artifacts`, or under a dedicated review directory that
is intended to be versioned?

Answer: Operational chr artifacts should live under `./.chr`. If an
`ops` directory is inside `.chr`, then the policy is already clear. If `ops` is
outside `.chr`, that should be reconsidered for this workflow. It is not correct
for chr's own operational activity to be reflected in directories outside
`./.chr`.

### Question 13

Question: Should review artifacts be versioned by default when they live under
`.chr`, or should they be local runtime evidence that may be ignored unless the
project opts into versioning them?

Answer: This should be configurable, with the default set to versioned review
artifacts.

### Question 14

Question: When review finds issues and returns work to development, should the
same ticket, branch, and pull request be reused, or should chr create a
new repair branch or pull request?

Answer: The same ticket, branch, and pull request should be reused. Traditional
review does not normally force a developer to create a new branch when changes
are requested. The developer updates the existing branch and the same pull
request carries the review cycle.

### Question 15

Question: What should be the default maximum number of review repair rounds
before chr blocks the ticket for human attention?

Answer: The default maximum should be one repair cycle. In an agent-based system
with established architecture, conventions, and linting, there is no ego-driven
review negotiation. More than one review repair cycle should not be acceptable by
default and should block for human attention.

### Question 16

Question: When a ticket blocks after the maximum repair round, should the whole
loop stop immediately, or may chr skip that ticket and continue with
later independent tickets?

Answer: The loop must stop. The workflow is intentionally sequential: a single,
steady cycle. chr should not continue advancing the slicing sequence
when one unit blocks.

### Question 17

Question: Should the merger phase perform only provider merge/preflight, or
should it also re-run the validation command immediately before merging?

Answer: The merger should not re-run validation just to add ceremony or consume
tokens. It should rely on valid prior evidence and perform only objective merge
preflight/provider integration checks.

### Question 18

Question: Should the formal reviewer phase always run, or should it run only
when project configuration disables the default automatic merge path?

Answer: This follows from the prior decisions. The formal reviewer phase should
not always run. The default path is automatic merge after strict development
validation and evidence. The formal reviewer/merger path is a configurable,
supported, well-tested mode for projects that want the more protocol-heavy
workflow.

### Question 19

Question: Should the current spec be revised to make automatic merge the default
path and formal reviewer/merger mode the configurable alternative?

Answer: The spec should explicitly reflect that automatic merge is the default
path and that formal reviewer/merger mode is a configurable alternative.
