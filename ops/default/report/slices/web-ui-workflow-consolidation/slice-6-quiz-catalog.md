# Slice 6: Scope-driven Quiz Catalog

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-6-quiz-catalog`
PR: `#403`

## Purpose

Deliver retained-answer catalog membership, named pending scopes and focused
selection under the approved [plan](../../../plan/web-ui-workflow-consolidation.md)
and [tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 5 was
verified integrated through PR #402 at
`303db92f10651f5eb36d730e25a742ff78d06cd6`.

## Delivered Behavior

Project Quiz always appears first, with No answers yet when empty. Existing
named cases follow in name order only when they have confirmed answered entries.
Partial, cancelled and completed sessions retain their answers. Skips, curated
file existence and runtime counters cannot create answered membership.

New Quiz lists only the project and existing named cases without answers. An
active choice opens its retained question; a cancelled choice offers Resume.
Absent scopes offer Start Quiz. A completed empty scope remains pending and
requires explicit Start new session, retaining previous entries. With every
scope answered the action disappears; adding a case restores that destination.
Catalog and focused GETs never start or replace a session.

The centered catalog uses at most 860 CSS px with 20px vertical and 16px
horizontal padding. Cards stay 268 × 178px, with 100px pale-green opening-content
previews, six clipped lines and no preview scrollbar. Three fixed desktop
tracks reduce to two below 860px and one below 590px. The 232px pending menu
supports native keyboard activation, Tab and Escape with focus return. Read
polling cannot replace an open menu or a focused catalog selection.

## Implementation Notes

- T6.1: `836b920dd8dcc805ea8bfd899594da309ba519c8` — shared read-only membership through the existing bounded quiz repository.
- T6.2: `14cd4d91361ab726e44c5bf30f3e56f9fb50ca9c` — dedicated Web projection, templates, scoped styling, pending navigation, explicit restart and owned browser evidence.

The shared projection observes at most the existing 4096-scope/record ceiling
and existing 1 MiB image bounds. It retains canonical scope/reference, observed
session and revision, derives membership from parsed validated entries, and
limits previews to 1024 runes before rendering. Focused actions re-observe and
use the existing session/question/revision/operation guard. Existing CLI/TUI
scope inventories and the shared generation policy continue through their owners.

## Contracts Added or Changed

`ReadQuizCatalog` separates answered cards from eligible unanswered destinations.
The Web projection includes named pending links instead of detached slug forms.
The focused view explicitly exposes Start new session after Done, preserving
raw entries and requiring the existing bound Start operation. Selecting a name
is read-only; mutations remain explicit focused actions.

## Files of Interest

- `internal/app/service_quiz_catalog.go`: shared membership and bounded preview.
- `internal/app/web_artifacts.go`, `web_quiz.go` and catalog HTTP/browser regressions: mapping, continuation and restart.
- `internal/webui/assets/templates/quiz_catalog.html`, `quiz.html`, `shell.html`, `assets/static/chr.css` and `quiz.js`: card/menu composition and focus behavior.
- `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: delivered usage.

## Validation

Successful development checks:

- `go test ./internal/app -run '^TestQuizCatalog' -count=1` — absent, active, skipped, partial, cancelled, completed, counter-only and curated-only states; retained answers without runtime; all-answered and new-case membership.
- `go test ./internal/app -run '^TestWebQuizCatalog' -count=1` — real HTTP selection without mutation, native session identity, explicit restart with unchanged raw entries, card anatomy and all-answered fragment.
- `go test ./internal/app ./internal/webui -run '(QuizCatalog|WebQuiz(StartResume|ReadOnly)|HandlerArtifactCatalogAndDetailRoutes|HandlerRendersGenerationAndWorkActionForms)' -count=1` — shared read model and affected HTTP/template checks passed.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_QUIZ_CATALOG_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-6" go test ./internal/app -run '^(TestWebQuizCatalogBrowser|TestWebQuizBrowser)$' -count=1 -v` — catalog journey, captures and focused answering regression passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app ./internal/webui` — zero issues.
- `git diff --check` — passed.

The browser uses real HTTP, rendering, filesystem state and owned Chromium;
only question generation uses the deterministic fixture model. It verifies
menu keyboard/pointer entry, selection preserving an unanswered session,
answer/cancel navigation, all-answered removal, adding a case, fixed geometry
across short/long/unbroken content and no horizontal page overflow. Each fixture
stops only its own HTTP server and Chromium after checks finish.

Final clean-HEAD `make verify`, focused quiz HTTP/service checks, catalog and
answering browser checks, and `git diff --check` are recorded through the
controller after the PR-reference follow-up. The controller's check record
binds the exact HEAD and ordered commands.

## Captured Views

- [Desktop closed](files/slice-6/desktop-closed.png): 1440 × 1000 CSS px, three tracks, empty project and short/long/unbroken answered cards.
- [Desktop open](files/slice-6/desktop-open.png): same viewport, 232px menu with Project Quiz and Pending named work.
- [All answered](files/slice-6/all-answered.png): same viewport, five cards and no New Quiz action.
- [Two columns](files/slice-6/width-768.png): 768 × 844 CSS px, menu open.
- [Narrow](files/slice-6/width-390.png): 390 × 844 CSS px, one column and menu within the viewport.
- [Small narrow](files/slice-6/width-320.png): 320 × 844 CSS px, same fixed card geometry.

All captures use 100% browser zoom, page scale 1 and DPR 1. Adjacent JSON files
retain CSS viewport, page width, container/card/preview/menu rectangles and track
counts. The 390px capture has a scrollbar-reduced document width of 380px.

## Risks and Follow-ups

None recorded.
