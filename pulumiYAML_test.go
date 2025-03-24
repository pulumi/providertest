package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/require"
)

func TestReplaceProgram(t *testing.T) {
	t.Parallel()

	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	test.WritePulumiYaml(t, `
name: yaml_empty
runtime: yaml
outputs:
    output: "output"`)

	res := test.Up(t)

	require.Equal(t, "output", res.Outputs["output"].Value)
}

func TestReadProgram(t *testing.T) {
	t.Parallel()

	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	program := `
name: yaml_empty
runtime: yaml
outputs:
    output: "output"`

	test.WritePulumiYaml(t, program)

	res := test.Up(t)
	require.Equal(t, "output", res.Outputs["output"].Value)

	require.Equal(t, program, test.ReadPulumiYaml(t))
}
