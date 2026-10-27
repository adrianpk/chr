package telegram

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ConversationAction enumerates local routing and capture operations.
type ConversationAction string

// Conversation actions change only receiver-owned state.
const (
	ConversationSelect  ConversationAction = "select"
	ConversationStart   ConversationAction = "start"
	ConversationFocus   ConversationAction = "focus"
	ConversationList    ConversationAction = "list"
	ConversationCapture ConversationAction = "capture"
	ConversationDone    ConversationAction = "done"
	ConversationCancel  ConversationAction = "cancel"
	ConversationShow    ConversationAction = "show"
	ConversationAuthor  ConversationAction = "author"
)

// ConversationInput contains current authority and a deterministic command.
// Projects must be read from the current registry through the application service.
type ConversationInput struct {
	Update        ReceivedUpdate
	Config        ReceiverConfig
	Projects      []projectcore.Project
	Action        ConversationAction
	Project       string
	Conversation  string
	Namespace     string
	Intent        string
	Text          string
	ContextDigest string
}

// Conversation freezes routing; it does not grant execution authority.
type Conversation struct {
	GuidedRunID          string
	ID                   string
	Peer                 PeerIdentity
	ProjectID            string
	ProjectBindingDigest string
	Namespace            string
	Intent               string
	Phase                ConversationPhase
	Topic                string
	Capture              string
	QuestionBinding      string
	AuthoringSessionRef  string
	AuthorizationRef     string
	InvocationID         string
	LastObservedRevision uint64
	Revision             int64
}

// ErrProjectSelectionAmbiguous identifies an unresolved current project choice.
var ErrProjectSelectionAmbiguous error = conversationDiagnosticError("Project selection is ambiguous. Use /projects then /project <project>.")

type conversationDiagnosticError string

func (e conversationDiagnosticError) Error() string { return string(e) }

// ConversationFailure returns only a domain-owned safe routing diagnostic.
func ConversationFailure(err error) (string, bool) {
	var diagnostic conversationDiagnosticError
	if errors.As(err, &diagnostic) {
		return string(diagnostic), true
	}

	return "", false
}

type conversationTransaction struct {
	ctx   context.Context
	tx    *sql.Tx
	input ConversationInput
	now   time.Time
}

// ApplyConversation acknowledges a retained source, changes local state
// and queues its response in one transaction. Diagnostics preserve existing state.
func (s *ReceiverStore) ApplyConversation(ctx context.Context, input ConversationInput, now time.Time) error {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return s.RejectPendingUpdate(ctx, input.Update, now)
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, input.Update)
		if err != nil || !pending {
			return err
		}

		operation := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := operation.ensurePeer(); err != nil {
			return err
		}

		conversationResult, err := operation.apply()
		text, conversation := conversationResult.Text, conversationResult.Conversation

		if err != nil {
			return err
		}

		if err := queuePresentation(ctx, tx, presentationReply{Source: input.Update, Message: operation.presentation(text, conversation), ConfigDigest: input.Config.Digest(), At: now}); err != nil {
			return err
		}

		return operation.bindResponse(conversation)
	})

	var diagnostic conversationDiagnosticError
	if errors.As(err, &diagnostic) {
		return s.QueuePresentationResponse(ctx, input.Update, tgmessage.Blocks{InitialText: string(diagnostic)}, input.Config.Digest(), now)
	}

	return err
}

func (op conversationTransaction) ensurePeer() error {
	p := op.input.Update.Peer

	_, err := op.tx.ExecContext(op.ctx, `INSERT INTO telegram_peers(bot_id,chat_id,topic_id,user_id,updated_at) VALUES(?,?,?,?,?) ON CONFLICT DO NOTHING`, p.BotID, p.ChatID, p.TopicID, p.UserID, database.FormatTime(op.now))
	if err != nil || op.input.ContextDigest == "" {
		return err
	}

	snapshot, err := op.intentSnapshot()
	if err != nil {
		return err
	}

	if snapshot.Digest != op.input.ContextDigest {
		return workflowstate.ErrConflict
	}

	return nil
}

func (op conversationTransaction) apply() (conversationReply, error) {
	switch op.input.Action {
	case ConversationSelect:
		return op.selectProject()
	case ConversationStart:
		return op.start()
	case ConversationFocus:
		return op.focus()
	case ConversationList:
		return op.list()
	default:
		return op.applyCapture()
	}
}

func (op conversationTransaction) bindResponse(conversation *Conversation) error {
	if conversation == nil {
		return nil
	}

	key := fmt.Sprintf("update:%d", op.input.Update.UpdateID)
	_, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_outbox SET conversation_id=?,conversation_revision=?,question_binding=? WHERE bot_id=? AND event_key=?`, conversation.ID, conversation.Revision, conversation.QuestionBinding, op.input.Update.Peer.BotID, key)

	return err
}

func (op conversationTransaction) allowedProjects() []projectcore.Project {
	user, _ := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)

	var projects []projectcore.Project

	for _, project := range op.input.Projects {
		if project.Enabled && slices.Contains(user.ProjectIDs, project.ID) {
			projects = append(projects, project)
		}
	}

	return projects
}

func exactProject(projects []projectcore.Project, target string) (projectcore.Project, error) {
	for _, project := range projects {
		if project.ID == target {
			return project, nil
		}
	}

	var matches []projectcore.Project

	for _, project := range projects {
		if project.Name == target {
			matches = append(matches, project)
		}
	}

	if len(matches) != 1 {
		return projectcore.Project{}, conversationDiagnosticError("That project is unavailable or its name is ambiguous. Send /projects, then select an exact ID or unique name with /project <project>.")
	}

	return matches[0], nil
}

func projectBinding(registered projectcore.Project) (string, error) {
	binding, err := projectcore.RootBindingDigest(registered)
	if errors.Is(err, projectcore.ErrRootUnavailable) {
		return "", conversationDiagnosticError("The project folder is unavailable. Saved conversations are unchanged. A local operator must restore the registered folder before you continue.")
	}

	return binding, err
}

func (op conversationTransaction) resolveProject(bound *Conversation) (projectcore.Project, error) {
	target := op.input.Project
	if target != "" {
		project, err := exactProject(op.allowedProjects(), target)
		if err != nil {
			return projectcore.Project{}, err
		}

		if bound != nil && bound.ProjectID != project.ID {
			return projectcore.Project{}, conversationDiagnosticError("The explicit project conflicts with this conversation. Saved details are unchanged. Use /sessions to choose a matching target.")
		}

		return project, nil
	}

	if bound != nil {
		return exactProject(op.allowedProjects(), bound.ProjectID)
	}

	selection, err := op.peerSelection()
	selected := selection.ProjectID

	if err != nil {
		return projectcore.Project{}, err
	}

	if selected != "" {
		return exactProject(op.allowedProjects(), selected)
	}

	return op.defaultProject()
}

func (op conversationTransaction) defaultProject() (projectcore.Project, error) {
	projects := op.allowedProjects()
	for _, project := range projects {
		if project.ID == op.input.Config.DefaultProjectID {
			return project, nil
		}
	}

	if len(projects) == 1 {
		return projects[0], nil
	}

	return projectcore.Project{}, ErrProjectSelectionAmbiguous
}

func (op conversationTransaction) peerSelection() (peerSelection, error) {
	p := op.input.Update.Peer

	var selected, focused string

	err := op.tx.QueryRowContext(op.ctx, `SELECT COALESCE(selected_project_id,''),COALESCE(focused_conversation_id,'') FROM telegram_peers WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, p.BotID, p.ChatID, p.TopicID, p.UserID).Scan(&selected, &focused)

	return peerSelection{ProjectID: selected, RootDigest: focused}, err
}

func (op conversationTransaction) selectProject() (conversationReply, error) {
	bound, err := op.explicitTarget()
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	project, err := op.resolveProject(bound)
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if op.input.Project == "" {
		return conversationReply{Text: "Selected project: " + project.Name, Conversation: nil}, nil
	}

	p := op.input.Update.Peer
	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_peers SET selected_project_id=?,revision=revision+1,updated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, project.ID, database.FormatTime(op.now), p.BotID, p.ChatID, p.TopicID, p.UserID)

	return conversationReply{Text: "Selected project: " + project.Name, Conversation: nil}, err
}

func (op conversationTransaction) start() (conversationReply, error) {
	bound, err := op.explicitTarget()
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	project, err := op.resolveProject(bound)
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	binding, err := projectBinding(project)
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if _, err := projectcore.ResolveDirectory(project.Path); err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	namespace, err := opsnamespace.Resolve(op.input.Namespace, nil)
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if err := op.checkAllocation(); err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	return op.insertConversation(project, binding, namespace)
}

func (op conversationTransaction) checkAllocation() error {
	p := op.input.Update.Peer

	var peerCount, botCount int

	err := op.tx.QueryRowContext(op.ctx, `SELECT COUNT(*),COALESCE(SUM(chat_id=? AND topic_id=? AND user_id=?),0) FROM telegram_conversations WHERE bot_id=? AND phase NOT IN ('completed','canceled','failed')`, p.ChatID, p.TopicID, p.UserID, p.BotID).Scan(&botCount, &peerCount)
	if err != nil {
		return err
	}

	if peerCount >= 8 || botCount >= 128 {
		return conversationDiagnosticError("Conversation capacity reached. Finish or cancel an existing conversation before starting another. Send /sessions to choose one.")
	}

	return nil
}

func (op conversationTransaction) insertConversation(project projectcore.Project, binding, namespace string) (conversationReply, error) {
	if !slices.Contains([]string{"explain", "prepare", "develop"}, op.input.Intent) || len(op.input.Text) > 512 || strings.TrimSpace(op.input.Text) == "" {
		return conversationReply{Text: "", Conversation: nil}, conversationDiagnosticError("Start with a nonempty topic of at most 512 UTF-8 bytes. This conversation was not created.")
	}

	c := Conversation{ID: conversationID(op.input.Update), Peer: op.input.Update.Peer, ProjectID: project.ID, ProjectBindingDigest: binding, Namespace: namespace, Intent: op.input.Intent, Phase: ConversationCollecting, Topic: op.input.Text, Revision: 1}
	c.AuthorizationRef = op.developmentInstructionRef(c)
	_, err := op.tx.ExecContext(op.ctx, `INSERT INTO telegram_conversations(id,bot_id,chat_id,topic_id,user_id,project_id,project_binding_digest,namespace,intent,phase,topic,authorization_ref,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, c.ID, c.Peer.BotID, c.Peer.ChatID, c.Peer.TopicID, c.Peer.UserID, c.ProjectID, c.ProjectBindingDigest, c.Namespace, c.Intent, c.Phase, c.Topic, c.AuthorizationRef, database.FormatTime(op.now), database.FormatTime(op.now))

	text := "Your request for " + conversationLabel(c.Topic) + " in " + conversationLabel(project.Name) + " is retained."
	text += "\n\nYou can add more details, or send /done " + c.ID + " to finish collecting them."

	return conversationReply{Text: text, Conversation: &c}, err
}

func conversationID(update ReceivedUpdate) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%d:%d", update.Peer.BotID, update.UpdateID))

	return "tgc-" + hex.EncodeToString(sum[:16])
}

func (op conversationTransaction) developmentInstructionRef(c Conversation) string {
	command := strings.Fields(op.input.Update.Text)
	if c.Intent != string(authoringstate.IntentDevelop) || len(command) < 2 || command[0] != "/develop" {
		return ""
	}

	user, allowed := op.input.Config.AuthorizedUser(c.Peer.ChatID, c.Peer.TopicID, c.Peer.UserID, op.input.Update.ChatType)
	if !allowed || user.OperatorID == "" {
		return ""
	}

	// A model proposal cannot grant authority. Only the original explicit control
	// creates this opaque source reference; the final scope still needs binding.
	data := fmt.Appendf(nil, "%s\x00%s\x00%s\x00%s\x00%s", c.ID, user.OperatorID, c.ProjectID, c.Namespace, op.input.Update.Text)

	return "instruction-" + contenthash.Sum(data)
}
