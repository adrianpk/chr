"""Owned protocol peer for generation integration tests; performs no network IO."""

import json
import os
import re
from pathlib import Path
import sys
import time
import uuid

fixture = Path(sys.argv[1])
arguments = sys.argv[2:]
if arguments == ["--version"]:
    (fixture / "version-called").touch()
    version = fixture / "version"
    print(version.read_text() if version.exists() else "1.3.0", flush=True)
    raise SystemExit(0)


def argument(name):
    return arguments[arguments.index(name) + 1]


provider = argument("--provider")
model = argument("--model")
thinking = argument("--thinking") if "--thinking" in arguments else "off"
state = Path(os.environ["PI_CODING_AGENT_DIR"])
credentials = json.loads((state / "auth.json").read_text())
assert list(credentials) == [provider]
assert credentials[provider]["type"] == "api_key"


def emit(value):
    sys.stdout.write(json.dumps(value, ensure_ascii=False) + "\n")
    sys.stdout.flush()


def prompt(command):
    directory = fixture / "calls"
    directory.mkdir(exist_ok=True)
    call = {
        "model": model,
        "provider": provider,
        "thinking": thinking,
        "prompt": command["message"],
        "argv": arguments,
        "agent_dir": str(state),
        "cwd": os.getcwd(),
        "environment": sorted(os.environ),
        "home": os.environ["HOME"],
        "auth_providers": list(credentials),
    }
    (directory / (uuid.uuid4().hex + ".json")).write_text(json.dumps(call))
    (fixture / "started").touch()
    deadline = time.monotonic() + 10
    while (fixture / "block").exists() and not (fixture / "release").exists():
        if time.monotonic() >= deadline:
            raise RuntimeError("owned fixture release deadline exceeded")
        time.sleep(0.005)
    failure = fixture / "fail-next"
    failed = failure.exists()
    if failed:
        failure.unlink()
    output = (fixture / "result.json").read_text()
    if (fixture / "plan-readiness").exists() and '"observation_hash"' in command["message"]:
        match = re.search(r'"observation_hash"\s*:\s*"([a-f0-9]+)"', command["message"])
        if match:
            output = json.dumps({"sufficient": True, "missing": [], "observation_hash": match.group(1)})
    emit({"type": "agent_start"})
    emit({"type": "message_end", "message": {"role": "user", "content": command["message"]}})
    message = {
        "role": "assistant",
        "provider": provider,
        "model": model,
        "stopReason": "error" if failed else "stop",
        "content": [{"type": "text", "text": output}],
        "usage": {"input": 17, "output": 8, "cacheRead": 0, "cacheWrite": 0},
    }
    emit({"type": "message_end", "message": message})
    emit({"type": "agent_end", "willRetry": False})
    emit({"type": "agent_settled", "aborted": False})


for line in sys.stdin:
    command = json.loads(line)
    kind = command["type"]
    data = None
    if kind == "set_model":
        assert command["provider"] == provider and command["modelId"] == model
        data = {"provider": provider, "id": model}
    elif kind in ("set_auto_retry", "set_auto_compaction"):
        assert command["enabled"] is False
    elif kind == "set_thinking_level":
        thinking = command["level"]
    elif kind == "get_state":
        data = {
            "model": {"provider": provider, "id": model},
            "thinkingLevel": thinking,
            "isStreaming": False,
            "isCompacting": False,
            "autoCompactionEnabled": False,
            "messageCount": 0,
            "pendingMessageCount": 0,
        }
    elif kind == "prompt":
        data = {"disposition": "started"}
    else:
        raise RuntimeError("unexpected host command")
    emit({"id": command["id"], "type": "response", "command": kind, "success": True, "data": data})
    if kind == "prompt":
        prompt(command)
