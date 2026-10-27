package harness

// CloneBinding preserves the admitted binding value and command contract.
func CloneBinding(binding *Binding) *Binding {
	if binding == nil {
		return nil
	}

	copy := *binding

	return &copy
}
