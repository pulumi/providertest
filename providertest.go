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

package providertest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/providertest/flags"
)

// Deprecated: Use pulumitest.PulumiTest instead. This will be removed in a future release.
type ProviderTest struct {
	dir              string
	providerStartups []StartProvider
	updateSteps      []UpdateStep
	config           map[string]string
	e2eOptions       []E2eOption
	skipSdk          map[string][]any
	upgradeOpts      providerUpgradeOpts
}

// NewProviderTest creates a new provider test with the initial directory to be tested.
//
// Deprecated: Use pulumitest.NewPulumiTest instead. This will be removed in a future release.
func NewProviderTest(dir string, opts ...Option) *ProviderTest {
	pt := &ProviderTest{
		dir:     dir,
		config:  map[string]string{},
		skipSdk: map[string][]any{},
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

type Option func(*ProviderTest)

func (pt *ProviderTest) Configure(opts ...Option) *ProviderTest {
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

func WithConfig(key, value string) Option {
	return func(pt *ProviderTest) {
		pt.config[key] = value
	}
}

func WithE2eOptions(opts ...E2eOption) Option {
	return func(pt *ProviderTest) {
		pt.e2eOptions = append(pt.e2eOptions, opts...)
	}
}

func WithSkipSdk(language string, reasonArgs ...any) Option {
	return func(pt *ProviderTest) {
		pt.skipSdk[language] = reasonArgs
	}
}

// WithUpdateStep adds a step to the test will be applied before then executing an update.
func WithUpdateStep(opts ...UpdateStepOption) Option {
	return func(pt *ProviderTest) {
		pt.updateSteps = append(pt.updateSteps, UpdateStep{pt: pt})
	}
}

type UpdateStep struct {
	// A reference to the parent provider test
	pt    *ProviderTest
	dir   *string
	clean bool
}

type UpdateStepOption func(*UpdateStep)

// UpdateStepDir fetches files from the dir before performing the update.
// If dir is a relative path, it will be resolved relative to the original test directory.
func UpdateStepDir(dir string) UpdateStepOption {
	return func(us *UpdateStep) {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(us.pt.dir, dir)
		}
		us.dir = &dir
	}
}

// UpdateStepClean will remove files from the directory which were removed in this step compared to the previous step's directory.
func UpdateStepClean() UpdateStepOption {
	return func(ed *UpdateStep) {
		ed.clean = true
	}
}

// StartProvider is a function that starts a provider and returns the name and port it is listening on.
// When the test is complete, the context will be cancelled and the provider should exit.
type StartProvider func(ctx context.Context) (*ProviderAttach, error)

type ProviderAttach struct {
	// Name of the provider e.g. "aws"
	Name string
	// Port the provider is listening on
	Port int
}

// WithProvider adds a provider to be started and attached for the test run.
func WithProvider(start StartProvider) Option {
	return func(pt *ProviderTest) {
		pt.providerStartups = append(pt.providerStartups, start)
	}
}

// Run starts executing the configured tests
func (pt *ProviderTest) Run(t *testing.T) {
	t.Helper()
	t.Run("e2e", func(t *testing.T) {
		t.Helper()
		if flags.SkipE2e.IsSet() {
			t.Skipf("Skipping e2e tests because %s", flags.SkipE2e.WhySet())
			return
		}
		pt.RunE2e(t, flags.E2e.IsSet(), pt.e2eOptions...)
	})
	t.Run("sdk-csharp", func(t *testing.T) {
		t.Helper()
		if reason, skip := pt.skipSdk["csharp"]; skip {
			t.Skip(reason...)
		}
		if !flags.SdkCsharp.IsSet() {
			t.Skipf("Skipping C# SDK tests because %s", flags.SdkCsharp.WhyNotSet())
			return
		}
		pt.RunSdk(t, "csharp")
	})
	t.Run("sdk-go", func(t *testing.T) {
		t.Helper()
		if reason, skip := pt.skipSdk["go"]; skip {
			t.Skip(reason...)
		}
		if !flags.SdkGo.IsSet() {
			t.Skipf("Skipping Go SDK tests because %s", flags.SdkGo.WhyNotSet())
			return
		}
		pt.RunSdk(t, "go")
	})
	t.Run("sdk-python", func(t *testing.T) {
		t.Helper()
		if reason, skip := pt.skipSdk["python"]; skip {
			t.Skip(reason...)
		}
		if !flags.SdkPython.IsSet() {
			t.Skipf("Skipping Python SDK tests because %s", flags.SdkPython.WhyNotSet())
			return
		}
		pt.RunSdk(t, "python")
	})
	t.Run("sdk-typescript", func(t *testing.T) {
		t.Helper()
		if reason, skip := pt.skipSdk["typescript"]; skip {
			t.Skip(reason...)
		}
		if !flags.SdkTypescript.IsSet() {
			t.Skipf("Skipping Typescript SDK tests because %s", flags.SdkTypescript.WhyNotSet())
			return
		}
		pt.RunSdk(t, "typescript")
	})

	t.Run("upgrade-snapshot", func(t *testing.T) {
		t.Helper()
		if flags.Snapshot.IsSet() {
			t.Logf("Recording baseline behavior because %s", flags.Snapshot.WhySet())
			pt.VerifyUpgradeSnapshot(t)
		} else {
			t.Skipf("Skip recording baseline behavior because %s", flags.Snapshot.WhyNotSet())
		}
	})

	for _, m := range UpgradeTestModes() {
		t.Run(fmt.Sprintf("upgrade-%s", m), func(t *testing.T) {
			t.Helper()
			pt.VerifyUpgrade(t, m)
		})
	}
}

func StartProviders(ctx context.Context, providerStartups ...StartProvider) ([]*ProviderAttach, error) {
	providers := make([]*ProviderAttach, len(providerStartups))
	for i, start := range providerStartups {
		provider, err := start(ctx)
		if err != nil {
			return nil, err
		}
		providers[i] = provider
	}
	return providers, nil
}

func GetProviderAttachEnv(runningProviders []*ProviderAttach) string {
	env := make([]string, 0, len(runningProviders))
	for _, rp := range runningProviders {
		env = append(env, fmt.Sprintf("%s:%d", rp.Name, rp.Port))
	}
	debugProviders := strings.Join(env, ",")
	return fmt.Sprintf("PULUMI_DEBUG_PROVIDERS=%s", debugProviders)
}
