# Ticket-Native Workflow Assurance

Run from the repository root:

```sh
make qa-ticket-native-workflow
```

The target validates all 24 acceptance bindings and four fault groups before
running their named Go tests once. It deduplicates shared evidence, rejects
missing functions and conflicting source mappings, and permits only the owned
core and application packages. Each manifest and test source is limited to
1 MiB, each criterion to 16 test references, and the suite to 128 distinct tests.
Go test has a 20-minute package deadline; the runner has a 30-minute process
deadline, including compilation. Any invalid binding or failing test fails QA.
Process assurance requires Linux. The runner adopts orphaned fixture descendants
and reaps them after the checks finish, including workers in separate process
groups. Cleanup uses owned children and PID handles; leftover fixtures fail QA.

[Acceptance bindings](acceptance.json) cover A01–A24. [Fault bindings](faults.json)
compose restart, crash, lost-response and manual-edit scenarios with the actual
filesystem, owned subprocesses, linked Git worktrees and bare remotes. The
composed lost-response queue loses each publication response, then confirms
actual canonical delivery across restarts without repeating workers or PRs.
Its native publisher uses actual SSH receive-pack report loss, retaining an
unknown transport response independently of verified Git application. Native
receipt restart, dispatch crash, member fanout crash, immutable-evidence edits
and stale-owner observations reuse the existing focused regression tests.
Journal and finalization faults use owned child processes at the write boundary;
projection and upgrade scenarios retain conflicting operator edits.

Fixtures use temporary projects and owning supervisors; they do not use the
operator's development repository, production project data or live credentials.
The suite does not contact a forge or mutate its PRs. Controlled provider facts
exercise the existing adapter contract alongside actual native atomic Git
publication, original-source attribution, receipts and frontier continuation.
They do not establish live provider capabilities.

`make qa-native-integration` separately validates the designated disposable
Forgejo and GitHub repositories using host-owned credentials and certified
method/transport profiles. See the [fixture boundary](../../../ops/default/report/native-integration-fixtures.md)
and [provider certification](../../../ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md).

The candidate gate requires `make verify`, `go test -race ./...`,
`make qa-ticket-native-workflow`, `make qa-native-integration` and
`git diff --check`, in that order on one recorded clean candidate. The delivery
controller also requires its exact integrated gate before either delivery set
can close.
