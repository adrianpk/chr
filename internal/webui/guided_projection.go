package webui

// GuidedFormFields freezes the displayed owner and question guards for one form.
type GuidedFormFields struct {
	RunID, Namespace, Revision, RequestID, QuestionID, InputDigest string
}

// GuidedRetryForm retains the exact failed step and original pending content.
type GuidedRetryForm struct {
	Fields        GuidedFormFields
	Step, Content string
}

// GuidedActionsSection contains only contextual controls and their form bindings.
type GuidedActionsSection struct {
	Visible, Review  bool
	RunID, Namespace string
	Controls         []GuidedUseCaseControl
	Retry            *GuidedRetryForm
	ReviewURL        string
}

// GuidedStatusSection contains the current status message and contextual actions.
type GuidedStatusSection struct {
	Message string
	Actions GuidedActionsSection
}

// GuidedRecoverySection distinguishes a blocked step from a failed attempt.
type GuidedRecoverySection struct {
	Blocked bool
	Actions GuidedActionsSection
}

// GuidedProgressSection presents observed work and its contextual actions.
type GuidedProgressSection struct {
	Message string
	Work    *GuidedUseCaseWork
	Actions GuidedActionsSection
}

// FormFields preserves the original displayed operation and question identity.
func (view GuidedUseCaseView) FormFields() GuidedFormFields {
	return GuidedFormFields{RunID: view.RunID, Namespace: view.Namespace, Revision: view.Revision, RequestID: view.RequestID, QuestionID: view.QuestionID, InputDigest: view.InputDigest}
}

// Actions projects controls without replacing their per-owner revisions.
func (view GuidedUseCaseView) Actions() GuidedActionsSection {
	section := GuidedActionsSection{
		Visible: view.ShowControls || view.ShowRetry || view.ShowReview,
		Review:  view.ShowReview, RunID: view.RunID, Namespace: view.Namespace,
		Controls: view.Controls,
	}
	if view.ShowRetry {
		step := view.PendingAction
		if step == "" {
			step = view.RetryAction
		}

		section.Retry = &GuidedRetryForm{Fields: view.FormFields(), Step: step, Content: view.PendingContent}
	}

	if view.Work != nil {
		section.ReviewURL = view.Work.PullRequestURL
	}

	return section
}

// StatusSection limits the status partial to its message and actions.
func (view GuidedUseCaseView) StatusSection() GuidedStatusSection {
	return GuidedStatusSection{Message: view.Message, Actions: view.Actions()}
}

// RecoverySection limits the recovery partial to its observed wait and actions.
func (view GuidedUseCaseView) RecoverySection() GuidedRecoverySection {
	return GuidedRecoverySection{Blocked: view.Blocker != "", Actions: view.Actions()}
}

// ProgressSection limits the progress partial to observed work and actions.
func (view GuidedUseCaseView) ProgressSection() GuidedProgressSection {
	return GuidedProgressSection{Message: view.Message, Work: view.Work, Actions: view.Actions()}
}
