package autotest

import (
	"context"
	"testing"

	"github.com/pulumi/providertest/providers"
)

type EnvBuilder struct {
	t                 *testing.T
	configPassphrase  string
	providers         map[string]providers.ProviderFactory
	useAmbientBackend bool
	custom            map[string]string
}

var defaultConfigPassphrase string = "correct horse battery staple"

func NewEnvBuilder(t *testing.T) *EnvBuilder {
	return &EnvBuilder{
		t:                 t,
		configPassphrase:  defaultConfigPassphrase,
		providers:         map[string]providers.ProviderFactory{},
		useAmbientBackend: false,
		custom:            map[string]string{},
	}
}

// AttachProvider will start the provider via the specified factory and attach it when running the program under test.
func (e *EnvBuilder) AttachProvider(name string, startProvider providers.ProviderFactory) *EnvBuilder {
	e.t.Helper()
	e.providers[name] = startProvider
	return e
}

// AttachProviderBinary adds a provider to be started and attached for the test run.
// Path can be a directory or a binary. If it is a directory, the binary will be assumed to be
// pulumi-resource-<name> in that directory.
func (e *EnvBuilder) AttachProviderBinary(name, path string) *EnvBuilder {
	e.t.Helper()
	startProvider, err := providers.LocalBinary(name, path)
	if err != nil {
		e.t.Fatalf("failed to create provider factory for %s: %v", name, err)
	}
	e.providers[name] = startProvider
	return e
}

// AttachProviderServer will start the specified and attach for the test run.
func (e *EnvBuilder) AttachProviderServer(name string, startProvider providers.ResourceProviderServerFactory) *EnvBuilder {
	e.t.Helper()
	startProviderFactory, err := providers.ResourceProviderFactory(startProvider)
	if err != nil {
		e.t.Fatalf("failed to create provider factory for %s: %v", name, err)
	}
	e.providers[name] = startProviderFactory
	return e
}

// AttachDownloadedPlugin installs the plugin via `pulumi plugin install` then will start the provider and attach it for the test run.
func (e *EnvBuilder) AttachDownloadedPlugin(name, version string) *EnvBuilder {
	e.t.Helper()
	binaryPath := providers.DownloadPluginBinary(e.t, name, version)
	return e.AttachProviderBinary(name, binaryPath)
}

// UseAmbientBackend configures the test to use the ambient backend rather than a local temporary directory.
func (e *EnvBuilder) UseAmbientBackend() *EnvBuilder {
	e.t.Helper()
	e.useAmbientBackend = true
	return e
}

// Set a custom environment variable to use when running the program under test.
func (e *EnvBuilder) Set(key, value string) *EnvBuilder {
	e.t.Helper()
	e.custom[key] = value
	return e
}

// Clear all custom environment variables.
func (e *EnvBuilder) Clear() map[string]string {
	e.t.Helper()
	e.custom = map[string]string{}
	return e.custom
}

// Delete a custom environment variable.
func (e *EnvBuilder) Delete(key string) *EnvBuilder {
	e.t.Helper()
	delete(e.custom, key)
	return e
}

// SetConfigPassword sets the config passphrase to use when running the program under test.
func (e *EnvBuilder) SetConfigPassword(passphrase string) *EnvBuilder {
	e.t.Helper()
	e.configPassphrase = passphrase
	return e
}

// GetEnv returns the environment variables to use when running the program under test.
func (e *EnvBuilder) GetEnv(ctx context.Context) map[string]string {
	e.t.Helper()

	env := map[string]string{
		"PULUMI_CONFIG_PASSPHRASE": e.configPassphrase,
	}

	if !e.useAmbientBackend {
		backendFolder := e.t.TempDir()
		env["PULUMI_BACKEND_URL"] = "file://" + backendFolder
	}

	if len(e.providers) > 0 {
		e.t.Log("starting providers")
		providerPorts, cancel, err := providers.StartProviders(ctx, e.providers)
		if err != nil {
			e.t.Fatalf("failed to start providers: %v", err)
		}
		e.t.Cleanup(func() {
			cancel()
		})
		env["PULUMI_DEBUG_PROVIDERS"] = providers.GetDebugProvidersEnv(providerPorts)
	}

	return env
}

func (e *EnvBuilder) Copy() *EnvBuilder {
	copy := *e
	copy.providers = map[string]providers.ProviderFactory{}
	for k, v := range e.providers {
		copy.providers[k] = v
	}
	copy.custom = map[string]string{}
	for k, v := range e.custom {
		copy.custom[k] = v
	}
	return &copy
}
