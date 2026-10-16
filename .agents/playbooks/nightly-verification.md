# Nightly Verification

## Schedule and Source

The Forgejo workflow `.forgejo/workflows/nightly.yml` reserves 00:17 in
`Europe/Warsaw` every day. It checks out `dev` once and records its exact commit
in `.build/nightly/result.json`. It never advances `main`, creates tags or
publishes a release.

The selected slot precedes the existing 04:17 nightly of Odra by
four hours. Admission is limited to 00:17-00:37. A delayed job skips its long
checks rather than running during the next slot. The exhaustive command budget is 200 minutes, leaving time to upload evidence.
The job deadline is 210 minutes, leaving at least ten minutes before 04:17 even at the latest admission.
An earlier active run or earlier attempt on the same local date also prevents
another automatic attempt. Admission failure does not authorize a local rerun.

Concurrency groups are repository-scoped. The staggered slots and bounded
start window protect this pair of scheduled jobs without relying on an
unverified shared runner capacity. Other heavy schedules must reserve their
own non-overlapping slots. Runner or instance timeout limits may be shorter
than the workflow deadline and must be checked during activation.

## Runner Prerequisites

Use an existing trusted Linux runner with the `arch` label, Go compatible with
`go.mod`, a C compiler for race instrumentation, Git, Make and Python 3.9 or
later, plus Chromium for the browser workflow fixture. The application tests use the standard library and existing module
dependencies; no additional test framework is required.

The scheduled workflow enables live certification. The runner account also
needs the previously authorized private QA fixtures, `tea`, `gh`, SSH access,
scoped authentication and a working `acta` activity recorder. Keep the fixture descriptor at
`$XDG_STATE_HOME/chr/qa/native-integration-fixtures.json`, or under the runner's
`~/.local/state/chr/qa/` when the variable is unset. Never commit credentials,
copy a developer's entire home directory, or upload the fixture descriptor.
Missing live prerequisites fail before expensive test execution and cannot
produce a complete green result.

When dedicated runner authentication exists under
`${XDG_CONFIG_HOME:-$HOME/.config}/chr/nightly/`, the live pass uses that isolated
tool configuration. Tea stores its fixture login below `tea/`; the fine-grained
GitHub QA token is stored in `github.token` with mode 0600 and supplied only to
the live child environment. Default runner logins remain separate. Forgejo's
automatic `GITHUB_TOKEN` alias is removed when it equals `FORGEJO_TOKEN`, so it
cannot replace GitHub authentication. Explicit distinct GitHub tokens remain
available. Admission requests identify the JSON API client explicitly.

`make nightly` without `CHR_NATIVE_INTEGRATION_QA=1` runs complete local normal
and race suites only. Its result explicitly records `live_enabled: false` and
does not certify providers. This diagnostic mode does not replace the complete
scheduled gate.

## Execution and Evidence

The nightly performs these checks in order and stops on failure:

1. Validate A01-A24 and the four fault-group source bindings.
2. Strict lint, CI script regressions and application build.
3. Uncached complete local tests, with JSON events.
4. Uncached complete race tests, with live certification disabled.
5. Browser workflow actions, once, with their JSON events.
6. Explicitly enabled live tests for the designated fixtures, once.
7. Whitespace check and unchanged candidate verification.

The acceptance command is not run again: every manifest-bound test is already
included in the complete suite. Race instrumentation is a separate validation
mode, not a third copy of normal acceptance checks.

`.build/nightly/` contains the exact-source result, each command's exit status
and elapsed time, raw test events, and sorted per-test timings. Forgejo retains
this directory as an artifact even after failure. The existing workflow QA
subreaper owns and cleans up its fixture descendants; leaked processes fail
the gate. No other controller or developer processes are signaled.

For ordinary development and controller configurations follow
`.agents/LINTING.md`. Keep historical reports and frozen running-controller
commands unchanged. The first scheduled JSON reports provide the full measured
inventory; package-level historical timings do not establish individual test
costs.

## Activation

Forgejo discovers scheduled workflows only from the repository's default
branch. This repository currently uses `main`. Preparing this workflow on
`dev` alone does not activate the schedule. Publication to `main` must follow
the explicit authorization boundary in `.agents/VERSIONING.md`; adding a
nightly does not authorize promoting application history or changing the
default branch.

Before activation, verify the runner prerequisites and the other reserved
nightly slots. Do not dispatch an exhaustive run during development merely to
check the YAML. Validate admission and failure propagation with:

```sh
python3 -m unittest discover -s scripts/qa -p 'test_nightly.py'
make verify
```
