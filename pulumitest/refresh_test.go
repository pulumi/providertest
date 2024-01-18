package pulumitest_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
)

func TestRefresh(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")
	test.Up()

	result := test.Refresh()

	assertrefresh.HasNoChanges(t, result)
}
