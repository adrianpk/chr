package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationIdentity preserves an object's identity and Git date.
type IntegrationIdentity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// IntegrationAttribution binds metadata to the complete commit object.
// MessageHash covers the original bytes, including trailing newlines.
type IntegrationAttribution struct {
	Commit      string              `json:"commit"`
	Tree        string              `json:"tree"`
	Parents     []string            `json:"parents"`
	Author      IntegrationIdentity `json:"author"`
	Committer   IntegrationIdentity `json:"committer"`
	MessageHash string              `json:"message_hash"`
	Encoding    string              `json:"encoding,omitempty"`
}

func (reader workflowIntegrationReader) commitAttribution(ctx context.Context, commit string) (IntegrationAttribution, error) {
	data, err := reader.RawOutput(ctx, "cat-file", "commit", commit)
	if err != nil {
		return IntegrationAttribution{}, err
	}

	return workflowCommitAttribution(commit, data)
}

func workflowCommitAttribution(commit, data string) (IntegrationAttribution, error) {
	header, message, ok := strings.Cut(data, "\n\n")
	if !ok {
		return IntegrationAttribution{}, workflowstate.ErrConflict
	}

	result := IntegrationAttribution{Commit: commit, Parents: []string{}, MessageHash: contenthash.Sum([]byte(message))}

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

func workflowIntegrationIdentity(value string) (IntegrationIdentity, error) {
	nameEnd := strings.LastIndex(value, " <")
	if nameEnd < 1 {
		return IntegrationIdentity{}, workflowstate.ErrConflict
	}

	email, date, ok := strings.Cut(value[nameEnd+2:], "> ")

	identity := IntegrationIdentity{Name: value[:nameEnd], Email: email, Date: date}
	if !ok || !validWorkflowCommitIdentity(identity) {
		return identity, workflowstate.ErrConflict
	}

	return identity, nil
}

func validWorkflowCommitIdentity(identity IntegrationIdentity) bool {
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

func validWorkflowIntegrationAttribution(entry IntegrationAttribution) bool {
	return !strings.ContainsAny(entry.Encoding, "\x00\r\n") && validWorkflowCommitIdentity(entry.Author) && validWorkflowCommitIdentity(entry.Committer) && validWorkflowAttributionObjects(entry)
}

func validWorkflowAttributionObjects(entry IntegrationAttribution) bool {
	if !workflowCommitID(entry.Commit) || !workflowCommitID(entry.Tree) || len(entry.Commit) != len(entry.Tree) || len(entry.Parents) > workflowstate.RecordLimit || !validWorkflowNativeFingerprint(entry.MessageHash) {
		return false
	}

	for _, parent := range entry.Parents {
		if !workflowCommitID(parent) || len(parent) != len(entry.Commit) {
			return false
		}
	}

	return true
}

func (reader workflowIntegrationReader) sourceAttributions(ctx context.Context, proof IntegrationProof) ([]IntegrationAttribution, error) {
	line, err := reader.Output(ctx, "rev-list", "--reverse", "--topo-order", "--max-count="+strconv.Itoa(workflowstate.RecordLimit+1), proof.OriginCommit+".."+proof.SourceCommit)

	commits := strings.Fields(line)
	if err != nil || len(commits) == 0 || len(commits) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("integration source manifest is empty or exceeds commit bounds")
	}

	entries := make([]IntegrationAttribution, 0, len(commits))
	size := 0

	for _, commit := range commits {
		entry, err := reader.commitAttribution(ctx, commit)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(entry)
		if err != nil || len(data) >= workflowstate.FileLimit-size {
			return nil, fmt.Errorf("integration source attribution exceeds file bounds")
		}

		size += len(data)

		entries = append(entries, entry)
	}

	return entries, nil
}
