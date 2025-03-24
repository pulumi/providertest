package pulumitest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumitest"
	"github.com/pulumi/pulumitest/assertpreview"
	"github.com/pulumi/pulumitest/optrun"
	"github.com/pulumi/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestRunInIsolation(t *testing.T) {
	t.Parallel()
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		test := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
		})

		preview := test.Preview(t)
		assertpreview.HasNoChanges(t, preview)
	})

	t.Run("additional options", func(t *testing.T) {
		t.Parallel()
		test := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		// Disable the gRPC log to ensure options are propagated correctly.
		test.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
			assert.Nil(t, test.GrpcLog(t), "expected no grpc logs to be captured")
		}, optrun.WithOpts(opttest.DisableGrpcLog()))
		test.Preview(t)
	})

	t.Run("cached state", func(t *testing.T) {
		t.Parallel()
		cacheDir := t.TempDir()
		cacheCalls := 0
		cachePath := filepath.Join(cacheDir, "stack.yaml")

		test1 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test1.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
			cacheCalls++
		}, optrun.WithCache(cachePath))
		preview1 := test1.Preview(t)

		test2 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test2.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
			cacheCalls++
		}, optrun.WithCache(cachePath))
		preview2 := test2.Preview(t)

		assert.Equal(t, 1, cacheCalls, "expected cached method to be called exactly once")
		assert.Equal(t, preview1.ChangeSummary, preview2.ChangeSummary, "expected uncached and cached preview to be the same")
	})

	t.Run("fix cached stack name", func(t *testing.T) {
		t.Parallel()
		cacheDir := t.TempDir()
		cachePath := filepath.Join(cacheDir, "stack.yaml")

		test1 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.StackName("stack-1"))
		test1.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
		}, optrun.WithCache(cachePath))
		preview1 := test1.Preview(t)
		assertpreview.HasNoChanges(t, preview1)

		test2 := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
		test2.Run(t, func(test *pulumitest.PulumiTest) {
			test.Up(t)
		}, optrun.WithCache(cachePath))
		preview2 := test2.Preview(t)
		assertpreview.HasNoChanges(t, preview2)
	})
}
