# Select Worker and Generation Harnesses

Use independently prepared profiles for worker execution and content generation.
Existing Codex-only projects need no Pi installation or configuration.

## Prepare private profiles

Create `harnesses.yml` in the [user configuration directory](../reference/user-environment.md).
Keep it and referenced resources private (directory mode 0700, file mode 0600).
Profiles contain host setup; the project stores only profile names.

A Codex profile can serve either role:

```yaml
profiles:
  code:
    harness: codex
    command: codex
    model: exact-model
```

Omit `model` to keep the CLI-managed default. Explicit Codex provider, thinking,
credential and model-resource options are not supported by this adapter; they
are rejected rather than ignored. Existing CLI authentication remains in use.

For native generation, prepare Pi **1.3.0** and check its version:

```sh
/opt/pi/bin/pi --version
```

Add a generation profile with an explicitly selected static API-key file:

```yaml
  writing:
    harness: pi
    command: /opt/pi/bin/pi
    model_provider: example-provider
    model: exact-model
    thinking: off
    credential_ref: pi-generation/auth.json
```

Use actual supported provider/model names. Edit the private credential file
locally; do not supply its contents through command arguments or project files.
Its selected provider entry has this shape:

```json
{"example-provider":{"type":"api_key","key":"replace-with-private-static-key"}}
```

Native generation does not support OAuth refresh or dynamic credential commands.
A worker may use OAuth through its separate refresh owner. Prepare its dedicated
credential file as described below, rather than reusing an unrelated Pi home.

## Prepare a Pi worker

Install missing sibling runtime assets in the project and explicitly build the
pinned image:

```sh
chr runtime init
docker build --file .chr/default/runtime/agent/pi/Dockerfile \
  --tag chr-agent-pi:1.3.0 .chr/default/runtime/agent/pi
```

From the selected CHR source revision, build the Linux bridge into a writable
installation directory:

```sh
CGO_ENABLED=0 go build -o /opt/chr/bin/chr ./cmd/chr
```

The worker profile selects that bridge; the generation profile selects native Pi:

```yaml
  development:
    harness: pi
    command: /opt/chr/bin/chr
    model_provider: example-provider
    model: exact-model
    thinking: off
    credential_ref: pi-worker/auth.json
    runtime_ref: chr-agent-pi:1.3.0
```

The worker auth file contains **only** the selected provider. For OAuth, its
private parent directory contains only that file and the managed refresh lock.
Keep model definitions outside the refresh directory. The refresh owner serializes
cross-process updates and atomically publishes complete rotated credentials.
Static-key files are mounted read-only. See [Pi runtime preparation](../reference/project-workspace.md#prepared-pi-worker-profile)
for image, mount and credential contracts.

For a custom endpoint, add `model_config_ref: pi-models.json` to the relevant
profile. A private definition can use this shape:

```json
{"providers":{"example-provider":{"api":"openai-completions","baseUrl":"http://model-gateway.example:8080/v1","models":[{"id":"exact-model","contextWindow":32768,"maxTokens":2048}]}}}
```

Use a container-reachable endpoint for workers. Literal loopback addresses are
rejected because they point inside the container. An explicit
`host.docker.internal` endpoint adds the host-gateway mapping, subject to the
host listener and firewall. Host networking is not enabled automatically.

## Select the registered project and defaults

Select the same registered project used by Web, TUI and Telegram:

```sh
chr project use example-project
chr harness status --json
```

Use this selected registry context for shared changes. `--dir` explicitly selects
a directory context with its directory-derived identity; it does not inherit a
registered project's ID. Keep `CHR_OPERATOR_ID` set to the authorized local
operator when changing defaults.

Copy the complete observed revision and choose a fresh request ID for the exact
change:

```sh
chr harness select --config-revision OBSERVED_REVISION \
  --request-id choose-worker-and-generation \
  --worker development --generation writing
```

Omit either role to leave it unchanged. Restore the generation action's legacy
dispatch without clearing the worker:

```sh
chr harness clear --config-revision OBSERVED_REVISION \
  --request-id restore-generation --generation
```

Use a new observed revision for each independent change. After a lost
acknowledgement, retry the **same** request, revision and choices to recover its
receipt. A stale request cannot silently adopt a newer image. Inspect current
defaults before deliberately submitting a new change.

The same operation is available through [Web Harnesses](../reference/project-workspace.md#web-harness-defaults),
[Project `h` in the TUI](../reference/project-workspace.md#tui-harness-defaults),
and [Telegram `/harness`](../reference/telegram-receiver.md#harness-defaults).
Both changed roles publish together. These controls select prepared names and
never provision credentials or start a model call. Other processes see the saved
image on their next operation without restarting the service.

## Verify and recover

`chr harness status` shows role-specific local readiness, configured model and
selection revision. Local readiness does not prove remote authentication, model
availability or the model's support for required tools/structured output.

| Observation | Required action |
| --- | --- |
| Unknown profile or missing executable | Correct the private profile and prepare its named executable; no fallback occurs. |
| Native Pi selected as worker | Select the static Linux CHR bridge and prepared image in a separate worker profile. |
| CHR bridge selected for generation | Select native Pi 1.3.0 in the generation profile. |
| Image unavailable or labels/platform rejected | Restore Docker access, explicitly build the pinned image for the host architecture and inspect status again. |
| Credential/resource rejected | Use a private regular file, static generation key, selected-provider-only worker auth and unchanged model definitions. |
| Provider rejects authentication/model or required capability | Resolve that exact provider/model setup privately; unsupported capability remains a failure. |
| Stale revision or changed authority | Inspect current settings and deliberately issue a new authorized request. |
| Interrupted call or lost response | Resume the original operation or retry its exact request; do not clear retained evidence. |
| Frozen executable/image/model definition unavailable | Restore the retained resources; editing defaults cannot repair or rebind that operation. |

Running, pending, paused, retried and resumed work retains its admitted binding.
Nested generation shares its owning operation's selection. A new independent
operation reads the new defaults. Compare the current default with the retained
invocation or `harness-run.json` selection; an explicit requested model and an
observed model are separate facts.

Observed input/output/cache counters remain unknown when absent. Pi observations
do not certify a hard whole-call token limit. Finite-token roles requiring a
certified driver reject an uncertified adapter, and Pi's unrestricted worker
cannot become a merger broker. Existing attempt, deadline, capacity, validation
and integration owners remain authoritative. One successful model does not
certify every provider/model combination.
