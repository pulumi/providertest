package pulumitest

import (
	"os"
	"strings"
)

var skipDestroyOnFailure = (func() func() bool {
	value, ok := os.LookupEnv("PULUMITEST_SKIP_DESTROY_ON_FAILURE")
	skipDestroy := ok && strings.EqualFold(value, "true")
	return func() bool { return skipDestroy }
})()

var runningInCI = (func() func() bool {
	_, ok := os.LookupEnv("CI")
	return func() bool { return ok }
})()
