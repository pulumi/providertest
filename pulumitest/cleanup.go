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
	if shouldAlwaysRetainFiles() {
		return true // Always retain if retaining everything
	}
	if value, ok := os.LookupEnv("PULUMITEST_RETAIN_FILES_ON_FAILURE"); ok {
		// Must be set explicitly to "false" to disable, set to anything else to enable
		return !strings.EqualFold(value, "false")
	}
	if skipDestroyOnFailure() {
		return true // Always retain files if we're skipping destroy
	}
	return !runningInCI() // Don't retain files by default in CI
}

// Determines whether files should always be retained, based on `PULUMITEST_RETAIN_FILES` environment variable.
func shouldAlwaysRetainFiles() bool {
	if value, isSet := os.LookupEnv("PULUMITEST_RETAIN_FILES"); isSet {
		// Must be set explicitly to "true" or empty to enable. Unset or set to anything else to disable.
		return value == "" || strings.EqualFold(value, "true")
	}
	return false
}
