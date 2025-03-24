package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestInstallStack(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program", opttest.SkipInstall(), opttest.SkipStackCreate())
	assert.Nil(t, test.CurrentStack())

	test.InstallStack(t, "teststack")

	assert.Equal(t, "teststack", test.CurrentStack().Name())
}
