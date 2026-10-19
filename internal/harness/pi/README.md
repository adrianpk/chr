# Pi adapter

Owns Pi 1.3.0 generation and JSONL command correlation through the shared bounded
process supervisor. A prompt acknowledgment or `agent_end` cannot complete a
call. Only a completed assistant answer from a settled run is normalized;
streaming deltas, reasoning and tool results are excluded. Failed exit, cleanup,
protocol or artifact IO discards output. Core retains schemas, validation,
publication, deadlines, operation identity and capacity.

Setup selects and verifies the exact provider/model and requested thinking
level. Automatic retry, compaction and cache warming are disabled. Intermediate
events do not reset host limits. Missing usage remains unknown; reported
counters never certify whole-call token enforcement.

Each native generation call uses a private temporary home, agent directory, credential copy,
settings and session state. Only the selected provider's static API-key entry
and credential-free model definition are copied. Model semantics retain their
frozen hash; credentials may rotate at the same explicit private reference.
OAuth generation, credential commands, environment expansion and credential-bearing model
settings fail with local preparation errors. No constructor or readiness check
spawns Pi. An owned version check requires 1.3.0 when a call starts.

Ordinary structured calls disable tools. Discovery callers explicitly request
read-only inspection: the adapter loads its embedded `readonly.mjs` and exposes
only `chr_inspect` with `list`, `find` and `read` actions. Shell, writes, MCP and
all discovered extensions, skills, templates, themes and context files remain
disabled. Inspection is confined to the project root and at most 16 exact
host-provided evidence files; selected credential/model files are excluded.
Opened descriptors are checked against that boundary. Files/results are bounded
to 64 KiB, scans to 4096 entries and one RPC run to 64 inspection calls.

Resource inputs must be private regular files of at most 64 KiB. Call artifacts
reject links, non-regular files and non-private modes before replacement. The
existing process owner reaps the owned group before private state is removed.

The wire contract is pinned to [Pi 1.3.0 RPC](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/rpc.md),
its [event stream](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/json.md)
and [explicit extensions](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/extensions.md).
The optional `CHR_TEST_PI_EXECUTABLE` fixture uses a prepared pinned binary with
an owned loopback model endpoint; it requires no paid provider or real key.

Worker transport has a separate explicit capability profile for the pinned
`read`, `bash`, `edit`, `write`, `grep`, `find` and `ls` builtins. Tool IDs and
names must match their registered calls; a settled run cannot have unfinished
tools. Its normalized output ceiling is 1 MiB. Generation keeps its existing
64 KiB ceiling and cannot acquire worker tools through its request mapping.

Worker resource preparation requires a private credential file containing only
the selected provider. OAuth additionally requires a dedicated private owner
directory. `WithWorkerCredential` excludes concurrent owners across processes,
creates a private call copy, and atomically publishes a valid refreshed entry
before releasing exclusion. Its action must reap Pi before returning. A valid
refresh survives a later call failure; malformed updates or an operator's
uncoordinated replacement cannot overwrite the retained source. Lock waits obey
the caller's context. This owner does not add model retries or capacity.

Selected Docker workers retain `WorkerRuntimeOptions` in their private host
command and freeze the image ID before admission. The host launcher removes
that argv from model context. It mounts only the execution checkout, Git
metadata, the admitted capacity files, selected profile resources and the
retained CHR bridge. The bridge and RPC process execute inside that container;
they receive private writable state and a replacement model environment.

The model process works from private state. Only the embedded `worker.mjs`
extension registers the pinned SDK's seven builtin implementations at the
host-selected checkout. This prevents project Pi settings from changing the
call. Project instructions arrive through explicit host context. Discovery,
MCP, retry, compaction and cache warming remain disabled.

The bridge holds shared references to the existing admitted slot and guard;
new admissions still require an exclusive slot. Host disconnection cancels the
RPC context, and adopted detached tools are stopped and reaped before capacity
or refreshed credentials are released. Normalized output remains distinct from
host task validation and publication. Protocol streams stay ephemeral; provider
failure details are redacted with the current credentials before excerpting.
