package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	res := test.Import("random:index/randomString:RandomString", "str", "importedString", "")

	require.Contains(t, res.Stdout, "type: random:RandomString")
}
