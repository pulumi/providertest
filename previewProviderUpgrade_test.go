package providertest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest"
	"github.com/pulumi/providertest/optproviderupgrade"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestPreviewUpgradeCached(t *testing.T) {
	t.Parallel()
	cacheDir := t.TempDir()
	test := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"),
		opttest.DownloadProviderVersion("random", "4.15.0"))

	uncachedPreviewResult := providertest.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach())
	assertpreview.HasNoReplacements(t, uncachedPreviewResult)
	assertpreview.HasNoChanges(t, uncachedPreviewResult)

	cachedPreviewResult := providertest.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach())
	assert.Equal(t, uncachedPreviewResult, cachedPreviewResult, "expected uncached and cached preview to be the same")
}

func TestPreviewUpgradeWithKnownSourceEdit(t *testing.T) {
	t.Parallel()
	cacheDir := t.TempDir()
	test := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"),
		opttest.DownloadProviderVersion("random", "4.15.0"))

	previewResult := providertest.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach(),
		optproviderupgrade.NewSourcePath(filepath.Join("pulumitest", "testdata", "yaml_program_updated")),
	)

	assert.Contains(t, previewResult.StdOut, "random:index:RandomPassword password create")
}
