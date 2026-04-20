package pulumitest_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestMissingProviderBinaryPath(t *testing.T) {
	t.Parallel()

	tt := mockT{T: t}
	pulumitest.NewPulumiTest(&tt, filepath.Join("testdata", "yaml_program"),
		opttest.LocalProviderPath("gcp", filepath.Join("provider_directory", "file_that_does_not_exist")),
	)

	assert.True(t, tt.Failed(), "expected test to fail")
}

func TestRequireYarnLinksEnforced(t *testing.T) {
	t.Parallel()

	tt := mockT{T: t}
	pulumitest.NewPulumiTest(&tt, filepath.Join("testdata", "nodejs_empty"),
		opttest.RequireYarnLinks(true),
	)

	assert.True(t, tt.Failed(), "expected test to fail")
}

func TestRequireYarnLinksSilenced(t *testing.T) {
	t.Parallel()

	tt := mockT{T: t}
	pulumitest.NewPulumiTest(&tt, filepath.Join("testdata", "nodejs_empty"),
		opttest.RequireYarnLinks(false),
	)

	assert.False(t, tt.Failed(), "expected test to pass")
}

func TestCleanupDestroyFailureWritesDestroyScript(t *testing.T) {
	t.Parallel()

	var tt *mockT
	var destroyScriptPath string

	t.Run("cleanup failure", func(t *testing.T) {
		tt = &mockT{T: t}
		test := pulumitest.NewPulumiTest(tt, filepath.Join("testdata", "yaml_program"),
			opttest.SkipInstall(),
			opttest.AttachProvider("random", func(context.Context, providers.PulumiTest) (providers.Port, error) {
				return 0, errors.New("boom")
			}),
		)

		destroyScriptPath = filepath.Join(test.WorkingDir(), "destroy.sh")
	})

	assert.True(t, tt.Failed(), "expected cleanup failure to mark the test as failed")
	_, err := os.Stat(destroyScriptPath)
	assert.NoError(t, err, "expected cleanup failure to leave a destroy script behind")
}
