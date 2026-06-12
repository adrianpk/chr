package workspace

import (
	"testing"
	"time"
)

func TestValidateAuthoringSessionAcceptsValidSession(t *testing.T) {
	session := validAuthoringSession()

	err := ValidateAuthoringSession(session)
	if err != nil {
		t.Fatalf("ValidateAuthoringSession returned error: %v", err)
	}
}

func TestValidateAuthoringSessionAcceptsFutureCompatibleModes(t *testing.T) {
	for _, mode := range []AuthoringSessionMode{
		AuthoringSessionModeMarkdownDirect,
		AuthoringSessionModeContextual,
		AuthoringSessionModeAssisted,
		AuthoringSessionModeTUI,
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
	session AuthoringSession
} {
	return []struct {
		name    string
		session AuthoringSession
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
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.ID = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "id",
		session: session,
	}
}

func missingAuthoringSessionProjectIDTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.ProjectID = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "project id",
		session: session,
	}
}

func missingAuthoringSessionTopicTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Topic = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "topic",
		session: session,
	}
}

func missingAuthoringSessionCreatedAtTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.CreatedAt = time.Time{}

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "created at",
		session: session,
	}
}

func missingAuthoringSessionUpdatedAtTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.UpdatedAt = time.Time{}

	return struct {
		name    string
		session AuthoringSession
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
	session AuthoringSession
} {
	return []struct {
		name    string
		session AuthoringSession
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
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Status = AuthoringSessionStatus("unknown")

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "status",
		session: session,
	}
}

func invalidAuthoringSessionModeTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Mode = AuthoringSessionMode("unknown")

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "mode",
		session: session,
	}
}

func invalidAuthoringSessionTargetTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Target = AuthoringTarget("unknown")

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "target",
		session: session,
	}
}

func invalidAuthoringInputKindTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Kind = AuthoringArtifactKind("unknown")

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "input kind",
		session: session,
	}
}

func invalidAuthoringInputRoleTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Role = AuthoringArtifactRoleOutput

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "input role",
		session: session,
	}
}

func invalidAuthoringOutputRoleTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Outputs[0].Role = AuthoringArtifactRoleInput

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "output role",
		session: session,
	}
}

func invalidAuthoringProvenanceRelationTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].Relation = AuthoringProvenanceRelation("unknown")

	return struct {
		name    string
		session AuthoringSession
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
	session AuthoringSession
} {
	return []struct {
		name    string
		session AuthoringSession
	}{
		missingAuthoringInputIDTest(),
		missingAuthoringInputPathTest(),
		missingAuthoringProvenanceSourcePathTest(),
		missingAuthoringProvenanceRecordedAtTest(),
	}
}

func missingAuthoringInputIDTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].ID = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "input id",
		session: session,
	}
}

func missingAuthoringInputPathTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Inputs[0].Path = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "input path",
		session: session,
	}
}

func missingAuthoringProvenanceSourcePathTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].Source.Path = ""

	return struct {
		name    string
		session AuthoringSession
	}{
		name:    "provenance source path",
		session: session,
	}
}

func missingAuthoringProvenanceRecordedAtTest() struct {
	name    string
	session AuthoringSession
} {
	session := validAuthoringSession()
	session.Provenance[0].RecordedAt = time.Time{}

	return struct {
		name    string
		session AuthoringSession
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

func validAuthoringSession() AuthoringSession {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	input := AuthoringArtifactRef{
		ID:          "dump-1",
		Kind:        AuthoringArtifactKindDump,
		Role:        AuthoringArtifactRoleInput,
		Path:        ".chc/user/dumps/auth-context.md",
		Description: "Initial operator context dump.",
	}
	output := AuthoringArtifactRef{
		ID:          "quiz-1",
		Kind:        AuthoringArtifactKindQuiz,
		Role:        AuthoringArtifactRoleOutput,
		Path:        ".chc/user/quizzes/auth-context-quiz.md",
		Description: "Questionnaire generated from project context.",
	}

	return AuthoringSession{
		ID:        "auth-20260521-120000",
		ProjectID: "project-1",
		Status:    AuthoringSessionStatusDraft,
		Mode:      AuthoringSessionModeContextual,
		Target:    AuthoringTargetQuiz,
		Topic:     "Markdown-first authoring workflow",
		Inputs:    []AuthoringArtifactRef{input},
		Outputs:   []AuthoringArtifactRef{output},
		Provenance: []AuthoringProvenanceLink{
			{
				Relation:   AuthoringProvenanceRelationGenerated,
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
