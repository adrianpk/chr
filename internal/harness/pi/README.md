# Pi adapter

Owns Pi 1.3.0 JSONL command correlation and session lifecycle through the shared
bounded process supervisor. A prompt acknowledgment or `agent_end` cannot
complete a call. Only the final completed assistant message from a settled run
is normalized; streaming deltas, reasoning and tool results are excluded.

The transport selects and verifies the exact provider/model and reasoning
level before prompting. Automatic retry and compaction are disabled so the
existing caller retains deadline and attempt authority. Reported intermediate
events never reset those limits. Missing usage remains unknown and observed
counters do not certify token enforcement. Discovery and generation assembly
belong to the subsequent tasks in this slice.

The wire contract is pinned to [Pi 1.3.0 RPC](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/rpc.md)
and its [event stream](https://github.com/earendil-works/pi/blob/v1.3.0/packages/coding-agent/docs/json.md).
