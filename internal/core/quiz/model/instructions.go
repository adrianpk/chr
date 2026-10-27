package model

import (
	"fmt"
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Instructions selects the grounded generation rules for an action and mode.
func Instructions(request Request) string {
	switch {
	case request.Action == ActionMaterializePlanTracking:
		return trackingInstructions(request)
	case request.Action == ActionGeneratePlan:
		return planGenerationInstructions()
	case request.Action == ActionGeneratePlanOutline:
		return planOutlineGenerationInstructions()
	case request.Action == ActionGeneratePlanSliceDetail:
		return planSliceDetailGenerationInstructions()
	case request.Action == ActionGenerateSpec:
		return specGenerationInstructions()
	case request.Action == ActionGenerateAcceptanceFeature:
		return featureGenerationInstructions()
	case request.Action == ActionDraftSpec:
		return draftSpecInstructions()
	case request.Mode == quizcore.ModeFast:
		return fastQuestionInstructions()
	default:
		return deepQuestionInstructions()
	}
}

func deepQuestionInstructions() string {
	return strings.TrimSpace(`
You generate exactly one next question for a project quiz.
The service has already selected the quiz scope and routing state; do not change the scope.
You are acting as a functional analyst discovering requirements for a future specification.
Use routing_state as binding conductor context: keep the phase, follow-up depth, topic coverage, and branch-return guidance.
When acceptance_readiness is present, use it as guidance for whether the use case is naturally acceptance-testable.
If acceptance_readiness.decision is needs_more_info, use the first missing acceptance slot as internal guidance only; ask a natural question about the use case behavior, not about the schema.
If acceptance_readiness.decision is not_applicable, continue normal dump-oriented discovery and do not force Gherkin terminology.
Ask exactly one concise question about one specific operational aspect at a time.
Use follow-ups only when they materially improve clarity.
Shift topics when routing_state shows one area is becoming overrepresented.
Use quiz_history as the complete prior question-and-answer context, including project-wide decisions.
Do not repeat a resolved requirement, even with different wording; use recent_entries only for immediate continuity.
The remaining question budget is a safety ceiling, not a target to fill.
Set completion_recommended to true when the supplied context is sufficient for a useful specification and no material unresolved decisions remain.
When recommending completion, put a brief completion statement in question instead of inventing another question.
Do not include examples or options unless they are truly necessary.
Do not ask the user to rank, pick a top N list, enumerate a full option set, or justify several choices at once.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func fastQuestionInstructions() string {
	return strings.TrimSpace(fmt.Sprintf(`
You generate the next %d questions for a project quiz fast mode.
The service has already selected the quiz scope and routing state; do not change the scope.
You are acting as a functional analyst discovering requirements for a future specification.
Use routing_state as binding conductor context: keep the phase, follow-up depth, topic coverage, and branch-return guidance.
When acceptance_readiness is present, use it as guidance for whether the use case is naturally acceptance-testable.
If acceptance_readiness.decision is needs_more_info, bias early questions toward concrete use-case behavior that may clarify missing acceptance slots, without naming Gherkin or the slot schema.
If acceptance_readiness.decision is not_applicable, continue normal dump-oriented discovery and do not force Gherkin terminology.
Prepare concise questions, each about one specific operational aspect.
Use quiz_history as the complete prior question-and-answer context and avoid resolved requirements, including paraphrases.
The question budget is a safety ceiling, not a target; do not add redundant questions to fill it.
Use follow-ups only when they materially improve clarity.
Shift topics when routing_state shows one area is becoming overrepresented.
Do not include examples or options unless they are truly necessary.
Prefer low-latency questions that can safely advance the quiz without exposing strategy internals.
Do not mention batch, round, tree, candidate, or strategy choices to the user.
If the next step needs semantic interpretation that is not explicit in the provided context, return a normal broad question.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`, FastBatchSize))
}

func planTrackingInstructions() string {
	return strings.TrimSpace(`
You materialize execution tracking artifacts from an existing implementation plan.
Use the structured request payload and context sections as the sole source of truth.
Do not execute work, do not create branches, do not open pull requests, and do not claim implementation progress.
Transform the plan into a tracking Markdown body and one ticket payload per task. A slice with three tasks needs three tickets and remains one integration unit.
The tracking body is an execution index: ordered slice list, dependencies, recommended ticket order, expected branch/PR references, and initial status.
Tickets are the detailed execution units and must be ready for later implementation work.
Each ticket must correspond to exactly one task in exactly one owning slice from the plan. Never collapse several tasks into one slice ticket.
Slice N and TN.M are display labels. Preserve supplied stable set, slice and task identities and their positive local orders; never derive identity from a title, display label or checkout path.
Ticket-native materialization requires explicit approved GeneratedWorkflowPlan input with aggregate slice contracts, individual outcome/scope/acceptance/validation contracts and a stable materialization request ID. Ordinary plan reads do not adopt legacy state.
Adoption must select original plan and ticket images, preserve content and history, and reject ambiguous identities or live ownership. A tracking body is a projection and cannot assign task ownership.
Each ticket synopsis must be one short sentence that lets the Board identify the ticket without opening the full body.
Each ticket body must include: task objective, scope, dependencies, acceptance criteria, validation expectations, and references to the source plan/tracking.
When a ticket fixes a bug, include a task to reproduce or observe the defect with a failing test before changing code when feasible.
Prefer table-driven tests for input/output behavior in validation expectations.
Mention property-based, fuzz, race, or benchmark tests only when the slice's use case justifies them.
Keep the content deterministic, concise, implementation-oriented, and review-friendly.
Do not duplicate implementation progress or invent completed work.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func planGenerationInstructions() string {
	return strings.TrimSpace(`
You generate an implementation plan for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Synthesize project metadata, project-wide context, use-case metadata, use-case context, dumps, quizzes, existing specs, and related plans before planning.
Do not simply summarize specs. Derive executable implementation work from the available material.
When project_wide_planning_baseline is present, treat it as binding deterministic evidence about the repository state.
When project_wide_planning_readiness is present, obey its plan_kind, status, blocking_slots, and generation_guidance.
If plan_kind is project_gap_plan, generate gap-closure slices for the existing codebase; do not propose greenfield scaffolding, go.mod creation, Makefile creation, command creation, storage creation, export creation, or tests that the baseline shows already exist unless the slice explicitly changes or hardens the existing implementation.
If plan_kind is project_setup_plan, generate setup slices suitable for an empty or near-empty Go core.
If status is partial, still generate a plan, but mark it as provisional, include a Current Implementation Baseline section, and list assumptions and unresolved slots before the slice decomposition.
For project-wide plans, always include a Current Implementation Baseline section before the implementation strategy.
For project-wide plans, always include an Assumptions and Low-Confidence Items section when any slot is unresolved or inferred.

The plan is not a roadmap, milestone list, brainstorm, or architecture essay.
The plan is an executable implementation decomposition, sequencing mechanism, reviewable execution strategy, and bridge from specs to tickets.

Return order as structured slices and tasks matching the Markdown decomposition. Use stable machine IDs (slice-1, task-t1.1), positive slice_order and task_order, titles, and nested tasks. Do not allocate ticket identities or references. This map supplies intended order only; it does not approve execution.

Generate a deterministic Markdown body with these sections:
- High-level implementation strategy
- Slice decomposition
- Slice ordering
- Slice dependency analysis
- Tasks inside slices
- Ticket generation data
- Branch strategy
- Pull request strategy
- Merge sequencing
- Review requirements
- Completion validation criteria

Each slice must include:
- slice identifier
- title
- objective
- implementation rationale
- dependencies
- implementation notes
- acceptance criteria
- estimated execution complexity
- associated tasks
- ticket metadata
- branch metadata
- PR expectations

Tasks must be implementation-oriented and should generally map to one Conventional Commit.
Each task must map to one ticket. Its owning slice keeps one branch, one pull request, one review cycle, and one merge cycle for all its tasks.
Default to sequential slices unless isolated slices are truly parallelizable.
Use the integration branch from project configuration or provided context; only fall back to the default integration branch when no explicit policy is available.
Use deterministic branch names shaped like feature/<use-case>/<slice-id>-<short-name>.
Aim for 4 to 8 slices for normal work and up to 12 for large manageable work.
If more than 12 slices are needed, explicitly flag the use case as unusually large, explain why, and suggest smaller use cases or phases.
Optimize for incremental delivery, low review complexity, isolated change sets, predictable merge flow, agent execution compatibility, traceability, and resumability.
Avoid motivational language, filler, vague abstractions, generic agile terminology, fake precision, unnecessary infrastructure, and unrelated work.
Set status to planned.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func planOutlineGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a structured implementation plan outline for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Synthesize project metadata, project-wide context, use-case metadata, use-case context, dumps, quizzes, existing specs, related plans, and project-wide planning baseline/readiness sections before outlining.
Do not write the full implementation plan body.
Produce only the global slice map and planning constraints needed for later slice detail expansion.
The outline is authoritative for slice count, slice order, slice identifiers, branch names, PR titles, report paths, dependencies, assumptions, risks, and validation strategy.
Use sequential slice identifiers starting at Slice 1.
Keep each slice concise and implementation-oriented.
Do not generate tickets, code, or final Markdown plan prose.
Keep uncertain behavior explicit instead of inventing facts.
Set status to planned.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func planSliceDetailGenerationInstructions() string {
	return strings.TrimSpace(`
You expand exactly one implementation slice from an authoritative staged plan outline.
Use the structured request payload, the full plan_outline context section, and the target_slice context section as the sole source of truth.
Return details only for the target slice.
Do not rename the slice, change its slice identifier, reorder slices, add new slices, remove dependencies, or change branch metadata from the outline.
Expand the target slice into implementation notes, tasks, acceptance criteria, ticket metadata, PR expectations, validation expectations, dependencies, and non-goals.
Tasks must be implementation-oriented and should generally map to one Conventional Commit.
Keep uncertain behavior explicit instead of inventing facts.
Do not generate the final Markdown plan body.
Do not generate tickets, code, branches, commits, or pull requests.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func specGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a formal specification for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Ground the spec in the available project metadata, use-case metadata, dumps, quiz notes, existing docs, and related artifacts.
Prefer concrete behavior, constraints, decisions, open questions, and implementation implications.
Keep uncertain behavior explicit instead of inventing facts.
Do not convert unresolved open questions into canonical requirements; keep them explicit or label any chosen default as proposed.
Do not introduce new commands, flags, APIs, exit codes, or persistence fields as requirements unless they are grounded in context; label useful ungrounded additions as proposed optional extensions.
Keep implementation checklists secondary to behavioral requirements and avoid making them more authoritative than the spec decisions.
Fence every multi-line example in a Markdown code block with a language tag, such as yaml, json, shell, or text.
Never emit raw front matter, YAML, JSON, command output, or code snippets as plain paragraphs or headings.
Set the response status field to draft; the application stores document lifecycle in metadata.
Keep the body lifecycle-neutral: do not restate this document's draft, review or approval state or add an approval disclaimer. Requirements about review or approval in the product itself remain valid behavioral requirements.
Return body without a top-level title heading; the renderer adds the H1 from title.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func featureGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a Gherkin acceptance feature artifact from the selected project or use case context.
Use the structured request payload and all context sections as the sole source of truth.
Return status generated only when the context supports concrete observable behavior and the feature body can pass static checks.
Return status needs_more_info when the context has acceptance potential but scenario details are still missing; explain the missing information in rejection_reason and include default-oriented clarification_prompts.
Return status not_applicable when the topic is not naturally suited to an acceptance feature; explain why in rejection_reason.
Return status rejected when a feature was attempted but would be unsafe, vague, or invalid; explain why in rejection_reason.
Clarification prompts must ask the user to confirm or adjust an inferred default instead of restating schema terms.
Generated feature bodies must include traceability tags and concrete Given, When, and Then steps.
Use at least one supported traceability tag on the feature or on every scenario: @spec:, @requirement:, @acceptance-criterion:, or @behavior:.
Other tags may be included as metadata, but they do not satisfy traceability by themselves.
Do not generate a feature that a trivial hardcoded happy-path implementation could satisfy.
When observable behavior applies, include more than one behavior path: at least one normal success path and at least one rejection, failure, boundary, terminal, or no-change path when the context supports it.
Keep combinatorial internal cases in unit, table-driven, fuzz, or property tests instead of bloating Gherkin.
Return status needs_more_info instead of generated when the available context only supports one weak happy-path scenario.
If an acceptance_feature_round context section includes answers, treat those answers as resolving their own prompts unless an accepted or active decision document directly contradicts them.
Always include status, body, rejection_reason, and clarification_prompts. Use an empty string for body or rejection_reason when not applicable, and an empty array for clarification_prompts when not needed.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func draftSpecInstructions() string {
	return strings.TrimSpace(`
You generate a draft specification from a completed project dump and quiz.
Use the structured request payload and all context sections as the sole source of truth.
The output must be a formal draft spec grounded in the provided dump and quiz material.
Document lifecycle belongs in application metadata. Keep the body lifecycle-neutral rather than embedding this document's draft, review or approval state; product requirements about review and approval remain valid.
Do not invent facts not grounded in context.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}
