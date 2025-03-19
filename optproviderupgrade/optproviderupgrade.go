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

// NewSourcePath sets the path to new source code to use for the target version of the upgrade.
// If not set, the original pulumitest program source is used.
// This is useful for where it's expected for a user to perform code changes during a migration.
func NewSourcePath(path string) PreviewProviderUpgradeOpt {
	return optionFunc(func(o *PreviewProviderUpgradeOptions) {
		o.NewSourcePath = path
	})
}

type PreviewProviderUpgradeOptions struct {
	CacheDirTemplate []string
	DisableAttach    bool
	BaselineOpts     []opttest.Option
	NewSourcePath    string
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
