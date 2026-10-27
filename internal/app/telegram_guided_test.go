package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func guidedTelegramFixture(t *testing.T) (*telegramConversationAppFixture, service) {
	t.Helper()

	if testing.Short() {
		t.Skip("real guided project, Git and Telegram transport fixture")
	}

	f := newTelegramConversationAppFixture(t)

	operations, ok := f.receiver.operations.(service)
	if !ok {
		t.Fatal("missing shared operation service")
	}

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("missing registered guided project", err)
	}

	root := projects[0].Path
	// Keep real project/Git owners while replacing only content-provider IO.
	if err := os.MkdirAll(filepath.Join(root, ".chr"), 0o755); err != nil {
		t.Fatal(err)
	}

	initializeGuidedWebGit(t, root)

	now := operations.utcNow()
	operations.quizModel = &fakeServiceQuizModel{responses: []quizmodel.Response{serviceQuizModelResponse("question", "question-1", now)}}
	f.receiver.operations = operations

	return f, operations
}

// TestGuidedUseCaseTelegramCapture routes ordinary messages through the same
// canonical run and proves that replay and stale peer replies cannot append twice.
func TestGuidedUseCaseTelegramCapture(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)

	start := f.send(t, "Develop a new use case", 0)
	if !strings.Contains(start.Text, "use-case name") {
		t.Fatal("natural start did not ask for a name", start.Text)
	}

	name := f.send(t, "Bounded audit", start.ID)
	if !strings.Contains(name.Text, "Stage: dump") {
		t.Fatal("name did not reach canonical capture", name.Text)
	}

	capture := f.send(t, "Retain exact validation and delivery evidence.", name.ID)
	c := f.sessions(t)[0]

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		t.Fatal(err)
	}

	scope := authoringcore.GuidedUseCaseScope{RunID: id, ProjectID: c.ProjectID, Namespace: c.Namespace}
	root := guidedTelegramProject(t, operations, c.ProjectID)

	before, err := authoringcore.ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil || strings.Count(before.Guided.Capture, "Retain exact") != 1 {
		t.Fatal("capture was not retained once", err)
	}

	assertGuidedTelegramReplay(t, f, root, scope, before, start.ID)

	if capture.ID == 0 {
		t.Fatal("capture response was not delivered")
	}
}

func assertGuidedTelegramReplay(t *testing.T, f *telegramConversationAppFixture, root string, scope authoringcore.GuidedUseCaseScope, before authoringcore.Session, startID int64) {
	t.Helper()

	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	stale := f.send(t, "Replace the old name", startID)
	if !strings.Contains(stale.Text, "Saved details are unchanged") {
		t.Fatal("stale prompt was not rejected", stale.Text)
	}

	after, err := authoringcore.ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil || after.Revision != before.Revision || after.Guided.Capture != before.Guided.Capture {
		t.Fatal("replay or stale prompt changed canonical state", err)
	}
}

func guidedTelegramProject(t *testing.T, operations service, id string) string {
	t.Helper()

	projects, err := operations.ListRegisteredProjects([]string{id})
	if err != nil || len(projects) != 1 {
		t.Fatal(err)
	}

	return projects[0].Path
}

// TestGuidedUseCaseTelegramStartIntent excludes incidental or negated model text
// from the deterministic development instruction accepted by the guided bridge.
func TestGuidedUseCaseTelegramStartIntent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		text string
		want bool
	}{
		{"Develop a new use case", true},
		{"Start guided development.", true},
		{"Do not develop a new use case", false},
		{"Explain how to develop a new use case", false},
		{"/develop manual topic", false},
	}
	for _, tt := range cases {
		t.Run(tt.text, func(t *testing.T) {
			if got := telegramGuidedStart(tt.text); got != tt.want {
				t.Fatal("development intent changed", got, tt.want)
			}
		})
	}
}

// TestGuidedUseCaseTelegramClarification follows current prompts through exact
// default approval; a later idea and pause cannot rewrite the committed scope.
func TestGuidedUseCaseTelegramClarification(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	model := guidedTelegramModel(t, &operations)
	f.receiver.operations = operations
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Bounded audit", start.ID)
	capture := f.send(t, "Deliver a bounded workflow.", name.ID)

	question := f.send(t, "Finish capture", capture.ID)
	if !strings.Contains(question.Text, "What should the workflow preserve") {
		t.Fatal("current clarification missing", question.Text)
	}

	refresh := f.send(t, "Show current step", question.ID)
	if len(model.requests) != 1 || !strings.Contains(refresh.Text, "What should the workflow preserve") {
		t.Fatal("inspection advanced clarification", refresh.Text)
	}

	response := f.send(t, "Keep the original scope.", refresh.ID)
	if !strings.Contains(response.Text, "Stage: preparation") || !strings.Contains(response.Text, "Configure native workflow version 1") {
		t.Fatal("retained prerequisite missing", response.Text)
	}

	c := f.sessions(t)[0]

	before := readGuidedTelegramSession(t, operations, c)
	if before.Guided.SpecApproval == nil || len(before.Guided.Checkpoints) != 3 || len(model.requests) != 3 {
		t.Fatal("generated specification was not approved once")
	}

	assertGuidedTelegramOriginals(t, f, operations, before)
	assertGuidedTelegramIdeaPause(t, f, operations, model, before, 0)
}

func assertGuidedTelegramIdeaPause(t *testing.T, f *telegramConversationAppFixture, operations service, model *fakeServiceQuizModel, before authoringcore.Session, reply int64) {
	t.Helper()

	idea := f.send(t, "Idea: add an unrelated export later", reply)
	paused := f.send(t, "Pause use case", idea.ID)

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if after.Guided.Status != "paused" || len(after.Guided.LaterIdeas) != 1 || after.Guided.Checkpoints[2].Commit != before.Guided.Checkpoints[2].Commit || len(model.requests) != 3 {
		t.Fatal("idea or pause replaced the agreed input", paused.Text)
	}
}

func guidedTelegramModel(t *testing.T, operations *service) *fakeServiceQuizModel {
	t.Helper()

	now := operations.utcNow()
	completion := serviceQuizModelResponse("completion", "complete", now)
	completion.Question.Required = false
	model := &fakeServiceQuizModel{responses: []quizmodel.Response{serviceQuizModelResponse("question", "question-1", now), completion, serviceQuizDraftSpecResponse("specification", now)}}
	operations.quizModel = model

	return model
}

func readGuidedTelegramSession(t *testing.T, operations service, c telegramcore.Conversation) authoringcore.Session {
	t.Helper()

	root := guidedTelegramProject(t, operations, c.ProjectID)

	session, err := authoringcore.ReadGuidedUseCase(t.Context(), root, authoringcore.GuidedUseCaseScope{RunID: c.GuidedRunID, ProjectID: c.ProjectID, Namespace: c.Namespace})
	if err != nil {
		t.Fatal(err)
	}

	return session
}

// TestGuidedUseCaseTelegramButtons consumes one retained choice through the
// existing action owner and binds its result prompt to the new guided revision.
func TestGuidedUseCaseTelegramButtons(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Bounded audit", start.ID)
	capture := f.send(t, "Deliver a bounded workflow.", name.ID)

	controls := f.send(t, "Show controls", capture.ID)
	if strings.Contains(controls.Text, "tgr-") || strings.Contains(controls.Text, "/answer") {
		t.Fatal("normal buttons exposed transport identifiers", controls.Text)
	}

	request, action := guidedTelegramAction(t, f, controls)

	callback := sendGuidedTelegramCallback(t, f, action, "finish", 100001)
	if !strings.Contains(callback.Text, "What should the workflow preserve") {
		t.Fatal("button did not reach current question", callback.Text)
	}

	before := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	duplicate := sendGuidedTelegramCallback(t, f, action, "finish", 100001)

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if after.Revision != before.Revision || !strings.Contains(duplicate.Text, "already confirmed") {
		t.Fatal("duplicate callback repeated the owner", duplicate.Text)
	}

	operationLookup, err := f.receiver.store.Operation(t.Context(), request.Source, 0)

	retained, found := operationLookup.Request, operationLookup.Found
	if err != nil || !found || retained.State != "confirmed" {
		t.Fatal("action receipt missing", err)
	}

	refresh := f.send(t, "Show current step", callback.ID)
	if !strings.Contains(refresh.Text, "What should the workflow preserve") {
		t.Fatal("button result did not bind the next input", refresh.Text)
	}
}

func guidedTelegramAction(t *testing.T, f *telegramConversationAppFixture, prompt telegramConversationNotice) (telegramcore.OperationRequest, telegramcore.Action) {
	t.Helper()

	var keyboard struct {
		Rows [][]struct {
			Text string `json:"text"`
			Data string `json:"callback_data"`
		} `json:"inline_keyboard"`
	}
	if err := json.Unmarshal(prompt.Keyboard, &keyboard); err != nil {
		t.Fatal(err)
	}

	if len(keyboard.Rows) == 0 || len(keyboard.Rows[0]) == 0 {
		t.Fatal("guided buttons missing")
	}

	button := keyboard.Rows[0][0]
	if button.Text != "Finish capture" {
		t.Fatal("current action label missing", button)
	}

	parts := strings.Split(button.Data, ":")
	if len(parts) != 3 {
		t.Fatal("invalid callback")
	}

	peer := f.sessions(t)[0].Peer

	request, err := f.receiver.store.ActionRequest(t.Context(), peer, telegramcore.ActionAnswer{Token: parts[1], Choice: parts[2]})
	if err != nil {
		t.Fatal(err)
	}

	action, err := f.receiver.store.RequestAction(t.Context(), request)
	if err != nil || action.GuardVersion != 1 {
		t.Fatal("sealed prompt binding missing", err)
	}

	return request, action
}

func sendGuidedTelegramCallback(t *testing.T, f *telegramConversationAppFixture, action telegramcore.Action, choice string, user int64) telegramConversationNotice {
	t.Helper()

	queueGuidedTelegramCallback(t, f, action, choice, user)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	select {
	case notice := <-f.notices:
		return notice
	default:
		return telegramConversationNotice{}
	}
}

func queueGuidedTelegramCallback(t *testing.T, f *telegramConversationAppFixture, action telegramcore.Action, choice string, user int64) {
	t.Helper()

	callback, err := telegramcore.ActionCallbackData(telegramcore.ActionAnswer{Token: action.Token, Choice: choice})
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal([]any{map[string]any{"update_id": f.next, "callback_query": map[string]any{"id": "guided-callback", "from": map[string]any{"id": user}, "data": callback, "message": map[string]any{"date": 1, "message_id": f.next, "chat": map[string]any{"id": 100001, "type": "private"}}}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- data
}

// TestGuidedUseCaseTelegramNames switches the existing peer focus by a unique
// visible name; duplicate names cannot choose an arbitrary canonical run.
func TestGuidedUseCaseTelegramNames(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	first := f.send(t, "Develop a new use case in example-project", 0)
	f.send(t, "First audit", first.ID)

	original := f.sessions(t)[0]
	second := f.send(t, "Develop a new use case", 0)
	f.send(t, "Second audit", second.ID)

	selected := f.send(t, "Show use case First audit", 0)
	if !strings.Contains(selected.Text, "First audit") {
		t.Fatal("visible name did not select its run", selected.Text)
	}

	f.send(t, "Details for the first run.", 0)

	after := readGuidedTelegramSession(t, operations, f.session(t, original.ID))
	if strings.TrimSpace(after.Guided.Capture) != "Details for the first run." {
		t.Fatal("name selection did not update peer focus")
	}

	duplicate := f.send(t, "Develop a new use case", 0)
	f.send(t, "First audit", duplicate.ID)

	rejected := f.send(t, "Show use case First audit", 0)
	if !strings.Contains(rejected.Text, "ambiguous") {
		t.Fatal("duplicate visible name picked an arbitrary run", rejected.Text)
	}

	unchanged := readGuidedTelegramSession(t, operations, f.session(t, original.ID))
	if unchanged.Revision != after.Revision {
		t.Fatal("ambiguous selection changed canonical input")
	}
}

type telegramGuidedLostAck struct {
	service
	lost  error
	calls int
}

func (s *telegramGuidedLostAck) ApplyTelegramGuided(ctx context.Context, target telegramcore.OperationTarget, input telegramGuidedInput) (authoringcore.Session, error) {
	session, err := s.service.ApplyTelegramGuided(ctx, target, input)
	if err == nil && input.Command.Action == "pause" {
		s.calls++
		if s.calls == 1 {
			return authoringcore.Session{}, s.lost
		}
	}

	return session, err
}

// TestGuidedUseCaseTelegramLostAck drops the application result after a real
// canonical pause, then replays the retained callback without another mutation.
func TestGuidedUseCaseTelegramLostAck(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Pause recovery", start.ID)
	controls := f.send(t, "Show controls", name.ID)
	_, action := guidedTelegramAction(t, f, controls)
	lost := errors.New("lost canonical guided acknowledgement")
	boundary := &telegramGuidedLostAck{service: operations, lost: lost}
	f.receiver.operations = boundary
	queueGuidedTelegramCallback(t, f, action, "pause", 100001)

	if _, err := f.receiver.step(t.Context()); !errors.Is(err, lost) {
		t.Fatal("lost acknowledgement was not retained", err)
	}

	before := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if before.Guided.Status != "paused" {
		t.Fatal("fault preceded the canonical effect")
	}

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if after.Revision != before.Revision || boundary.calls != 2 {
		t.Fatal("recovery changed original control effect")
	}

	notice := telegramOperationNotice(t, f)
	if !strings.Contains(notice.Text, "paused") {
		t.Fatal("original result was not delivered", notice.Text)
	}
}

// TestGuidedUseCaseTelegramProjectChoice requests a name only when the current
// allowed registry actually has multiple candidates and retains the chosen root.
func TestGuidedUseCaseTelegramProjectChoice(t *testing.T) {
	t.Parallel()

	f, _ := guidedTelegramFixture(t)

	config, err := telegramcore.LoadReceiverConfig(f.receiver.env)
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = append(config.AllowedUsers[0].ProjectIDs, "private-project")
	if err := telegramcore.SaveReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	selection := f.send(t, "Develop a new use case", 0)
	if !strings.Contains(selection.Text, "example-project") || !strings.Contains(selection.Text, "private-project") || strings.Contains(selection.Text, "/project") || len(f.sessions(t)) != 0 {
		t.Fatal("ambiguous project did not request a recognizable name", selection.Text)
	}

	assertGuidedTelegramChosenProject(t, f, selection.ID)
}

func assertGuidedTelegramChosenProject(t *testing.T, f *telegramConversationAppFixture, reply int64) {
	t.Helper()

	chosen := f.send(t, "Develop a new use case in example-project", reply)

	sessions := f.sessions(t)
	if len(sessions) != 1 || sessions[0].ProjectID != "example-project" || !strings.Contains(chosen.Text, "use-case name") {
		t.Fatal("named choice did not retain exact project", chosen.Text)
	}

	rejected := f.send(t, "Develop a new use case in disabled-project", 0)
	if !strings.Contains(rejected.Text, "unavailable") || len(f.sessions(t)) != 1 {
		t.Fatal("disabled project was adopted", rejected.Text)
	}
}

type telegramInvalidGuidedScope struct{ service }

func (s telegramInvalidGuidedScope) InspectTelegramGuided(ctx context.Context, target telegramcore.OperationTarget, id string) (GuidedUseCaseInspection, error) {
	observed, err := s.service.InspectTelegramGuided(ctx, target, id)
	for i := range observed.Runs {
		if observed.Runs[i].Guided != nil && observed.Runs[i].Topic == "First guarded scope" {
			state := *observed.Runs[i].Guided
			state.SchemaVersion = 99
			observed.Runs[i].Topic = "Unbound guarded scope"
			observed.Runs[i].Guided = &state
		}
	}

	return observed, err
}

// A narrow routing projection cannot bypass complete guided scope validation.
// Rejected name attachment must leave the peer on its current canonical run.
func TestGuidedUseCaseTelegramScope(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	first := f.send(t, "Develop a new use case in example-project", 0)
	f.send(t, "First guarded scope", first.ID)

	original := f.sessions(t)[0]
	before := readGuidedTelegramSession(t, operations, original)
	second := f.send(t, "Develop a new use case", 0)
	f.send(t, "Current guarded scope", second.ID)

	current := f.sessions(t)[0]
	if current.ID == original.ID {
		t.Fatal("fixture did not select a distinct current run")
	}

	f.receiver.operations = telegramInvalidGuidedScope{operations}
	f.send(t, "Show use case Unbound guarded scope", 0)

	f.receiver.operations = operations
	f.send(t, "Details stay on the current valid run.", 0)

	unchanged := readGuidedTelegramSession(t, operations, f.session(t, original.ID))

	after := readGuidedTelegramSession(t, operations, f.session(t, current.ID))
	if unchanged.Revision != before.Revision || strings.TrimSpace(after.Guided.Capture) != "Details stay on the current valid run." {
		t.Fatal("invalid domain scope changed peer focus or canonical progression", unchanged.Revision, before.Revision, after.Guided.Capture)
	}
}
