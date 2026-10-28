package state

// ValidFingerprint checks the exact lowercase SHA-256 spelling used by retained inputs.
func ValidFingerprint(value string) bool {
	return len(value) == 64 && ValidCommitID(value)
}

// ValidCommitID accepts the existing full lowercase Git object identity formats.
func ValidCommitID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}

	return true
}
