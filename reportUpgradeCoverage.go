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
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// This is a temporary helper method to assess upgrade resource coverage until better methods for
// tracking coverage are built. Run with -test.v to see the data logged. This finds all recorded
// GRPC states and traverses them to find the union of all resources used. It does not take into
// account if the corresponding tests are skipped or passing.
func ReportUpgradeCoverage(t *testing.T) {
	t.Helper()
	u := &upgradeCoverage{}
	dir := filepath.Join("testdata", "recorded", "TestProviderUpgrade")

	states := findFiles(t, dir, func(fn string) bool {
		filename := filepath.Base(fn)
		// Check for both the old name (state) from PulumiTest and the current name (stack).
		return filename == "stack.json" || filename == "state.json"
	})

	for _, s := range states {
		u.checkStateFile(t, s)
	}

	covered := u.resources
	t.Logf("Resources covered: %d", len(covered))

	sorted := []string{}
	for k := range covered {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	for _, s := range sorted {
		t.Logf("- %s", s)
	}
}

// Tracks resource coverage through upgrade tests.
type upgradeCoverage struct {
	resources map[string]struct{}
}

func findFiles(t *testing.T, dir string, matches func(string) bool) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if matches(path) {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	return files
}

func (u *upgradeCoverage) checkStateFile(t *testing.T, stateFile string) {
	type stack struct {
		Deployment struct {
			Resources []struct {
				Type string `json:"type"`
			} `json:"resources"`
		} `json:"deployment"`
	}
	b, err := os.ReadFile(stateFile)
	if err != nil {
		return // perhaps it did not exist, no matter
	}

	var st stack
	require.NoError(t, json.Unmarshal(b, &st))

	if u.resources == nil {
		u.resources = map[string]struct{}{}
	}

	for _, r := range st.Deployment.Resources {
		if strings.Contains(r.Type, "providers") {
			continue
		}
		if strings.Contains(r.Type, "Stack") {
			continue
		}
		u.resources[r.Type] = struct{}{}
	}
}
