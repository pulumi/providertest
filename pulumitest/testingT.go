package pulumitest

import (
	"time"
)

type T interface {
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
