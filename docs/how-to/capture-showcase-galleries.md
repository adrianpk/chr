# Capture the showcase galleries

The [Web gallery](../explanation/gallery-web.md) and
[TUI gallery](../explanation/gallery-tui.md) follow the same **CLI gameplay**
use case in the tictactoe example. Capture real application state before
selecting images for either page.

## Prepare the example

Set `SHOWCASE_PROJECT`, `SHOWCASE_HOST` and `CAPTURES` to the isolated
project, host-state and output directories outside this checkout.

The starting point is the incomplete Go project in `qa/ref/003/tictactoe/`.
Keep this reference unchanged. In a separate disposable Git repository, copy
its `go.mod`, `cmd/`, `internal/` and `README.md`. Initialize chr with
`chr init --dir PATH`, name the project Tictactoe, and configure a generation
harness and the native workflow before starting the wizard.

For actual publication, use an owned QA repository and a dedicated target
branch. Configure its publisher as described in
[Native integration](configure-native-integration.md). Fetch only that branch;
fetching every historical QA branch is unnecessary. Keep credentials and host
publisher configuration outside both the example snapshot and the images.

Before starting the captured run, publish the repository's legitimate
`ops/default/coordinator.json` marker to that target branch. Explicit native
authoring creates the marker and its private host registration. The successful
capture uses a fresh linked worktree of that registered repository, so the
coordinator remains the same. See
[Coordinator registration](../reference/workflow-execution.md#coordinator-registration).
Copying the private registration to another clone does not establish ownership.

Open Cases, choose **Wizard**, enter **CLI gameplay**, and submit the
[saved Dump](../../qa/showcase/tictactoe/inputs/cli-gameplay.json). Answer the quiz
through its normal controls. The [retained answers](../../qa/showcase/tictactoe/inputs/quiz-answers.md)
record this run's decisions. The JSON `dump` property preserves exact spaces
and line endings. To extract it from the repository root:

```sh
mkdir -p "$CAPTURES"
node -e 'process.stdout.write(require("./qa/showcase/tictactoe/inputs/cli-gameplay.json").dump)' > "$CAPTURES/dump.txt"
```

A fresh model run can choose different questions
and produce different code; these inputs are a scenario, not a deterministic
model fixture.

Let chr generate the specification, plan and tickets, then run and integrate
the admitted work. Inspect retained validation and delivery receipts before
capturing a completed result. A successful worker or an open PR alone does
not prove delivery. Preserve the generated game separately from the seed.

The recorded run needed explicit CLI continuation after admission: the managed
host left it pending without launching a worker. The operator used
`chr work continue` on the same retained invocation, with its original task
membership and budgets. The images document that real assisted run; they do
not establish that the wizard completed without operator intervention.

Git publication applied the generated commit to both refs, but the forge did
not automatically acknowledge the existing PR. With maintainer approval, the
operator confirmed that PR at the already-applied commit through the forge
and restored its temporary manual-merge option. Native reconciliation then
retained delivery and closing receipts. The capture record lists this assistance.

## Capture the Web UI

Start the pinned chr build against the example. The capture session uses a
local, token-free server; its address is not exposed outside the host:

```sh
chr web serve --dir "$SHOWCASE_PROJECT" --dev-root "$SHOWCASE_HOST" \
  --host 127.0.0.1 --port 8791 --require-token=false
```

The capture helper uses Playwright and system Chromium. Keep capture-tool
dependencies outside the repository and set `NODE_PATH` to their `node_modules`
directory. The recorded tool versions are in the example's
[capture record](../../qa/showcase/tictactoe/capture.json).

For an isolated tool installation:

```sh
CAPTURE_TOOLS="$PWD/../chr-capture-tools"
npm install --prefix "$CAPTURE_TOOLS" --no-audit --no-fund \
  playwright@1.64.0-alpha-1790635538000 @xterm/xterm@6.0.0
export NODE_PATH="$CAPTURE_TOOLS/node_modules"
```

Set `CHR_CAPTURE_BROWSER` when Chromium is not at `/usr/bin/chromium`.

```sh
node scripts/assets/capture-gallery.cjs web \
  --url 'http://127.0.0.1:8791/board?case=cli-gameplay&namespace=default' \
  --output "$CAPTURES/web/board.png" --width 1440 --height 1000 --top 84
```

The helper waits for fonts and rendering, then captures the viewport without
browser or desktop chrome. The gallery trims the 84 px project header, retaining
the navigation and content in a 1440 × 916 frame. `--scroll 350` scrolls a long
document before capture. `--wait SELECTOR` can require a unique visible element.
`--form FILE` fills a JSON mapping of selectors to draft text; it never
submits a form. Complete workflow actions through the real UI before capturing
the next saved state.

Use the existing Cases, Dumps, Quizzes, Specs, Plans, Board and Dash screens.
Keep the functional case in the URL when inspecting a guided stage. Capture
comparable milestones in both interfaces rather than recreating their content.

## Capture the TUI

Use a dedicated tmux socket, a fixed terminal size, and a truecolor environment.
Remove `NO_COLOR` when the calling environment sets it:

```sh
tmux -L chr-showcase -f /dev/null new-session -d -s tui -x 160 -y 44 \
  "exec env -u NO_COLOR TERM=tmux-256color COLORTERM=truecolor \
  CLICOLOR_FORCE=1 'COLORFGBG=15;0' chr tui \
  --dir '$SHOWCASE_PROJECT' --dev-root '$SHOWCASE_HOST'"
```

Operate the real TUI with `tmux send-keys`, or attach to it normally. In Cases,
`w` opens Wizard and `o` reopens the selected case's retained workflow. The
ordinary F1–F7 tabs inspect the same saved artifacts; `c` returns to the current
stage. Wait for the requested view to render before taking its image.

```sh
node scripts/assets/capture-gallery.cjs tui \
  --socket chr-showcase --target tui:0.0 \
  --output "$CAPTURES/tui/board.png"
```

The helper reads the real pane's ANSI cells and geometry, then renders them
with xterm.js. It does not construct UI content. The recorded setup uses
CaskaydiaMono Nerd Font Mono at 18 px, line height 1.1, and a 16 px frame.
The 160 × 44 terminal produces a 1720 × 1044 image with that font. Other fonts
can change the image dimensions.

## Select and verify images

Keep each gallery concise: Wizard entry, Dump, Quiz with history, Spec, Plan,
Board, execution/results and the working game. Prefer a useful saved state over
multiple nearly identical progress frames. Preserve each interface's native
colors and layout.

Inspect every full-size image for legibility and incidental private paths,
accounts, hostnames or network details. Verify the captions against the actual
state and the recorded source revision. Open both rendered gallery pages and
check every image and link. After capture, stop the owned idle UI processes;
do not terminate an active worker or validation command to tidy the session.
