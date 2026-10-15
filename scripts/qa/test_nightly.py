"""Protect nightly admission, exact-source evidence and failure propagation."""

from datetime import datetime
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import nightly
from nightly_admission import decision


class NightlyTests(unittest.TestCase):
    def setUp(self):
        console = patch.object(nightly, "print", create=True)
        console.start()
        self.addCleanup(console.stop)
        browser = patch.object(nightly, "browser_path", return_value="/tmp/test-browser")
        browser.start()
        self.addCleanup(browser.stop)

    def test_failure_retains_source_and_stops_later_checks(self):
        calls = []

        def runner(command, **options):
            calls.append(command)
            if "-json" in command:
                options["stdout"].write(json.dumps({"Test": "TestBoundary", "Package": "core",
                                                  "Action": "pass", "Elapsed": 4.2}) + "\n")
            return 1 if "-race" in command else 0

        with tempfile.TemporaryDirectory() as directory, \
                patch.object(nightly, "source_commit", return_value="a" * 40), \
                patch.object(nightly, "preflight"):
            output = Path(directory)
            self.assertEqual(nightly.run(output, live=True, runner=runner), 1)
            evidence = json.loads((output / "result.json").read_text())
            self.assertEqual(evidence["commit"], "a" * 40)
            self.assertEqual(evidence["status"], "failed")
            self.assertEqual([step["name"] for step in evidence["checks"]],
                             ["lint", "scripts", "build", "normal", "race"])
            self.assertEqual(json.loads((output / "normal-timings.json").read_text())[0]["Elapsed"], 4.2)
            self.assertEqual(len(calls), 5)

    def test_live_is_isolated_from_local_full_and_race(self):
        modes = []

        def runner(command, **options):
            modes.append((command, options["env"]["CHR_NATIVE_INTEGRATION_QA"]))
            for name in (nightly.LIVE_NAMES if options["env"]["CHR_NATIVE_INTEGRATION_QA"] == "1" else ("TestWebWorkflowBrowser",)):
                options["stdout"].write(json.dumps({"Test": name, "Action": "pass", "Elapsed": 0}) + "\n")
            return 0

        with tempfile.TemporaryDirectory() as directory, \
                patch.object(nightly, "source_commit", return_value="a" * 40), \
                patch.object(nightly, "preflight"):
            self.assertEqual(nightly.run(Path(directory), live=True, runner=runner), 0)
        self.assertEqual([mode for _, mode in modes], ["0", "0", "0", "0", "0", "0", "1", "0"])
        self.assertEqual(sum("-race" in command for command, _ in modes), 1)
        full = [command for command, _ in modes if "./..." in command]
        self.assertEqual(len(full), 2)
        self.assertTrue(all("-short" not in command and "-run" not in command for command in full))

    def test_missing_browser_test_cannot_get_green_evidence(self):
        with self.assertRaisesRegex(ValueError, "incomplete test evidence"):
            nightly.check_selected_tests("browser", [])
        with self.assertRaisesRegex(ValueError, "incomplete test evidence"):
            nightly.check_selected_tests("live", [{"Test": name, "Action": "skip"} for name in nightly.LIVE_NAMES])

    def test_candidate_change_cannot_produce_green_evidence(self):
        def runner(command, **options):
            options["stdout"].write(json.dumps({"Test": "TestWebWorkflowBrowser", "Action": "pass", "Elapsed": 0}) + "\n")
            return 0

        with tempfile.TemporaryDirectory() as directory, \
                patch.object(nightly, "source_commit", side_effect=["a" * 40, "b" * 40]), \
                patch.object(nightly, "preflight"):
            output = Path(directory)
            self.assertEqual(nightly.run(output, runner=runner), 1)
            evidence = json.loads((output / "result.json").read_text())
            self.assertEqual(evidence["status"], "failed")
            self.assertIn("candidate changed", evidence["error"])

    def test_missing_live_authorization_fails_before_expensive_checks(self):
        with tempfile.TemporaryDirectory() as directory, \
                patch.dict("os.environ", {"XDG_STATE_HOME": directory}):
            with self.assertRaisesRegex(ValueError, "authorized fixture descriptor"):
                nightly.preflight(True)


class AdmissionTests(unittest.TestCase):
    def test_daily_admission_and_delayed_queue(self):
        current = {"id": 20, "workflow_id": "nightly.yml", "status": "running",
                   "created": "2026-10-15T22:17:00Z"}
        now = datetime.fromisoformat("2026-10-16T00:20:00+02:00")
        cases = [
            ("first", [], now, True),
            ("too late", [], now.replace(hour=1), False),
            ("too early", [], now.replace(minute=10), False),
            ("previous day done", [dict(current, id=19, status="success", created="2026-10-14T22:17:00Z")], now, True),
            ("previous day active", [dict(current, id=19, created="2026-10-14T22:17:00Z")], now, False),
            ("same day failed", [dict(current, id=19, status="failure")], now, False),
            ("same day passed", [dict(current, id=19, status="success")], now, False),
            ("unrelated", [dict(current, id=19, workflow_id="quick.yml")], now, True),
        ]
        for name, earlier, clock, expected in cases:
            with self.subTest(name=name):
                self.assertEqual(decision([current, *earlier], 20, clock)[0], expected)


if __name__ == "__main__":
    unittest.main()
