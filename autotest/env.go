package autotest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/pulumi/providertest/providerfactory"
)

type EnvBuilder struct {
	t                 *testing.T
	configPassword    string
	providers         map[string]providerfactory.ProviderFactory
	useAmbientBackend bool
}

func NewEnvBuilder(t *testing.T) *EnvBuilder {
	return &EnvBuilder{
		t:                 t,
		configPassword:    defaultConfigPassword,
		providers:         map[string]providerfactory.ProviderFactory{},
		useAmbientBackend: false,
	}
}

func (e *EnvBuilder) AttachProvider(name string, startProvider providerfactory.ProviderFactory) *EnvBuilder {
	e.t.Helper()
	e.providers[name] = startProvider
	return e
}

// AttachProviderBinary adds a provider to be started and attached for the test run.
// Path can be a directory or a binary. If it is a directory, the binary will be assumed to be
// pulumi-resource-<name> in that directory.
func (e *EnvBuilder) AttachProviderBinary(name, path string) *EnvBuilder {
	e.t.Helper()
	startProvider, err := providerfactory.LocalBinary(name, path)
	if err != nil {
		e.t.Fatalf("failed to create provider factory for %s: %v", name, err)
	}
	e.providers[name] = startProvider
	return e
}

// AttachProviderServer adds a provider to be started and attached for the test run.
func (e *EnvBuilder) AttachProviderServer(name string, startProvider providerfactory.ResourceProviderServerFactory) *EnvBuilder {
	e.t.Helper()
	startProviderFactory, err := providerfactory.ResourceProviderFactory(startProvider)
	if err != nil {
		e.t.Fatalf("failed to create provider factory for %s: %v", name, err)
	}
	e.providers[name] = startProviderFactory
	return e
}

// AttachDownloadedPlugin installs the plugin via `pulumi plugin install` and adds it to the test environment.
func (e *EnvBuilder) AttachDownloadedPlugin(name, version string) *EnvBuilder {
	e.t.Helper()
	binaryPath := providerfactory.DownloadPluginBinary(e.t, name, version)
	return e.AttachProviderBinary(name, binaryPath)
}

// UseAmbientBackend configures the test to use the ambient backend rather than a local temporary directory.
func (e *EnvBuilder) UseAmbientBackend() *EnvBuilder {
	e.t.Helper()
	e.useAmbientBackend = true
	return e
}

var defaultConfigPassword string = "correct horse battery staple"

func (e *EnvBuilder) GetEnv() map[string]string {
	e.t.Helper()

	env := map[string]string{
		"PULUMI_CONFIG_PASSPHRASE": defaultConfigPassword,
	}

	if !e.useAmbientBackend {
		backendFolder := e.t.TempDir()
		env["PULUMI_BACKEND_URL"] = "file://" + backendFolder
	}

	if len(e.providers) > 0 {
		providerContext, cancel := context.WithCancel(context.Background())
		providerNames := make([]string, 0, len(e.providers))
		for providerName := range e.providers {
			providerNames = append(providerNames, providerName)
		}
		sort.Strings(providerNames)
		portMappings := make([]string, 0, len(e.providers))
		for _, providerName := range providerNames {
			start := e.providers[providerName]
			port, err := start(providerContext)
			if err != nil {
				e.t.Fatalf("failed to start provider %s: %v", providerName, err)
			}
			portMappings = append(portMappings, fmt.Sprintf("%s:%d", providerName, port))
		}
		env["PULUMI_DEBUG_PROVIDERS"] = strings.Join(portMappings, ",")
		e.t.Cleanup(func() {
			cancel()
		})
	}

	return env
}
