package core

import (
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func TestValidateAuthoringSessionAcceptsValidSession(t *testing.T) {
	session := validAuthoringSession()

	err := ValidateAuthoringSession(session)
	if err != nil {
		t.Fatalf("ValidateAuthoringSession returned error: %v", err)
	}
}

func TestValidateAuthoringSessionAcceptsFutureCompatibleModes(t *testing.T) {
	for _, mode := range []authoringcore.SessionMode{
		authoringcore.SessionModeMarkdownDirect,
		authoringcore.SessionModeContextual,
		authoringcore.SessionModeAssisted,
		authoringcore.SessionModeTUI,
	} {
		t.Run(string(mode), func(t *testing.T) {
			session := validAuthoringSession()
			session.Mode = mode

			err := ValidateAuthoringSession(session)
			if err != nil {
				t.Fatalf("ValidateAuthoringSession returned error: %v", err)
			}
		})
	}
}

func TestValidateAuthoringSessionRejectsMissingRequiredFields(t *testing.T) {
	for _, test := range missingRequiredAuthoringSessionTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAuthoringSession(test.session)
			if err == nil {
				t.Fatal("ValidateAuthoringSession returned nil error")
			}
		})
	}
}

func missingRequiredAuthoringSessionTests() []struct {
	name    string
	session authoringcore.Session
} {
	return []struct {
		name    string
		session authoringcore.Session
	}{
		missingAuthoringSessionIDTest(),
		missingAuthoringSessionProjectIDTest(),
		missingAuthoringSessionTopicTest(),
		missingAuthoringSessionCreatedAtTest(),
		missingAuthoringSessionUpdatedAtTest(),
	}
}

func missingAuthoringSessionIDTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.ID = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "id",
		session: session,
	}
}

func missingAuthoringSessionProjectIDTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.ProjectID = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "project id",
		session: session,
	}
}

func missingAuthoringSessionTopicTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Topic = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "topic",
		session: session,
	}
}

func missingAuthoringSessionCreatedAtTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.CreatedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "created at",
		session: session,
	}
}

func missingAuthoringSessionUpdatedAtTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.UpdatedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "updated at",
		session: session,
	}
}

func TestValidateAuthoringSessionRejectsInvalidEnums(t *testing.T) {
	for _, test := range invalidAuthoringSessionEnumTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAuthoringSession(test.session)
			if err == nil {
				t.Fatal("ValidateAuthoringSession returned nil error")
			}
		})
	}
}

func invalidAuthoringSessionEnumTests() []struct {
	name    string
	session authoringcore.Session
} {
	return []struct {
		name    string
		session authoringcore.Session
	}{
		invalidAuthoringSessionStatusTest(),
		invalidAuthoringSessionModeTest(),
		invalidAuthoringSessionTargetTest(),
		invalidAuthoringInputKindTest(),
		invalidAuthoringInputRoleTest(),
		invalidAuthoringOutputRoleTest(),
		invalidAuthoringProvenanceRelationTest(),
	}
}

func invalidAuthoringSessionStatusTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Status = authoringcore.SessionStatus("unknown")

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "status",
		session: session,
	}
}

func invalidAuthoringSessionModeTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Mode = authoringcore.SessionMode("unknown")

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "mode",
		session: session,
	}
}

func invalidAuthoringSessionTargetTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Target = authoringcore.Target("unknown")

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "target",
		session: session,
	}
}

func invalidAuthoringInputKindTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Inputs[0].Kind = authoringcore.ArtifactKind("unknown")

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "input kind",
		session: session,
	}
}

func invalidAuthoringInputRoleTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Inputs[0].Role = authoringcore.ArtifactRoleOutput

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "input role",
		session: session,
	}
}

func invalidAuthoringOutputRoleTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Outputs[0].Role = authoringcore.ArtifactRoleInput

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "output role",
		session: session,
	}
}

func invalidAuthoringProvenanceRelationTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Provenance[0].Relation = authoringcore.ProvenanceRelation("unknown")

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "provenance relation",
		session: session,
	}
}

func TestValidateAuthoringSessionRejectsInvalidArtifactRefs(t *testing.T) {
	for _, test := range invalidAuthoringArtifactRefTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAuthoringSession(test.session)
			if err == nil {
				t.Fatal("ValidateAuthoringSession returned nil error")
			}
		})
	}
}

func invalidAuthoringArtifactRefTests() []struct {
	name    string
	session authoringcore.Session
} {
	return []struct {
		name    string
		session authoringcore.Session
	}{
		missingAuthoringInputIDTest(),
		missingAuthoringInputPathTest(),
		missingAuthoringProvenanceSourcePathTest(),
		missingAuthoringProvenanceRecordedAtTest(),
	}
}

func missingAuthoringInputIDTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Inputs[0].ID = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "input id",
		session: session,
	}
}

func missingAuthoringInputPathTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Inputs[0].Path = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "input path",
		session: session,
	}
}

func missingAuthoringProvenanceSourcePathTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Provenance[0].Source.Path = ""

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "provenance source path",
		session: session,
	}
}

func missingAuthoringProvenanceRecordedAtTest() struct {
	name    string
	session authoringcore.Session
} {
	session := validAuthoringSession()
	session.Provenance[0].RecordedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.Session
	}{
		name:    "provenance recorded at",
		session: session,
	}
}

func TestValidateAuthoringSessionRejectsFinishedBeforeStarted(t *testing.T) {
	session := validAuthoringSession()
	startedAt := session.CreatedAt
	finishedAt := startedAt.Add(-time.Minute)
	session.StartedAt = &startedAt
	session.FinishedAt = &finishedAt

	err := ValidateAuthoringSession(session)
	if err == nil {
		t.Fatal("ValidateAuthoringSession returned nil error")
	}
}

func validAuthoringSession() authoringcore.Session {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	input := authoringcore.ArtifactRef{
		ID:          "dump-1",
		Kind:        authoringcore.ArtifactKindDump,
		Role:        authoringcore.ArtifactRoleInput,
		Path:        ".chr/default/dumps/auth-context.md",
		Description: "Initial operator context dump.",
	}
	output := authoringcore.ArtifactRef{
		ID:          "quiz-1",
		Kind:        authoringcore.ArtifactKindQuiz,
		Role:        authoringcore.ArtifactRoleOutput,
		Path:        ".chr/default/quizzes/auth-context-quiz.md",
		Description: "Questionnaire generated from project context.",
	}

	return authoringcore.Session{
		ID:        "auth-20260521-120000",
		ProjectID: "project-1",
		Status:    authoringcore.SessionStatusDraft,
		Mode:      authoringcore.SessionModeContextual,
		Target:    authoringcore.TargetQuiz,
		Topic:     "Markdown-first authoring workflow",
		Inputs:    []authoringcore.ArtifactRef{input},
		Outputs:   []authoringcore.ArtifactRef{output},
		Provenance: []authoringcore.ProvenanceLink{
			{
				Relation:   authoringcore.ProvenanceRelationGenerated,
				Source:     input,
				Target:     output,
				Decision:   "Generate a quiz from project context.",
				RecordedAt: now,
			},
		},
		CreatedAt:   now,
		UpdatedAt:   now,
		Description: "Capture the initial static questionnaire workflow.",
	}
}
