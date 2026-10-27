package core

import "forge.adrianpk.com/chr/internal/core/database"

// remoteCommandRow retains encoded loop values and timestamp inputs beside the domain record.
type remoteCommandRow struct {
	record                                RemoteCommandRecord
	status, intent, receivedAt, createdAt string
	loop                                  remoteCommandLoopColumns
}

const remoteCommandColumns = `id, status, channel, sender_external_id, sender_display_name,
    project_selector, intent, ticket_key, comment, loop_statuses,
    loop_priorities, loop_severities, loop_project_areas, loop_tags,
    loop_types, loop_use_cases, loop_max_iterations, raw_text,
    handoff_command, rejection_reason, received_at, created_at`

func (row remoteCommandRow) values() []any {
	record := row.record
	request := record.Request

	return []any{
		request.ID,
		string(record.Status),
		request.Sender.Channel,
		request.Sender.ExternalID,
		request.Sender.DisplayName,
		request.ProjectSelector,
		string(request.Intent),
		request.TicketKey,
		request.Comment,
		mustEncodeRemoteCommandValues(request.Loop.Statuses),
		mustEncodeRemoteCommandValues(request.Loop.Priorities),
		mustEncodeRemoteCommandValues(request.Loop.Severities),
		mustEncodeRemoteCommandValues(request.Loop.ProjectAreas),
		mustEncodeRemoteCommandValues(request.Loop.Tags),
		mustEncodeRemoteCommandValues(request.Loop.Types),
		mustEncodeRemoteCommandValues(request.Loop.UseCases),
		request.Loop.MaxIterations,
		request.RawText,
		record.HandoffCommand,
		record.RejectionReason,
		database.FormatTime(request.ReceivedAt),
		database.FormatTime(request.CreatedAt),
	}
}

func (row *remoteCommandRow) scanTargets() []any {
	record := &row.record

	return []any{
		&record.Request.ID,
		&row.status,
		&record.Request.Sender.Channel,
		&record.Request.Sender.ExternalID,
		&record.Request.Sender.DisplayName,
		&record.Request.ProjectSelector,
		&row.intent,
		&record.Request.TicketKey,
		&record.Request.Comment,
		&row.loop.Statuses,
		&row.loop.Priorities,
		&row.loop.Severities,
		&row.loop.ProjectAreas,
		&row.loop.Tags,
		&row.loop.Types,
		&row.loop.UseCases,
		&record.Request.Loop.MaxIterations,
		&record.Request.RawText,
		&record.HandoffCommand,
		&record.RejectionReason,
		&row.receivedAt,
		&row.createdAt,
	}
}

func insertRemoteCommandRecordSQL() string {
	return `INSERT INTO remote_command_records (` + remoteCommandColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func selectRemoteCommandRecordSQL() string {
	return `SELECT ` + remoteCommandColumns + ` FROM remote_command_records`
}

func (row remoteCommandRow) decoded() (RemoteCommandRecord, error) {
	record := row.record

	var err error

	record.Status = RemoteCommandRecordStatus(row.status)
	record.Request.Intent = RemoteCommandIntent(row.intent)

	record.Request.ReceivedAt, err = database.ParseTime(row.receivedAt)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	record.Request.CreatedAt, err = database.ParseTime(row.createdAt)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	record.Request.Loop, err = decodeRemoteCommandLoop(row.loop, record.Request.Loop.MaxIterations)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	return record, nil
}
