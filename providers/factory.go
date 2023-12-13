package providers

import "context"

// ProviderName is the name of a provider without the "pulumi-" prefix. e.g. "aws" or "azure"
type ProviderName string

// Port is the port that a provider is listening on.
type Port int

// ProviderFactory is a function that starts a provider and returns the port it is listening on.
// The function should return an error if the provider fails to start.
// When the test is complete, the context will be cancelled and the provider should exit.
type ProviderFactory func(ctx context.Context) (Port, error)
