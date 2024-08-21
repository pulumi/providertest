package pulumitest

import "path/filepath"

// Copy files from a source directory to the current program directory.
func (a *PulumiTest) UpdateSource(pathElems ...string) {
	a.t.Helper()

	path := filepath.Join(pathElems...)
	a.logf("updating source from %s", path)
	err := copyDirectory(path, a.workingDir)
	if err != nil {
		a.t.Log(err)
		a.t.FailNow()
	}
}
