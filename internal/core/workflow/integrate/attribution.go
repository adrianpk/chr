package integrate

import (
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowCommitAttribution parses exact Git headers and the original message hash without changing object bytes.
func WorkflowCommitAttribution(commit, data string) (workflowstate.IntegrationAttribution, error) {
	header, message, ok := strings.Cut(data, "\n\n")
	if !ok {
		return workflowstate.IntegrationAttribution{}, workflowstate.ErrConflict
	}

	result := workflowstate.IntegrationAttribution{Commit: commit, Parents: []string{}, MessageHash: contenthash.Sum([]byte(message))}

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

func workflowIntegrationIdentity(value string) (workflowstate.IntegrationIdentity, error) {
	nameEnd := strings.LastIndex(value, " <")
	if nameEnd < 1 {
		return workflowstate.IntegrationIdentity{}, workflowstate.ErrConflict
	}

	email, date, ok := strings.Cut(value[nameEnd+2:], "> ")

	identity := workflowstate.IntegrationIdentity{Name: value[:nameEnd], Email: email, Date: date}
	if !ok || !validWorkflowCommitIdentity(identity) {
		return identity, workflowstate.ErrConflict
	}

	return identity, nil
}

func validWorkflowCommitIdentity(identity workflowstate.IntegrationIdentity) bool {
	if identity.Name == "" || identity.Email == "" || strings.ContainsAny(identity.Name+identity.Email, "\x00\r\n<>") {
		return false
	}

	return ValidWorkflowCommitDate(identity.Date)
}

// ValidWorkflowCommitDate checks the existing signed Git timestamp and timezone grammar.
func ValidWorkflowCommitDate(date string) bool {
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

func validWorkflowIntegrationAttribution(entry workflowstate.IntegrationAttribution) bool {
	return !strings.ContainsAny(entry.Encoding, "\x00\r\n") && validWorkflowCommitIdentity(entry.Author) && validWorkflowCommitIdentity(entry.Committer) && validWorkflowAttributionObjects(entry)
}

func validWorkflowAttributionObjects(entry workflowstate.IntegrationAttribution) bool {
	if !workflowstate.ValidCommitID(entry.Commit) || !workflowstate.ValidCommitID(entry.Tree) || len(entry.Commit) != len(entry.Tree) || len(entry.Parents) > workflowstate.RecordLimit || !workflowstate.ValidFingerprint(entry.MessageHash) {
		return false
	}

	for _, parent := range entry.Parents {
		if !workflowstate.ValidCommitID(parent) || len(parent) != len(entry.Commit) {
			return false
		}
	}

	return true
}
