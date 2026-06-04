package chc

// DumpDirectory executes the corresponding package operation.
func DumpDirectory(projectDir string) string {
	return UserPath(projectDir, "dumps")
}
