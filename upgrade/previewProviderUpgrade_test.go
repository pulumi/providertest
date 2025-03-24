package upgrade_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumitest"
	"github.com/pulumi/pulumitest/assertpreview"
	"github.com/pulumi/pulumitest/optproviderupgrade"
	"github.com/pulumi/pulumitest/opttest"
	"github.com/pulumi/pulumitest/providers"
	"github.com/pulumi/pulumitest/upgrade"
	"github.com/stretchr/testify/assert"
)

func TestPreviewUpgradeCached(t *testing.T) {
	t.Parallel()
	cacheDir := t.TempDir()
	test := pulumitest.NewPulumiTest(t, filepath.Join("..", "testdata", "yaml_program"),
		opttest.DownloadProviderVersion("random", "4.15.0"))

	uncachedPreviewResult := upgrade.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach())
	assertpreview.HasNoReplacements(t, uncachedPreviewResult)
	assertpreview.HasNoChanges(t, uncachedPreviewResult)

	cachedPreviewResult := upgrade.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach())
	assert.Equal(t, uncachedPreviewResult, cachedPreviewResult, "expected uncached and cached preview to be the same")
}

func TestPreviewUpgradeWithKnownSourceEdit(t *testing.T) {
	t.Parallel()
	cacheDir := t.TempDir()
	test := pulumitest.NewPulumiTest(t, filepath.Join("..", "testdata", "yaml_program"),
		opttest.DownloadProviderVersion("random", "4.15.0"))

	previewResult := upgrade.PreviewProviderUpgrade(t, test, "random", "4.5.0",
		optproviderupgrade.CacheDir(cacheDir, "{programName}", "{baselineVersion}"),
		optproviderupgrade.DisableAttach(),
		optproviderupgrade.NewSourcePath(filepath.Join("..", "testdata", "yaml_program_updated")),
	)

	assert.Contains(t, previewResult.StdOut, "random:index/randomPassword:RandomPassword::password")
	assert.Contains(t, previewResult.StdOut, "+ 1 to create\n")
}

func TestPreviewWithInvokeReplayed(t *testing.T) {
	t.Parallel()
	cacheDir := t.TempDir()
	commandProvider := providers.DownloadPluginBinaryFactory("command", "1.0.1")
	// Intercept all invokes and replay them from a gRPC log during the preview.
	commandProvider = commandProvider.ReplayInvokes(filepath.Join(cacheDir, "grpc.json"), false)
	test := pulumitest.NewPulumiTest(t, filepath.Join("..", "testdata", "yaml_command_invoke"),
		opttest.AttachProvider("command", commandProvider))

	// We're not changing the version, but if the preview doesn't re-use the captured invoke the value will be different.
	// This will cause a resource update which we can assert against.
	previewResult := upgrade.PreviewProviderUpgrade(t, test, "command", "1.0.1",
		optproviderupgrade.CacheDir(cacheDir))

	assertpreview.HasNoChanges(t, previewResult)
}
