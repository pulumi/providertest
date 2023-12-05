package autotest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
)

type EnvBuilder struct {
	t               *testing.T
	configPassword  string
	providers       map[string]ProviderFactory
	useLocalBackend bool
}

func NewEnvBuilder(t *testing.T) *EnvBuilder {
	return &EnvBuilder{
		t:               t,
		configPassword:  defaultConfigPassword,
		providers:       map[string]ProviderFactory{},
		useLocalBackend: false,
	}
}

// ProviderFactory is a function that starts a provider and returns the port it is listening on.
// The function should return an error if the provider fails to start.
// When the test is complete, the context will be cancelled and the provider should exit.
type ProviderFactory func(ctx context.Context) (int, error)

func (e *EnvBuilder) AttachProvider(name string, startProvider ProviderFactory) *EnvBuilder {
	e.t.Helper()
	e.providers[name] = startProvider
	return e
}

// AttachProviderBinary adds a provider to be started and attached for the test run.
// Path can be a directory or a binary. If it is a directory, the binary will be assumed to be
// pulumi-resource-<name> in that directory.
func (e *EnvBuilder) AttachProviderBinary(name, path string) *EnvBuilder {
	e.t.Helper()
	startProvider, err := LocalProviderBinary(name, path)
	if err != nil {
		e.t.Fatalf("failed to create provider factory for %s: %v", name, err)
	}
	e.providers[name] = startProvider
	return e
}

func (e *EnvBuilder) UseLocalBackend() *EnvBuilder {
	e.t.Helper()
	e.useLocalBackend = true
	return e
}

var defaultConfigPassword string = "correct horse battery staple"

func (e *EnvBuilder) GetEnv() map[string]string {
	e.t.Helper()

	env := map[string]string{
		"PULUMI_CONFIG_PASSPHRASE": defaultConfigPassword,
	}

	if e.useLocalBackend {
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
