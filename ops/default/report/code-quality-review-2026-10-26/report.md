# Code Quality Review

Status: reviewed; accepted findings remediated; integrated acceptance completed

Source baseline: `28c3db33` on `dev`, inspected on 2026-10-26.

Detections: [`detections.md`](detections.md)

Scope: general quality pass over production Go code in `cmd/` and
`internal/`, after the previous deep review and its repairs. This report
records findings, maintainer decisions and the proposed remedy for each
finding. It changes no code, lint configuration or workflow policy.

## Method

1. Deterministic `go/ast` scan of all non-test Go files in `cmd/` and
   `internal/`: parameter count, result shape, nesting depth, function and
   file length, struct width and interface size. Lists derived from this scan
   are complete for the baseline and are in `detections.md`.
2. One-off `golangci-lint` run with linters that the repository gate does not
   enable: `dupl` (threshold 120), `unparam`, `unused`, `goconst`, `gocritic`
   (diagnostic, performance), `interfacebloat` (8), `errname`, and `revive`
   rules `argument-limit`, `function-result-limit`, `confusing-results`,
   `flag-parameter` and `max-control-nesting`. Test files were excluded. These
   results are also deterministic for the baseline, but they depend on linter
   heuristics and thresholds.
3. Manual reading of samples. Conclusions from reading are judgments, not
   complete lists.

The repository gate was not rerun.

## Baseline

| Metric | Value |
| --- | --- |
| Production Go functions | 8,286 |
| Nesting depth 0 / 1 / 2 / 3 / 4 | 1,837 / 4,534 / 1,766 / 147 / 2 |
| Functions with 5 parameters | 230 |
| Functions with more than 5 parameters | 121 |
| Functions and interface methods outside the result policy | 420 |
| Files above 800 lines | 14 |
| `WorkflowStore` methods | 909 |
| Complexity suppressions in production code | 11 |

The existing gate already enforces `cyclop` 10, `gocognit` 15, `nestif` 3 and
`funlen` 60/40. Cyclomatic complexity and nesting are under control. The
remaining problems are structural.

## Decision Summary

| # | Finding | Maintainer decision |
| --- | --- | --- |
| 1 | Parameter counts | List all detections; fix; enforce |
| 2 | Result shapes | Allow only one value plus optional `error` or `bool`; fix all others |
| 3 | `internal/core` size and naming | Split into packages with plain, exact names |
| 4 | Oversized files | Apply the most useful remedy that is not worse than the problem |
| 5 | Wide structs | Follow the proposed remedy |
| 6 | Duplication | Remove where semantics stay identical; verify against specs and plans |
| 7 | Dead production code | Remove unless it has a real future use |
| 8 | Command dependency wiring | Use one idiomatic, consistent pattern |
| 9 | Positional SQL | Follow the proposed remedy |
| 10 | Minor findings | Fix what is practical in this round; defer the rest explicitly |

## Findings

### 1. Parameter counts

Hard limit: 7. Target: 5.

Detections are complete for the baseline: 121 functions take more than 5
parameters (19 above 7) and 230 take exactly 5. See `detections.md`, sections
A and B. 86 of the 121 are in `internal/core`, 31 in `internal/app`.

Causes:

1. Workflow and Telegram store helpers pass the same group through each step:
   `ctx`, `tx`, `claim`, `prepared`, plus step-specific values. The merger
   driver passes `claim, prepared, tools, call, gate, permit, index, stdout,
   stderr` unchanged from its caller.
2. Ticket loop helpers in `internal/app/ticket.go` repeat `env, repository,
   projectContext, record, eventSink, assetsFS` in each step.
3. Row inserts receive each column as a parameter, for example
   `insertTelegramWorkNotice`.
4. HTTP senders receive each message field as a parameter, for example
   `TelegramReceiverAPI.SendDocument`.

Remedy:

1. Repeated groups become one value per operation, and the steps become
   methods on it. Examples: `mergeRun` with `claim`, `prepared`, `tools`,
   `call`, `gate`, `permit`, `index` and the outputs; `ticketLoop` with `env`,
   `repository`, `projectContext`, `eventSink` and `assetsFS`. `ctx` stays an
   explicit first parameter and is never stored in the struct.
2. Row inserts take the row value. This is the same change as finding 9.
3. HTTP senders take a request value, for example `SendDocumentRequest`.
4. Do not hide many unrelated values in a generic `options` bag. Each grouping
   must name one concept.

Enforcement, deterministic and permanent:

1. Add `revive` `argument-limit` to `.golangci.yml`. Set it to 7 at once with
   `issues.new-from-merge-base: dev`, so new and changed code cannot exceed
   the hard limit.
2. Fix the 19 functions above 7 and remove the merge-base exception for this
   rule.
3. Lower the limit to 6, fix, then to 5, fix. Each step is one delivery unit.

This changes repository lint configuration. Conventions require approval
before new build tools; `revive` is already enabled, so only its rule set
changes.

### 2. Result shapes

Policy: a function returns nothing, one value, `(T, error)` or `(T, bool)`.
Every other shape is a defect.

Detections are complete for the baseline: 420 functions and interface methods
violate the policy. 118 return `(T, bool, error)`, 177 return two values plus
`error`, 43 return 4 or more values, 18 return three values without `error`,
and 64 return two values where the second is not `error` or `bool`. See `detections.md`, section C.

Main shapes and their remedy:

| Shape | Count | Remedy |
| --- | --- | --- |
| `(T, bool, error)` | 118 | If `false` means "not found", return `(T, error)` with a sentinel such as `ErrNotFound` and test it with `errors.Is`. If `bool` reports an outcome such as "applied" or "changed", put it in the result type: `Result{Value, Applied}`. |
| `(A, B, error)`, for example `(string, string, error)` | 177 | Return one named struct. If the two values are independent, split the function. |
| 4 or more results, or 3 results without `error` | 61 | Return a struct that names the values, for example `TelegramTarget{Conversation, Operation}`. Often the struct already exists and is unpacked, as in `runtimeArtifactPaths`. |
| `([]Item, Summary)` | 3 | Return one report type with `Items` and `Summary`. |
| `(int, int)`, `(string, string)`, `(int64, int64)` | 11 | Return a struct with named fields, for example `Size{Width, Height}`. |
| `(tea.Model, tea.Cmd)` in TUI helpers | 31 | Only `Model.Update` must have this shape. Helpers become pointer-receiver methods that change the model and return `tea.Cmd`. See finding 5. |
| `(tea.Model, tea.Cmd, bool)` | 6 | Same as above, returning `(tea.Cmd, bool)`. |
| `(context.Context, context.CancelFunc)` and paired channels | 2 | Return a struct, or restructure so the caller derives the context. Keeping the `context.WithCancel` shape is the only defensible exception; the maintainer decides. |

Worst cases: `runtimeArtifactPaths` (`internal/core/runtime_plan.go:186`)
returns six `string` values and an `error`, after it builds a
`runtimeArtifactPathSet` and unpacks it. `prepareWorkflowIntegrationValidation`
(`internal/core/workflow_preparation_claim.go:12`) returns 6 values.

Exceptions: only signatures imposed by an external interface, today only
`tea.Model.Update`. Record each exception in the check's allow list.

Enforcement, deterministic and permanent:

1. Enable `revive` `function-result-limit` with 2. This rejects 3 or more
   results.
2. No standard linter rejects `(T, U)` where `U` is not `error` or `bool`.
   Add a small repository-owned check with `go/ast` from the standard library,
   for example under `scripts/`, and run it in `make verify`. It reports any
   two-result signature whose second type is not `error` or `bool`, except
   allow-listed interface implementations. This is a new tool and needs
   approval under the conventions.
3. Use `issues.new-from-merge-base: dev` at first, then remove it after the
   detections are fixed.

### 3. `internal/core` size and naming

`internal/core` is one package with 670 production files and 116,785 lines.
323 files start with `workflow_`, 70 with `telegram_`. `WorkflowStore` has 909
methods. File-name prefixes act as package boundaries that the compiler does
not enforce, and identifiers repeat the prefix to avoid collisions, for
example `WorkflowMergerAcceptance` and `workflowMergerDriverResult`.

Naming rule for the split. Use the names a careful systems programmer would
choose when every name must be exact and must survive without explanation:

1. A package name is one plain noun for the thing the package owns:
   `telegram`, `outbox`, `merge`, `review`, `ticket`. No metaphors, no
   abbreviations a new reader cannot expand, no `util`, `common` or `helpers`.
2. The package name qualifies its identifiers. Do not repeat it:
   `merge.Driver`, not `merge.MergeDriver` or `core.WorkflowMergerDriver`.
3. Functions are a verb and an object: `merge.Run`, `outbox.Enqueue`,
   `review.Accept`. Types are nouns: `merge.Attempt`, `review.Decision`.
4. One concept, one word. Do not alternate "merger", "merge" and
   "integration" for the same step.
5. Rename internal jargon that does not say what it is. Candidates:
   `native` (forge pull-request integration; propose `forge`), `frontier`,
   `adoption`, `loose`, `continuation`, `amendment`, `closing`. The maintainer
   confirms each meaning before rename, because specs and plans use these
   terms.
6. Record each term in a glossary under `docs/reference/` with its package.

Coupling measured by identifier references:

| Prefix group | Files | Distinct `workflow*` identifiers used | `WorkflowStore` mentions |
| --- | --- | --- | --- |
| `telegram_` | 70 | 3 | 0 |
| `harness_` | 15 | 6 | 0 |
| `quiz_` | 14 | 4 | 0 |
| `authoring_` | 19 | 5 | 0 |
| `project_` | 14 | 4 | 0 |
| `notification_` | 7 | 1 | 0 |
| `gate_` | 6 | 0 | 0 |
| `ticket_` | 30 | 13 | 11 |
| `plan_` | 19 | 14 | 3 |
| `guided_` | 16 | 22 | 7 |

Proposed target layout, still under `internal/core` as
`.agents/ARCHITECTURE.md` requires:

| Package | Source files | Order |
| --- | --- | --- |
| `core/gate` | `gate_*` | 1 |
| `core/notify` | `notification_*` | 1 |
| `core/telegram` | `telegram_*` | 2 |
| `core/quiz`, `core/quiz/openai` | `quiz_*` | 2 |
| `core/project` | `project_*` | 2 |
| `core/authoring` | `authoring_*` | 3 |
| `internal/harness` or `core/harness` | `harness_*` | 3; merge into the existing `internal/harness` if the dependency direction allows it |
| `core/ticket` | `ticket_*` | 4 |
| `core/workflow` | coordinator, store, claim, lifecycle, state | 5 |
| `core/workflow/merge` | `workflow_merger_*` (46) | 5 |
| `core/workflow/integrate` | `workflow_integration_*` (38), `workflow_preparation_*` (15) | 5 |
| `core/workflow/forge` | `workflow_native_*` (25), `workflow_submission_*`, `workflow_pull_*`, `workflow_publication_*` | 5 |
| `core/workflow/review` | `workflow_review_*` | 5 |
| `core/workflow/closeout` | `workflow_closing_*` (16) | 5 |
| `core` | shared kernel: project paths, time format, identifiers, hashing, shared errors | remains |

The main obstacle is `WorkflowStore`. Go cannot define methods of one type in
several packages. Remedy: `WorkflowStore` keeps persistence and transactions.
Each stage package defines the small interface it needs (for example
`merge.Store` with the 5 to 10 operations the stage uses) and owns its logic
as functions or types over that interface. `core/workflow` implements the
interfaces and coordinates the stages. Stage packages never import
`core/workflow`.

Guard against a remedy worse than the problem:

1. Move one package per delivery unit with `git mv`, so history follows the
   files. Each unit compiles, passes `make verify` and changes no behavior.
2. Move first, rename second. Do not mix a move and a semantic change in one
   commit.
3. If a move forces many unexported symbols to become exported, the boundary
   is wrong. Stop and record why instead of exporting the internals.
4. This needs its own plan and tracker. It is not a ticket.

### 4. Oversized files

| Lines | Functions | File | Remedy |
| --- | --- | --- | --- |
| 2,962 | 142 | `internal/app/ticket.go` | Split by subcommand, then move orchestration to the service (see below) |
| 1,940 | 87 | `internal/tui/model.go` | Split after the model change in finding 5 |
| 1,626 | 70 | `internal/core/quiz_openai_model.go` | Moves into `core/quiz/openai` (finding 3); split schema, request and response there |
| 1,558 | 89 | `internal/tui/shell.go` | Split by screen region |
| 1,348 | 76 | `internal/core/work_readiness_service.go` | Split after removing its internal duplicate (finding 6) |
| 1,342 | 72 | `internal/app/author.go` | Split by subcommand |
| 1,217 | — | `internal/core/quiz_model_payload.go` | Moves with the quiz package |

The most useful remedy for `internal/app/ticket.go`, in order:

1. Mechanical split by subcommand: `ticket_loop.go`, `ticket_review.go`,
   `ticket_merge.go`, `ticket_evidence.go`, `ticket_start.go`,
   `ticket_print.go`. No behavior change; low risk; review is easy.
2. Move orchestration out of the CLI adapter. Today the file constructs core
   repositories and adapters (`core.NewExecutionRepository`,
   `core.NewRemoteReviewAdapter`, `core.NewTicketReconciler`,
   `core.LoadWorkspaceConfig` six times) and changes loop-run records
   (`repository.UpdateLoopRun`). The conventions require thin adapters and
   reusable operations in the operation service. Move loop execution, review,
   merge and start preflight into `service_ticket_*.go` methods. The CLI keeps
   argument parsing and output.
3. Check whether the TUI, web or Telegram adapters repeat these flows; route
   them to the same service methods.

Do not split a cohesive file only to meet a line count. A schema literal or a
single parser can stay long. Use about 800 lines as a review signal, not as a
lint failure.

### 5. Wide structs

| Fields | Struct |
| --- | --- |
| 85 | `tui.Model` — `internal/tui/model.go:111` |
| 48 | `core.WorkflowInvocation` — `internal/core/workflow_invocation.go:83` |
| 47 | `webui.GuidedUseCaseView` — `internal/webui/guided_use_case.go:24` |
| 40 | `core.IndexedTicket` — `internal/core/indexed_ticket.go:6` |
| 39 | `core.WorkflowMergerInput` — `internal/core/workflow_merger.go:37` |
| 37 | `app.WorkflowWorkStatus` — `internal/app/service_workflow_status.go:30` |

Remedy:

1. `tui.Model` (about 49.8 KB, value receivers in 216 methods). Bubble Tea
   accepts a pointer model. Make `*Model` the `tea.Model`, use pointer
   receivers, and group the 85 fields into named sub-states: navigation,
   active interaction or modal, layout, background work, and one state per
   view (several view structs already exist). The update helpers then change
   the model and return `tea.Cmd`, which also removes 37 result-shape
   detections from finding 2. Add a small benchmark for `Update` and `View`
   before and after.
2. Core records (`WorkflowInvocation`, `IndexedTicket`,
   `WorkflowMergerInput`). Group cohesive fields into embedded structs, for
   example identity, Git references, timing and outcome. `encoding/json`
   flattens embedded structs, so the JSON shape stays the same. Verify each
   persisted or wire shape with an existing round-trip test, or add one before
   the change. SQL scans are explicit and need matching updates.
3. View models (`GuidedUseCaseView`, `WorkflowWorkStatus`). Split by page
   section, so each template partial receives only its section.
4. Do not act on the 4,374 general `hugeParam` reports. Most are value
   semantics by design. Change only measured hot paths.

### 6. Duplication

`dupl` found 11 duplicate pairs. Remedy for each, without semantic change:

| # | Duplicate | Remedy |
| --- | --- | --- |
| 1 | `internal/tui/plans.go:42-102` and `internal/tui/specs.go:42-102` | One document list view parameterized by artifact kind |
| 2 | `internal/core/plan_catalog.go:13-52` and `internal/core/spec_catalog.go:15-54` | One document catalog parameterized by kind and location |
| 3 | `internal/core/work_readiness_service.go:108-150` and `:182-224` | One helper with the differing step as a parameter |
| 4 | `internal/app/service_spec.go:305-340` and `internal/app/service_acceptance_feature.go:195-230` | Part of the shared generation pipeline below |
| 5 | `internal/app/service_plan.go:677-701` and `internal/app/service_spec.go:202-226` | Part of the shared generation pipeline below |
| 6 | `internal/app/telegram_explanation_dispatch.go:12-39` and `internal/app/telegram_work_dispatch.go:13-40` | One dispatch helper; each file keeps its handler table |
| 7 | `internal/app/service_guided_admission.go:11-29` and `internal/app/service_guided_preparation.go:14-32` | One shared guided preflight |
| 8 | `internal/core/workflow_control_receipt.go:62-87` and `internal/core/workflow_mutation.go:136-161` | One receipt writer |
| 9 | `internal/core/authoring_contextual_quiz.go:160-181` and `internal/core/authoring_promotion.go:61-82` | One shared authoring step |
| 10 | `internal/app/dump_cli.go:225-255` and `:257-287` | One function with the differing value as a parameter |
| 11 | `internal/tui/shell.go:636-656` and `:658-678` | One function with the differing value as a parameter |

`service.generateSpec` (97 lines) and `service.generateAcceptanceFeature` (99
lines) have the same shape, and the plan generator shares a block with spec
generation. Extract one generation pipeline: load context, call model,
validate, record, log. This removes duplicates 4 and 5 and two `funlen`
suppressions.

Semantic guard:

1. Before each extraction, read the governing spec or plan for both callers
   and confirm that the duplicated behavior is the same rule, not two rules
   that are equal today.
2. If the rules differ by intent, keep the copies and record why in one line.
3. One pair per commit, with the existing tests unchanged and passing.
4. After cleanup, add `dupl` to the gate with a threshold of about 150.

### 7. Dead production code

`unused` reports 7 production symbols that only tests reference. Decision:
remove, unless a real future use exists.

| Symbol | Remedy |
| --- | --- |
| `service.dumpDirectory` — `internal/app/service_dump.go:123` | Delete with its test references |
| `service.newQuizModelRequest` — `internal/app/service_quiz.go:174` | Delete; point tests at the production request path |
| `telegramGuidedResultExport` — `internal/app/telegram_guided_document.go:61` | Delete with its test |
| `telegramHelpReply` — `internal/app/telegram_help.go:17` | Delete with its test, after confirming help replies use another path |
| `timePtr` — `internal/core/ticket_index.go:565` | Move into a `_test.go` helper |
| `WorkflowStore.prepareWorkflowTask` — `internal/core/workflow_execution_prepare.go:18` | It is a one-line wrapper over `prepareWorkflowTaskMode(..., "automatic")`. Move it into a `_test.go` helper, or make the tests pass the mode explicitly |
| `withoutWorkflowPublicationProfile` — `internal/core/workflow_upgrade_journal.go:279` | Delete, or move to a test helper if the test still needs it |

Correction to the earlier draft: `prepareWorkflowTask` delegates to the
production path, so its tests do exercise production code. The risk is lower
than first stated.

Unused parameters from `unparam` are in finding 8 and in item 10.4.

### 8. Command dependency wiring

Two paths deliver the same dependencies to commands:

1. 15 constructors such as `newTicketCommand(stdout, deps)` receive `deps`
   and ignore it.
2. Actions read `cmd.Root().Metadata[appDependenciesMetadataKey]` through
   `appDependenciesFromCommand` and type-assert it at run time
   (`internal/app/app.go:23`). This is a service locator over an untyped map.
   A wiring mistake becomes a run-time error instead of a compile error.

Remedy: constructor injection through one value with method actions.

```go
type commands struct {
	stdout io.Writer
	deps   appDependencies
}

func (c commands) ticket() *cli.Command {
	return &cli.Command{
		Name: "ticket",
		Commands: []*cli.Command{
			{Name: "review", Action: c.ticketReview},
		},
	}
}

func (c commands) ticketReview(ctx context.Context, cmd *cli.Command) error {
	// c.deps and c.stdout are available; no metadata lookup.
}
```

Method values satisfy the action signature of `urfave/cli`. Each command gets
its dependencies at construction, the compiler checks them, and
`appDependenciesFromCommand` and the metadata key are removed. This also
removes the `stdout` and `deps` parameters from many action helpers, which
reduces finding 1.

### 9. Positional SQL

Several statements bind 16 to 27 positional arguments, for example
`internal/core/execution_repository.go:95` and `:136` (27),
`internal/core/remote_command_record_repository.go:89` (24),
`internal/core/telegram_document_outbox.go:74` (22). Swapping two values of
the same type compiles and can pass shallow tests.

Remedy:

1. Per table, one row type and one column list, side by side:

   ```go
   var executionColumns = []string{"id", "ticket_key", "status" /* ... */}

   func (r executionRow) values() []any {
       return []any{r.ID, r.TicketKey, r.Status /* ... */}
   }

   func (r *executionRow) targets() []any {
       return []any{&r.ID, &r.TicketKey, &r.Status /* ... */}
   }
   ```

2. Build `INSERT`, `UPDATE` and `SELECT` text from the column list with two
   small shared functions. Do not add a query builder or ORM dependency.
3. A table test checks that `len(columns) == len(values()) == len(targets())`
   for every row type.
4. One round-trip test per repository writes distinct values to every column
   and reads them back.
5. Apply when each repository is touched, starting with the statements above
   16 arguments.

### 10. Minor findings

Fix in this round:

1. Dispatch switches with complexity suppressions:
   `tuiPlanActionExecutor.ExecutePlanAction`
   (`internal/app/tui_adapters_spec_plan_ticket.go:93`, 109 lines),
   `internal/tui/command_palette.go:247` and `:353`, and
   `internal/tui/model.go:495`. Replace with a table from action to handler.
   Express key hints (`internal/tui/model.go:1394`, `:1441`) as data per view
   and per mode. Keep the schema-literal suppression in
   `internal/core/quiz_openai_model.go:1246`.
2. Error type names: rename the 8 error types without the `...Error` suffix,
   for example `telegramServiceFailure` in `internal/app/telegram_failure.go:9`.
3. Bare suppressions: give each of the 23 `nolint:errcheck` directives without
   a reason one. For read-only handles, use one `closeQuietly(io.Closer)`
   helper.
4. Unused parameters and results: `stdout` in `newTelegramServeCommand`
   (`internal/app/telegram_serve.go:28`), `ctx` in
   `WorkflowStore.workflowClosingInvocationFiles`
   (`internal/core/workflow_closing_invocation.go:65`), and result 1 of
   `serviceFromCommand` (`internal/app/metadata_cli.go:169`).
5. State literals: `goconst` reports state values repeated many times, for
   example `"blocked"` 82, `"pending"` 50, `"paused"` 46, `"running"` 45,
   `"completed"` 42, `"integrated"` 38, `"confirmed"` 38. Define a typed string
   per state machine with constants, as `core.TicketStatus*` already does.
   Leave SQL column names and JSON keys as literals.
6. Stringly typed modes: `prepareWorkflowTaskMode` takes `mode string` with
   values such as `"automatic"`. Use a typed mode.

Defer to a later round:

1. Control flags: 88 boolean control parameters (`dryRun`, `clearChoice`,
   `replace`, `explicit`). Split a function where the flag selects a different
   behavior. Review case by case.
2. Legacy surface: about 300 production references to legacy paths and
   several `*_legacy*.go` files. Each needs a recorded removal condition before
   deletion.
3. General `hugeParam` and `rangeValCopy` reports, except measured hot paths.

## Proposed Order

1. Findings 7, 8 and 10 (fix list): small, safe and high signal.
2. Enforcement for findings 1 and 2 with the merge-base ratchet, so no new
   detections appear during the remaining work.
3. Finding 2 fixes, starting with `runtimeArtifactPaths` and the
   `(T, bool, error)` shapes.
4. Finding 1 fixes: parameter values for merge, ticket loop and Telegram
   senders.
5. Finding 4 for `ticket.go`, then finding 5 for `tui.Model`.
6. Finding 6.
7. Finding 9 per repository.
8. Finding 3 as its own delivery set with plan and tracker.

## Resolution

Record each fixed finding here with the commit, the validation command and the
result. Report the detection count before and after for findings 1 and 2.

| # | Status | Commit | Validation |
| --- | --- | --- | --- |
| 1 | open | — | — |
| 2 | open | — | — |
| 3 | open | — | — |
| 4 | open | — | — |
| 5 | open | — | — |
| 6 | open | — | — |
| 7 | open | — | — |
| 8 | open | — | — |
| 9 | open | — | — |
| 10 | open | — | — |

## Risks and Follow-ups

1. Metrics come from a scan of `28c3db33`. They are not a gate result and
   will drift until enforcement is active.
2. Test code was excluded; test quality was not reviewed.
3. Behavior, security and concurrency were not reviewed in this pass.
4. Finding 3 renames domain terms used in specs and plans. Renames need the
   maintainer's confirmation of each term's meaning.
5. Findings 1 and 2 enforcement and the result-shape check change lint
   configuration and add a repository tool. Both need approval before
   implementation.


## Remediation closure

The approved sixteen-slice remediation is integrated into dev through canonical
PRs #438 through #453. All 97 tasks are delivered. The controller full gate
passed on `2a4390aeb70d783d40a132fe6583d8dfb428fa7d`, receipt
`KSTGSZNOX5ESTLZXEPAAPJKJ6L`, in 102.538 seconds. The
[completed cycle report](../cycles/code-quality-remediation.md) records exact
commands and outcomes. The
[final source reconciliation](../../plan/code-quality-remediation-slice-16-findings.md)
links every original signature/secondary identity, structural finding and minor
item to actual current code, owning commits and validation. Supported compatibility,
imposed interfaces, intentionally distinct rules and measured heuristic exclusions
have individual evidence; none is unknown migration debt.

The original review body above remains a baseline snapshot. Closure changes its
status and adds this appendix; it does not rewrite original detections, counts,
maintainer decisions or proposed remedies. The detections and baseline inventory
remain unchanged. Later ledgers and the cycle report supply final dispositions.
