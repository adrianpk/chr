#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
port="${CHR_WEB_SMOKE_PORT:-7340}"
project_dir="${CHR_WEB_SMOKE_PROJECT:-}"
created_project=""

if [ -z "$project_dir" ]; then
	project_dir="$(mktemp -d)"
	created_project="yes"
fi

server_binary="$(mktemp)"
server_log="$(mktemp)"
body_file="$(mktemp)"
screenshot_dir="${CHR_WEB_SMOKE_SCREENSHOTS:-/tmp/chr-web-ui-smoke-$$}"
server_pid=""

cleanup() {
	if [ -n "$server_pid" ]; then
		kill "$server_pid" 2>/dev/null || true
		wait "$server_pid" 2>/dev/null || true
	fi
	rm -f "$server_binary" "$server_log" "$body_file"
	if [ -n "$created_project" ]; then
		rm -rf "$project_dir"
	fi
}
trap cleanup EXIT HUP INT TERM

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "web ui smoke failed: missing command $1" >&2
		exit 1
	fi
}

fetch_expect() {
	path="$1"
	want="$2"

	if ! curl -fsS "http://127.0.0.1:$port$path" >"$body_file"; then
		echo "web ui smoke failed: GET $path failed" >&2
		cat "$server_log" >&2
		exit 1
	fi

	if ! grep -q "$want" "$body_file"; then
		echo "web ui smoke failed: GET $path did not contain $want" >&2
		exit 1
	fi
}

fetch_hx_expect() {
	path="$1"
	want="$2"

	if ! curl -fsS -H "HX-Request: true" "http://127.0.0.1:$port$path" >"$body_file"; then
		echo "web ui smoke failed: htmx GET $path failed" >&2
		cat "$server_log" >&2
		exit 1
	fi

	if ! grep -q "$want" "$body_file"; then
		echo "web ui smoke failed: htmx GET $path did not contain $want" >&2
		exit 1
	fi
}

require_command curl

mkdir -p "$screenshot_dir"

(cd "$repo_root" && go run ./cmd/chr init --dir "$project_dir" >/dev/null)
(cd "$repo_root" && go run ./cmd/chr metadata project set --dir "$project_dir" \
	--name "Web Smoke Project" \
	--description "Temporary Web UI smoke project." \
	--context "Web UI smoke context." >/dev/null)
(cd "$repo_root" && go run ./cmd/chr metadata case set web-smoke --dir "$project_dir" \
	--name "Web Smoke Case" \
	--description "Temporary Web UI smoke case." \
	--context "Case context for Web UI smoke." >/dev/null)
(cd "$repo_root" && go run ./cmd/chr dump new --dir "$project_dir" \
	--title "Smoke Notes" \
	--content "Initial Web UI smoke dump." >/dev/null)
(cd "$repo_root" && go run ./cmd/chr ticket new "Smoke ticket" --dir "$project_dir" \
	--synopsis "Web UI smoke ticket." \
	--type task \
	--priority normal >/dev/null)

(cd "$repo_root" && go build -o "$server_binary" ./cmd/chr)
"$server_binary" web serve --dir "$project_dir" --port "$port" --require-token false >"$server_log" 2>&1 &
server_pid="$!"

ready=""
for _ in 1 2 3 4 5 6 7 8 9 10; do
	if curl -fsS "http://127.0.0.1:$port/healthz" >/dev/null 2>&1; then
		ready="yes"
		break
	fi
	sleep 1
done

if [ -z "$ready" ]; then
	echo "web ui smoke failed: server did not become ready" >&2
	cat "$server_log" >&2
	exit 1
fi

fetch_expect "/" "chr"
fetch_expect "/project" "$project_dir"
fetch_expect "/board" "Smoke ticket"
fetch_expect "/dash" "Readiness"
fetch_expect "/cases" "Web Smoke Case"
fetch_expect "/dumps" "Smoke Notes"
fetch_expect "/quizzes" "Quizzes"
fetch_expect "/specs" "Generate Spec"
fetch_expect "/plans" "Generate Plan"
fetch_hx_expect "/dumps/fragment" "Smoke Notes"
fetch_expect "/assets/chr.css" ".app-shell"
fetch_expect "/assets/chr-logo.png" ""

if ! curl -fsS -X POST "http://127.0.0.1:$port/actions" \
	-d "action=dump.create" \
	-d "family=dumps" \
	-d "confirm=true" \
	-d "title=Created From Web Smoke" \
	-d "content=Created through the Web UI smoke script." >/dev/null; then
	echo "web ui smoke failed: dump.create action failed" >&2
	cat "$server_log" >&2
	exit 1
fi

fetch_expect "/dumps" "Created From Web Smoke"

browser=""
for candidate in chromium google-chrome google-chrome-stable; do
	if command -v "$candidate" >/dev/null 2>&1; then
		browser="$candidate"
		break
	fi
done

if [ -n "$browser" ]; then
	(cd "$repo_root" && CHR_WEB_BROWSER="$(command -v "$browser")" go test ./internal/app -run '^TestWebWorkflowBrowser$' -count=1)
	"$browser" --headless=new --disable-gpu --no-sandbox --window-size=1280,900 \
		--screenshot="$screenshot_dir/board.png" "http://127.0.0.1:$port/board" >/dev/null 2>&1
	"$browser" --headless=new --disable-gpu --no-sandbox --window-size=390,900 \
		--screenshot="$screenshot_dir/dumps-mobile.png" "http://127.0.0.1:$port/dumps" >/dev/null 2>&1

	if [ ! -s "$screenshot_dir/board.png" ] || [ ! -s "$screenshot_dir/dumps-mobile.png" ]; then
		echo "web ui smoke failed: browser screenshots were not created" >&2
		exit 1
	fi

	echo "web ui smoke ok: pages, fragments, write action, native workflow browser actions, and screenshots passed"
	echo "screenshots: $screenshot_dir"
else
	echo "web ui smoke ok: pages, fragments, and write action passed"
	echo "browser screenshots skipped: chromium or chrome not found"
fi
