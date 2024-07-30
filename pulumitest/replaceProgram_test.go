package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/require"
)

func TestReplaceProgram(t *testing.T) {
	t.Parallel()

	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	test.ReplaceProgram(`
name: yaml_empty
runtime: yaml
outputs:
    output: "output"`)

	res := test.Up()

	require.Equal(t, "output", res.Outputs["output"].Value)
}
