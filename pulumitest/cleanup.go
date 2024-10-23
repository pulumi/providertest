package pulumitest

import (
	"os"
	"strings"
)

func skipDestroyOnFailure() bool {
	value, ok := os.LookupEnv("PULUMITEST_SKIP_DESTROY_ON_FAILURE")
	skipDestroy := ok && strings.EqualFold(value, "true")
	return skipDestroy
}

func runningInCI() bool {
	_, ok := os.LookupEnv("CI")
	return ok
}

func shouldRetainFilesOnFailure() bool {
	if value, ok := os.LookupEnv("PULUMITEST_RETAIN_FILES_ON_FAILURE"); ok {
		return !strings.EqualFold(value, "false")
	}
	if skipDestroyOnFailure() {
		return true
	}
	return !runningInCI()
}
