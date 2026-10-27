package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationIdentity preserves an object's identity and Git date.
type WorkflowIntegrationIdentity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// WorkflowIntegrationAttribution binds metadata to the complete commit object.
// MessageHash covers the original bytes, including trailing newlines.
type WorkflowIntegrationAttribution struct {
	Commit      string                      `json:"commit"`
	Tree        string                      `json:"tree"`
	Parents     []string                    `json:"parents"`
	Author      WorkflowIntegrationIdentity `json:"author"`
	Committer   WorkflowIntegrationIdentity `json:"committer"`
	MessageHash string                      `json:"message_hash"`
	Encoding    string                      `json:"encoding,omitempty"`
}

func (reader workflowIntegrationReader) commitAttribution(ctx context.Context, commit string) (WorkflowIntegrationAttribution, error) {
	data, err := reader.rawOutput(ctx, "cat-file", "commit", commit)
	if err != nil {
		return WorkflowIntegrationAttribution{}, err
	}

	return workflowCommitAttribution(commit, data)
}

func workflowCommitAttribution(commit, data string) (WorkflowIntegrationAttribution, error) {
	header, message, ok := strings.Cut(data, "\n\n")
	if !ok {
		return WorkflowIntegrationAttribution{}, workflowstate.ErrConflict
	}

	result := WorkflowIntegrationAttribution{Commit: commit, Parents: []string{}, MessageHash: contenthash.Sum([]byte(message))}

	for _, line := range strings.Split(header, "\n") {
		key, value, _ := strings.Cut(line, " ")

		var err error

		switch key {
		case "tree":
			result.Tree = value
		case "parent":
			result.Parents = append(result.Parents, value)
		case "encoding":
			result.Encoding = value
		case "author":
			result.Author, err = workflowIntegrationIdentity(value)
		case "committer":
			result.Committer, err = workflowIntegrationIdentity(value)
		}

		if err != nil {
			return result, err
		}
	}

	if !validWorkflowIntegrationAttribution(result) {
		return result, workflowstate.ErrConflict
	}

	return result, nil
}

func workflowIntegrationIdentity(value string) (WorkflowIntegrationIdentity, error) {
	nameEnd := strings.LastIndex(value, " <")
	if nameEnd < 1 {
		return WorkflowIntegrationIdentity{}, workflowstate.ErrConflict
	}

	email, date, ok := strings.Cut(value[nameEnd+2:], "> ")

	identity := WorkflowIntegrationIdentity{Name: value[:nameEnd], Email: email, Date: date}
	if !ok || !validWorkflowCommitIdentity(identity) {
		return identity, workflowstate.ErrConflict
	}

	return identity, nil
}

func validWorkflowCommitIdentity(identity WorkflowIntegrationIdentity) bool {
	if identity.Name == "" || identity.Email == "" || strings.ContainsAny(identity.Name+identity.Email, "\x00\r\n<>") {
		return false
	}

	return validWorkflowCommitDate(identity.Date)
}

func validWorkflowCommitDate(date string) bool {
	fields := strings.Fields(date)
	if len(fields) != 2 || !validWorkflowCommitTimezone(fields[1]) {
		return false
	}

	_, secondsErr := strconv.ParseInt(fields[0], 10, 64)

	return secondsErr == nil
}

func validWorkflowCommitTimezone(zone string) bool {
	if len(zone) != 5 || (zone[0] != '+' && zone[0] != '-') {
		return false
	}

	hours, hoursErr := strconv.Atoi(zone[1:3])
	minutes, minutesErr := strconv.Atoi(zone[3:])

	return hoursErr == nil && minutesErr == nil && hours >= 0 && hours < 24 && minutes >= 0 && minutes < 60
}

func validWorkflowIntegrationAttribution(entry WorkflowIntegrationAttribution) bool {
	return !strings.ContainsAny(entry.Encoding, "\x00\r\n") && validWorkflowCommitIdentity(entry.Author) && validWorkflowCommitIdentity(entry.Committer) && validWorkflowAttributionObjects(entry)
}

func validWorkflowAttributionObjects(entry WorkflowIntegrationAttribution) bool {
	if !workflowCommitID(entry.Commit) || !workflowCommitID(entry.Tree) || len(entry.Commit) != len(entry.Tree) || len(entry.Parents) > workflowstate.WorkflowRecordLimit || !validWorkflowNativeFingerprint(entry.MessageHash) {
		return false
	}

	for _, parent := range entry.Parents {
		if !workflowCommitID(parent) || len(parent) != len(entry.Commit) {
			return false
		}
	}

	return true
}

func (reader workflowIntegrationReader) sourceAttributions(ctx context.Context, proof WorkflowIntegrationProof) ([]WorkflowIntegrationAttribution, error) {
	line, err := reader.output(ctx, "rev-list", "--reverse", "--topo-order", "--max-count="+strconv.Itoa(workflowstate.WorkflowRecordLimit+1), proof.OriginCommit+".."+proof.SourceCommit)

	commits := strings.Fields(line)
	if err != nil || len(commits) == 0 || len(commits) > workflowstate.WorkflowRecordLimit {
		return nil, fmt.Errorf("integration source manifest is empty or exceeds commit bounds")
	}

	entries := make([]WorkflowIntegrationAttribution, 0, len(commits))
	size := 0

	for _, commit := range commits {
		entry, err := reader.commitAttribution(ctx, commit)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(entry)
		if err != nil || len(data) >= workflowstate.WorkflowFileLimit-size {
			return nil, fmt.Errorf("integration source attribution exceeds file bounds")
		}

		size += len(data)

		entries = append(entries, entry)
	}

	return entries, nil
}
