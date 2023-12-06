package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

func StartProviders(ctx context.Context, factories map[string]ProviderFactory) (map[string]int, context.CancelFunc, error) {
	if len(factories) == 0 {
		return nil, nil, nil
	}

	providerContext, cancel := context.WithCancel(context.Background())
	providerNames := make([]string, 0, len(factories))
	for providerName := range factories {
		providerNames = append(providerNames, providerName)
	}
	sort.Strings(providerNames)
	portMappings := map[string]int{}
	for _, providerName := range providerNames {
		factory := factories[providerName]
		port, err := factory(providerContext)
		if err != nil {
			cancel() // Cancel the context if any provider fails to start.
			return nil, nil, fmt.Errorf("failed to start provider %s: %v", providerName, err)
		}
		portMappings[providerName] = port
	}
	return portMappings, cancel, nil
}

func GetDebugProvidersEnv(runningProviders map[string]int) string {
	// Sort the provider names so that the order is deterministic.
	providerNames := make([]string, 0, len(runningProviders))
	for providerName := range runningProviders {
		providerNames = append(providerNames, providerName)
	}
	sort.Strings(providerNames)
	mappings := make([]string, 0, len(runningProviders))
	for _, providerName := range providerNames {
		port := runningProviders[providerName]
		mapping := fmt.Sprintf("%s:%d", providerName, port)
		mappings = append(mappings, mapping)
	}
	return strings.Join(mappings, ",")
}
