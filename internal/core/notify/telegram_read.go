package notify

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

// ResolveTelegramTarget uses the same canonical project-root binding
// as receiver views. The caller retains its existing notification authorization.
func ResolveTelegramTarget(project projectcore.Project, namespace, destination, digest string) (TelegramTarget, error) {
	binding, err := projectcore.RootBindingDigest(project)
	if err != nil {
		return TelegramTarget{}, err
	}

	if !project.Enabled || !opsnamespace.Valid(namespace) || !validTelegramDigest(digest) {
		return TelegramTarget{}, tgmessage.ErrDocumentUnavailable
	}

	return TelegramTarget{ProjectID: project.ID, Namespace: namespace, ProjectBindingDigest: binding, Destination: destination, ConfigDigest: contenthash.Sum([]byte(digest + "\x00" + binding + "\x00" + namespace))}, nil
}

type telegramManifest struct {
	project                                      string
	payload, hash, document, destination, digest sql.NullString
	namespace, root                              sql.NullString
}

// ReadTelegram reopens the frozen view without reading a source path
// or resetting transport state. A NULL manifest is an untouched legacy event.
func (r *sqliteRepository) ReadTelegram(ctx context.Context, id string, target TelegramTarget) (TelegramLookup, error) {
	var (
		delivery TelegramDelivery
		m        telegramManifest
	)

	err := r.db.QueryRowContext(ctx, `SELECT project_id,telegram_presentation,telegram_presentation_hash,telegram_document_id,telegram_destination,telegram_config_digest,telegram_namespace,telegram_project_binding_digest,telegram_attempts,telegram_plain_fallback,telegram_uncertain,COALESCE(telegram_message_id,0),COALESCE(telegram_file_id,''),status FROM notification_events WHERE id=? AND channel='telegram'`, id).Scan(&m.project, &m.payload, &m.hash, &m.document, &m.destination, &m.digest, &m.namespace, &m.root, &delivery.Attempts, &delivery.PlainFallback, &delivery.Uncertain, &delivery.Acknowledgement.MessageID, &delivery.Acknowledgement.FileID, &delivery.Status)
	if err != nil {
		return TelegramLookup{Delivery: delivery, Frozen: false}, err
	}

	if m.project != target.ProjectID {
		return TelegramLookup{Delivery: delivery, Frozen: false}, ErrTelegramTargetUnavailable
	}

	if !m.payload.Valid {
		if !m.isLegacy() {
			return TelegramLookup{Delivery: delivery, Frozen: true}, tgmessage.ErrDocumentUnavailable
		}

		return TelegramLookup{Delivery: delivery, Frozen: false}, nil
	}

	manifestLookup, err := decodeTelegramManifest(m, target)

	presentation, frozen := manifestLookup.Presentation, manifestLookup.Frozen
	if err != nil || !frozen {
		return TelegramLookup{Delivery: delivery, Frozen: frozen}, err
	}

	delivery.Presentation = presentation
	if m.document.Valid {
		document, err := r.readTelegramDocument(ctx, id, m.document.String, target, presentation.DocumentHash)
		if err != nil {
			return TelegramLookup{Delivery: delivery, Frozen: true}, fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
		}

		delivery.Document = &document
	}

	return TelegramLookup{Delivery: delivery, Frozen: true}, nil
}

func decodeTelegramManifest(m telegramManifest, target TelegramTarget) (telegramManifestLookup, error) {
	var presentation tgmessage.Presentation
	if len(m.payload.String) > 64<<10 {
		return telegramManifestLookup{Presentation: presentation, Frozen: true}, tgmessage.ErrDocumentUnavailable
	}

	if err := json.Unmarshal([]byte(m.payload.String), &presentation); err != nil {
		return telegramManifestLookup{Presentation: presentation, Frozen: true}, fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
	}

	if presentation.Version == 0 && m.isLegacy() {
		return telegramManifestLookup{Presentation: presentation, Frozen: false}, nil
	}

	if !m.matchesTarget(target) {
		return telegramManifestLookup{Presentation: presentation, Frozen: true}, ErrTelegramTargetUnavailable
	}

	if !m.hash.Valid {
		return telegramManifestLookup{Presentation: presentation, Frozen: true}, tgmessage.ErrDocumentUnavailable
	}

	return telegramManifestLookup{Presentation: presentation, Frozen: true}, validateTelegramManifest(m, presentation)
}

func (m telegramManifest) isLegacy() bool {
	return !m.hash.Valid && !m.document.Valid && !m.destination.Valid && !m.digest.Valid && !m.namespace.Valid && !m.root.Valid
}

func (m telegramManifest) matchesTarget(target TelegramTarget) bool {
	return m.destination.String == target.Destination && m.digest.String == target.ConfigDigest && m.namespace.String == target.Namespace && m.root.String == target.ProjectBindingDigest
}

func validateTelegramManifest(m telegramManifest, presentation tgmessage.Presentation) error {
	if _, err := tgmessage.DecodeRetained(tgmessage.RetainedInput{Payload: m.payload.String, Hash: m.hash.String, Text: presentation.PlainFallback, DocumentID: m.document.String}); err != nil {
		return fmt.Errorf("%w: %w", tgmessage.ErrDocumentUnavailable, err)
	}

	return validateTelegramActions(presentation)
}

func validateTelegramActions(p tgmessage.Presentation) error {
	if p.ActionRef != "" || p.ScopeDigest != "" || len(p.ActionChoices) != 0 || len(p.ActionLabels) != 0 {
		return tgmessage.ErrDocumentUnavailable
	}

	return nil
}

func (r *sqliteRepository) readTelegramDocument(ctx context.Context, owner, id string, target TelegramTarget, hash string) (tgmessage.Document, error) {
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
