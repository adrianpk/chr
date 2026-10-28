package state

// ProcessIdentity binds a PID to one Linux boot and process birth.
type ProcessIdentity struct {
	PID        int    `json:"pid"`
	BootID     string `json:"boot_id"`
	StartTicks string `json:"start_ticks"`
}
