package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ResolveTelegramNotificationTarget uses the same canonical project-root binding
// as receiver views. The caller retains its existing notification authorization.
func ResolveTelegramNotificationTarget(project projectcore.Project, namespace, destination, digest string) (TelegramNotificationTarget, error) {
	binding, err := telegramProjectBinding(project)
	if err != nil {
		return TelegramNotificationTarget{}, err
	}

	if !project.Enabled || !opsnamespace.Valid(namespace) || !validTelegramDigest(digest) {
		return TelegramNotificationTarget{}, tgmessage.ErrDocumentUnavailable
	}

	return TelegramNotificationTarget{ProjectID: project.ID, Namespace: namespace, ProjectBindingDigest: binding, Destination: destination, ConfigDigest: contenthash.Sum([]byte(digest + "\x00" + binding + "\x00" + namespace))}, nil
}

type telegramNotificationManifest struct {
	project                                      string
	payload, hash, document, destination, digest sql.NullString
	namespace, root                              sql.NullString
}

// ReadTelegramNotification reopens the frozen view without reading a source path
// or resetting transport state. A NULL manifest is an untouched legacy event.
func (r *sqliteNotificationEventRepository) ReadTelegramNotification(ctx context.Context, id string, target TelegramNotificationTarget) (TelegramNotificationDelivery, bool, error) {
	var (
		delivery TelegramNotificationDelivery
		m        telegramNotificationManifest
	)

	err := r.db.QueryRowContext(ctx, `SELECT project_id,telegram_presentation,telegram_presentation_hash,telegram_document_id,telegram_destination,telegram_config_digest,telegram_namespace,telegram_project_binding_digest,telegram_attempts,telegram_plain_fallback,telegram_uncertain,COALESCE(telegram_message_id,0),COALESCE(telegram_file_id,''),status FROM notification_events WHERE id=? AND channel='telegram'`, id).Scan(&m.project, &m.payload, &m.hash, &m.document, &m.destination, &m.digest, &m.namespace, &m.root, &delivery.Attempts, &delivery.PlainFallback, &delivery.Uncertain, &delivery.Acknowledgement.MessageID, &delivery.Acknowledgement.FileID, &delivery.Status)
	if err != nil {
		return delivery, false, err
	}

	if m.project != target.ProjectID {
		return delivery, false, ErrTelegramNotificationTargetUnavailable
	}

	if !m.payload.Valid {
		if !m.isLegacy() {
			return delivery, true, tgmessage.ErrDocumentUnavailable
		}

		return delivery, false, nil
	}

	presentation, frozen, err := decodeTelegramNotificationManifest(m, target)
	if err != nil || !frozen {
		return delivery, frozen, err
	}

	delivery.Presentation = presentation
	if m.document.Valid {
		document, err := r.readTelegramNotificationDocument(ctx, id, m.document.String, target, presentation.DocumentHash)
		if err != nil {
			return delivery, true, fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
		}

		delivery.Document = &document
	}

	return delivery, true, nil
}

func decodeTelegramNotificationManifest(m telegramNotificationManifest, target TelegramNotificationTarget) (tgmessage.Presentation, bool, error) {
	var presentation tgmessage.Presentation
	if len(m.payload.String) > 64<<10 {
		return presentation, true, tgmessage.ErrDocumentUnavailable
	}

	if err := json.Unmarshal([]byte(m.payload.String), &presentation); err != nil {
		return presentation, true, fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
	}

	if presentation.Version == 0 && m.isLegacy() {
		return presentation, false, nil
	}

	if !m.matchesTarget(target) {
		return presentation, true, ErrTelegramNotificationTargetUnavailable
	}

	if !m.hash.Valid {
		return presentation, true, tgmessage.ErrDocumentUnavailable
	}

	return presentation, true, validateTelegramNotificationManifest(m, presentation)
}

func (m telegramNotificationManifest) isLegacy() bool {
	return !m.hash.Valid && !m.document.Valid && !m.destination.Valid && !m.digest.Valid && !m.namespace.Valid && !m.root.Valid
}

func (m telegramNotificationManifest) matchesTarget(target TelegramNotificationTarget) bool {
	return m.destination.String == target.Destination && m.digest.String == target.ConfigDigest && m.namespace.String == target.Namespace && m.root.String == target.ProjectBindingDigest
}

func validateTelegramNotificationManifest(m telegramNotificationManifest, presentation tgmessage.Presentation) error {
	check := TelegramReceiverDelivery{Text: presentation.PlainFallback, DocumentID: m.document.String}
	if err := decodeTelegramPresentationHash(m.payload.String, m.hash.String, &check); err != nil {
		return fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
	}

	return validateTelegramNotificationActions(presentation)
}

func validateTelegramNotificationActions(p tgmessage.Presentation) error {
	if p.ActionRef != "" || p.ScopeDigest != "" || len(p.ActionChoices) != 0 || len(p.ActionLabels) != 0 {
		return tgmessage.ErrDocumentUnavailable
	}

	return nil
}

func (r *sqliteNotificationEventRepository) readTelegramNotificationDocument(ctx context.Context, owner, id string, target TelegramNotificationTarget, hash string) (tgmessage.Document, error) {
	var (
		d          tgmessage.Document
		storedHash string
	)

	err := r.db.QueryRowContext(ctx, `SELECT project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,payload FROM telegram_notification_documents WHERE id=? AND owner_id=? AND destination=? AND config_digest=? AND payload IS NOT NULL`, id, owner, target.Destination, target.ConfigDigest).Scan(&d.ProjectID, &d.Namespace, &d.ProjectBindingDigest, &d.SourceRef, &d.SourceHash, &storedHash, &d.Filename, &d.MediaType, &d.Payload)
	if err != nil {
		return d, err
	}

	if d.ProjectID != target.ProjectID || d.Namespace != target.Namespace || d.ProjectBindingDigest != target.ProjectBindingDigest || storedHash != hash || contenthash.Sum(d.Payload) != hash {
		return tgmessage.Document{}, tgmessage.ErrDocumentUnavailable
	}

	return d, tgmessage.ValidateDocument(d)
}
