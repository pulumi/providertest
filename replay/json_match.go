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

package replay

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testingT interface {
	assert.TestingT
	require.TestingT
}

// Assert that a given JSON document structurally matches a pattern.
//
// The pattern language supports the following constructs:
//
// "*" matches anything.
//
// {"\\": x} matches only JSON documents strictly equal to x. This pattern essentially escapes the sub-tree, for example
// use {"\\": "*"} to match only the literal string "*".
//
// An object pattern {"key1": "pattern1", "key2": "pattern2"} matches objects in a natural manner. By default it will
// only match objects with the exact set of keys specified. To tolerate extraneous keys, a catch-all pattern can be
// specified as follows, to match against all unspecified keys:
//
//	{"key1": "pattern1", "key2": "pattern2", "*": "catch-all-pattern"}
//
// In particular this can be used to ignore all extraneous keys:
//
//	{"key1": "pattern1", "key2": "pattern2", "*": "*"}
//
// It is possible to escape keys in an object pattern by prefixing them with "\\", for example this pattern:
//
//	{"\\*": "foo"}
//
// This pattern will only match the object {"*": "foo"}, that is the wildcard is interpreted literally and not as the
// catch-all pattern.
func AssertJSONMatchesPattern(
	t *testing.T,
	expectedPattern json.RawMessage,
	actual json.RawMessage,
) {
	if len(expectedPattern) == 0 {
		require.Fail(t, "Expected response was missing")
	}
	assertJSONMatchesPattern(t, expectedPattern, actual)
}

func assertJSONMatchesPattern(
	t testingT,
	expectedPattern json.RawMessage,
	actual json.RawMessage,
) {
	var p, a interface{}

	if err := json.Unmarshal(expectedPattern, &p); err != nil {
		require.NoError(t, err)
	}

	if err := json.Unmarshal(actual, &a); err != nil {
		require.NoError(t, err)
	}

	match(t, "#", p, a)
}

func match(t testingT, path string, p, a interface{}) {
	switch pp := p.(type) {
	case string:
		if pp != "*" {
			assertJSONEquals(t, path, p, a)
		}
	case []interface{}:
		aa, ok := a.([]interface{})
		if !ok {
			t.Errorf("[%s]: expected an array, but got %s", path, prettyJSON(t, a))
			return
		}
		if len(aa) != len(pp) {
			t.Errorf("[%s]: expected an array of length %d, but got %s",
				path, len(pp), prettyJSON(t, a))
			return
		}
		for i, pv := range pp {
			av := aa[i]
			match(t, fmt.Sprintf("%s[%d]", path, i), pv, av)
		}
	case map[string]interface{}:
		matchObjectPattern(t, path, pp, a)
	default:
		assertJSONEquals(t, path, p, a)
	}
}

type objectPattern struct {
	keyPatterns        map[string]any
	catchAllPattern    any
	hasCatchAllPattern bool
}

func (p *objectPattern) sortedKeyUnion(value map[string]any) []string {
	var keys []string
	for k := range p.keyPatterns {
		keys = append(keys, k)
	}
	for k := range value {
		if _, seen := p.keyPatterns[k]; seen {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func compileObjectPattern(pattern map[string]any) (objectPattern, error) {
	o := objectPattern{
		keyPatterns: map[string]any{},
	}

	var err error
	for k, v := range pattern {
		if k == "*" {
			o.hasCatchAllPattern = true
			o.catchAllPattern = v
			continue
		}

		// Keys in object patterns may be escaped.
		cleanKey := strings.TrimPrefix(k, "\\")

		if _, conflict := o.keyPatterns[cleanKey]; conflict {
			err = errors.Join(err, fmt.Errorf("object key pattern %q specified more than once", cleanKey))
		}

		o.keyPatterns[cleanKey] = v
	}

	if err != nil {
		return objectPattern{}, err
	}

	return o, nil
}

func matchObjectPattern(t testingT, path string, pattern map[string]any, value any) {
	if esc, isEsc := detectEscape(pattern); isEsc {
		assertJSONEquals(t, path, esc, value)
		return
	}

	objPattern, err := compileObjectPattern(pattern)
	if err != nil {
		t.Errorf("[%s]: %v", err)
		return
	}

	aa, ok := value.(map[string]interface{})
	if !ok {
		t.Errorf("[%s]: expected an object, but got %s", path, prettyJSON(t, value))
		return
	}

	for _, k := range objPattern.sortedKeyUnion(aa) {
		pv, gotPV := objPattern.keyPatterns[k]
		av, gotAV := aa[k]
		subPath := fmt.Sprintf("%s[%q]", path, k)
		switch {
		case gotPV && gotAV:
			match(t, subPath, pv, av)
		case !gotPV && gotAV && !objPattern.hasCatchAllPattern:
			t.Errorf("[%s] unexpected value %s", subPath, prettyJSON(t, av))
		case !gotPV && gotAV && objPattern.hasCatchAllPattern:
			match(t, subPath, objPattern.catchAllPattern, av)
		case gotPV && !gotAV:
			t.Errorf("[%s] missing a required value", subPath)
		}
	}
}

func detectEscape(m map[string]interface{}) (interface{}, bool) {
	if len(m) != 1 {
		return nil, false
	}
	for k, v := range m {
		if k == "\\" {
			return v, true
		}
	}
	return nil, false
}

func assertJSONEquals(t testingT, path string, expected, actual interface{}) {
	assert.Equalf(t, prettyJSON(t, expected), prettyJSON(t, actual), "at %s", path)
}

func prettyJSON(t testingT, msg interface{}) string {
	bytes, err := json.MarshalIndent(msg, "", "  ")
	assert.NoError(t, err)
	return string(bytes)
}
