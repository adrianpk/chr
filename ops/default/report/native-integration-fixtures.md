# Native Integration Certification Fixtures

Status: designated; access verified; publisher certification pending

Date: 2026-10-11

Delivery set: native-integration-queue

Plan: [Native Integration Corrective Delivery](../plan/native-integration-queue.md)

Tracker: [Native Integration Tracker](../tracker/native-integration-queue.md)

## Designation and Authority

The maintainer authorized creating and configuring both private disposable
repositories and using synthetic pushes, pull requests and merges for the
approved certification. The maintainer created the empty Forgejo repository;
the agent made it private and initialized its README on `dev`. The agent
created and configured the private GitHub repository.

| Forge | Designated repository | Target ref | Initial target |
| --- | --- | --- | --- |
| Forgejo | [adrianpk/chr-integration-qa](https://forge.adrianpk.com/adrianpk/chr-integration-qa) | `refs/heads/dev` | `9617688a200b517cb7396e3a7387a1629c67be74` |
| GitHub | [adrianpk/chr-integration-qa](https://github.com/adrianpk/chr-integration-qa) | `refs/heads/dev` | `e3be8f8731fb804bff700acecdf8d3cb7e8073b1` |

These identities are required execution inputs. Use these repositories only.
The concrete host-owned descriptor is
`${XDG_STATE_HOME:-$HOME/.local/state}/chr/qa/native-integration-fixtures.json`.
It records SSH remote locations, existing credential references, roles, refs,
policy and access evidence. It contains no credential values and is a delivery
fixture descriptor, not a product configuration schema.

## Baseline Profile

- Both repositories are private, owned by `adrianpk`, with default branch `dev`.
- The existing host role has repository admin and push access on both forges.
  Forgejo API access uses the existing `tea` login `forge`; GitHub uses existing
  `gh` authentication. SSH uses existing host configuration and agent access.
- Tests own source refs under `refs/heads/chr/work/`. Independent cases may use
  target refs under `refs/heads/chr/qa/`; the baseline target is `refs/heads/dev`.
- The baseline target has no native branch protection, required native
  approvals or signing rule. CHR must validate and review the exact prepared
  candidate before its explicitly authorized fast-forward target publication.
- All three methods, merge, rebase and squash, are enabled. Automatic source
  deletion is disabled to retain certification evidence.
- Rewriting the exact CHR-owned test source branch is authorized only against
  its expected prior SHA. Preserve original source and prepared-result evidence.
- Denial profiles and contention cases must stay within these synthetic
  repositories. Account permissions and production branch policies remain
  outside this authority. Retain case refs until evidence is complete.

## Verified Setup Evidence

- Authenticated API metadata confirms repository identity, privacy, default
  branch, enabled methods and admin/push roles on both forges.
- Branch API metadata confirms both baseline targets are unprotected.
- Forgejo accepted the README commit and an SSH push to `dev`; the API and SSH
  ref read return the initial target above. The README is readable through the
  authenticated API.
- GitHub accepted creation of `dev` and repository settings through its API;
  authenticated HTTPS and SSH ref reads return the initial target above.

This satisfies the Slice 2 repository designation and setup-access prerequisite.
It does not certify guarded atomic publication or canonical PR merged-state
observation. Slice 2 must prove default rebase behavior on each forge, and Slice
12 must produce the complete method and denial matrix. Recheck current roles,
refs and policy before each live case. Stop on access drift or unsupported
behavior rather than substituting mocks or weakening the accepted specification.
