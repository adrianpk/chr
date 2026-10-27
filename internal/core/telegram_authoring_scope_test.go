package core

import (
	"errors"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramAuthoringScope(t *testing.T) AuthoringWorkScope {
	t.Helper()

	digest := contenthash.Sum([]byte("canonical source"))

	scope, err := SealAuthoringWorkScope(AuthoringWorkScope{Binding: WorkflowMutationBinding{RequestID: "description", Operator: "maintainer"}, Intent: AuthoringIntentDevelop, Instruction: "Develop the bounded retry fix.", ProjectID: "example", Namespace: "team", SessionID: "session", SessionRef: ".chr/authoring/sessions/session/session.md", SessionRevision: 3, SourceHash: digest, DescriptionRef: "ops/team/spec/00000-source/artifact.md", DescriptionHash: digest, ContractHash: digest, RecordHash: digest})
	if err != nil {
		t.Fatal(err)
	}

	return scope
}

// TestTelegramAuthoringScopeAuthority binds development to the displayed scope
// and operator; preparing or explaining cannot acquire authority from a proposal.
func TestTelegramAuthoringScopeAuthority(t *testing.T) {
	scope := telegramAuthoringScope(t)
	decision := WorkflowMutationBinding{RequestID: "explicit-decision", Operator: "maintainer"}

	if err := ValidateAuthoringWorkAuthorization(scope, nil); !errors.Is(err, ErrAuthoringAuthorityRequired) {
		t.Fatalf("missing authority: %v", err)
	}

	authority, err := AuthorizeAuthoringWork(scope, decision, scope.Hash)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		change func(*AuthoringWorkScope, *AuthoringWorkAuthorization)
	}{
		{"instruction", func(s *AuthoringWorkScope, _ *AuthoringWorkAuthorization) {
			s.Instruction += " Also redesign scheduling."
		}},
		{"source", func(s *AuthoringWorkScope, _ *AuthoringWorkAuthorization) {
			s.SourceHash = contenthash.Sum([]byte("changed"))
		}},
		{"scope", func(_ *AuthoringWorkScope, a *AuthoringWorkAuthorization) {
			a.ScopeHash = contenthash.Sum([]byte("changed"))
		}},
		{"operator", func(_ *AuthoringWorkScope, a *AuthoringWorkAuthorization) { a.Binding.Operator = "other" }},
		{"prepare", func(s *AuthoringWorkScope, _ *AuthoringWorkAuthorization) { s.Intent = AuthoringIntentPrepare }},
		{"explain", func(s *AuthoringWorkScope, _ *AuthoringWorkAuthorization) { s.Intent = AuthoringIntentExplain }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed, grant := scope, authority
			tt.change(&changed, &grant)

			if err := ValidateAuthoringWorkAuthorization(changed, &grant); err == nil {
				t.Fatal("changed scope or authority accepted")
			}
		})
	}

	changed := scope
	changed.SourceHash = contenthash.Sum([]byte("new requirements"))

	changed, err = SealAuthoringWorkScope(changed)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateAuthoringWorkAuthorization(changed, &authority); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("old decision accepted a refreshed scope: %v", err)
	}

	if _, err := AuthorizeAuthoringWork(changed, decision, changed.Hash); err != nil {
		t.Fatal("explicit changed-scope decision rejected", err)
	}
}

// TestTelegramAuthoringScopePreparation permits artifact/work preparation
// without treating it as development authority.
func TestTelegramAuthoringScopePreparation(t *testing.T) {
	scope := telegramAuthoringScope(t)
	scope.Intent = AuthoringIntentPrepare

	scope, err := SealAuthoringWorkScope(scope)
	if err != nil || ValidateAuthoringWorkAuthorization(scope, nil) != nil {
		t.Fatal("preparation unexpectedly needs execution authority", err)
	}
}

// TestTelegramAuthoringInstructionReference reopens the real SQLite store and
// distinguishes the original explicit development command from a model proposal.
func TestTelegramAuthoringInstructionReference(t *testing.T) {
	tests := []struct {
		name, text, intent string
		authorized         bool
	}{
		{"develop", "/develop Keep retries bounded", "develop", true},
		{"prepare", "/prepare Keep retries bounded", "prepare", false},
		{"explain", "/explain How are retries bounded?", "explain", false},
		{"proposal", "Explain how retries should work", "develop", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) {
				input.Project, input.Intent = "alpha", tt.intent
				input.Update = telegramRetainedTestUpdate(t, input.Update.UpdateID, tt.text)
			})

			before := f.sessions(t)[0]
			if (before.AuthorizationRef != "") != tt.authorized || before.Intent != tt.intent {
				t.Fatalf("instruction binding=%+v", before)
			}

			if before.AuthorizationRef != "" && !strings.HasPrefix(before.AuthorizationRef, "instruction-") {
				t.Fatal("nonopaque instruction reference", before.AuthorizationRef)
			}

			if err := f.store.Close(); err != nil {
				t.Fatal(err)
			}

			f.store = openTelegramReceiverTestStore(t, f.env)
			if after := f.sessions(t)[0]; after != before {
				t.Fatalf("instruction binding lost on reopening: %+v", after)
			}

			telegramAuthoringCheckInstruction(t, f, before, tt.text)
		})
	}
}

func telegramAuthoringCheckInstruction(t *testing.T, f *telegramConversationFixture, c TelegramConversation, original string) {
	t.Helper()

	input := f.input
	input.Conversation, input.Project = c.ID, c.ProjectID
	input.Update = telegramRetainedTestUpdate(t, 100, "Inspect original instruction")

	retained, err := f.store.ReadTelegramConversationInstruction(t.Context(), input)
	if err != nil || retained.Text != original || retained.DevelopmentRef != c.AuthorizationRef {
		t.Fatalf("original instruction changed: %+v err=%v", retained, err)
	}

	input.Config.AllowedUsers[0].OperatorID = "replacement-operator"

	changed, err := f.store.ReadTelegramConversationInstruction(t.Context(), input)
	if err != nil || changed.Text != original || changed.DevelopmentRef != "" {
		t.Fatalf("operator remapping reused authority: %+v err=%v", changed, err)
	}
}
