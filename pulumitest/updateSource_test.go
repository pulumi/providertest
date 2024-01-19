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
