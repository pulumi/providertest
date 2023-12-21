package providertest_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

func TestPreviewUpgradeCached(t *testing.T) {
	cacheDir := t.TempDir()
	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"), opttest.TestInPlace())
	baselineProviderConfig := providertest.ProviderConfiguration{
		Name:                  "random",
		DownloadBinaryVersion: "4.15.0",
		DisableAttach:         true,
	}

	uncachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{
		CacheDir: cacheDir,
	})
	assertpreview.HasNoChanges(t, uncachedPreviewResult)

	cachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{
		CacheDir: cacheDir,
	})
	assertpreview.HasNoChanges(t, cachedPreviewResult)
}

func TestPreviewUpgradeNoTest(t *testing.T) {
	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"), opttest.TestInPlace())
	baselineProviderConfig := providertest.ProviderConfiguration{
		Name:                  "random",
		DownloadBinaryVersion: "4.15.0",
		DisableAttach:         true,
	}

	uncachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{})
	assertpreview.HasNoChanges(t, uncachedPreviewResult)
}

func TestRecordSnapshot(t *testing.T) {
	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"))
	snapshot := providertest.SnapshotUp(source, providertest.ProviderConfiguration{
		Name:                  "random",
		DownloadBinaryVersion: "4.15.0",
		DisableAttach:         true,
	})
	assert.NotNil(t, snapshot)
	previewResult := providertest.PreviewUpdateFromSnapshot(source, snapshot)
	assertpreview.HasNoChanges(t, previewResult)
}
