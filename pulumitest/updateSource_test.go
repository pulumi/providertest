package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSource(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")
	test.Up()

	test.UpdateSource("testdata/yaml_program_updated")
	updated := test.Up()

	changes := *updated.Summary.ResourceChanges
	assert.Equal(t, 1, changes["create"])

}

func TestUpdateSourceError(t *testing.T) {
	t.Parallel()

	tt := &mockT{T: t}
	test := pulumitest.NewPulumiTest(tt, "testdata/yaml_program")
	test.UpdateSource(tt, "this-should-fail")

	assert.True(tt, tt.Failed())
}

type mockT struct {
	*testing.T
	failed bool
}

func (m *mockT) Fail() {
	m.failed = true
}

func (m *mockT) FailNow() {
	m.failed = true
}

func (m *mockT) Failed() bool {
	return m.failed
}
