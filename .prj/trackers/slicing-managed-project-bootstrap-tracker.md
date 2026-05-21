# Managed Project Bootstrap Slicing Tracker

Status: completed

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Review existing delivered plans and implementation state.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Decide whether the pending bootstrap safety pass should become an implementation slice.

## Slice Status

| Slice | Status | Notes |
| --- | --- | --- |
| CLI Bootstrap | Delivered | Root CLI command and command family structure exist. |
| Project-Local Initialization | Delivered | `chc init` creates `.chc/` in current directory or `--dir`. |
| Asset Profiles | Delivered | Embedded `standard`, `personal`, and `minimal` profiles exist. |
| Root Agents File | Delivered | Root `AGENTS.md` is materialized and existing root file is preserved. |
| Protocol Defaults | Delivered | Standard profile includes protocol files and playbooks. |
| Bootstrap Safety Pass | Delivered | Repository PR #116 merged. |

## Pending Slice Candidate: Bootstrap Safety Pass

Candidate tasks:

- [x] Review current `chc init` CLI behavior against the acceptance criteria.
- [x] Add CLI-boundary tests for dry-run, existing files, and invalid targets if missing.
- [x] Improve invalid target or permission errors if current messages are unclear.
- [x] Verify real embedded profile assets, not only synthetic test assets.
- [x] Run `make verify`.
- [x] Open one branch and one PR for the pending slice if implementation is approved.

## Validation Record

- `go test ./internal/chcapp -run 'TestInit'`: pass.
- `make verify`: pass.

## Notes

- This tracker covers the slicing level, not a single implementation branch.
- Most candidate slices in this slicing were already delivered before this slicing map existed.
- The only expected remaining implementation branch is the bootstrap safety pass, unless review finds another concrete bootstrap gap.
