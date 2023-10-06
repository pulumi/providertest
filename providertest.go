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
	"strings"
	"testing"

	"github.com/pulumi/providertest/flags"
)

type ProviderTest struct {
	dir              string
	providerStartups []StartProvider
	editDirs         []EditDir
}

// NewProviderTest creates a new provider test with the initial directory to be tested.
func NewProviderTest(dir string, opts ...Option) *ProviderTest {
	pt := &ProviderTest{
		dir: dir,
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

type Option func(*ProviderTest)

func (pt *ProviderTest) Configure(opts ...Option) {
	for _, opt := range opts {
		opt(pt)
	}
}

// WithEditDir adds a step to the test which will overwrite the contents of the directory then execute an update.
// If this is a relative path, it will be resolved relative to the original test directory.
func WithEditDir(dir string, opts ...EditDir) Option {
	return func(pt *ProviderTest) {
		pt.editDirs = append(pt.editDirs, EditDir{dir: dir})
	}
}

type EditDir struct {
	dir   string
	clean bool
}

type EditDirOption func(*EditDir)

// WithClean will remove files from the directory which are not in the edit directory but were in the original directory.
func WithClean() EditDirOption {
	return func(ed *EditDir) {
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
		if flags.SkipE2e() {
			t.Skip("Skipping e2e tests due to either -provider-skip-e2e or PULUMI_PROVIDER_TEST_MODE=skip-e2e being set")
		}
		pt.RunE2e(t, flags.IsE2e())
	})
}

func (pt *ProviderTest) StartProviders(ctx context.Context) ([]*ProviderAttach, error) {
	providers := make([]*ProviderAttach, len(pt.providerStartups))
	for i, start := range pt.providerStartups {
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
