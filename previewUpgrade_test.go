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

// func TestPreviewUpgradeCached(t *testing.T) {
// 	cacheDir := t.TempDir()
// 	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"), opttest.TestInPlace())
// 	baselineProviderConfig := providertest.ProviderConfiguration{
// 		Name:                  "random",
// 		DownloadBinaryVersion: "4.15.0",
// 		DisableAttach:         true,
// 	}

// 	uncachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{
// 		CacheDir: cacheDir,
// 	})
// 	assertpreview.HasNoChanges(t, uncachedPreviewResult)

// 	cachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{
// 		CacheDir: cacheDir,
// 	})
// 	assertpreview.HasNoChanges(t, cachedPreviewResult)
// }

// func TestPreviewUpgradeNoTest(t *testing.T) {
// 	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"), opttest.TestInPlace())
// 	baselineProviderConfig := providertest.ProviderConfiguration{
// 		Name:                  "random",
// 		DownloadBinaryVersion: "4.15.0",
// 		DisableAttach:         true,
// 	}

// 	uncachedPreviewResult := providertest.PreviewUpgrade(source, baselineProviderConfig, providertest.UpgradePreviewOpts{})
// 	assertpreview.HasNoChanges(t, uncachedPreviewResult)
// }

func TestCachedSnapshot(t *testing.T) {
	cacheDir := t.TempDir()
	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"))
	source.InstallStack("test")
	snapshot1Created := false
	snapshot1 := providertest.CachedSnapshot(source, cacheDir, func(test *pulumitest.PulumiTest) {
		test.Up()
		snapshot1Created = true
	})
	assert.True(t, snapshot1Created, "snapshot1 should have been created from scratch")

	snapshot2Created := false
	snapshot2 := providertest.CachedSnapshot(source, cacheDir, func(test *pulumitest.PulumiTest) {
		test.Up()
		snapshot2Created = true
	})
	assert.False(t, snapshot2Created, "snapshot2 should have been loaded from cache")

	// Can't compare the whole snapshot because the json.RawMessage gets reformatted.
	assert.Equal(t, snapshot1.StackExport.Version, snapshot2.StackExport.Version, "snapshot2 should have the same version as snapshot1")
	assert.Equal(t, snapshot1.GrpcLog, snapshot2.GrpcLog, "snapshot2 should have the same grpc log as snapshot1")
}

func TestRecordSnapshotLowLevel(t *testing.T) {
	source := pulumitest.NewPulumiTest(t, filepath.Join("pulumitest", "testdata", "yaml_program"))
	snapshot := providertest.CaptureSnapshot(source, func(test *pulumitest.PulumiTest) {
		test.Up()
	}, opttest.DownloadProviderVersion("random", "4.15.0"))
	assert.NotNil(t, snapshot)
	previewResult := providertest.PreviewUpdateFromSnapshot(source, snapshot)
	assertpreview.HasNoChanges(t, previewResult)
}
