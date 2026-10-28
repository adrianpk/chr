package state

// StopStatus maps the existing stop intent to its quiescent native status.
func StopStatus(intent string) InvocationStatus {
	if intent == "cancel" {
		return InvocationCancelled
	}

	return InvocationPaused
}
