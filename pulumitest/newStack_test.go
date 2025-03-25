package pulumitest_test

import (
	"path/filepath"
	"testing"

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
