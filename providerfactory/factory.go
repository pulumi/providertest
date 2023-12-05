package providerfactory

import "context"

// ProviderFactory is a function that starts a provider and returns the port it is listening on.
// The function should return an error if the provider fails to start.
// When the test is complete, the context will be cancelled and the provider should exit.
type ProviderFactory func(ctx context.Context) (int, error)
