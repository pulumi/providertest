package pulumitest

// Copy files from a source directory to the current program directory.
func (a *PulumiTest) UpdateSource(source string) {
	a.t.Helper()

	a.t.Log("updating source")
	copyDirectory(source, a.source)
}
