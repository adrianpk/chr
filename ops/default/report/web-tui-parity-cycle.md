# Web and TUI parity cycle

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../tracker/web-tui-parity.md)
Coverage: [All source IDs](../plan/web-tui-parity-coverage.csv)
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on dev.
Integrated candidate: pending Slice 14 integration.
Last verified integrated slice: PR #467 at `91df4d01d2838180430f9259eeab1569ed6618e5`.
Final delivery-set gate candidate: pending controller evidence.
Closure-only commit: pending; separate from the gate candidate.

## Delivered outcomes

The terminal opens on the seven case-oriented tabs. Project, Harness defaults,
contextual Help and retained auxiliary views remain accessible. Shared manual
forms cover cases, dumps, quizzes, specifications, plans and tickets, including
comments and review references. Input and functional scope survive rejected or
stale writes; replacing unsent drafts requires explicit confirmation.

Dash and Board share canonical work observations and six status destinations.
They retain independent visible selections, queue exclusions, paging, recent
activity and evidence. Work previews capture the exact scope and policy before
confirmation; retained workflow controls keep revision, review and budget guards.

Named wizards reuse the existing tabs, forms and owners. Exact original inputs,
checkpointed preparation, native work, review and confirmed delivery remain
separate stages. Results expose delivery receipts and validation commands/exit
codes. Cancellation keeps saved documents and completed changes. Read-only
polling preserves drafts, focus and navigation while explicit retries retain
operation identity. Unsupported checklist writes now explain their guard before
confirmation; no new writer is introduced.

## Integration evidence

Thirteen slices are integrated and closed. Slice 14 has implementation and owned
task evidence in its report; its PR, final exact-HEAD checks and integration are
recorded during the remaining handoff. This report does not claim cycle closure.

| Slice | Canonical PR | Integrated commit | Report | Successful validation |
| --- | --- | --- | --- | --- |
| 1 | [#455](https://forge.adrianpk.com/adrianpk/chr/pulls/455) | `161622891a376efe97e3e2d0b3bbebaaf8235660` | [Slice 1 report](slices/web-tui-parity/slice-1-safe-actions.md) | `YCBB7LWNNQR5OLKQDKXKLQ62HN` |
| 2 | [#456](https://forge.adrianpk.com/adrianpk/chr/pulls/456) | `322e8f0a7a7b17536f1cbe72b72e8023d2a4efb0` | [Slice 2 report](slices/web-tui-parity/slice-2-navigation.md) | `HCHOQPMEAG7XZXQ6XGFNLY3BFU` |
| 3 | [#457](https://forge.adrianpk.com/adrianpk/chr/pulls/457) | `c91ba4feac5b4296548a854e26b8d9cb59170ed7` | [Slice 3 report](slices/web-tui-parity/slice-3-feedback.md) | `7IDEK3RU26UI7CUQSTNJ7T3NAK` |
| 4 | [#458](https://forge.adrianpk.com/adrianpk/chr/pulls/458) | `c4451b766e3a5a92e3b9be4a4dfc1a952401b00e` | [Slice 4 report](slices/web-tui-parity/slice-4-dash.md) | `MQWCF6KW2NQXNMRL3FEXHGLGIH` |
| 5 | [#459](https://forge.adrianpk.com/adrianpk/chr/pulls/459) | `1bc34edaac9b751810f776efd5ea89535908760f` | [Slice 5 report](slices/web-tui-parity/slice-5-case-dump.md) | `YV7TF3QYUUICJ24UVUIWR7TTC3` |
| 6 | [#460](https://forge.adrianpk.com/adrianpk/chr/pulls/460) | `81e4331da3a711dd3dd8f1fb5e2bbd93083b85c3` | [Slice 6 report](slices/web-tui-parity/slice-6-specs.md) | `BZNEGXNC3AUIASKTADDVRE26ZI` |
| 7 | [#461](https://forge.adrianpk.com/adrianpk/chr/pulls/461) | `abc766b3e15ed14cd16d215ac823d1a16d25e582` | [Slice 7 report](slices/web-tui-parity/slice-7-quizzes.md) | `JRP44LHIFOTCAKMSHR3ELP542E` |
| 8 | [#462](https://forge.adrianpk.com/adrianpk/chr/pulls/462) | `59108deadad1848bc0ffbfda9def9fb97ea81f52` | [Slice 8 report](slices/web-tui-parity/slice-8-plans.md) | `32RLE7DHDBWJYZRIOXRECPQMSH` |
| 9 | [#463](https://forge.adrianpk.com/adrianpk/chr/pulls/463) | `be5cfd9ae8a1fa2899517feddcf1b87dc32ecea6` | [Slice 9 report](slices/web-tui-parity/slice-9-tickets.md) | `RNRFX3HTEIWFXQJHIKQSD4HET4` |
| 10 | [#464](https://forge.adrianpk.com/adrianpk/chr/pulls/464) | `728f5dd5bcc59f172ba63b9e62e386a084f86c2d` | [Slice 10 report](slices/web-tui-parity/slice-10-board-workflows.md) | `JQCPMJ55MJPOB64VPBLMV6N4Q3` |
| 11 | [#465](https://forge.adrianpk.com/adrianpk/chr/pulls/465) | `6ee42aac701f6b2e6f479cbd639601965b29602b` | [Slice 11 report](slices/web-tui-parity/slice-11-wizard-entry.md) | `COHC2KP7RZY5ONJTIC4UAQFWZJ` |
| 12 | [#466](https://forge.adrianpk.com/adrianpk/chr/pulls/466) | `a0e8ec258860a56fcc6d9d1c5e5fefee9eb39e8f` | [Slice 12 report](slices/web-tui-parity/slice-12-wizard-authoring.md) | `TKFQDKEODV6WVKUYFMQMOX3ZWP` |
| 13 | [#467](https://forge.adrianpk.com/adrianpk/chr/pulls/467) | `91df4d01d2838180430f9259eeab1569ed6618e5` | [Slice 13 report](slices/web-tui-parity/slice-13-wizard-results.md) | `BV6G4MNWIBE7TQXVHIDFMORIVQ` |
| 14 | pending | Integration pending | [Slice 14 report](slices/web-tui-parity/slice-14-acceptance.md) | Final PR-reference validation recorded by controller |

## Source coverage

The ledger retains exactly 38 G findings and four I findings. Baseline acceptance
text, evidence locations and the five runtime-unverified assessments are
unchanged. Forty-one rows have delivered owning slices and immutable report,
merge, command and assertion references. I-003 is implemented and reviewing in
Slice 14; its delivered disposition follows verified integration. No source item
is excluded, deferred or closed from static provenance alone.

The Slice 14 report cross-checks G-008, G-013, G-014, G-018 and G-024 against
production terminal and real HTTP/Chromium evidence. It also reconciles Project,
Harness, workspace-index health, 17-field filtering, source/external editing,
acceptance features, readiness repair, queue controls, runtime events/logs,
resizing/maximize and the existing wizard extras.

## Acceptance and fixture boundaries

Task checks passed for pre-confirmation unavailability, retained legacy access,
manual resize/input and a fresh wizard through real native delivery and restart.
Chromium polling, focused input, resize, declined cancellation and the existing
full guided journey also passed. Exact commands appear in the Slice 14 report.

Final clean-HEAD acceptance runs after the PR-reference follow-up commit through
the controller, in this order:

- `make verify`
- `go test -race ./internal/tui -count=1`
- `go test ./internal/app -run '^Test(GuidedUseCaseParityNative|GuidedUseCaseTUITerminal|TUIWebParity)$' -count=1 -timeout=5m -v`
- `CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^Test(GuidedUseCaseWebBrowser|WebQuizBrowser|WebSpecBrowser|WebDumpEditorBrowser|WebBoardBrowser|WebDashBrowser)$' -count=1 -timeout=5m -v`
- `git diff --check`

Verbose logs must show the three application roots, all 14 named parity
scenarios and all six browser roots executing without skips. The controller
ledger and Acta bind results to the exact clean PR head and ordered commands.
The controller's later final `make verify` evidence is recorded separately from
this slice batch and any documentation closure commit.

All mutations occur in owned temporary projects, stores, Git remotes and
providers. Terminal input uses actual Bubble Tea processes and PTYs; Web checks
use real HTTP and Chromium. Only model and external forge responses are
deterministic. Actual canonical writes, worker processes, Git publication,
receipts, revisions and read-only re-entry are asserted. Child/fixture/browser
limits and original safety assertions remain enabled; only owned children stop
after their checks finish. These focused checks do not claim exhaustive nightly
or live-forge certification.

## Residual limitations

No unresolved source finding or approved exception is recorded. Slice 14
integration, final controller gate evidence and documentation closure remain
pending. The separate broader Web feedback ticket remains outside this delivery
set: `TKT-20261023210422`.
