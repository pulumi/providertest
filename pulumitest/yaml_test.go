package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/providertest/pulumiyaml"
	"github.com/stretchr/testify/assert"
)

func TestYaml(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		pt := pulumitest.NewInlinePulumiTest(t, pulumiyaml.NewProgram())
		pt.Preview(t)
		pt.Up(t)
		pt.Destroy(t)
	})
	t.Run("random resource", func(t *testing.T) {
		program := pulumiyaml.NewProgram()
		randomString := pulumiyaml.Resource{
			Type: "random:RandomString",
			Properties: map[string]any{
				"length":  10,
				"special": true,
			},
		}
		err := program.AddResource("aString", randomString)
		assert.NoError(t, err)

		pt := pulumitest.NewInlinePulumiTest(t, program)

		upResult := pt.Up(t)
		assertup.CreateCountEquals(t, upResult, 2)

		randomString.Properties["length"] = 20
		_, err = program.UpdateResource("aString", randomString)
		assert.NoError(t, err)
		pt.WriteProject(t, program)
		upResult2 := pt.Up(t)
		assertup.ReplaceCountEquals(t, upResult2, 1)

		_, err = program.RemoveResource("aString")
		assert.NoError(t, err)
		pt.WriteProject(t, program)
		upResult3 := pt.Up(t)
		assertup.DeleteCountEquals(t, upResult3, 1)
	})
}
