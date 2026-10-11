# Slice 1: Guarded Git transport

Status: reviewing
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Corrective Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-1-native-git-transport`
PR: pending

## Purpose

Publish the exact prepared result through one advertisement-aware atomic transaction guarding both the original source and target base.

## Delivered Behavior

The trusted Git port sends source S->P and target B->P as two explicit full-ref commands, negotiating `atomic` and `report-status` after verifying the server's advertised values. Local receive-pack, host-authenticated SSH and smart HTTPS use the same exchange. A source or target already at P requires observation before commands; stale refs and missing atomic support cannot fall back to another push.

Publication requires an explicitly rewrite-authorized source in the same canonical repository, a sealed candidate revision, verified method/tree/history proof and independently checked B->P fast-forward ancestry. Results distinguish `not_sent`, `rejected`, `applied` and `unknown`, retaining original, advertised and published source identities separately. Transport application does not create a delivery receipt.

## Implementation Notes

Host Git generates a complete non-thin object pack and independently verifies the supplied prepared history. Process groups, cancellation and a three-minute deadline bound local and SSH IO. Advertisement and report data are limited to 1 MiB, individual packets to 65,520 bytes and packed objects to 64 MiB.

SSH uses argument vectors and one quoted remote repository argument. Host keys, keys and agents remain owned by SSH. HTTPS requires TLS, rejects redirects, checks smart service and response types, and uses the host Git credential helper after an authentication challenge. Credentials and subprocess stderr are never emitted as publication evidence.

Protocol safety evidence follows the [Git pack protocol](https://git-scm.com/docs/pack-protocol), [smart HTTP protocol](https://git-scm.com/docs/http-protocol) and [Git credential interface](https://git-scm.com/docs/git-credential). Tests capture both commands at a real receive-pack hook, preserve external writers after advertisement, exercise server policy rejection and prove the standard CLI no-op lease trap. Malformed frames, duplicate capabilities/refs, oversized advertisements and incomplete/mixed reports cannot become success. Real SSH and TLS fixtures truncate replies after both refs have actually been updated; the result remains unknown.

## Contracts Added or Changed

`WorkflowGitPublisher` is a narrow host publication port. `WorkflowGitPublicationRevision` binds canonical repository and source ownership, full refs, original source, base, result and existing method/history proof. The result uses `chr-git-publication-result/v1` and makes uncertainty explicit without retrying publication.

Coverage: AC-036 through AC-043, AC-051 through AC-055, and AC-096 through AC-098; INV-07 through INV-10, INV-14 and INV-15. Positive transport proof covers merge, rebase and squash objects. Admission to a live forge profile and canonical PR merge observation remain owned by Slice 2.

## Files of Interest

- `internal/core/workflow_git_publication.go`: publication port and atomic exchange.
- `internal/core/workflow_git_publication_proof.go`: immutable ownership and history binding.
- `internal/core/workflow_git_protocol.go`: bounded packet, advertisement and report parsing.
- `internal/core/workflow_git_transport.go`, `workflow_git_ssh.go`, `workflow_git_http.go`, `workflow_git_credential.go`: host transport, lifecycle and authentication.
- `internal/core/workflow_git_contention_test.go`, `workflow_git_response_test.go`: real-process concurrency and lost-response evidence.

## Validation

Passed during task implementation with Go 1.27.1:

- `go test ./internal/core -run 'TestWorkflowGit(Atomic|Advertisement)$' -count=1`
- `go test ./internal/core -run 'TestWorkflow(GitSSH|SSHArguments)$' -count=1`
- `go test ./internal/core -run 'TestWorkflow(GitHTTPS|HTTPIdentity)$' -count=1`
- `go test ./internal/core -run 'TestWorkflowGit(Atomic|Candidate|AlreadyPrepared|FastForward)$' -count=1`
- `go test ./internal/core -run 'TestWorkflowGit' -count=1`

The configured final candidate commands are `make verify`, `go test ./internal/core ./internal/app -count=1`, `go test -race ./internal/core ./internal/app -count=1`, and `git diff --check`, in that order. Looper records their results against the exact clean PR HEAD after the PR-reference commit. Task checks are local transport evidence; live provider certification is not claimed here.

## Risks and Follow-ups

The owned SSH fixture requires OpenSSH client/server tools. Repositories exceeding the explicit transfer bounds are rejected before commands. The publisher exposes uncertain outcome; durable reconciliation and retry admission are delivered by the later approved recovery slices. Slice 2 still requires designated disposable Forgejo and GitHub repositories and policy/access profiles.
