package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
)

func TestDestroy(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")
	test.Up()
	test.Destroy()
}
