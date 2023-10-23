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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// Code in this file is copied from the ProgramTest framework and modified to introduce
// extensibility points. It should eventually be upstreamed ideally so this is not needed here.
type programTestWrapper struct {
	pt *integration.ProgramTester
}

// Behaves just like ProgramTester.TestLifeCycleInitAndDestroy() but with custom inner test logic.
// This function was obtained by inlining TestLifeCycleInitAndDestroy implementation and
// generalizing it.
func (wrapper *programTestWrapper) lifecycleInitAndDestroy(
	t *testing.T,
	opts integration.ProgramTestOptions,
	customTest func() error,
) error {
	pt := wrapper.pt
	assert.Falsef(t, opts.RunUpdateTest, "RunUpdateTest is not supported")

	err := pt.TestLifeCyclePrepare()
	if err != nil {
		return fmt.Errorf("copying test to temp dir %s: %w", "<tmpdir>", err)
	}

	pt.TestFinished = false
	if opts.DestroyOnCleanup {
		t.Cleanup(pt.TestCleanUp)
	} else {
		defer pt.TestCleanUp()
	}

	err = pt.TestLifeCycleInitialize()
	if err != nil {
		return fmt.Errorf("initializing test project: %w", err)
	}

	destroyStack := func() {
		destroyErr := pt.TestLifeCycleDestroy()
		assert.NoError(t, destroyErr)
	}
	if opts.DestroyOnCleanup {
		// Allow other tests to refer to this stack until the test is complete.
		t.Cleanup(destroyStack)
	} else {
		// Ensure that before we exit, we attempt to destroy and remove the stack.
		defer destroyStack()
	}

	if err = customTest(); err != nil {
		return err
	}

	pt.TestFinished = true
	return nil
}

// Utility to load up Pulumi.yaml so we know things like what language the project is.
func getProjinfo(projectDir string) (*engine.Projinfo, error) {
	projfile := filepath.Join(projectDir, workspace.ProjectFile+".yaml")
	proj, err := workspace.LoadProject(projfile)
	if err != nil {
		return nil, err
	}
	return &engine.Projinfo{Proj: proj, Root: projectDir}, nil
}
