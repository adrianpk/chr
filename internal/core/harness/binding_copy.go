package harness

// CloneHarnessBinding preserves the admitted binding value and command contract.
func CloneHarnessBinding(binding *Binding) *Binding {
	if binding == nil {
		return nil
	}

	copy := *binding

	return &copy
}
