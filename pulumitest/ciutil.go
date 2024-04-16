package pulumitest

import (
	"os"
)

var runningInCI = (func() func() bool {
	_, ok := os.LookupEnv("CI")
	return func() bool { return ok }
})()
