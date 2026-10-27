package core

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

func (r *fileAuthoringSessionRepository) prepareAuthoringContinuation(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, draft AuthoringDraft) ([]workflowstate.FileChange, error) {
	request := binding.Request
	session := snapshot.Session
	session.Revision++
	session.Target = AuthoringTarget(request.Target)
	session.UpdatedAt = time.Now().UTC()

	artifacts, events, err := r.prepareAuthoringContent(ctx, binding, snapshot, &session, draft)
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

	content, err := RenderAuthoringSession(session, relative)
	if err != nil {
		return nil, err
	}

	transcript, err := r.authoringTranscriptImage(binding, snapshot, relative.TranscriptPath, events)
	if err != nil {
		return nil, err
	}

	changes := []workflowstate.FileChange{{Path: relative.SessionPath, Before: snapshot.SessionHash, After: []byte(content)}, {Path: relative.TranscriptPath, Before: snapshot.TranscriptHash, After: transcript}}

	return append(changes, artifacts...), nil
}

func (r *fileAuthoringSessionRepository) prepareAuthoringContent(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, session *AuthoringSession, draft AuthoringDraft) ([]workflowstate.FileChange, []AssistedTranscriptEvent, error) {
	request := binding.Request

	var (
		changes []workflowstate.FileChange
		events  []AssistedTranscriptEvent
	)

	if request.Action == "create" {
		capture := AuthoringContinuationArtifact(request.Namespace, request.SessionID, AuthoringArtifactKindDump, true)
		session.Inputs = []AuthoringArtifactRef{capture}
		raw := renderAuthoringCapture(*session, request.Capture, request.OriginalInstruction)

		prepared, err := r.authoringArtifactChanges(binding, snapshot, capture, raw)
		if err != nil {
			return nil, nil, err
		}

		if request.Target != AuthoringArtifactKindDump || draft.Question != "" {
			changes = append(changes, prepared...)
		}
	}

	if request.Action == "answer" {
		return changes, []AssistedTranscriptEvent{authoringAnswerEvent(snapshot, request, session.Revision)}, nil
	}

	if request.Action == "respond" {
		events = append(events, authoringAnswerEvent(snapshot, request, session.Revision))
	}

	if draft.Question != "" {
		event := authoringQuestionEvent(snapshot, session, draft)
		if request.Target == AuthoringArtifactKindQuiz {
			prepared, err := r.prepareAuthoringOutput(ctx, binding, snapshot, session, renderAuthoringQuiz(*session, event))

			return append(changes, prepared...), append(events, event), err
		}

		return changes, append(events, event), nil
	}

	prepared, err := r.prepareAuthoringOutput(ctx, binding, snapshot, session, draft.Body)

	return append(changes, prepared...), events, err
}

func (r *fileAuthoringSessionRepository) authoringTranscriptImage(binding authoringRequestBinding, snapshot AuthoringSnapshot, path string, events []AssistedTranscriptEvent) ([]byte, error) {
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

func authoringQuestionEvent(snapshot AuthoringSnapshot, session *AuthoringSession, draft AuthoringDraft) AssistedTranscriptEvent {
	index := 1

	for _, event := range snapshot.Events {
		if event.Type == AssistedTranscriptEventQuestion {
			index++
		}
	}

	turn := fmt.Sprintf("turn-%d", index)

	return AssistedTranscriptEvent{ID: turn + "-question", SessionID: session.ID, TurnID: turn, TurnIndex: index, Revision: session.Revision, Type: AssistedTranscriptEventQuestion, Intent: AssistedAuthoringIntentClarify, Status: AssistedAuthoringTurnAsked, Question: AssistedAuthoringQuestion{ID: fmt.Sprintf("question-%d", index), Prompt: draft.Question, Rationale: draft.Rationale, Required: true}, CreatedAt: session.UpdatedAt}
}

func authoringAnswerEvent(snapshot AuthoringSnapshot, request AuthoringContinuationRequest, revision int64) AssistedTranscriptEvent {
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

func renderAuthoringQuiz(session AuthoringSession, event AssistedTranscriptEvent) string {
	result := AuthoringQuizGenerationResult{Provider: "configured", Strategy: "bounded_continuation", Questions: []AuthoringQuizQuestion{{ID: event.Question.ID, Text: event.Question.Prompt}}}

	return RenderContextualQuizMarkdown(session, AuthoringContinuationArtifact(session.Namespace, session.ID, AuthoringArtifactKindQuiz, false), AuthoringQuizGenerationRequest{ID: event.ID, SessionID: session.ID, Topic: session.Topic, Target: session.Target}, result)
}

func renderAuthoringCapture(session AuthoringSession, capture, instruction string) string {
	body := "---\ntitle: " + opsmetadata.Quote(session.Topic+" Raw Capture") + "\nstatus: \"raw\"\ntype: \"dump_raw\"\n---\n\n# " + session.Topic + " Raw Capture\n"
	if instruction != "" {
		body += "\n## Original Instruction\n\n" + instruction + "\n"
	}

	return body + "\n## Raw Capture\n\n" + capture + "\n"
}

// AuthoringContinuationArtifact reserves a deterministic canonical artifact ref.
// Its identity is derived from the shared session, never from model output.
func AuthoringContinuationArtifact(namespace, sessionID string, kind AuthoringArtifactKind, capture bool) AuthoringArtifactRef {
	suffix := string(kind)
	role := AuthoringArtifactRoleOutput

	if capture {
		suffix = "capture"
		role = AuthoringArtifactRoleInput
	}

	id := "00000-" + opscore.ArtifactSlugOrDefault(sessionID+"-"+suffix)

	return AuthoringArtifactRef{ID: id, Kind: kind, Role: role, Path: filepath.ToSlash(filepath.Join(opscore.DirectoryName, namespace, authoringArtifactType(kind), id, opscore.ArtifactBodyFile())), Description: "Shared authoring " + suffix + " artifact."}
}

func (r *fileAuthoringSessionRepository) prepareAuthoringOutput(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, session *AuthoringSession, body string) ([]workflowstate.FileChange, error) {
	ref, changes, err := r.prepareAuthoringOutputChanges(ctx, binding, snapshot, session, body)
	if err != nil {
		return nil, err
	}

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
	relation := AuthoringProvenanceRelationGenerated

	if binding.Request.Action == "refine" {
		source = ref
		relation = AuthoringProvenanceRelationRefined
	}

	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{Relation: relation, Source: source, Target: ref, Decision: "Bound authoring request " + binding.Request.Binding.RequestID, RecordedAt: session.UpdatedAt})

	return changes, nil
}

func (r *fileAuthoringSessionRepository) prepareAuthoringOutputChanges(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, session *AuthoringSession, body string) (AuthoringArtifactRef, []workflowstate.FileChange, error) {
	if binding.Request.Target == AuthoringArtifactKindDump {
		return r.prepareAuthoringDump(ctx, binding, snapshot, *session, body)
	}

	ref := AuthoringContinuationArtifact(session.Namespace, session.ID, binding.Request.Target, false)
	changes, err := r.authoringArtifactChanges(binding, snapshot, ref, body)

	return ref, changes, err
}
