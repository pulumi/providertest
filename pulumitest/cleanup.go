package pulumitest

import (
	"os"
	"strings"
)

// We use this interesting contraption of immediately invoked callbacks to
// eagerly calculate these variables but ensure they're not mutated.

var skipDestroyOnFailure = (func() func() bool {
	value, ok := os.LookupEnv("PULUMITEST_SKIP_DESTROY_ON_FAILURE")
	skipDestroy := ok && strings.EqualFold(value, "true")
	return func() bool { return skipDestroy }
})()

var runningInCI = (func() func() bool {
	_, ok := os.LookupEnv("CI")
	return func() bool { return ok }
})()

var shouldRetainFilesOnFailure = (func() func() bool {
	if value, ok := os.LookupEnv("PULUMITEST_RETAIN_FILES_ON_FAILURE"); ok {
		if strings.EqualFold(value, "false") {
			return func() bool { return false }
		}
		return func() bool { return true }
	}
	if skipDestroyOnFailure() {
		return func() bool { return true }
	}
	return func() bool { return !runningInCI() }
})()
