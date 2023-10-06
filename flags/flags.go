package flags

import (
	"flag"
	"os"
	"strings"
)

// Environment variables override flags.
// Flags are prefixed `-provider-`
// Environment variable can have multiple modes set e.g.
// PULUMI_PROVIDER_TEST_MODE=e2e,sdk-python
// PULUMI_PROVIDER_TEST_MODE=skip-e2e
func parseFlag(modeName string, flagVal *bool) bool {
	env, ok := os.LookupEnv("PULUMI_PROVIDER_TEST_MODE")
	if ok {
		modes := strings.Split(env, ",")
		for _, mode := range modes {
			if mode == modeName {
				return true
			}
		}
	}
	return flagVal != nil && *flagVal
}

var (
	skipE2e       = flag.Bool("provider-skip-e2e", false, "Skip e2e provider tests")
	e2e           = flag.Bool("provider-e2e", false, "Enable full e2e provider tests, otherwise uses quick mode by default")
	sdkCsharp     = flag.Bool("provider-sdk-csharp", false, "Enable C# SDK provider tests")
	sdkPython     = flag.Bool("provider-sdk-python", false, "Enable Python SDK provider tests")
	sdkGo         = flag.Bool("provider-sdk-go", false, "Enable Go SDK provider tests")
	sdkTypescript = flag.Bool("provider-sdk-typescript", false, "Enable TypeScript SDK provider tests")
	snapshot      = flag.Bool("provider-snapshot", false, "Create snapshots for use with quick e2e tests")
)

func SkipE2e() bool {
	return parseFlag("skip-e2e", skipE2e)
}

func IsE2e() bool {
	return parseFlag("e2e", e2e)
}

func IsSdkCsharp() bool {
	return parseFlag("sdk-csharp", sdkCsharp)
}

func IsSdkPython() bool {
	return parseFlag("sdk-python", sdkPython)
}

func IsSdkGo() bool {
	return parseFlag("sdk-go", sdkGo)
}

func IsSdkTypescript() bool {
	return parseFlag("sdk-typescript", sdkTypescript)
}

func IsSnapshot() bool {
	return parseFlag("snapshot", snapshot)
}
