package pulumitest_test

import (
	"testing"

	"github.com/pulumi/pulumitest"
	"github.com/pulumi/pulumitest/assertrefresh"
)

func TestRefresh(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")
	test.Up(t)

	result := test.Refresh(t)

	assertrefresh.HasNoChanges(t, result)
}
