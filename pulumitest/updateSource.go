package pulumitest

import "path/filepath"

// Copy files from a source directory to the current program directory.
func (a *PulumiTest) UpdateSource(pathElems ...string) {
	a.t.Helper()

	path := filepath.Join(pathElems...)
	a.t.Logf("updating source from %s", path)
	copyDirectory(path, a.source)
}
