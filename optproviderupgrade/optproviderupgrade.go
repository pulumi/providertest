package optproviderupgrade

import (
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// DisableAttach will configure the provider binary in the program's Pulumi.yaml rather than attaching the running provider.
func DisableAttach() PreviewProviderUpgradeOpt {
	return optionFunc(func(o *PreviewProviderUpgradeOptions) {
		o.DisableAttach = true
	})
}

// ProgramName is replaced with the name of the program under test based on the program's directory name.
var ProgramName string = "{programName}"

// BaselineVersion is replaced with the version of the provider used for the baseline.
var BaselineVersion string = "{baselineVersion}"

// BaselineOptions sets the options to use when creating the baseline stack.
func BaselineOpts(opts ...opttest.Option) PreviewProviderUpgradeOpt {
	return optionFunc(func(o *PreviewProviderUpgradeOptions) {
		o.BaselineOpts = opts
	})
}

// CacheDir sets the path to the directory to use for caching the stack state and grpc log.
// The path can contain the following placeholders:
//
// - {programName}: the name of the program under test based on the program's directory name
//
// - {baselineVersion}: the version of the provider used for the baseline
//
// Calculated path elements are joined with filepath.Join.
func CacheDir(elem ...string) PreviewProviderUpgradeOpt {
	return optionFunc(func(o *PreviewProviderUpgradeOptions) {
		o.CacheDirTemplate = elem
	})
}

type PreviewProviderUpgradeOptions struct {
	CacheDirTemplate []string
	DisableAttach    bool
	BaselineOpts     []opttest.Option
}

type PreviewProviderUpgradeOpt interface {
	Apply(*PreviewProviderUpgradeOptions)
}

func Defaults() PreviewProviderUpgradeOptions {
	return PreviewProviderUpgradeOptions{
		CacheDirTemplate: []string{"testdata", "recorded", "TestProviderUpgrade", ProgramName, BaselineVersion},
	}
}

type optionFunc func(*PreviewProviderUpgradeOptions)

func (o optionFunc) Apply(opts *PreviewProviderUpgradeOptions) {
	o(opts)
}
