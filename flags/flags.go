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
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	testModeEnvVarName          = "PULUMI_PROVIDER_TEST_MODE"
	testModeEnv, testModeEnvSet = os.LookupEnv(testModeEnvVarName)
)

type Flag interface {
	IsSet() bool

	// Explains why the flag is considered to be IsSet().
	WhySet() string

	// Explains why the flag is not considered to be IsSet().
	WhyNotSet() string
}

type stdFlag struct {
	flag     string
	flagVal  *bool
	modeName string
}

func newFlag(modeName, description string) *stdFlag {
	fl := fmt.Sprintf("provider-%s", modeName)
	return &stdFlag{
		flag:     fl,
		modeName: modeName,
		flagVal:  flag.Bool(fl, false, description),
	}
}

func (f *stdFlag) IsSet() bool {
	return f.WhySet() != ""
}

// Environment variables override flags.
// Flags are prefixed `-provider-`
// Environment variable can have multiple modes set e.g.
// PULUMI_PROVIDER_TEST_MODE=e2e,sdk-python
// PULUMI_PROVIDER_TEST_MODE=skip-e2e
func (f *stdFlag) WhySet() string {
	if testModeEnvSet {
		modes := strings.Split(testModeEnv, ",")
		for _, mode := range modes {
			if mode == f.modeName {
				return fmt.Sprintf("%s=%q contains %q",
					testModeEnvVarName, testModeEnv, f.modeName)
			}
		}
	}
	if f.flagVal != nil && *f.flagVal {
		return fmt.Sprintf("-%s flag was set", f.flag)
	}
	return ""
}

func (f *stdFlag) WhyNotSet() string {
	if f.IsSet() {
		return ""
	}
	var reasons []string
	reasons = append(reasons, fmt.Sprintf("-%s flag is unset", f.flag))
	if testModeEnvSet {
		reasons = append(reasons, fmt.Sprintf("%s=%q does not contain %q",
			testModeEnvVarName, testModeEnv, f.modeName))
	} else {
		reasons = append(reasons, fmt.Sprintf("%s=%q is unset",
			testModeEnvVarName, f.modeName))
	}
	return strings.Join(reasons, " and ")
}

type orFlag struct {
	a Flag
	b Flag
}

func (f *orFlag) IsSet() bool {
	return f.WhySet() != ""
}

func (f *orFlag) WhySet() string {
	if reason := f.a.WhySet(); reason != "" {
		return reason
	}
	if reason := f.b.WhySet(); reason != "" {
		return reason
	}
	return ""
}

func (f *orFlag) WhyNotSet() string {
	if f.IsSet() {
		return ""
	}
	mixed := fmt.Sprintf("%s and %s", f.a.WhyNotSet(), f.b.WhyNotSet())
	parts := strings.Split(mixed, " and ")
	sort.Strings(parts)
	return fmt.Sprintf("\n%s", strings.Join(parts, "\n"))
}

var (
	SkipE2e       = newFlag("skip-e2e", "Skip e2e provider tests")
	E2e           = newFlag("e2e", "Enable full e2e provider tests, otherwise uses quick mode by default")
	Sdk           = newFlag("sdk", "Enable all SDK provider tests")
	SdkCsharp     = &orFlag{newFlag("sdk-csharp", "Enable C# SDK provider tests"), Sdk}
	SdkGo         = &orFlag{newFlag("sdk-go", "Enable Go SDK provider tests"), Sdk}
	SdkPython     = &orFlag{newFlag("sdk-python", "Enable Python SDK provider tests"), Sdk}
	SdkTypescript = &orFlag{newFlag("sdk-typescript", "Enable TypeScript SDK provider tests"), Sdk}
	Snapshot      = newFlag("snapshot", "Create snapshots for use with quick e2e tests")
)
