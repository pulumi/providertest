package pulumitest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/optrun"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestRunInIsolation(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		test := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test.Run(func(test *pulumitest.PulumiTest) {
			test.Up()
		})
		preview := test.Preview()
		assertpreview.HasNoChanges(t, preview)
	})

	t.Run("additional options", func(t *testing.T) {
		test := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		// Disable the gRPC log to ensure options are propagated correctly.
		test.Run(func(test *pulumitest.PulumiTest) {
			test.Up()
			assert.Nil(t, test.GrpcLog(), "expected no grpc logs to be captured")
		}, optrun.WithOpts(opttest.DisableGrpcLog()))
		test.Preview()
	})

	t.Run("cached state", func(t *testing.T) {
		cacheDir := t.TempDir()
		cacheCalls := 0
		cachePath := filepath.Join(cacheDir, "stack.yaml")

		test1 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test1.Run(func(test *pulumitest.PulumiTest) {
			test.Up()
			cacheCalls++
		}, optrun.WithCache(cachePath))
		preview1 := test1.Preview()

		test2 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test2.Run(func(test *pulumitest.PulumiTest) {
			test.Up()
			cacheCalls++
		}, optrun.WithCache(cachePath))
		preview2 := test2.Preview()

		assert.Equal(t, 1, cacheCalls, "expected cached method to be called exactly once")
		assert.Equal(t, preview1.ChangeSummary, preview2.ChangeSummary, "expected uncached and cached preview to be the same")
	})
}
