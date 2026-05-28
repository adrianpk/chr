package chc

func DumpDirectory(projectDir string) string {
	return UserPath(projectDir, "dumps")
}
