package pulumitest

import (
	"time"
)

// A subset of *testing.T functionality used by pulumitest.
type PT interface {
	TempDir() string
	Log(...any)
	Fail()
	FailNow()
	Cleanup(func())
	Helper()
	Deadline() (time.Time, bool)
}
