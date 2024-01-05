package providertest

import (
	"path/filepath"

	"github.com/pulumi/providertest/optproviderupgrade"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/optrun"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// PreviewProviderUpgrade captures the state of a stack from a baseline provider configuration, then previews the stack
// with the current provider configuration.
// Uses a default cache directory of "testdata/recorded/TestProviderUpgrade/{programName}/{baselineVersion}".
func PreviewProviderUpgrade(pulumiTest *pulumitest.PulumiTest, providerName string, baselineVersion string, opts ...optproviderupgrade.PreviewProviderUpgradeOpt) auto.PreviewResult {
	pulumiTest.T().Helper()
	options := optproviderupgrade.PreviewProviderUpgradeOptions{}
	for _, opt := range opts {
		opt.Apply(&options)
	}
	programName := filepath.Base(pulumiTest.Source())
	cacheDir := getCacheDir(options, programName, baselineVersion)
	pulumiTest.Run(
		func(test *pulumitest.PulumiTest) {
			test.T().Helper()
			test.Up()
			grptLog := test.GrpcLog()
			grpcLogPath := filepath.Join(cacheDir, "grpc.json")
			test.T().Logf("writing grpc log to %s", grpcLogPath)
			grptLog.WriteTo(grpcLogPath)
		},
		optrun.WithCache(filepath.Join(cacheDir, "stack.json")),
		optrun.WithOpts(options.BaselineOpts...))
	return pulumiTest.Preview()
}

func getCacheDir(options optproviderupgrade.PreviewProviderUpgradeOptions, programName string, baselineVersion string) string {
	if len(options.CacheDirTemplate) == 0 {
		options.CacheDirTemplate = []string{"testdata", "recorded", "TestProviderUpgrade", "{programName}", "{baselineVersion}"}
	}
	var cacheDir string
	for _, pathTemplateElement := range options.CacheDirTemplate {
		switch pathTemplateElement {
		case "{programName}":
			cacheDir = filepath.Join(cacheDir, programName)
		case "{baselineVersion}":
			cacheDir = filepath.Join(cacheDir, baselineVersion)
		default:
			cacheDir = filepath.Join(cacheDir, pathTemplateElement)
		}
	}
	return cacheDir
}
