// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	sdkGo         = flag.Bool("provider-sdk-go", false, "Enable Go SDK provider tests")
	sdkPython     = flag.Bool("provider-sdk-python", false, "Enable Python SDK provider tests")
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

func IsSdkGo() bool {
	return parseFlag("sdk-go", sdkGo)
}

func IsSdkPython() bool {
	return parseFlag("sdk-python", sdkPython)
}

func IsSdkTypescript() bool {
	return parseFlag("sdk-typescript", sdkTypescript)
}

func IsSnapshot() bool {
	return parseFlag("snapshot", snapshot)
}
