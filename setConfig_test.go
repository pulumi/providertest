package pulumitest_test

import (
	"testing"

	"github.com/pulumi/pulumitest"
	"github.com/stretchr/testify/assert"
)

func TestSetConfig(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program_with_config")
	test.SetConfig(t, "passwordLength", "7")
	result := test.Up(t)

	outputs := result.Outputs
	assert.Len(t, outputs["password"].Value, 7)
}
