package pulumitest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestCacheStackState(t *testing.T) {
	cacheDir := t.TempDir()
	cacheCalls := 0
	cachePath := filepath.Join(cacheDir, "stack.yaml")

	test1 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
	test1.CacheStackState(cachePath, func(test *pulumitest.PulumiTest) {
		test.Up()
		cacheCalls++
	})
	preview1 := test1.Preview()

	test2 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
	test2.CacheStackState(cachePath, func(test *pulumitest.PulumiTest) {
		test.Up()
		cacheCalls++
	})
	preview2 := test2.Preview()

	assert.Equal(t, 1, cacheCalls, "expected cached method to be called exactly once")
	assert.Equal(t, preview1.ChangeSummary, preview2.ChangeSummary, "expected uncached and cached preview to be the same")
}

func TestRenamingStackInState(t *testing.T) {
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "stack.yaml")

	test1 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.StackName("stack-1"))
	test1.CacheStackState(cachePath, func(test *pulumitest.PulumiTest) {
		test.Up()
	})
	preview1 := test1.Preview()
	assertpreview.HasNoChanges(t, preview1)

	test2 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
	test2.CacheStackState(cachePath, func(test *pulumitest.PulumiTest) {
		test.Up()
	})
	preview2 := test2.Preview()
	assertpreview.HasNoChanges(t, preview2)
}
