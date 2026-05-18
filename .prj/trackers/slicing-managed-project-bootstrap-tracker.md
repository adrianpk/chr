# Managed Project Bootstrap Slicing Tracker

Status: active

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Review existing delivered plans and implementation state.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [ ] Commit plan and tracker directly to `dev`.
- [ ] Decide whether the pending bootstrap safety pass should become an implementation slice.

## Slice Status

| Slice | Status | Notes |
| --- | --- | --- |
| CLI Bootstrap | Delivered | Root CLI command and command family structure exist. |
| Project-Local Initialization | Delivered | `chc init` creates `.chc/` in current directory or `--dir`. |
| Asset Profiles | Delivered | Embedded `standard`, `personal`, and `minimal` profiles exist. |
| Root Agents File | Delivered | Root `AGENTS.md` is materialized and existing root file is preserved. |
| Protocol Defaults | Delivered | Standard profile includes protocol files and playbooks. |
| Bootstrap Safety Pass | Pending | Candidate final implementation slice for this slicing. |

## Pending Slice Candidate: Bootstrap Safety Pass

Candidate tasks:

- [ ] Review current `chc init` CLI behavior against the acceptance criteria.
- [ ] Add CLI-boundary tests for dry-run, existing files, and invalid targets if missing.
- [ ] Improve invalid target or permission errors if current messages are unclear.
- [ ] Verify real embedded profile assets, not only synthetic test assets.
- [ ] Run `make verify`.
- [ ] Open one branch and one PR for the pending slice if implementation is approved.

## Validation Record

No implementation validation has been run for this tracker yet.

## Notes

- This tracker covers the slicing level, not a single implementation branch.
- Most candidate slices in this slicing were already delivered before this slicing map existed.
- The only expected remaining implementation branch is the bootstrap safety pass, unless review finds another concrete bootstrap gap.
