#!/usr/bin/env python3
"""Admit one scheduled attempt per local date within the reserved night slot."""

from datetime import datetime, time
import json
import os
from pathlib import Path
import urllib.request
from zoneinfo import ZoneInfo


ZONE = ZoneInfo("Europe/Warsaw")


def decision(runs, current_id, now):
    local = now.astimezone(ZONE)
    if not time(0, 17) <= local.time() < time(0, 37):
        return False, "outside the reserved 00:17-00:37 start window"
    current = next(run for run in runs if run["id"] == current_id)
    created = datetime.fromisoformat(current["created"].replace("Z", "+00:00"))
    if created.astimezone(ZONE).date() != local.date():
        return False, "the queued attempt belongs to an earlier local date"
    for run in runs:
        if run["workflow_id"] != current["workflow_id"] or run["id"] >= current_id:
            continue
        if run["status"] in ("running", "waiting", "blocked"):
            return False, "an earlier nightly is still active"
        day = datetime.fromisoformat(run["created"].replace("Z", "+00:00")).astimezone(ZONE).date()
        if day == local.date():
            return False, "this local date already has a nightly attempt"
    return True, "first nightly attempt in the reserved slot"


def main():
    server = os.environ["FORGEJO_SERVER_URL"].rstrip("/")
    repository = os.environ["FORGEJO_REPOSITORY"]
    request = urllib.request.Request(
        f"{server}/api/v1/repos/{repository}/actions/runs?limit=50",
        headers={"Authorization": "Bearer " + os.environ["CHR_ACTIONS_TOKEN"]},
    )
    with urllib.request.urlopen(request, timeout=20) as response:
        data = response.read(1024 * 1024 + 1)
    if len(data) > 1024 * 1024:
        raise ValueError("nightly admission response exceeds 1 MiB")
    admitted, reason = decision(json.loads(data)["workflow_runs"],
                                int(os.environ["FORGEJO_RUN_ID"]), datetime.now(ZONE))
    with Path(os.environ["FORGEJO_OUTPUT"]).open("a") as output:
        output.write("admitted=" + ("yes" if admitted else "no") + "\n")
    print("Nightly admission: " + reason)


if __name__ == "__main__":
    main()
