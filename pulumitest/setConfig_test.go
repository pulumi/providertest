package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/assert"
)

func TestSetConfig(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program_with_config")
	test.SetConfig("passwordLength", "7")
	result := test.Up()

	outputs := result.Outputs
	assert.Len(t, outputs["password"].Value, 7)
}
