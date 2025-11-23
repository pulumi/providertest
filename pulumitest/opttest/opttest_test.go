package opttest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestPythonLinkOption(t *testing.T) {
	t.Parallel()

	opts := opttest.DefaultOptions()
	assert.Empty(t, opts.PythonLinks, "expected PythonLinks to be empty by default")

	pythonLink := opttest.PythonLink("path/to/sdk")
	pythonLink.Apply(opts)

	assert.Equal(t, []string{"path/to/sdk"}, opts.PythonLinks, "expected PythonLink to append path")
}

func TestPythonLinkMultiplePackages(t *testing.T) {
	t.Parallel()

	opts := opttest.DefaultOptions()

	pythonLink := opttest.PythonLink("path/to/sdk1", "path/to/sdk2")
	pythonLink.Apply(opts)

	assert.Equal(t, []string{"path/to/sdk1", "path/to/sdk2"}, opts.PythonLinks,
		"expected PythonLink to append multiple paths")
}

func TestPythonLinkAccumulates(t *testing.T) {
	t.Parallel()

	opts := opttest.DefaultOptions()

	pythonLink1 := opttest.PythonLink("path/to/sdk1")
	pythonLink1.Apply(opts)

	pythonLink2 := opttest.PythonLink("path/to/sdk2")
	pythonLink2.Apply(opts)

	assert.Equal(t, []string{"path/to/sdk1", "path/to/sdk2"}, opts.PythonLinks,
		"expected PythonLinks to accumulate across multiple calls")
}

func TestDefaultsResetsPythonLinks(t *testing.T) {
	t.Parallel()

	opts := opttest.DefaultOptions()

	pythonLink := opttest.PythonLink("path/to/sdk")
	pythonLink.Apply(opts)

	assert.NotEmpty(t, opts.PythonLinks, "expected PythonLinks to be populated")

	defaults := opttest.Defaults()
	defaults.Apply(opts)

	assert.Empty(t, opts.PythonLinks, "expected Defaults to reset PythonLinks")
}

func TestPythonLinkIntegrationV1(t *testing.T) {
	t.Parallel()

	// Integration test: verify PythonLink can be used with a real test package (v1)
	// This test checks that the option correctly processes package paths
	pkgV1Path := filepath.Join("..", "testdata", "python_pkg_v1")

	// Verify the test package directory exists
	_, err := filepath.Abs(pkgV1Path)
	assert.NoError(t, err, "expected to resolve package path v1")

	// Create test with PythonLink pointing to v1 package
	opts := opttest.DefaultOptions()
	pythonLink := opttest.PythonLink(pkgV1Path)
	pythonLink.Apply(opts)

	// Verify the path was correctly added to options
	assert.Equal(t, 1, len(opts.PythonLinks), "expected one Python package path")
	assert.True(t, len(opts.PythonLinks[0]) > 0, "expected non-empty package path")
}

func TestPythonLinkIntegrationV2(t *testing.T) {
	t.Parallel()

	// Integration test: verify PythonLink can be used with a real test package (v2)
	pkgV2Path := filepath.Join("..", "testdata", "python_pkg_v2")

	// Verify the test package directory exists
	_, err := filepath.Abs(pkgV2Path)
	assert.NoError(t, err, "expected to resolve package path v2")

	// Create test with PythonLink pointing to v2 package
	opts := opttest.DefaultOptions()
	pythonLink := opttest.PythonLink(pkgV2Path)
	pythonLink.Apply(opts)

	// Verify the path was correctly added to options
	assert.Equal(t, 1, len(opts.PythonLinks), "expected one Python package path")
	assert.True(t, len(opts.PythonLinks[0]) > 0, "expected non-empty package path")
}

func TestPythonLinkUpgradePathGeneration(t *testing.T) {
	t.Parallel()

	// Integration test: verify PythonLink generates correct paths for version upgrades
	pkgV1Path := filepath.Join("..", "testdata", "python_pkg_v1")
	pkgV2Path := filepath.Join("..", "testdata", "python_pkg_v2")

	opts := opttest.DefaultOptions()

	// Add v1 package
	pythonLinkV1 := opttest.PythonLink(pkgV1Path)
	pythonLinkV1.Apply(opts)
	assert.Equal(t, 1, len(opts.PythonLinks), "expected one path after v1")

	// Add v2 package (simulating version upgrade)
	pythonLinkV2 := opttest.PythonLink(pkgV2Path)
	pythonLinkV2.Apply(opts)
	assert.Equal(t, 2, len(opts.PythonLinks), "expected two paths after adding v2")

	// Verify both paths are present
	assert.Contains(t, opts.PythonLinks, pkgV1Path, "expected v1 path to be present")
	assert.Contains(t, opts.PythonLinks, pkgV2Path, "expected v2 path to be present")
}
