package tui

import "forge.adrianpk.com/chc/internal/core"

// WorkReadinessView is the TUI projection of work readiness state.
type WorkReadinessView struct {
	Status            string
	Mode              string
	IntegrationBranch string
	CanonicalRemote   string
	CurrentBranch     string
	Diagnostics       []WorkReadinessDiagnosticView
	RepairsApplied    []WorkReadinessAppliedRepairView
}

// WorkReadinessDiagnosticView is the TUI projection of one readiness diagnostic.
type WorkReadinessDiagnosticView struct {
	Code            string
	Severity        string
	Detail          string
	RepairCode      string
	RepairAvailable bool
	RepairSafe      bool
	RepairDetail    string
}

// WorkReadinessAppliedRepairView is the TUI projection of one applied repair.
type WorkReadinessAppliedRepairView struct {
	Code   string
	Detail string
}

// NewWorkReadinessView projects a core readiness report into the TUI model.
func NewWorkReadinessView(report core.WorkReadinessReport) WorkReadinessView {
	return WorkReadinessView{
		Status:            string(report.Status),
		Mode:              string(report.Policy.Mode),
		IntegrationBranch: report.Policy.IntegrationBranch,
		CanonicalRemote:   report.Policy.CanonicalRemote,
		CurrentBranch:     report.CurrentBranch,
		Diagnostics:       newWorkReadinessDiagnosticViews(report.Diagnostics),
	}
}

// NewWorkReadinessRepairView projects a core repair result into the TUI model.
func NewWorkReadinessRepairView(result core.WorkReadinessRepairResult) WorkReadinessView {
	view := NewWorkReadinessView(result.After)
	view.RepairsApplied = newWorkReadinessAppliedRepairViews(result.Applied)

	return view
}

func newWorkReadinessDiagnosticViews(diagnostics []core.WorkReadinessDiagnostic) []WorkReadinessDiagnosticView {
	views := make([]WorkReadinessDiagnosticView, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		views = append(views, WorkReadinessDiagnosticView{
			Code:            string(diagnostic.Code),
			Severity:        string(diagnostic.Severity),
			Detail:          diagnostic.Detail,
			RepairCode:      string(diagnostic.Repair.Code),
			RepairAvailable: diagnostic.Repair.Available,
			RepairSafe:      diagnostic.Repair.Safe,
			RepairDetail:    diagnostic.Repair.Detail,
		})
	}

	return views
}

func newWorkReadinessAppliedRepairViews(repairs []core.WorkReadinessAppliedRepair) []WorkReadinessAppliedRepairView {
	views := make([]WorkReadinessAppliedRepairView, 0, len(repairs))
	for _, repair := range repairs {
		views = append(views, WorkReadinessAppliedRepairView{
			Code:   string(repair.Code),
			Detail: repair.Detail,
		})
	}

	return views
}

func (view WorkReadinessView) ready() bool {
	return view.Status == string(core.WorkReadinessStatusReady)
}

func (view WorkReadinessView) blocked() bool {
	return view.Status == string(core.WorkReadinessStatusBlocked)
}

func (view WorkReadinessView) safeRepairAvailable() bool {
	for _, diagnostic := range view.Diagnostics {
		if diagnostic.RepairAvailable && diagnostic.RepairSafe {
			return true
		}
	}

	return false
}
