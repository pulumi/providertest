package pulumitest

import (
	"fmt"
	"time"
)

// A subset of *testing.T functionality used by pulumitest.
// This is used to allow pulumitest to be used with other testing frameworks and to be mocked out in our own unit tests.
type PT interface {
	Name() string
	TempDir() string
	Log(...any)
	Fail()
	FailNow()
	Cleanup(func())
	Helper()
	Deadline() (time.Time, bool)
}

func ptErrorF(t PT, format string, args ...any) {
	t.Helper()
	t.Log(fmt.Sprintf(format, args...))
	t.Fail()
}

func ptFatalF(t PT, format string, args ...any) {
	t.Helper()
	t.Log(fmt.Sprintf(format, args...))
	t.FailNow()
}

func ptFailed(t PT) bool {
	if tF, tFailedSupported := t.(interface {
		Failed() bool
	}); tFailedSupported && tF.Failed() {
		return true
	}
	return false
}

func ptLogF(t PT, format string, args ...any) {
	t.Helper()
	t.Log(fmt.Sprintf(format, args...))
}

func ptError(t PT, args ...any) {
	t.Helper()
	t.Log(args...)
	t.Fail()
}

func ptFatal(t PT, args ...any) {
	t.Helper()
	t.Log(args...)
	t.FailNow()
}
