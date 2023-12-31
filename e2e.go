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
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func (pt *ProviderTest) RunE2e(t *testing.T, runFullTest bool, options ...E2eOption) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Logf("starting providers")
	providers, err := StartProviders(ctx, pt.providerStartups...)
	if err != nil {
		t.Errorf("failed to start providers: %v", err)
		return
	}
	opts := buildProgramTestOptions(t, pt, providers)
	// If we're not running full E2E test, we want to only run the non-effecting steps.
	if !runFullTest {
		// TODO: We can't currently do preview only, so this is as close as we can get.
		opts.SkipEmptyPreviewUpdate = true
		opts.SkipExportImport = true
		opts.SkipRefresh = true
		opts.SkipUpdate = true
	}
	for _, opt := range options {
		opt(&opts)
	}
	integration.ProgramTest(t, &opts)
}

type E2eOption func(*integration.ProgramTestOptions)

func buildProgramTestOptions(t *testing.T, pt *ProviderTest, runningProviders []*ProviderAttach) integration.ProgramTestOptions {
	editDirs := make([]integration.EditDir, len(pt.updateSteps))
	for i, ed := range pt.updateSteps {
		if ed.dir == nil {
			t.Errorf("update step %d has no changes specified", i+1)
		}
		editDirs[i] = integration.EditDir{
			Dir:      *ed.dir,
			Additive: !ed.clean,
		}
	}
	env := []string{GetProviderAttachEnv(runningProviders)}
	return integration.ProgramTestOptions{
		Dir:      pt.dir,
		EditDirs: editDirs,
		Env:      env,
		Config:   pt.config,
	}
}
