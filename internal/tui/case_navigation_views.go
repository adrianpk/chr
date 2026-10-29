package tui

func (m *Model) visibleDumps() []DumpView {
	if !m.caseContextAvailable() {
		return nil
	}

	if m.caseContext.CaseSlug == "" {
		return m.dumps
	}

	var items []DumpView

	for _, item := range m.dumps {
		if item.Scope.CaseSlug == m.caseContext.CaseSlug {
			items = append(items, item)
		}
	}

	return items
}

func (m *Model) visibleQuizzes() []QuizView {
	if !m.caseContextAvailable() {
		return nil
	}

	if m.caseContext.CaseSlug == "" {
		return m.quizzes
	}

	var items []QuizView

	for _, item := range m.quizzes {
		if item.Scope.CaseSlug == m.caseContext.CaseSlug {
			items = append(items, item)
		}
	}

	return items
}

func (m *Model) visibleSpecs() []SpecView {
	if !m.caseContextAvailable() {
		return nil
	}

	if m.caseContext.CaseSlug == "" {
		return m.specs
	}

	var items []SpecView

	for _, item := range m.specs {
		if item.Scope.CaseSlug == m.caseContext.CaseSlug {
			items = append(items, item)
		}
	}

	return items
}

func (m *Model) visiblePlans() []PlanView {
	if !m.caseContextAvailable() {
		return nil
	}

	if m.caseContext.CaseSlug == "" {
		return m.plans
	}

	var items []PlanView

	for _, item := range m.plans {
		if item.Scope.CaseSlug == m.caseContext.CaseSlug {
			items = append(items, item)
		}
	}

	return items
}

func (m *Model) caseTickets() []TicketView {
	if !m.caseContextAvailable() {
		return nil
	}

	if m.caseContext.CaseSlug == "" {
		return m.tickets
	}

	var items []TicketView

	for _, item := range m.tickets {
		if item.UseCase == m.caseContext.CaseSlug {
			items = append(items, item)
		}
	}

	return items
}
