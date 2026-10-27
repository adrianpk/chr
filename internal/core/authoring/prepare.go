package authoring

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileSessionRepository) prepareAuthoringContinuation(ctx context.Context, binding authoringRequestBinding, snapshot Snapshot, draft Draft) ([]workflowstate.FileChange, error) {
	request := binding.Request
	session := snapshot.Session
	session.Revision++
	session.Target = Target(request.Target)
	session.UpdatedAt = time.Now().UTC()

	prepared, err := r.prepareAuthoringContent(ctx, binding, snapshot, &session, draft)
	if err != nil {
		return nil, err
	}

	paths, err := r.authoringSessionPaths(session.ID)
	if err != nil {
		return nil, err
	}

	relative, err := authoringRelativePaths(r.projectDir, paths)
	if err != nil {
		return nil, err
	}

	content, err := RenderSession(session, relative)
	if err != nil {
		return nil, err
	}

	transcript, err := r.authoringTranscriptImage(binding, snapshot, relative.TranscriptPath, prepared.Events)
	if err != nil {
		return nil, err
	}

	changes := []workflowstate.FileChange{{Path: relative.SessionPath, Before: snapshot.SessionHash, After: []byte(content)}, {Path: relative.TranscriptPath, Before: snapshot.TranscriptHash, After: transcript}}

	return append(changes, prepared.Changes...), nil
}

// authoringContentImages joins changes and transcript events for one journal.
type authoringContentImages struct {
	Changes []workflowstate.FileChange
	Events  []AssistedTranscriptEvent
}

type authoringOutputImages struct {
	Ref     ArtifactRef
	Changes []workflowstate.FileChange
}

func (r *fileSessionRepository) prepareAuthoringContent(ctx context.Context, binding authoringRequestBinding, snapshot Snapshot, session *Session, draft Draft) (authoringContentImages, error) {
	request := binding.Request

	var (
		changes []workflowstate.FileChange
		events  []AssistedTranscriptEvent
	)

	if request.Action == "create" {
		capture := ContinuationArtifact(request.Namespace, request.SessionID, ArtifactKindDump, true)
		session.Inputs = []ArtifactRef{capture}
		raw := renderAuthoringCapture(*session, request.Capture, request.OriginalInstruction)

		prepared, err := r.authoringArtifactChanges(binding, snapshot, capture, raw)
		if err != nil {
			return authoringContentImages{}, err
		}

		if request.Target != ArtifactKindDump || draft.Question != "" {
			changes = append(changes, prepared...)
		}
	}

	if request.Action == "answer" {
		return authoringContentImages{Changes: changes, Events: []AssistedTranscriptEvent{authoringAnswerEvent(snapshot, request, session.Revision)}}, nil
	}

	if request.Action == "respond" {
		events = append(events, authoringAnswerEvent(snapshot, request, session.Revision))
	}

	if draft.Question != "" {
		event := authoringQuestionEvent(snapshot, session, draft)
		if request.Target == ArtifactKindQuiz {
			prepared, err := r.prepareAuthoringOutput(ctx, binding, snapshot, session, renderAuthoringQuiz(*session, event))

			return authoringContentImages{Changes: append(changes, prepared...), Events: append(events, event)}, err
		}

		return authoringContentImages{Changes: changes, Events: append(events, event)}, nil
	}

	prepared, err := r.prepareAuthoringOutput(ctx, binding, snapshot, session, draft.Body)

	return authoringContentImages{Changes: append(changes, prepared...), Events: events}, err
}

func (r *fileSessionRepository) authoringTranscriptImage(binding authoringRequestBinding, snapshot Snapshot, path string, events []AssistedTranscriptEvent) ([]byte, error) {
	text := "# Transcript\n\n"

	if binding.Request.Action != "create" {
		content, err := r.authoringRead(path)
		if err != nil {
			return nil, err
		}

		if contenthash.Sum(content) != snapshot.TranscriptHash {
			return nil, workflowstate.ErrConflict
		}

		text = string(content)
	}

	for _, event := range events {
		rendered, err := RenderAssistedTranscriptEvent(event)
		if err != nil {
			return nil, err
		}

		text += "\n" + rendered
	}

	return []byte(text), nil
}

func authoringQuestionEvent(snapshot Snapshot, session *Session, draft Draft) AssistedTranscriptEvent {
	index := 1

	for _, event := range snapshot.Events {
		if event.Type == AssistedTranscriptEventQuestion {
			index++
		}
	}

	turn := fmt.Sprintf("turn-%d", index)

	return AssistedTranscriptEvent{ID: turn + "-question", SessionID: session.ID, TurnID: turn, TurnIndex: index, Revision: session.Revision, Type: AssistedTranscriptEventQuestion, Intent: AssistedAuthoringIntentClarify, Status: AssistedAuthoringTurnAsked, Question: AssistedAuthoringQuestion{ID: fmt.Sprintf("question-%d", index), Prompt: draft.Question, Rationale: draft.Rationale, Required: true}, CreatedAt: session.UpdatedAt}
}

func authoringAnswerEvent(snapshot Snapshot, request ContinuationRequest, revision int64) AssistedTranscriptEvent {
	var question AssistedTranscriptEvent

	for _, event := range snapshot.Events {
		if event.Type == AssistedTranscriptEventQuestion && event.Question.ID == request.Question.QuestionID {
			question = event
		}
	}

	question.ID = question.TurnID + "-answer"
	question.Type = AssistedTranscriptEventAnswer
	question.Status = AssistedAuthoringTurnAnswered
	question.Revision = revision
	question.CreatedAt = time.Now().UTC()
	question.Answer = AssistedAuthoringAnswer{QuestionID: question.Question.ID, Text: request.Instructions, AnsweredAt: question.CreatedAt}

	return question
}

func renderAuthoringQuiz(session Session, event AssistedTranscriptEvent) string {
	result := QuizGenerationResult{Provider: "configured", Strategy: "bounded_continuation", Questions: []QuizQuestion{{ID: event.Question.ID, Text: event.Question.Prompt}}}

	return RenderContextualQuizMarkdown(session, ContinuationArtifact(session.Namespace, session.ID, ArtifactKindQuiz, false), QuizGenerationRequest{ID: event.ID, SessionID: session.ID, Topic: session.Topic, Target: session.Target}, result)
}

func renderAuthoringCapture(session Session, capture, instruction string) string {
	body := "---\ntitle: " + opsmetadata.Quote(session.Topic+" Raw Capture") + "\nstatus: \"raw\"\ntype: \"dump_raw\"\n---\n\n# " + session.Topic + " Raw Capture\n"
	if instruction != "" {
		body += "\n## Original Instruction\n\n" + instruction + "\n"
	}

	return body + "\n## Raw Capture\n\n" + capture + "\n"
}

// ContinuationArtifact reserves a deterministic canonical artifact ref.
// Its identity is derived from the shared session, never from model output.
func ContinuationArtifact(namespace, sessionID string, kind ArtifactKind, capture bool) ArtifactRef {
	suffix := string(kind)
	role := ArtifactRoleOutput

	if capture {
		suffix = "capture"
		role = ArtifactRoleInput
	}

	id := "00000-" + opscore.ArtifactSlugOrDefault(sessionID+"-"+suffix)

	return ArtifactRef{ID: id, Kind: kind, Role: role, Path: filepath.ToSlash(filepath.Join(opscore.DirectoryName, namespace, authoringArtifactType(kind), id, opscore.ArtifactBodyFile())), Description: "Shared authoring " + suffix + " artifact."}
}

func (r *fileSessionRepository) prepareAuthoringOutput(ctx context.Context, binding authoringRequestBinding, snapshot Snapshot, session *Session, body string) ([]workflowstate.FileChange, error) {
	images, err := r.prepareAuthoringOutputChanges(ctx, binding, snapshot, session, body)
	if err != nil {
		return nil, err
	}

	ref := images.Ref
	found := false

	for _, existing := range session.Outputs {
		if existing.Path == ref.Path {
			found = true
		}
	}

	if !found {
		session.Outputs = append(session.Outputs, ref)
	}

	source := session.Inputs[0]
	relation := ProvenanceRelationGenerated

	if binding.Request.Action == "refine" {
		source = ref
		relation = ProvenanceRelationRefined
	}

	session.Provenance = append(session.Provenance, ProvenanceLink{Relation: relation, Source: source, Target: ref, Decision: "Bound authoring request " + binding.Request.Binding.RequestID, RecordedAt: session.UpdatedAt})

	return images.Changes, nil
}

func (r *fileSessionRepository) prepareAuthoringOutputChanges(ctx context.Context, binding authoringRequestBinding, snapshot Snapshot, session *Session, body string) (authoringOutputImages, error) {
	if binding.Request.Target == ArtifactKindDump {
		return r.prepareAuthoringDump(ctx, binding, snapshot, *session, body)
	}

	ref := ContinuationArtifact(session.Namespace, session.ID, binding.Request.Target, false)
	changes, err := r.authoringArtifactChanges(binding, snapshot, ref, body)

	return authoringOutputImages{Ref: ref, Changes: changes}, err
}
