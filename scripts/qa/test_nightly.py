"""Protect nightly admission, exact-source evidence and failure propagation."""

from datetime import datetime
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import nightly
import nightly_admission
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
            for name in (nightly.LIVE_NAMES if options["env"]["CHR_NATIVE_INTEGRATION_QA"] == "1" else nightly.BROWSER_NAMES):
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
        cases = [[], [{"Test": "TestWebWorkflowBrowser", "Action": "pass"}],
                 [{"Test": name, "Action": "skip" if name == "TestWebConsolidatedBrowser" else "pass"}
                  for name in nightly.BROWSER_NAMES]]
        for evidence in cases:
            with self.subTest(evidence=evidence), self.assertRaisesRegex(ValueError, "incomplete test evidence"):
                nightly.check_selected_tests("browser", evidence)
        with self.assertRaisesRegex(ValueError, "incomplete test evidence"):
            nightly.check_selected_tests("live", [{"Test": name, "Action": "skip"} for name in nightly.LIVE_NAMES])

    def test_candidate_change_cannot_produce_green_evidence(self):
        def runner(command, **options):
            for name in nightly.BROWSER_NAMES:
                options["stdout"].write(json.dumps({"Test": name, "Action": "pass", "Elapsed": 0}) + "\n")
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

    def test_live_uses_dedicated_config_and_preserves_github_tokens(self):
        cases = [("workflow alias", "workflow-token", None),
                 ("explicit github token", "github-token", "github-token")]
        for name, token, expected in cases:
            with self.subTest(name=name), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                config = root / "config"
                dedicated = config / "chr/nightly"
                (dedicated / "tea").mkdir(parents=True)
                (dedicated / "tea/config.yml").write_text("{}")
                (dedicated / "github.token").write_text("dedicated-test-token\n")
                observed = []

                def runner(command, **options):
                    env = options["env"]
                    self.assertEqual(env.get("GITHUB_TOKEN"), expected)
                    if env["CHR_NATIVE_INTEGRATION_QA"] == "1":
                        observed.append(env["XDG_CONFIG_HOME"])
                        self.assertEqual(env["GH_TOKEN"], "dedicated-test-token")
                    for test in nightly.LIVE_NAMES + nightly.BROWSER_NAMES:
                        options["stdout"].write(json.dumps({"Test": test, "Action": "pass", "Elapsed": 0}) + "\n")
                    return 0

                with patch.dict("os.environ", {"FORGEJO_TOKEN": "workflow-token", "GITHUB_TOKEN": token,
                                               "XDG_CONFIG_HOME": str(config)}), \
                        patch.object(nightly, "source_commit", return_value="a" * 40), \
                        patch.object(nightly, "preflight"):
                    self.assertEqual(nightly.run(root / "output", live=True, runner=runner), 0)
                self.assertEqual(observed, [str(dedicated)])


class AdmissionTests(unittest.TestCase):
    def test_admission_identifies_its_authenticated_api_request(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "output"
            env = {"FORGEJO_SERVER_URL": "https://forge.example.test", "FORGEJO_REPOSITORY": "team/project",
                   "FORGEJO_RUN_ID": "20", "CHR_ACTIONS_TOKEN": "test-token", "FORGEJO_OUTPUT": str(output)}
            with patch.dict("os.environ", env), \
                    patch.object(nightly_admission.urllib.request, "urlopen") as request_open, \
                    patch.object(nightly_admission, "decision", return_value=(True, "first attempt")):
                request_open.return_value.__enter__.return_value.read.return_value = b'{"workflow_runs": []}'
                nightly_admission.main()
                request = request_open.call_args.args[0]
                self.assertEqual(request.get_header("User-agent"), "chr-nightly/1.0")
                self.assertEqual(request.get_header("Accept"), "application/json")
                self.assertEqual(request.get_header("Authorization"), "Bearer test-token")
            self.assertEqual(output.read_text(), "admitted=yes\n")

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
