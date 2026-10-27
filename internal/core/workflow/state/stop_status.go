package state

// StopStatus maps the existing stop intent to its quiescent native status.
func StopStatus(intent string) string {
	if intent == "cancel" {
		return "cancelled"
	}

	return "paused"
}
