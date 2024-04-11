package providertest

import (
	"fmt"
	"path/filepath"

	"github.com/pulumi/providertest/optproviderupgrade"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/optnewstack"
	"github.com/pulumi/providertest/pulumitest/optrun"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// PreviewProviderUpgrade captures the state of a stack from a baseline provider configuration, then previews the stack
// with the current provider configuration.
// Uses a default cache directory of "testdata/recorded/TestProviderUpgrade/{programName}/{baselineVersion}".
func PreviewProviderUpgrade(t pulumitest.PT, pulumiTest *pulumitest.PulumiTest, providerName string, baselineVersion string, opts ...optproviderupgrade.PreviewProviderUpgradeOpt) auto.PreviewResult {
	t.Helper()
	previewTest := pulumiTest.CopyToTempDir(opttest.NewStackOptions(optnewstack.DisableAutoDestroy()))
	options := optproviderupgrade.Defaults()
	for _, opt := range opts {
		opt.Apply(&options)
	}
	programName := filepath.Base(pulumiTest.Source())
	cacheDir := getCacheDir(options, programName, baselineVersion)
	previewTest.Run(
		func(test *pulumitest.PulumiTest) {
			t.Helper()
			test.Up()
			grptLog := test.GrpcLog()
			grpcLogPath := filepath.Join(cacheDir, "grpc.json")
			t.Log(fmt.Sprintf("writing grpc log to %s", grpcLogPath))
			grptLog.WriteTo(grpcLogPath)
		},
		optrun.WithCache(filepath.Join(cacheDir, "stack.json")),
		optrun.WithOpts(
			opttest.NewStackOptions(optnewstack.EnableAutoDestroy()),
			baselineProviderOpt(options, providerName, baselineVersion)),
		optrun.WithOpts(options.BaselineOpts...),
	)
	return previewTest.Preview()
}

func baselineProviderOpt(options optproviderupgrade.PreviewProviderUpgradeOptions, providerName string, baselineVersion string) opttest.Option {
	if options.DisableAttach {
		return opttest.DownloadProviderVersion(providerName, baselineVersion)
	} else {
		return opttest.AttachDownloadedPlugin(providerName, baselineVersion)
	}
}

func getCacheDir(options optproviderupgrade.PreviewProviderUpgradeOptions, programName string, baselineVersion string) string {
	var cacheDir string
	for _, pathTemplateElement := range options.CacheDirTemplate {
		switch pathTemplateElement {
		case optproviderupgrade.ProgramName:
			cacheDir = filepath.Join(cacheDir, programName)
		case optproviderupgrade.BaselineVersion:
			cacheDir = filepath.Join(cacheDir, baselineVersion)
		default:
			cacheDir = filepath.Join(cacheDir, pathTemplateElement)
		}
	}
	return cacheDir
}
