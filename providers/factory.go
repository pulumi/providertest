package providers

import "context"

// ProviderName is the name of a provider without the "pulumi-" prefix. e.g. "aws" or "azure"
type ProviderName string

// Port is the port that a provider is listening on.
type Port int

// ProviderOptions is a set of options to be passed to the provider factory.
type ProviderOptions struct {
	// WorkDir is the Pulumi workspace directory.
	WorkDir string
}

// ProviderFactory is a function that starts a provider and returns the port it is listening on.
// The function should return an error if the provider fails to start.
// When the test is complete, the context will be cancelled and the provider should exit.
type ProviderFactory func(ctx context.Context, opts ProviderOptions) (Port, error)
