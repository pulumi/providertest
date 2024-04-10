package pulumitest

import (
	"time"
)

// A subset of *testing.T functionality used by pulumitest.
type PT interface {
	TempDir() string
	Fatal(...any)
	Fatalf(string, ...any)
	Log(...any)
	Logf(string, ...any)
	Errorf(string, ...any)
	Cleanup(func())
	Helper()
	Deadline() (time.Time, bool)
}
