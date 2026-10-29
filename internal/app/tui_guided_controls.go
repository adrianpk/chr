package app

import "slices"

// Match the existing execution-only Pause presentation. The canonical guided
// and native owners still validate every control and retained revision.
func tuiGuidedControlChoices(observed GuidedUseCaseInspection) []string {
	controls := guidedControlChoices(observed)
	if tuiGuidedCanPause(observed) {
		return controls
	}

	return slices.DeleteFunc(controls, func(action string) bool { return action == "pause" })
}

func tuiGuidedCanPause(observed GuidedUseCaseInspection) bool {
	if observed.Session.Guided.Stage != "execution" || observed.Work == nil {
		return false
	}

	return slices.Contains([]string{"pending", "running", "waiting_review", "waiting_integration"}, observed.Work.Status)
}
