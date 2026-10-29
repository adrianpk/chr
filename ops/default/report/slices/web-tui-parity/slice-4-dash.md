# Slice 4: Shared queue observation and stable refresh

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-4-tui-dash`
PR: [#458](https://forge.adrianpk.com/adrianpk/chr/pulls/458)
Integration: Rebase + Fast-forward into dev at `c4451b766e3a5a92e3b9be4a4dfc1a952401b00e`.
Base: `545d08312b8897b25fe56287d1cea51aa6f93674` on dev, after verified PR #457 integration and Slice 3 closure

## Purpose

Resolve G-032, G-035, G-036 and G-037 through shared work observation,
independent selection, retained activity and exact preview confirmation.

## Delivered Behavior

- Dash uses the same ObserveWork queue eligibility, totals, readiness reasons,
  exclusions, current contexts, results and events as Web. Locally ready-looking
  tickets cannot override application policy. Priority and canonical order apply
  before queue pagination; brackets navigate pages of at most 128 rows.
- Board and Dash retain independent selected keys through refresh and reorder.
  Removed keys choose the first visible row; empty pages have no executable
  target. Queue paging retains the full independent Board catalog. Replies for
  an earlier filter or page cannot replace the current observation.
- Current and Recent panes expose independent work contexts, results and events
  without the running-work overlay. Native validation remains distinct from
  confirmed delivery and harness success. Up to 128 contexts, 128 events and 16
  results remain inspectable through scrolling and pane maximization.
- Native `v` Preview Work reads the selected extent, snapshot, target, policy
  and execution limits. Separate Enter confirmation permits one checkpoint;
  Esc cancels without admission. `w` Workflow details reads retained status
  without continuing it. Changed ticket revisions, scope, snapshot or policy
  require a new preview. Existing priority/order, Run all ready, safe repair,
  stop and running-work log/report details remain available.

## Implementation Notes

T4.1 is `0910c75b747f267e3c13ba981c61fc5aad73601e`. T4.2 is
`863d09bdbf8c8d1ba9075d12d4afea33b1290e76`, with the complete Board catalog preserved in
`63f66c326325c6e1cdd91939891bb1d09242fbda`. T4.3 is
`fed68882824878cb90735e3c381dd4203114f5ea`. T4.4 is
`f265da0e8acfc14bc67a5163d44dae9c318582a5`.
T4.2/T4.4 share the query-generation correction in `e242a9a67b20f4b3a7ba9c0409225ffef9b275d8`.

Initial regressions reproduced a Board selection changing after row reorder
and an older queue response installing after a filter change. Replacement now
reconciles the previous Board key against its own case, column and filter,
independently of the visible Dash queue. New queue queries disable execution
until their observation arrives; old replies release their own read and cannot
install obsolete rows. A monotonic query generation also rejects older reads
and previews when navigation returns to the same filter/order/page values.
Document reply bindings remain independent of unrelated
queue refreshes.

A policy-change regression also reproduced admission under a policy different
from the preview. Its initial fixture inherited the default worker; the owned
worker group was stopped and the fixture now pins a local deterministic command.
The passing regression requires a canonical conflict and unchanged operational
and Git bytes. Normal guarded execution is independently exercised by the
existing owned host-validation fixture, which confirms no early publication.
Read-only preflight rejects changed snapshots and policy before target hydration;
canonical admission rechecks the exact guard under its existing owner.

The independent Board catalog is read through canonical pages with one revision,
a 15-second aggregate deadline and the existing 4,096-record source bound.
A changed page revision rejects the read. This preserves access beyond the
first queue page without rebuilding the legacy index. Existing running-work
observability continues to provide log and report references.

## Contracts Added or Changed

`internal/app/work` owns the transport-neutral observation values and queue
request. Existing application names alias those values; production service
files do not import terminal types. ObserveWork retains its existing default
entry point, while ObserveWorkPage applies the shared filter and canonical
ordering before bounded pagination. Neither operation grants execution authority.

ExecuteConfirmedWorkExtent accepts the displayed native admission guard and
performs only a bounded checkpoint through the existing application and domain
owners. TUI requests retain the original request identity, snapshot, selected
revision, publication policy, execution policy and ready-ticket guard. Domain
persistence, admission rules and delivery evidence retain their existing owners.

## Files of Interest

- `internal/app/work/observation.go`, `service_work_observation*.go`: shared
  queue and observation contracts, source-owned readiness and pre-page selection.
- `internal/app/tui_ticket_snapshot.go`, `tui_adapters_agent_work.go`: bounded
  canonical Board reads and periodic work observations.
- `internal/tui/dash_*.go`, `ticket_refresh.go`, `agent_work_refresh.go`: queue
  paging, stable selection, activity, details and late-read rejection.
- `internal/app/service_workflow_confirmation.go`, `tui_adapters_workflow.go`:
  exact policy guard and canonical checkpoint execution.
- `internal/tui/dash_*_test.go`, `internal/app/tui_dash_*_test.go`: policy,
  pagination, independent selection and actual production-loader PTY regressions.
- User Guide, TUI reference, action matrix and changelog: delivered controls.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including existing catalog frames.
- `go test ./internal/app -run '^Test(TUIDash(Observation|QueuePages)|WorkObservation.*|WebDashQueuePages|TUINavigationTerminal)$' -count=1 -timeout=3m` — passed, including native/legacy adapter agreement and empty pages.
- `go test -race ./internal/tui -run '^Test(Dash.*|Model(RefreshesDash.*|SkipsDash.*|Workflow.*)|Operation.*|BusyQuit)$' -count=1` — passed.
- `go test -race ./internal/app -run '^Test(TUIDash.*|TUIWorkflow(Preview|Stale|Controls|Start|Approval|Refresh))$' -count=1 -timeout=3m` — passed without short mode.
- `go test ./internal/app -run '^TestTUIDashQueuePages$' -count=1 -timeout=3m` — passed with all 130 Board tickets retained on every queue page.
- `go test ./internal/app -run '^Test(TUIDash(NativeTerminal|PolicyBinding|ScopeBinding)|TUIWorkflowPreview)$' -count=1 -timeout=3m` — passed with visible limits and actual terminal controls.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app ./internal/app/work` — passed after formatting and complexity corrections.
- `git diff --check` — passed.

Production PTY fixtures use the root loader and managed host, actual Bubble Tea
input, owned Git/SQLite/HTTP state and the existing cleanup deadlines. Preview,
cancellation, retained detail and passive observation compare canonical bytes;
children are stopped and joined. Native policy/scope rejection also compares
Git bytes. No acceptance assertions or resource limits were weakened.

The initial controller batch `7KJGJSIR7AP5F2T5FB72U7D5NK` passed make verify
and focused terminal race checks before being interrupted for a reproduced
query-generation defect. New regressions reproduced acceptance of both an older observation
and an older preview after cycling back to the same order. Both pass after
binding reads and previews to the monotonic query generation. This earlier
batch is not final-candidate evidence.

Controller validation `MQWCF6KW2NQXNMRL3FEXHGLGIH` passed all four ordered checks
on clean HEAD `c4451b766e3a5a92e3b9be4a4dfc1a952401b00e`:

- `make verify`
- `go test -race ./internal/tui -run '^Test(Dash.*|Model(RefreshesDash.*|SkipsDash.*|Workflow.*|Ownership|AsyncRefresh)|Operation.*|BusyQuit)$' -count=1`
- `go test -race ./internal/app -run '^Test(TUIDash.*|TUIWorkflow(Preview|Stale|Controls|Start|Approval|Refresh)|WorkObservation.*|WebDash(Observation|Revision|QueuePages)|TUINavigationTerminal|TUIOperation.*)$' -count=1 -timeout=3m`
- `git diff --check`

Canonical Forgejo and Acta event `01M63TY93EHS3XRP9YDP8AYNPG` confirm
PR #458 integrated at the validated commit. No checks were repeated for closure.

## Risks and Follow-ups

None recorded.
