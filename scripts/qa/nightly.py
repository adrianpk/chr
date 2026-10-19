#!/usr/bin/env python3
"""Run the exhaustive gate once, retaining exact-source and per-test evidence."""

from datetime import datetime, timezone
import json
import os
from pathlib import Path
import runpy
import shutil
import subprocess
import sys
import time


ROOT = Path(__file__).resolve().parents[2]
QA = runpy.run_path(str(ROOT / "scripts/qa/ticket-native-workflow.py"))
LIVE_NAMES = ("TestNativeForgejoMethods", "TestNativeGitHubMethods", "TestNativeHTTPSMethods",
              "TestNativePolicyDenials", "TestNativePRChanges", "TestNativeTransportLoss",
              "TestWorkflowNativeRebase")
LIVE_TESTS = "^(" + "|".join(LIVE_NAMES) + ")$"
BROWSER_NAMES = ("TestWebWorkflowBrowser", "TestWebConsolidatedBrowser", "TestWebLoginBrowser")
BROWSER_TESTS = "^(" + "|".join(BROWSER_NAMES) + ")$"
BUDGET_SECONDS = 200 * 60


def commands(go, live):
    test = [go, "test", "-count=1", "-timeout=90m", "-json"]
    checks = [
        ("lint", ["make", "lint-strict"]),
        ("scripts", ["make", "test-ci"]),
        ("build", ["make", "build"]),
        ("normal", [*test, "./..."]),
        ("race", [*test, "-race", "./..."]),
        ("browser", [go, "test", "-count=1", "-timeout=5m", "-json",
                     "./internal/app", "-run", BROWSER_TESTS]),
    ]
    if live:
        checks.append(("live", [go, "test", "-count=1", "-timeout=30m", "-json",
                                "./internal/core", "-run", LIVE_TESTS]))
    checks.append(("whitespace", ["git", "diff", "--check"]))
    return checks


def test_timings(path):
    timings = []
    with path.open() as stream:
        for line in stream:
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if event.get("Test") and event.get("Action") in ("pass", "fail", "skip"):
                timings.append({key: event.get(key) for key in
                                ("Package", "Test", "Action", "Elapsed")})
    return sorted(timings, key=lambda row: row["Elapsed"] or 0, reverse=True)


def check_selected_tests(name, timings):
    if name not in ("live", "browser"):
        return
    expected = LIVE_NAMES if name == "live" else BROWSER_NAMES
    actual = {row["Test"]: row["Action"] for row in timings if "/" not in row["Test"]}
    if any(actual.get(test) != "pass" for test in expected):
        raise ValueError(f"{name} produced incomplete test evidence")


def source_commit():
    status = subprocess.check_output(["git", "status", "--porcelain"], cwd=ROOT, text=True)
    if status.strip():
        raise ValueError("nightly requires a clean candidate checkout")
    return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT,
                                   text=True).strip()


def browser_path():
    configured = os.environ.get("CHR_WEB_BROWSER")
    for name in ([configured] if configured else ["chromium", "google-chrome", "google-chrome-stable"]):
        resolved = shutil.which(name)
        if resolved:
            return resolved
    raise ValueError("nightly browser checks require Chromium in the runner account")


def preflight(live):
    QA["scenario_names"]()
    if live:
        state = Path(os.environ.get("XDG_STATE_HOME") or
                     str(Path.home() / ".local/state"))
        profile = state / "chr/qa/native-integration-fixtures.json"
        if not profile.is_file():
            raise ValueError("live nightly requires the authorized fixture descriptor in the runner account")
        for tool in ("tea", "gh", "ssh", "acta"):
            if shutil.which(tool) is None:
                raise ValueError(f"live nightly requires {tool} in the runner account")
    browser_path()


def run(output, go="go", live=False, runner=None):
    output.mkdir(parents=True, exist_ok=True)
    evidence = {"version": 1, "commit": None, "ref": "dev", "live_enabled": live,
                "started_at": datetime.now(timezone.utc).isoformat(),
                "status": "running", "checks": []}
    runner = runner or QA["run_scenarios"]
    deadline = time.monotonic() + BUDGET_SECONDS
    try:
        head = source_commit()
        evidence["commit"] = head
        preflight(live)
        for name, command in commands(go, live):
            env = dict(os.environ, CHR_NATIVE_INTEGRATION_QA="1" if name == "live" else "0",
                       GOFLAGS="", GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL="/dev/null",
                       CHR_WEB_BROWSER=browser_path() if name == "browser" else "")
            # Forgejo aliases its repository token as GITHUB_TOKEN; it is not GitHub authentication.
            if env.get("FORGEJO_TOKEN") and env.get("GITHUB_TOKEN") == env["FORGEJO_TOKEN"]:
                env.pop("GITHUB_TOKEN")
            if name == "live":
                config = Path(os.environ.get("XDG_CONFIG_HOME") or str(Path.home() / ".config")) / "chr/nightly"
                if (config / "tea/config.yml").is_file() or (config / "gh/hosts.yml").is_file():
                    env["XDG_CONFIG_HOME"] = str(config)
                    env["GH_CONFIG_DIR"] = str(config / "gh")
                    token = config / "github.token"
                    if token.is_file():
                        env["GH_TOKEN"] = token.read_text().strip()
            path = output / (name + (".jsonl" if name in ("normal", "race", "browser", "live") else ".log"))
            started = time.monotonic()
            print(f"Nightly {name}: {' '.join(command)}; evidence: {path}", flush=True)
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError("nightly exhausted its 200-minute execution budget")
            with path.open("w") as stream:
                status = runner(command, timeout=min(6000, remaining), env=env, stdout=stream)
            evidence["checks"].append({"name": name, "command": command,
                                       "status": status, "elapsed_seconds": time.monotonic() - started})
            if path.suffix == ".jsonl":
                timings = test_timings(path)
                (output / (name + "-timings.json")).write_text(json.dumps(timings, indent=2) + "\n")
                if status == 0:
                    check_selected_tests(name, timings)
            if status:
                raise ValueError(f"{name} failed with exit {status}; see {path}")
        if source_commit() != head:
            raise ValueError("candidate changed during verification")
        evidence["status"] = "passed"
        return 0
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        evidence["status"], evidence["error"] = "failed", str(error)
        print(f"Nightly failed: {error}", file=sys.stderr)
        return 1
    finally:
        evidence["finished_at"] = datetime.now(timezone.utc).isoformat()
        (output / "result.json").write_text(json.dumps(evidence, indent=2) + "\n")


if __name__ == "__main__":
    sys.exit(run(ROOT / ".build/nightly", sys.argv[1] if len(sys.argv) == 2 else "go",
                 os.environ.get("CHR_NATIVE_INTEGRATION_QA") == "1"))
