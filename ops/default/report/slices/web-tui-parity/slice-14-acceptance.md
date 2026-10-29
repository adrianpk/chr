# Slice 14: Parity acceptance and capability inventory

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `test/slice-14-tui-acceptance`
PR: [#468](https://forge.adrianpk.com/adrianpk/chr/pulls/468)
Integrated commit: `d94d5af3dedad35b430f6fc3952b5efce88c75c4` on dev.

## Purpose

Reconcile all 38 gaps and four incidental findings with delivered assertions,
gate unsupported palette writes, and verify the integrated terminal/Web journeys.

## Delivered Behavior

- The coverage ledger preserves every original finding, acceptance statement and
  static/runtime assessment. Its 42 delivered rows now link immutable delivered
  reports, canonical merges, successful controller evidence and exact assertion
  families. I-003 is closed with this slice's verified integration.
- Both checklist palette entries visibly explain unavailability. Selecting them
  opens blocked feedback before confirmation, even without a selected ticket;
  Enter dispatches no writer. Shared confirmation construction cannot advertise
  these unsupported actions. Existing source editing, checklist display and all
  other supported ticket operations remain available.
- `TestTUIWebParity` selects 14 owned scenarios for manual forms, retained
  workflows, legacy access and the wizard. A fresh production terminal now goes
  from Name/Create and exact capture through clarification, native review and
  approval, real local Git publication, verified results, JSON and restart.
- Terminal case editing preserves focus and complete text through narrow/wide
  resize. Chromium observes a peer's saved idea while an unsent answer remains
  focused; resize, continued typing and declined cancellation preserve it.
  Canonical revisions, question binding and document images prove no hidden
  submission or progression. The offered reload remains an explicit choice.
- The User Guide and reference describe pre-confirmation unavailability. The
  changelog consolidates the delivered TUI capabilities by user task, retaining
  unrelated entries. The cycle report records all 14 integrations and acceptance
  evidence; its separate final controller gate remains pending.

## Implementation Notes

T14.1: `6c7920ddce29df8a6ede341c84cc81e7294d3d74`.
T14.2: `80b1dcdcf47c65a58e53747a1561f6d30ac04816`.
T14.3: `e941712f0c977722a961ff4f68991927a371dff5`.
T14.4 report introduction: `c2a07937d9c2a9ac23b2de155989df450a072b8c`.
T14.5 delivery metadata: `45ef4cb4f8156d67c59f67c3f1b948488450705d`.
PR-reference follow-up: `d94d5af3dedad35b430f6fc3952b5efce88c75c4`.

The new family composes existing production fixtures rather than substituting
reducers for persistence or terminal IO. The fresh wizard child uses the root
loader, managed host, existing deterministic model/forge response boundaries,
real SQLite and a temporary bare Git remote. It checks exact capture, four
checkpoints, finite native membership/policy, one worker attempt, committed
originals, verified completion and unchanged completed-run revision on restart.

Fixture corrections follow the approved interaction: bind generation through the
same test provider boundary as existing fixtures; use Continue when the retained
stage advances without redirecting the selected tab; inspect native approval
before its separate exact confirmation; maximize the detail pane to read complete
wrapped evidence. Browser typing explicitly focuses the real textarea before
CDP input. These corrections preserve the observable assertions and original
40-second child, 45-second terminal fixture and two-minute browser limits.

The case form's original description/context assertions remain, with an exact
additional `AFTER_RESIZE` marker proving typing stayed in the focused Context
field. Existing cancellation, stale/reload, external-editor, retry identity,
unsaved-unload, native review and delivery assertions remain enabled.

## Contracts Added or Changed

No domain writer, schema, dependency, integration policy or resource allowance
changes. Unsupported checklist presentation now stops before confirmation.
Acceptance adds one named selection family and extends owned IO evidence.

## Files of Interest

- `ops/default/plan/web-tui-parity-coverage.csv`: all 42 source dispositions.
- `internal/tui/ticket_actions.go`, `command_palette.go`,
  `palette_capabilities_test.go`: pre-confirmation checklist guard.
- `internal/app/tui_web_parity_linux_test.go`: named 14-scenario selection.
- `internal/app/tui_parity_wizard_terminal_linux_test.go`: complete production
  terminal delivery and read-only completed restart.
- `internal/app/tui_forms_terminal_linux_test.go`,
  `web_tui_parity_browser_test.go`: actual focused input through resize/polling.
- `ops/default/report/web-tui-parity-cycle.md`: integrated delivery accounting.

## Validation

Passed task checks with local owned fixtures:

- `go test ./internal/tui -run '^Test(PaletteChecklist|BoardMoveCommands|ModelPlansTicketActionCommand)$' -count=1`
- `go test ./internal/app -run '^Test(TUIParityCapabilities|TUINavigationTerminal|TUIQuizTerminal)$' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestTUIWebParity$/(case_and_dump_forms|wizard_verified_delivery)$' -count=1 -timeout=2m`
- `CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^TestGuidedUseCaseWebBrowser$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — zero issues.
- `git diff --check`

Controller validation `4OVYDKWV4PUPDT57DT7IE3UMT2` passed all five checks on
clean PR-reference HEAD `d94d5af3dedad35b430f6fc3952b5efce88c75c4`, using this
ordered command list:

- `make verify`
- `go test -race ./internal/tui -count=1`
- `go test ./internal/app -run '^Test(GuidedUseCaseParityNative|GuidedUseCaseTUITerminal|TUIWebParity)$' -count=1 -timeout=5m -v`
- `CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^Test(GuidedUseCaseWebBrowser|WebQuizBrowser|WebSpecBrowser|WebDumpEditorBrowser|WebBoardBrowser|WebDashBrowser)$' -count=1 -timeout=5m -v`
- `git diff --check`

The retained verbose log confirms three application roots, all 14 named parity
scenarios and six Chromium roots ran and passed without skips. The controller
ledger binds these results to the exact clean head and command order. Acta note
`01M64FNVNGK034565BYWC0ZQ1N` records validation; merge event
`01M64FPHBRCSXTN5Q8YG2K1ZS6` and canonical PR #468 confirm the same head was
integrated into dev. The controller's separate final delivery-set gate remains
pending; these results do not establish a pass for later documentation commits.

Runtime uncertainties are cross-checked by these actual boundaries:

| Finding | Terminal evidence | Web reference |
| --- | --- | --- |
| G-008 | Exact capture and retained replay; fresh full wizard PTY; existing restarted journey | Chromium finish-to-Quiz submission and original inputs |
| G-013 | Production Pause, declined/confirmed cancellation, exact retained originals/checkpoints/worker commit | Chromium cancellation warning and unchanged unsent answer/run |
| G-014 | Production polling observes peer ideas while typing and reading an earlier tab | Real polling feedback, retained textarea/focus/view, same canonical revision/question after peer write |
| G-018 | Actual concurrent Dump Save, retained draft and confirmed reload | Chromium lost acknowledgement, stale Save and explicit reload |
| G-024 | Manual/generated/project-wide canonical Spec catalog and production editor | Chromium manual/generated catalog and revision-checked editing |

The preserved capability inventory remains reachable and covered:

| Capability | Entry and assertions |
| --- | --- |
| Project metadata, Harness defaults, unresolved references and artifact counts | `alt+p`, Cases; harness read/save/refresh tests, workspace graph/count/diagnostic tests |
| Workspace-index health and runtime information | Palette Runtime; `TestTUINavigationTerminal`, `TestModelShowsWorkspaceIndexHealthInDash`, Agent observability tests |
| Seventeen ticket filter fields | Board filter; `TestTicketFilterMatchesMetadata` |
| External editors, source inspection and checklist display | Dump `alt+e`, ticket source; forms/ticket PTYs and selected-source/checklist rendering tests |
| Acceptance-feature generation and safe readiness repair | Specs and Dash/palette; acceptance-feature adapter tests and readiness repair/control tests |
| Queue filters/order and Run all ready | Dash; shared queue/paging, filter and Run all ready binding tests |
| Running-work events and log locations | Dash/Agent; activity, independent running-work and observability render tests |
| Pane focus, splits, footer clicks, resize and maximize | Shared shell; mouse/split/footer tests and production resized form/Board detail |
| Later ideas, originals and JSON | Existing wizard secondary actions; polling, cancellation, original-byte and verified full-result PTYs |

## Risks and Follow-ups

None recorded.
