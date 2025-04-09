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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonMatch(t *testing.T) {
	t.Parallel()
	AssertJSONMatchesPattern(t, []byte(`1`), []byte(`1`))
	AssertJSONMatchesPattern(t, []byte(`"*"`), []byte(`1`))
	AssertJSONMatchesPattern(t, []byte(`"*"`), []byte(`2`))
	AssertJSONMatchesPattern(t, []byte(`{"\\": "*"}`), []byte(`"*"`))
	AssertJSONMatchesPattern(t, []byte(`[1, "*", 3]`), []byte(`[1, 2, 3]`))
	AssertJSONMatchesPattern(t, []byte(`{"foo": "*", "bar": 3}`), []byte(`{"foo": 1, "bar": 3}`))
}

func TestJsonListLengthMistmatch(t *testing.T) {
	mt := &mockTestingT{}
	assertJSONMatchesPattern(mt, []byte(`[1, 3]`), []byte(`[1, 2, 3]`))
	require.NotEmpty(t, mt.errors)
	require.Equal(t, 1, len(mt.errors))
	require.Equal(t, "[#]: expected an array of length 2, but got [\n  1,\n  2,\n  3\n]", mt.errors[0])
}

func TestCatchAllPattern(t *testing.T) {
	t.Parallel()

	t.Run("Accept extraneous keys", func(t *testing.T) {
		t.Parallel()

		assertJSONMatchesPattern(t,
			[]byte(`{"foo": "bar", "*": "*"}`),
			[]byte(`{"foo": "bar", "baz": 1}`))
	})

	t.Run("Reject extraneous keys that do not match", func(t *testing.T) {
		t.Parallel()

		mt := &mockTestingT{}
		assertJSONMatchesPattern(mt,
			[]byte(`{"foo": "bar", "*": "2"}`),
			[]byte(`{"foo": "bar", "baz": 1, "x": "2"}`))
		require.NotEmpty(t, mt.errors)
		require.Equal(t, 1, len(mt.errors))
		require.Contains(t, mt.errors[0], `#["baz"]`)
	})
}

type mockTestingT struct {
	errors []string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) FailNow() {
	panic("FailNow")
}

var _ testingT = (*mockTestingT)(nil)
