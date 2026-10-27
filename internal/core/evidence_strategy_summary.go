package core

import workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"

// EvidenceStrategySummaryStatus describes a ticket evidence checklist summary.
type EvidenceStrategySummaryStatus string

const (
	// EvidenceStrategySummaryStatusMissing means no anchored checklist exists.
	EvidenceStrategySummaryStatusMissing EvidenceStrategySummaryStatus = "missing"
	// EvidenceStrategySummaryStatusMalformed means the anchored checklist cannot parse.
	EvidenceStrategySummaryStatusMalformed EvidenceStrategySummaryStatus = "malformed"
	// EvidenceStrategySummaryStatusBlocked means at least one checklist item is blocked.
	EvidenceStrategySummaryStatusBlocked EvidenceStrategySummaryStatus = "blocked"
	// EvidenceStrategySummaryStatusIncomplete means evidence exists but is not complete.
	EvidenceStrategySummaryStatusIncomplete EvidenceStrategySummaryStatus = "incomplete"
	// EvidenceStrategySummaryStatusComplete means evidence checklist items are complete.
	EvidenceStrategySummaryStatusComplete EvidenceStrategySummaryStatus = "complete"
)

// EvidenceStrategySummaryItem describes one summarized checklist row.
type EvidenceStrategySummaryItem struct {
	Label   string
	Status  string
	Checked bool
	Detail  string
}

// EvidenceStrategySummary describes the concise evidence state for a ticket.
type EvidenceStrategySummary struct {
	Status EvidenceStrategySummaryStatus
	Found  bool
	Error  string
	Items  []EvidenceStrategySummaryItem
}

// SummarizeEvidenceStrategyChecklist parses and summarizes ticket evidence.
func SummarizeEvidenceStrategyChecklist(content string) EvidenceStrategySummary {
	checklist, err := ParseEvidenceStrategyChecklist(content)
	if err != nil {
		return EvidenceStrategySummary{
			Status: EvidenceStrategySummaryStatusMalformed,
			Error:  err.Error(),
		}
	}

	if !checklist.Found {
		return EvidenceStrategySummary{Status: EvidenceStrategySummaryStatusMissing}
	}

	return NewEvidenceStrategySummary(checklist)
}

// NewEvidenceStrategySummary summarizes an already parsed evidence checklist.
func NewEvidenceStrategySummary(checklist EvidenceStrategyChecklist) EvidenceStrategySummary {
	items := evidenceStrategySummaryItems(checklist)
	status := evidenceStrategySummaryStatus(checklist)

	return EvidenceStrategySummary{
		Status: status,
		Found:  true,
		Items:  items,
	}
}

func evidenceStrategySummaryItems(checklist EvidenceStrategyChecklist) []EvidenceStrategySummaryItem {
	items := make([]EvidenceStrategySummaryItem, 0, len(EvidenceStrategyChecklistLabels()))

	for _, label := range EvidenceStrategyChecklistLabels() {
		item, ok := checklist.Item(label)
		if !ok {
			continue
		}

		items = append(items, EvidenceStrategySummaryItem{
			Label:   item.Label,
			Status:  item.Status,
			Checked: item.Checked,
			Detail:  item.Detail,
		})
	}

	return items
}

func evidenceStrategySummaryStatus(checklist EvidenceStrategyChecklist) EvidenceStrategySummaryStatus {
	readiness := EvaluateEvidenceReadiness(EvidenceReadinessRequest{
		Config:    workspaceconfig.DefaultEvidenceStrategyConfig(),
		Checklist: &checklist,
	})
	if readiness.Allowed {
		return EvidenceStrategySummaryStatusComplete
	}

	if evidenceReadinessHasDiagnostic(readiness, EvidenceStrategyDiagnosticBlocked) {
		return EvidenceStrategySummaryStatusBlocked
	}

	return EvidenceStrategySummaryStatusIncomplete
}

func evidenceReadinessHasDiagnostic(readiness EvidenceReadinessResult, code EvidenceStrategyDiagnosticCode) bool {
	for _, diagnostic := range readiness.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}

	return false
}
