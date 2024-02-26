package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// StartProviders starts each of the given providers and returns a map of provider names to the ports they are listening on.
// The context should be cancelled when the test is complete to shut down the providers.
func StartProviders(ctx context.Context, factories map[ProviderName]ProviderFactory, opts PulumiTest) (map[ProviderName]Port, error) {
	if len(factories) == 0 {
		return nil, nil
	}

	providerNames := make([]ProviderName, 0, len(factories))
	for providerName := range factories {
		providerNames = append(providerNames, providerName)
	}
	sort.SliceStable(providerNames, func(i, j int) bool {
		return providerNames[i] < providerNames[j]
	})
	portMappings := map[ProviderName]Port{}
	for _, providerName := range providerNames {
		factory := factories[providerName]
		port, err := factory(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to start provider %s: %v", providerName, err)
		}
		portMappings[providerName] = port
	}
	return portMappings, nil
}

// GetDebugProvidersEnv returns a comma-separated list of provider names and ports in the format expected by the
// PULUMI_DEBUG_PROVIDERS environment variable.
func GetDebugProvidersEnv(runningProviders map[ProviderName]Port) string {
	// Sort the provider names so that the order is deterministic.
	providerNames := make([]ProviderName, 0, len(runningProviders))
	for providerName := range runningProviders {
		providerNames = append(providerNames, providerName)
	}
	sort.SliceStable(providerNames, func(i, j int) bool {
		return providerNames[i] < providerNames[j]
	})
	mappings := make([]string, 0, len(runningProviders))
	for _, providerName := range providerNames {
		port := runningProviders[providerName]
		mapping := fmt.Sprintf("%s:%d", providerName, port)
		mappings = append(mappings, mapping)
	}
	return strings.Join(mappings, ",")
}
