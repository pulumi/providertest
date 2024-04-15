package pulumitest

import (
	"fmt"
	"time"
)

// A subset of *testing.T functionality used by pulumitest.
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
	t.Log(fmt.Sprintf(format, args...))
	t.Fail()
}

func ptFatalF(t PT, format string, args ...any) {
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
