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
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
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
	t.Run("e2e", pt.runE2eTest)
}

func (pt *ProviderTest) runE2eTest(t *testing.T) {
	t.Helper()
	if flags.SkipE2e() {
		t.Skip("Skipping e2e tests due to either -provider-skip-e2e or PULUMI_PROVIDER_TEST_MODE=skip-e2e being set")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Logf("starting providers")
	providers, err := startProviders(ctx, pt)
	if err != nil {
		t.Errorf("failed to start providers: %v", err)
		return
	}
	opts := buildProgramTestOptions(pt, providers)
	// If we're not running full E2E test, we want to only run the non-effecting steps.
	if !flags.IsE2e() {
		// We can't currently do preview only, so this is as close as we can get.
		opts.SkipEmptyPreviewUpdate = true
		opts.SkipExportImport = true
		opts.SkipRefresh = true
		opts.SkipUpdate = true
	}
	integration.ProgramTest(t, &opts)
}

func startProviders(ctx context.Context, pt *ProviderTest) ([]*ProviderAttach, error) {
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

func buildProgramTestOptions(pt *ProviderTest, runningProviders []*ProviderAttach) integration.ProgramTestOptions {
	editDirs := make([]integration.EditDir, len(pt.editDirs))
	for i, ed := range pt.editDirs {
		editDirs[i] = integration.EditDir{
			Dir:      ed.dir,
			Additive: !ed.clean,
		}
	}
	env := []string{getProviderAttachEnv(runningProviders)}
	return integration.ProgramTestOptions{
		Dir:      pt.dir,
		EditDirs: editDirs,
		Env:      env,
	}
}

func getProviderAttachEnv(runningProviders []*ProviderAttach) string {
	env := make([]string, 0, len(runningProviders))
	for _, rp := range runningProviders {
		env = append(env, fmt.Sprintf("%s:%d", rp.Name, rp.Port))
	}
	debugProviders := strings.Join(env, ",")
	return fmt.Sprintf("PULUMI_DEBUG_PROVIDERS=%s", debugProviders)
}
