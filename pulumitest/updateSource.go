package pulumitest

import "path/filepath"

// Copy files from a source directory to the current program directory.
// Any files in the current program directory that are not in the source directory will remain unchanged.
func (pt *PulumiTest) UpdateSource(t PT, pathElems ...string) {
	t.Helper()

	path := filepath.Join(pathElems...)
	ptLogF(t, "updating source from %s", path)
	err := copyDirectory(path, pt.workingDir)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
