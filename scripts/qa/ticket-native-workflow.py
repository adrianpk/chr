#!/usr/bin/env python3
"""Run bounded local workflow assurance with checked acceptance/fault bindings."""

import ctypes
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import sys
import time


ROOT = Path(__file__).resolve().parents[2]
SCENARIOS = ROOT / "qa/scenarios/ticket-native-workflow"
PACKAGES = ("./internal/core", "./internal/app")
FILE_LIMIT = 1024 * 1024


def read_bounded(path):
    with path.open("rb") as stream:
        data = stream.read(FILE_LIMIT + 1)
    if len(data) > FILE_LIMIT:
        raise ValueError("scenario manifest or test source exceeds 1 MiB")
    return data


def load_bindings(filename, collection, expected):
    data = read_bounded(SCENARIOS / filename)
    manifest = json.loads(data)
    if manifest["version"] != 1 or manifest["evidence_kind"] != "deterministic_local":
        raise ValueError("unsupported scenario schema or evidence kind")
    entries = manifest[collection]
    if [entry["id"] for entry in entries] != expected:
        raise ValueError("scenario manifest has missing, repeated or reordered IDs")
    tests = {}
    for entry in entries:
        if not 1 <= len(entry["tests"]) <= 16:
            raise ValueError("scenario test count exceeds bounds")
        for test in entry["tests"]:
            validate_binding(test)
            key = (test["package"], test["test"])
            if key in tests and tests[key] != test["source"]:
                raise ValueError("a test has conflicting source bindings")
            tests[key] = test["source"]
    return tests


def validate_binding(test):
    package, name, source = test["package"], test["test"], test["source"]
    if package not in PACKAGES or not re.fullmatch(r"Test[A-Za-z0-9_]+", name):
        raise ValueError("invalid package or test name")
    path = Path(source)
    expected_parent = Path(package.removeprefix("./"))
    if path.parent != expected_parent or not path.name.endswith("_test.go"):
        raise ValueError("test source is outside its owning package")
    data = read_bounded(ROOT / path)
    if not re.search(
        rb"^func " + name.encode("ascii") + rb"\(t \*testing\.T\)", data, re.MULTILINE
    ):
        raise ValueError("scenario test function is absent or its source exceeds bounds")


def own_fixture_children():
    if sys.platform != "linux":
        raise ValueError("workflow process assurance requires Linux")
    # Adopt descendants if a Go test or fixture owner crashes, including workers
    # in their own process groups. This changes only this QA runner's ownership.
    libc = ctypes.CDLL(None, use_errno=True)
    if libc.prctl(36, 1, 0, 0, 0) != 0:  # PR_SET_CHILD_SUBREAPER
        raise OSError(ctypes.get_errno(), "cannot supervise owned fixture descendants")


def remaining_children():
    while True:
        try:
            pid, _ = os.waitpid(-1, os.WNOHANG)
        except ChildProcessError:
            return []
        if pid == 0:
            break
    path = Path(f"/proc/self/task/{os.getpid()}/children")
    return [int(pid) for pid in path.read_text().split()]


def stop_fixture_children():
    leaked = False
    for attempt in range(30):
        children = remaining_children()
        if not children:
            return leaked
        leaked = True
        for pid in children:
            try:
                descriptor = os.pidfd_open(pid)
            except ProcessLookupError:
                continue
            try:
                # Direct children are owned by this subreaper. pidfd prevents
                # PID reuse from redirecting cleanup to another process.
                action = signal.SIGTERM if attempt < 10 else signal.SIGKILL
                signal.pidfd_send_signal(descriptor, action)
            except ProcessLookupError:
                pass
            finally:
                os.close(descriptor)
        time.sleep(0.1)
    raise ValueError("owned fixtures did not stop within the cleanup deadline")


def run_scenarios(command, timeout=1800, env=None, stdout=None):
    own_fixture_children()
    try:
        result = subprocess.run(
            command, cwd=ROOT, timeout=timeout, check=False, env=env, stdout=stdout
        )
    finally:
        leaked = stop_fixture_children()
    if leaked:
        raise ValueError("owned fixtures remained after checks; cleanup completed")
    return result.returncode


def scenario_names():
    tests = load_bindings("acceptance.json", "criteria", [f"A{i:02}" for i in range(1, 25)])
    faults = load_bindings(
        "faults.json", "faults", ["restart", "crash", "lost_response", "manual_edit"]
    )
    for key, source in faults.items():
        if key in tests and tests[key] != source:
            raise ValueError("fault and acceptance bindings disagree")
        tests[key] = source
    names = sorted({name for _, name in tests})
    if len(names) > 128:
        raise ValueError("scenario suite exceeds 128 distinct tests")
    return names


def main():
    if len(sys.argv) > 2:
        raise ValueError("usage: ticket-native-workflow.py [go-executable]")
    names = scenario_names()
    command = [sys.argv[1] if len(sys.argv) == 2 else "go", "test", *PACKAGES]
    command += ["-run", "^(" + "|".join(names) + ")$", "-count=1", "-timeout=20m", "-v"]
    print(f"Validated A01-A24 and four fault groups; running {len(names)} distinct tests.", flush=True)
    status = run_scenarios(command)
    if status:
        return status
    print("Local workflow assurance passed; live provider/method validation remains separate.")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, KeyError, TypeError, subprocess.TimeoutExpired) as error:
        print(f"workflow assurance failed: {error}", file=sys.stderr)
        sys.exit(1)
