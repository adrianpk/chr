package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// TelegramConversationAction enumerates local routing and capture operations.
type TelegramConversationAction string

// Conversation actions change only receiver-owned state.
const (
	TelegramConversationSelect  TelegramConversationAction = "select"
	TelegramConversationStart   TelegramConversationAction = "start"
	TelegramConversationFocus   TelegramConversationAction = "focus"
	TelegramConversationList    TelegramConversationAction = "list"
	TelegramConversationCapture TelegramConversationAction = "capture"
	TelegramConversationDone    TelegramConversationAction = "done"
	TelegramConversationCancel  TelegramConversationAction = "cancel"
	TelegramConversationShow    TelegramConversationAction = "show"
	TelegramConversationAuthor  TelegramConversationAction = "author"
)

// TelegramConversationInput contains current authority and a deterministic command.
// Projects must be read from the current registry through the application service.
type TelegramConversationInput struct {
	Update        TelegramReceivedUpdate
	Config        TelegramReceiverConfig
	Projects      []Project
	Action        TelegramConversationAction
	Project       string
	Conversation  string
	Namespace     string
	Intent        string
	Text          string
	ContextDigest string
}

// TelegramConversation freezes routing; it does not grant execution authority.
type TelegramConversation struct {
	ID                   string
	Peer                 TelegramPeerIdentity
	ProjectID            string
	ProjectBindingDigest string
	Namespace            string
	Intent               string
	Phase                string
	Topic                string
	Capture              string
	QuestionBinding      string
	AuthoringSessionRef  string
	AuthorizationRef     string
	InvocationID         string
	LastObservedRevision uint64
	Revision             int64
}

type telegramConversationDiagnostic string

func (e telegramConversationDiagnostic) Error() string { return string(e) }

// TelegramConversationFailure returns only a domain-owned safe routing diagnostic.
func TelegramConversationFailure(err error) (string, bool) {
	var diagnostic telegramConversationDiagnostic
	if errors.As(err, &diagnostic) {
		return string(diagnostic), true
	}

	return "", false
}

type telegramConversationTransaction struct {
	ctx   context.Context
	tx    *sql.Tx
	input TelegramConversationInput
	now   time.Time
}

// ApplyTelegramConversation acknowledges a retained source, changes local state
// and queues its response in one transaction. Diagnostics preserve existing state.
func (s *TelegramReceiverStore) ApplyTelegramConversation(ctx context.Context, input TelegramConversationInput, now time.Time) error {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return s.RejectPendingUpdate(ctx, input.Update, now)
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, input.Update)
		if err != nil || !pending {
			return err
		}

		operation := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := operation.ensurePeer(); err != nil {
			return err
		}

		text, conversation, err := operation.apply()
		if err != nil {
			return err
		}

		if err := queueTelegramResponse(ctx, tx, input.Update, BoundTelegramReceiverText(text), input.Config.Digest(), now); err != nil {
			return err
		}

		return operation.bindResponse(conversation)
	})

	var diagnostic telegramConversationDiagnostic
	if errors.As(err, &diagnostic) {
		return s.QueueReadResponse(ctx, input.Update, BoundTelegramReceiverText(string(diagnostic)), input.Config.Digest(), now)
	}

	return err
}

func (op telegramConversationTransaction) ensurePeer() error {
	p := op.input.Update.Peer

	_, err := op.tx.ExecContext(op.ctx, `INSERT INTO telegram_peers(bot_id,chat_id,topic_id,user_id,updated_at) VALUES(?,?,?,?,?) ON CONFLICT DO NOTHING`, p.BotID, p.ChatID, p.TopicID, p.UserID, formatProjectTime(op.now))
	if err != nil || op.input.ContextDigest == "" {
		return err
	}

	snapshot, err := op.intentSnapshot()
	if err != nil {
		return err
	}

	if snapshot.Digest != op.input.ContextDigest {
		return ErrWorkflowConflict
	}

	return nil
}

func (op telegramConversationTransaction) apply() (string, *TelegramConversation, error) {
	switch op.input.Action {
	case TelegramConversationSelect:
		return op.selectProject()
	case TelegramConversationStart:
		return op.start()
	case TelegramConversationFocus:
		return op.focus()
	case TelegramConversationList:
		return op.list()
	default:
		return op.applyCapture()
	}
}

func (op telegramConversationTransaction) bindResponse(conversation *TelegramConversation) error {
	if conversation == nil {
		return nil
	}

	key := fmt.Sprintf("update:%d", op.input.Update.UpdateID)
	_, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_outbox SET conversation_id=?,conversation_revision=?,question_binding=? WHERE bot_id=? AND event_key=?`, conversation.ID, conversation.Revision, conversation.QuestionBinding, op.input.Update.Peer.BotID, key)

	return err
}

func (op telegramConversationTransaction) allowedProjects() []Project {
	user, _ := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)

	var projects []Project

	for _, project := range op.input.Projects {
		if project.Enabled && slices.Contains(user.ProjectIDs, project.ID) {
			projects = append(projects, project)
		}
	}

	return projects
}

func exactTelegramProject(projects []Project, target string) (Project, error) {
	for _, project := range projects {
		if project.ID == target {
			return project, nil
		}
	}

	var matches []Project

	for _, project := range projects {
		if project.Name == target {
			matches = append(matches, project)
		}
	}

	if len(matches) != 1 {
		return Project{}, telegramConversationDiagnostic("Select an exact authorized project ID or unique registered name with /project <project>.")
	}

	return matches[0], nil
}

func telegramProjectBinding(project Project) (string, error) {
	root, err := filepath.Abs(project.Path)
	if err != nil {
		return "", err
	}

	registeredRoot := root

	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", telegramConversationDiagnostic("The registered project root is unavailable. Resolve the registry binding before continuing.")
	}

	sum := sha256.Sum256([]byte(project.ID + "\x00" + registeredRoot + "\x00" + root))

	return hex.EncodeToString(sum[:]), nil
}

func (op telegramConversationTransaction) resolveProject(bound *TelegramConversation) (Project, error) {
	target := op.input.Project
	if target != "" {
		project, err := exactTelegramProject(op.allowedProjects(), target)
		if err != nil {
			return Project{}, err
		}

		if bound != nil && bound.ProjectID != project.ID {
			return Project{}, telegramConversationDiagnostic("The explicit project conflicts with the bound conversation. Choose a matching target.")
		}

		return project, nil
	}

	if bound != nil {
		return exactTelegramProject(op.allowedProjects(), bound.ProjectID)
	}

	selected, _, err := op.peerSelection()
	if err != nil {
		return Project{}, err
	}

	if selected != "" {
		return exactTelegramProject(op.allowedProjects(), selected)
	}

	return op.defaultProject()
}

func (op telegramConversationTransaction) defaultProject() (Project, error) {
	projects := op.allowedProjects()
	for _, project := range projects {
		if project.ID == op.input.Config.DefaultProjectID {
			return project, nil
		}
	}

	if len(projects) == 1 {
		return projects[0], nil
	}

	return Project{}, telegramConversationDiagnostic("Project selection is ambiguous. Use /projects then /project <project>.")
}

func (op telegramConversationTransaction) peerSelection() (string, string, error) {
	p := op.input.Update.Peer

	var selected, focused string

	err := op.tx.QueryRowContext(op.ctx, `SELECT COALESCE(selected_project_id,''),COALESCE(focused_conversation_id,'') FROM telegram_peers WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, p.BotID, p.ChatID, p.TopicID, p.UserID).Scan(&selected, &focused)

	return selected, focused, err
}

func (op telegramConversationTransaction) selectProject() (string, *TelegramConversation, error) {
	bound, err := op.explicitTarget()
	if err != nil {
		return "", nil, err
	}

	project, err := op.resolveProject(bound)
	if err != nil {
		return "", nil, err
	}

	if op.input.Project == "" {
		return "Selected project: " + project.ID, nil, nil
	}

	p := op.input.Update.Peer
	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_peers SET selected_project_id=?,revision=revision+1,updated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, project.ID, formatProjectTime(op.now), p.BotID, p.ChatID, p.TopicID, p.UserID)

	return "Selected project: " + project.ID, nil, err
}

func (op telegramConversationTransaction) start() (string, *TelegramConversation, error) {
	bound, err := op.explicitTarget()
	if err != nil {
		return "", nil, err
	}

	project, err := op.resolveProject(bound)
	if err != nil {
		return "", nil, err
	}

	binding, err := telegramProjectBinding(project)
	if err != nil {
		return "", nil, err
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: project.Path, Namespace: op.input.Namespace})
	if err != nil {
		return "", nil, err
	}

	if err := op.checkAllocation(); err != nil {
		return "", nil, err
	}

	return op.insertConversation(project, binding, layout.Namespace())
}

func (op telegramConversationTransaction) checkAllocation() error {
	p := op.input.Update.Peer

	var peerCount, botCount int

	err := op.tx.QueryRowContext(op.ctx, `SELECT COUNT(*),COALESCE(SUM(chat_id=? AND topic_id=? AND user_id=?),0) FROM telegram_conversations WHERE bot_id=? AND phase NOT IN ('completed','canceled','failed')`, p.ChatID, p.TopicID, p.UserID, p.BotID).Scan(&botCount, &peerCount)
	if err != nil {
		return err
	}

	if peerCount >= 8 || botCount >= 128 {
		return telegramConversationDiagnostic("Conversation capacity reached. Finish or cancel an existing session before creating another.")
	}

	return nil
}

func (op telegramConversationTransaction) insertConversation(project Project, binding, namespace string) (string, *TelegramConversation, error) {
	if !slices.Contains([]string{"explain", "prepare", "develop"}, op.input.Intent) || len(op.input.Text) > 512 || strings.TrimSpace(op.input.Text) == "" {
		return "", nil, telegramConversationDiagnostic("A conversation requires an intent and a topic of at most 512 UTF-8 bytes.")
	}

	c := TelegramConversation{ID: telegramConversationID(op.input.Update), Peer: op.input.Update.Peer, ProjectID: project.ID, ProjectBindingDigest: binding, Namespace: namespace, Intent: op.input.Intent, Phase: "collecting", Topic: op.input.Text, Revision: 1}
	c.AuthorizationRef = op.developmentInstructionRef(c)
	_, err := op.tx.ExecContext(op.ctx, `INSERT INTO telegram_conversations(id,bot_id,chat_id,topic_id,user_id,project_id,project_binding_digest,namespace,intent,phase,topic,authorization_ref,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, c.ID, c.Peer.BotID, c.Peer.ChatID, c.Peer.TopicID, c.Peer.UserID, c.ProjectID, c.ProjectBindingDigest, c.Namespace, c.Intent, c.Phase, c.Topic, c.AuthorizationRef, formatProjectTime(op.now), formatProjectTime(op.now))

	return "Collecting session " + c.ID + " for " + c.ProjectID + " (" + c.Intent + "). Reply with text; /done finishes capture. Use /sessions or /focus <conversation> to target it.", &c, err
}

func telegramConversationID(update TelegramReceivedUpdate) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%d:%d", update.Peer.BotID, update.UpdateID))

	return "tgc-" + hex.EncodeToString(sum[:16])
}

func (op telegramConversationTransaction) developmentInstructionRef(c TelegramConversation) string {
	command := strings.Fields(op.input.Update.Text)
	if c.Intent != string(AuthoringIntentDevelop) || len(command) < 2 || command[0] != "/develop" {
		return ""
	}

	user, allowed := op.input.Config.AuthorizedUser(c.Peer.ChatID, c.Peer.TopicID, c.Peer.UserID, op.input.Update.ChatType)
	if !allowed || user.OperatorID == "" {
		return ""
	}

	// A model proposal cannot grant authority. Only the original explicit control
	// creates this opaque source reference; the final scope still needs binding.
	data := fmt.Appendf(nil, "%s\x00%s\x00%s\x00%s\x00%s", c.ID, user.OperatorID, c.ProjectID, c.Namespace, op.input.Update.Text)

	return "instruction-" + workflowHash(data)
}
