package opttest

import (
	"path/filepath"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// TestInPlace will run the program under test from its current location, rather than firstly copying to a temporary directory.
func TestInPlace() Option {
	return optionFunc(func(o *Options) {
		o.TestInPlace = true
	})
}

// AttachProvider will start the provider via the specified factory and attach it when running the program under test.
func AttachProvider(name string, startProvider providers.ProviderFactory) Option {
	return optionFunc(func(o *Options) {
		o.ProviderFactories[providers.ProviderName(name)] = startProvider
	})
}

// AttachProviderBinary adds a provider to be started and attached for the test run.
// Path can be a directory or a binary. If it is a directory, the binary will be assumed to be
// pulumi-resource-<name> in that directory.
func AttachProviderBinary(name, path string) Option {
	return optionFunc(func(o *Options) {
		o.ProviderFactories[providers.ProviderName(name)] = providers.LocalBinary(name, path)
	})
}

// AttachProviderServer will start the specified and attach for the test run.
func AttachProviderServer(name string, startProvider providers.ResourceProviderServerFactory) Option {
	return optionFunc(func(o *Options) {
		o.ProviderFactories[providers.ProviderName(name)] = providers.ResourceProviderFactory(startProvider)
	})
}

// AttachDownloadedPlugin installs the plugin via `pulumi plugin install` then will start the provider and attach it for the test run.
func AttachDownloadedPlugin(name, version string) Option {
	return optionFunc(func(o *Options) {
		o.ProviderFactories[providers.ProviderName(name)] = providers.DownloadPluginBinaryFactory(name, version)
	})
}

// LocalProviderPath sets the path to the local provider binary to use when running the program under test.
// This sets the `plugins.providers` property in the project settings (Pulumi.yaml).
func LocalProviderPath(name string, path ...string) Option {
	return optionFunc(func(o *Options) {
		o.ProviderPluginPaths[name] = filepath.Join(path...)
	})
}

// YarnLink specifies packages which are linked via `yarn link` and should be used when running the program under test.
// Each package is called with `yarn link <package>` on stack creation.
func YarnLink(packages ...string) Option {
	return optionFunc(func(o *Options) {
		o.YarnLinks = append(o.YarnLinks, filepath.Join(packages...))
	})
}

// GoModReplacement specifies replacements to be add to the go.mod file when running the program under test.
// Each replacement is added to the go.mod file with `go mod edit -replace <replacement>` on stack creation.
func GoModReplacement(packageSpecifier string, replacementPathElem ...string) Option {
	return optionFunc(func(o *Options) {
		o.GoModReplacements[packageSpecifier] = filepath.Join(replacementPathElem...)
	})
}

// UseAmbientBackend skips setting `PULUMI_BACKEND_URL` to a local temporary directory which overrides any backend configuration which might have been done on the local environment via `pulumi login`.
// Using this option will cause the program under test to use whatever backend configuration has been set via `pulumi login` or an existing `PULUMI_BACKEND_URL` value.
func UseAmbientBackend() Option {
	return optionFunc(func(o *Options) {
		o.UseAmbientBackend = true
	})
}

// Set a custom environment variable to use when running the program under test.
func Env(key, value string) Option {
	return optionFunc(func(o *Options) {
		o.CustomEnv[key] = value
	})
}

// SetConfigPassword sets the config passphrase to use when running the program under test.
func ConfigPassphrase(passphrase string) Option {
	return optionFunc(func(o *Options) {
		o.ConfigPassphrase = passphrase
	})
}

// WorkspaceOptions sets additional options to pass to the workspace when running the program under test.
func WorkspaceOptions(opts ...auto.LocalWorkspaceOption) Option {
	return optionFunc(func(o *Options) {
		o.ExtraWorkspaceOptions = opts
	})
}

type Options struct {
	TestInPlace           bool
	ConfigPassphrase      string
	ProviderFactories     map[providers.ProviderName]providers.ProviderFactory
	ProviderPluginPaths   map[string]string
	UseAmbientBackend     bool
	YarnLinks             []string
	GoModReplacements     map[string]string
	CustomEnv             map[string]string
	ExtraWorkspaceOptions []auto.LocalWorkspaceOption
}

var defaultConfigPassphrase string = "correct horse battery staple"

func DefaultOptions() *Options {
	return &Options{
		ConfigPassphrase:    defaultConfigPassphrase,
		ProviderFactories:   make(map[providers.ProviderName]providers.ProviderFactory),
		ProviderPluginPaths: make(map[string]string),
		GoModReplacements:   make(map[string]string),
		CustomEnv:           make(map[string]string),
	}
}

type Option interface {
	Apply(*Options)
}

type optionFunc func(*Options)

func (o optionFunc) Apply(opts *Options) {
	o(opts)
}
