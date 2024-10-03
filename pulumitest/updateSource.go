package pulumitest

import "path/filepath"

// Copy files from a source directory to the current program directory.
func (a *PulumiTest) UpdateSource(t PT, pathElems ...string) {
	t.Helper()

	path := filepath.Join(pathElems...)
	ptLogF(t, "updating source from %s", path)
	err := copyDirectory(path, a.workingDir)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
