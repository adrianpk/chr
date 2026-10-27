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
	for _, mode := range []authoringcore.AuthoringSessionMode{
		authoringcore.AuthoringSessionModeMarkdownDirect,
		authoringcore.AuthoringSessionModeContextual,
		authoringcore.AuthoringSessionModeAssisted,
		authoringcore.AuthoringSessionModeTUI,
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
	session authoringcore.AuthoringSession
} {
	return []struct {
		name    string
		session authoringcore.AuthoringSession
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
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.ID = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "id",
		session: session,
	}
}

func missingAuthoringSessionProjectIDTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.ProjectID = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "project id",
		session: session,
	}
}

func missingAuthoringSessionTopicTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Topic = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "topic",
		session: session,
	}
}

func missingAuthoringSessionCreatedAtTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.CreatedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "created at",
		session: session,
	}
}

func missingAuthoringSessionUpdatedAtTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.UpdatedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.AuthoringSession
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
	session authoringcore.AuthoringSession
} {
	return []struct {
		name    string
		session authoringcore.AuthoringSession
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
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Status = authoringcore.AuthoringSessionStatus("unknown")

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "status",
		session: session,
	}
}

func invalidAuthoringSessionModeTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Mode = authoringcore.AuthoringSessionMode("unknown")

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "mode",
		session: session,
	}
}

func invalidAuthoringSessionTargetTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Target = authoringcore.AuthoringTarget("unknown")

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "target",
		session: session,
	}
}

func invalidAuthoringInputKindTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Kind = authoringcore.AuthoringArtifactKind("unknown")

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "input kind",
		session: session,
	}
}

func invalidAuthoringInputRoleTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Role = authoringcore.AuthoringArtifactRoleOutput

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "input role",
		session: session,
	}
}

func invalidAuthoringOutputRoleTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Outputs[0].Role = authoringcore.AuthoringArtifactRoleInput

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "output role",
		session: session,
	}
}

func invalidAuthoringProvenanceRelationTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].Relation = authoringcore.AuthoringProvenanceRelation("unknown")

	return struct {
		name    string
		session authoringcore.AuthoringSession
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
	session authoringcore.AuthoringSession
} {
	return []struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		missingAuthoringInputIDTest(),
		missingAuthoringInputPathTest(),
		missingAuthoringProvenanceSourcePathTest(),
		missingAuthoringProvenanceRecordedAtTest(),
	}
}

func missingAuthoringInputIDTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].ID = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "input id",
		session: session,
	}
}

func missingAuthoringInputPathTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Path = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "input path",
		session: session,
	}
}

func missingAuthoringProvenanceSourcePathTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].Source.Path = ""

	return struct {
		name    string
		session authoringcore.AuthoringSession
	}{
		name:    "provenance source path",
		session: session,
	}
}

func missingAuthoringProvenanceRecordedAtTest() struct {
	name    string
	session authoringcore.AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].RecordedAt = time.Time{}

	return struct {
		name    string
		session authoringcore.AuthoringSession
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

func validAuthoringSession() authoringcore.AuthoringSession {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	input := authoringcore.AuthoringArtifactRef{
		ID:          "dump-1",
		Kind:        authoringcore.AuthoringArtifactKindDump,
		Role:        authoringcore.AuthoringArtifactRoleInput,
		Path:        ".chr/default/dumps/auth-context.md",
		Description: "Initial operator context dump.",
	}
	output := authoringcore.AuthoringArtifactRef{
		ID:          "quiz-1",
		Kind:        authoringcore.AuthoringArtifactKindQuiz,
		Role:        authoringcore.AuthoringArtifactRoleOutput,
		Path:        ".chr/default/quizzes/auth-context-quiz.md",
		Description: "Questionnaire generated from project context.",
	}

	return authoringcore.AuthoringSession{
		ID:        "auth-20260521-120000",
		ProjectID: "project-1",
		Status:    authoringcore.AuthoringSessionStatusDraft,
		Mode:      authoringcore.AuthoringSessionModeContextual,
		Target:    authoringcore.AuthoringTargetQuiz,
		Topic:     "Markdown-first authoring workflow",
		Inputs:    []authoringcore.AuthoringArtifactRef{input},
		Outputs:   []authoringcore.AuthoringArtifactRef{output},
		Provenance: []authoringcore.AuthoringProvenanceLink{
			{
				Relation:   authoringcore.AuthoringProvenanceRelationGenerated,
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
